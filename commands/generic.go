package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// listFilter declares a resource-specific list filter mapped to an API query param.
type listFilter struct {
	Flag   string   // CLI flag name, e.g. "status"
	Query  string   // API query parameter name, e.g. "status"
	Usage  string   // help text
	Values []string // optional allowed values (used for shell completion)
}

// resourceSpec declares how to expose one resource as a CRUD command group. A new
// resource is this spec plus an api type + accessor — no edits to shared code.
type resourceSpec[T any] struct {
	Use     string
	Aliases []string
	Short   string
	Long    string

	// New returns the typed resource handle for a client.
	New func(*api.Client) *api.Resource[T]

	Columns     []string
	ListFilters []listFilter

	NoCreate bool
	NoUpdate bool
	NoDelete bool
	NoGet    bool // resource has no GET-by-id endpoint (e.g. variables)

	// IDFields are the JSON keys used to match a record when `get` is served from
	// the list (NoGet resources). Defaults to {"id"}.
	IDFields []string

	// Extra attaches custom subcommands (activate, retry, transfer, ...).
	Extra func(parent *cobra.Command, sp resourceSpec[T])
}

// registerResource builds the command group for a resource and attaches it to root.
func registerResource[T any](sp resourceSpec[T]) {
	cmd := buildResourceCmd(sp)
	rootCmd.AddCommand(cmd)
}

// buildResourceCmd constructs the parent command and its subcommands.
func buildResourceCmd[T any](sp resourceSpec[T]) *cobra.Command {
	parent := &cobra.Command{
		Use:     sp.Use,
		Aliases: sp.Aliases,
		Short:   sp.Short,
		Long:    sp.Long,
	}

	parent.AddCommand(newListCmd(sp))
	if !sp.NoGet || len(sp.IDFields) > 0 {
		parent.AddCommand(newGetCmd(sp))
	}
	if !sp.NoCreate {
		parent.AddCommand(newCreateCmd(sp))
	}
	if !sp.NoUpdate {
		parent.AddCommand(newUpdateCmd(sp))
	}
	if !sp.NoDelete {
		parent.AddCommand(newDeleteCmd(sp))
	}
	if sp.Extra != nil {
		sp.Extra(parent, sp)
	}
	return parent
}

// --- list ---

func newListCmd[T any](sp resourceSpec[T]) *cobra.Command {
	var (
		limit      int
		cursor     string
		all        bool
		maxPages   int
		filterVals = map[string]*string{}
		rawParams  []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List " + sp.Use,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			params := api.ListParams{Limit: limit, Cursor: cursor, Extra: url.Values{}}
			for _, f := range sp.ListFilters {
				if v := filterVals[f.Flag]; v != nil && *v != "" {
					params.Extra.Set(f.Query, *v)
				}
			}
			for _, kv := range rawParams {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid --param %q (want key=value)", kv)
				}
				params.Extra.Add(k, v)
			}

			res := sp.New(client)
			if all {
				items, truncated, err := res.ListAllChecked(cmd.Context(), params, maxPages)
				if err != nil {
					if api.IsDryRun(err) {
						return nil
					}
					return err
				}
				if truncated && !flagQuiet {
					effCap := maxPages
					if effCap <= 0 {
						effCap = api.DefaultMaxPages
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: results truncated at %d pages; narrow filters or raise --max-pages\n", effCap)
				}
				return render(cmd, items, sp.Columns...)
			}
			items, next, err := res.List(cmd.Context(), params)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			if next != "" && !flagQuiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "more results available — re-run with --cursor %s (or --all)\n", next)
			}
			return render(cmd, items, sp.Columns...)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "max items to return (page size, capped at 250)")
	cmd.Flags().StringVar(&cursor, "cursor", "", "pagination cursor from a previous response")
	cmd.Flags().BoolVar(&all, "all", false, "fetch every page (auto-paginate)")
	cmd.Flags().IntVar(&maxPages, "max-pages", 0, "page cap for --all (0 = safety default)")
	cmd.Flags().StringArrayVar(&rawParams, "param", nil, "extra query parameter key=value (repeatable)")
	for _, f := range sp.ListFilters {
		s := new(string)
		filterVals[f.Flag] = s
		cmd.Flags().StringVar(s, f.Flag, "", f.Usage)
		if len(f.Values) > 0 {
			_ = cmd.RegisterFlagCompletionFunc(f.Flag, fixedCompletions(f.Values))
		}
	}
	return cmd
}

// --- get ---

func newGetCmd[T any](sp resourceSpec[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single " + singular(sp.Use) + " by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			res := sp.New(client)
			if sp.NoGet {
				return getViaList(cmd, sp, res, args[0])
			}
			item, err := res.Get(cmd.Context(), args[0], nil)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return render(cmd, item)
		},
	}
	return cmd
}

// getViaList serves `get` for resources without a GET-by-id endpoint by scanning
// the full list and matching one of IDFields.
func getViaList[T any](cmd *cobra.Command, sp resourceSpec[T], res *api.Resource[T], id string) error {
	items, err := res.ListAll(cmd.Context(), api.ListParams{}, 0)
	if err != nil {
		if api.IsDryRun(err) {
			return nil
		}
		return err
	}
	fields := sp.IDFields
	if len(fields) == 0 {
		fields = []string{"id"}
	}
	for _, it := range items {
		m := toMap(it)
		for _, f := range fields {
			if fmt.Sprintf("%v", m[f]) == id {
				return render(cmd, it)
			}
		}
	}
	return fmt.Errorf("no %s matched %q", singular(sp.Use), id)
}

// --- create ---

func newCreateCmd[T any](sp resourceSpec[T]) *cobra.Command {
	var bf bodyFlags
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a " + singular(sp.Use),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			body, err := bf.build()
			if err != nil {
				return err
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			created, err := sp.New(client).Create(cmd.Context(), body)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, created)
		},
	}
	bf.register(cmd, "request body")
	return cmd
}

// --- update ---

func newUpdateCmd[T any](sp resourceSpec[T]) *cobra.Command {
	var bf bodyFlags
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a " + singular(sp.Use),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := bf.build()
			if err != nil {
				return err
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			updated, err := sp.New(client).Update(cmd.Context(), args[0], body)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, updated)
		},
	}
	bf.register(cmd, "fields to update")
	return cmd
}

// --- delete ---

func newDeleteCmd[T any](sp resourceSpec[T]) *cobra.Command {
	var yes bool
	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a " + singular(sp.Use),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !yes && !flagDryRun {
				if !confirm(cmd, fmt.Sprintf("Delete %s %q?", singular(sp.Use), args[0])) {
					fmt.Fprintln(cmd.ErrOrStderr(), "aborted")
					return nil
				}
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := sp.New(client).Delete(cmd.Context(), args[0]); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "deleted %s %s\n", singular(sp.Use), args[0])
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}

// --- body flags shared by create/update ---

type bodyFlags struct {
	data string
	file string
	sets []string
}

func (bf *bodyFlags) register(cmd *cobra.Command, what string) {
	cmd.Flags().StringVar(&bf.data, "data", "", "inline JSON "+what+` (e.g. '{"name":"x"}')`)
	cmd.Flags().StringVar(&bf.file, "file", "", "read JSON "+what+" from a file ('-' for stdin)")
	cmd.Flags().StringArrayVar(&bf.sets, "set", nil, "set a field key=value (repeatable; value parsed as JSON when possible)")
}

// build assembles the request body from --file/--data/--set, requiring at least one.
func (bf *bodyFlags) build() (json.RawMessage, error) {
	var base map[string]any

	switch {
	case bf.file != "":
		raw, err := readFileOrStdin(bf.file)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &base); err != nil {
			// Allow non-object JSON (e.g. arrays) to pass straight through when no --set.
			if len(bf.sets) == 0 {
				return json.RawMessage(raw), nil
			}
			return nil, fmt.Errorf("parsing --file JSON: %w", err)
		}
	case bf.data != "":
		if err := json.Unmarshal([]byte(bf.data), &base); err != nil {
			if len(bf.sets) == 0 {
				return json.RawMessage(bf.data), nil
			}
			return nil, fmt.Errorf("parsing --data JSON: %w", err)
		}
	}

	if base == nil {
		base = map[string]any{}
	}
	for _, kv := range bf.sets {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --set %q (want key=value)", kv)
		}
		base[k] = inferValue(v)
	}
	if len(base) == 0 {
		return nil, fmt.Errorf("no body provided — use --data, --file, or --set")
	}
	return json.Marshal(base)
}

// inferValue parses v as JSON (number, bool, null, object, array, quoted string)
// and falls back to the raw string when it is not valid JSON.
func inferValue(v string) any {
	var parsed any
	if err := json.Unmarshal([]byte(v), &parsed); err == nil {
		return parsed
	}
	return v
}

func readFileOrStdin(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path) //nolint:gosec // path is user-supplied by design
}

// --- helpers ---

func singular(s string) string { return strings.TrimSuffix(s, "s") }

func confirm(cmd *cobra.Command, prompt string) bool {
	fmt.Fprintf(cmd.ErrOrStderr(), "%s [y/N] ", prompt)
	line, _ := stdinReader().ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

func toMap(v any) map[string]any {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}

func fixedCompletions(values []string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
