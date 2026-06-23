package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/wffile"
	"github.com/jjuanrivvera/n8n-cli/internal/wflint"
)

// addWorkflowGitopsCmds wires the workflows-as-code subcommands onto `workflows`.
func addWorkflowGitopsCmds(parent *cobra.Command) {
	parent.AddCommand(readOnlyHints(workflowConvertCmd()), readOnlyHints(workflowLintCmd()), writeHints(workflowApplyCmd()), readOnlyHints(workflowDiffCmd()))
}

// --- helpers shared by apply/diff ---

// readWorkflowDir reads every workflow file in dir (.json/.yaml/.yml), re-inlining
// externalized $ref fields relative to dir. Returns workflows keyed by name.
func readWorkflowDir(dir string) (map[string]*api.Workflow, error) {
	out := map[string]*api.Workflow{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	loader := wffile.DirLoader(dir)
	for _, e := range entries {
		if e.IsDir() || !isWorkflowFile(e.Name()) {
			continue
		}
		raw, rerr := os.ReadFile(filepath.Join(dir, e.Name())) //nolint:gosec // within the user-supplied dir
		if rerr != nil {
			return nil, rerr
		}
		wf, derr := wffile.DecodeWithFiles(raw, wffile.FormatFromPath(e.Name()), loader)
		if derr != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), derr)
		}
		if wf.Name == "" {
			return nil, fmt.Errorf("%s: workflow has no name", e.Name())
		}
		out[wf.Name] = wf
	}
	return out, nil
}

func isWorkflowFile(name string) bool {
	l := strings.ToLower(name)
	return strings.HasSuffix(l, ".json") || strings.HasSuffix(l, ".yaml") || strings.HasSuffix(l, ".yml")
}

// canonical returns a stable JSON form of a workflow's writable fields, for
// comparison and diffing (read-only fields like id/active/version are dropped).
func canonical(wf *api.Workflow) string {
	b, _ := json.MarshalIndent(workflowCreateBody(wf), "", "  ")
	return string(b) + "\n"
}

// --- convert ---

func workflowConvertCmd() *cobra.Command {
	var to, outDir string
	var externalize int
	cmd := &cobra.Command{
		Use:   "convert <file...> --to json|yaml",
		Short: "Convert workflow files between JSON and YAML (local)",
		Long: "Convert workflow definition files between JSON and YAML on disk. With\n" +
			"--externalize, long code fields (jsCode, query, jsonBody, ...) are split into\n" +
			"sibling files for cleaner review.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := wffile.Format(strings.ToLower(to))
			if format != wffile.JSON && format != wffile.YAML {
				return fmt.Errorf("--to must be json or yaml")
			}
			for _, in := range args {
				raw, err := os.ReadFile(in) //nolint:gosec // user-supplied path
				if err != nil {
					return err
				}
				wf, err := wffile.Decode(raw, wffile.FormatFromPath(in))
				if err != nil {
					return fmt.Errorf("%s: %w", in, err)
				}
				stem := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in))
				main, subfiles, err := wffile.EncodeExternalized(wf, format, stem, externalize)
				if err != nil {
					return err
				}
				dir := outDir
				if dir == "" {
					dir = filepath.Dir(in)
				}
				dst := filepath.Join(dir, stem+"."+string(format))
				// #nosec G703 G304 -- convert writes to the user's own --out dir / input path, not untrusted input
				if err := os.WriteFile(dst, main, 0o600); err != nil {
					return err
				}
				for rel, content := range subfiles {
					sub := filepath.Join(dir, filepath.FromSlash(rel))
					if err := os.MkdirAll(filepath.Dir(sub), 0o755); err != nil { //nolint:gosec // user dir
						return err
					}
					if err := os.WriteFile(sub, content, 0o600); err != nil {
						return err
					}
				}
				if !flagQuiet {
					fmt.Fprintf(cmd.OutOrStdout(), "converted %s -> %s\n", in, dst)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "target format: json or yaml (required)")
	cmd.Flags().StringVar(&outDir, "out", "", "output directory (default: alongside each input)")
	cmd.Flags().IntVar(&externalize, "externalize", 0, "externalize code fields longer than N lines (0 = off)")
	return cmd
}

// --- lint ---

func workflowLintCmd() *cobra.Command {
	var dir string
	var files []string
	var remote, listRules bool
	var disable []string
	cmd := &cobra.Command{
		Use:   "lint [--dir <dir> | -f <file>... | --remote]",
		Short: "Lint workflow definitions for common mistakes",
		Long: "Static checks over workflow files (or live workflows with --remote):\n" +
			"required fields, dangling connections, orphaned nodes, missing webhookId,\n" +
			"and expression strings missing the leading '='. Exits non-zero on errors.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if listRules {
				for _, r := range wflint.Rules {
					fmt.Fprintf(cmd.OutOrStdout(), "%-22s %s\n%-22s  basis: %s\n", r.Name, r.Desc, "", r.Basis)
				}
				return nil
			}
			disabled := map[string]bool{}
			for _, d := range disable {
				disabled[d] = true
			}

			workflows := map[string]*api.Workflow{}
			switch {
			case remote:
				client, err := getReadClient(cmd)
				if err != nil {
					return err
				}
				items, err := client.Workflows().ListAll(cmd.Context(), api.ListParams{}, 0)
				if err != nil {
					if api.IsDryRun(err) {
						return nil
					}
					return err
				}
				for i := range items {
					workflows[items[i].Name] = &items[i]
				}
			case dir != "":
				m, err := readWorkflowDir(dir)
				if err != nil {
					return err
				}
				workflows = m
			case len(files) > 0:
				for _, f := range files {
					raw, err := os.ReadFile(f) //nolint:gosec // user path
					if err != nil {
						return err
					}
					wf, err := wffile.Decode(raw, wffile.FormatFromPath(f))
					if err != nil {
						return fmt.Errorf("%s: %w", f, err)
					}
					workflows[f] = wf
				}
			default:
				return fmt.Errorf("provide --dir, -f/--file, or --remote")
			}

			var results []lintResult
			totalErrors := 0
			keys := sortedKeys(workflows)
			for _, k := range keys {
				fs := wflint.Lint(workflows[k], disabled)
				totalErrors += wflint.Errors(fs)
				if len(fs) > 0 {
					results = append(results, lintResult{Workflow: k, Findings: fs})
				}
			}

			if jsonOutput() {
				if err := render(cmd, results); err != nil {
					return err
				}
			} else {
				printLintText(cmd, results)
			}
			if totalErrors > 0 {
				return fmt.Errorf("lint found %d error(s)", totalErrors)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "lint all workflow files in a directory")
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "lint specific files")
	cmd.Flags().BoolVar(&remote, "remote", false, "lint live workflows from the instance")
	cmd.Flags().BoolVar(&listRules, "list-rules", false, "list available rules and exit")
	cmd.Flags().StringSliceVar(&disable, "disable-rule", nil, "rules to disable (comma-separated)")
	return cmd
}

// lintResult groups a workflow's lint findings.
type lintResult struct {
	Workflow string           `json:"workflow"`
	Findings []wflint.Finding `json:"findings"`
}

func printLintText(cmd *cobra.Command, results []lintResult) {
	for _, r := range results {
		for _, f := range r.Findings {
			loc := r.Workflow
			if f.Node != "" {
				loc += " · " + f.Node
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s: %s (%s)\n", symbolFor(f.Severity), loc, f.Message, f.Rule)
		}
	}
	if len(results) == 0 && !flagQuiet {
		fmt.Fprintln(cmd.OutOrStdout(), "no issues found")
	}
}

// --- apply (GitOps reconcile) ---

func workflowApplyCmd() *cobra.Command {
	var dir string
	var prune, activate bool
	cmd := &cobra.Command{
		Use:   "apply --dir <dir>",
		Short: "Reconcile a directory of workflow files into the instance (GitOps)",
		Long: "Treat a directory of workflow files (JSON/YAML) as the desired state and\n" +
			"apply it: create new workflows, update existing ones (matched by name), and\n" +
			"with --prune, delete instance workflows not present in the directory.\n\n" +
			"Combine with profiles to promote the same desired state across instances:\n" +
			"  n8nctl --profile staging workflows apply --dir ./workflows\n" +
			"  n8nctl --profile prod    workflows apply --dir ./workflows --prune\n\n" +
			"Workflows are matched by name (the only stable handle the API exposes), so\n" +
			"renaming a file creates a new workflow and, with --prune, deletes the old one.\n" +
			"Duplicate names on the instance are skipped to avoid acting on the wrong one.\n" +
			"Reconcile covers name, nodes, connections and settings; runtime-only fields\n" +
			"(pinData, meta) are not managed.\n\n" +
			"Always preview with --dry-run first, especially with --prune.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dir == "" {
				return fmt.Errorf("--dir is required")
			}
			local, err := readWorkflowDir(dir)
			if err != nil {
				return err
			}
			client, err := getReadClient(cmd)
			if err != nil {
				return err
			}
			remoteList, err := client.Workflows().ListAll(cmd.Context(), api.ListParams{}, 0)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			// n8n does not enforce unique workflow names. Reconcile matches by name,
			// so ambiguous (duplicate-name) remote workflows are unsafe to update or
			// prune — acting on an arbitrary one could lose or delete the wrong
			// workflow. Detect duplicates and skip those names entirely.
			remote := map[string]*api.Workflow{}
			ambiguous := map[string]bool{}
			for i := range remoteList {
				n := remoteList[i].Name
				if _, ok := remote[n]; ok {
					ambiguous[n] = true
				}
				remote[n] = &remoteList[i]
			}
			if len(ambiguous) > 0 {
				dups := make([]string, 0, len(ambiguous))
				for n := range ambiguous {
					dups = append(dups, n)
				}
				sort.Strings(dups)
				fmt.Fprintf(cmd.ErrOrStderr(),
					"warning: %d duplicate workflow name(s) on the instance (%s); apply will skip them to avoid acting on the wrong workflow\n",
					len(ambiguous), strings.Join(dups, ", "))
			}

			out := cmd.OutOrStdout()
			var created, updated, pruned, unchanged, skipped int
			for _, name := range sortedKeys(local) {
				lwf := local[name]
				body := workflowCreateBody(lwf)
				if ambiguous[name] {
					fmt.Fprintf(out, "! skip %s (ambiguous: duplicate name on instance)\n", name)
					skipped++
					continue
				}
				if ex, ok := remote[name]; ok {
					// Update only if the writable content differs.
					full, gerr := client.Workflows().Get(cmd.Context(), ex.ID.String(), nil)
					if gerr == nil && canonical(full) == canonical(lwf) {
						unchanged++
						continue
					}
					if flagDryRun {
						fmt.Fprintf(out, "~ update %s\n", name)
					} else if _, uerr := client.Workflows().Update(cmd.Context(), ex.ID.String(), body); uerr != nil {
						return fmt.Errorf("updating %q: %w", name, uerr)
					}
					updated++
				} else {
					if flagDryRun {
						fmt.Fprintf(out, "+ create %s\n", name)
					} else {
						res, cerr := client.Workflows().Create(cmd.Context(), body)
						if cerr != nil {
							return fmt.Errorf("creating %q: %w", name, cerr)
						}
						if activate && res != nil && res.ID != "" {
							_, _ = client.ActivateWorkflow(cmd.Context(), res.ID.String())
						}
					}
					created++
				}
			}
			if prune {
				for _, name := range sortedKeys(remote) {
					if _, ok := local[name]; ok {
						continue
					}
					if ambiguous[name] {
						fmt.Fprintf(out, "! skip prune of %s (ambiguous: duplicate name on instance)\n", name)
						skipped++
						continue
					}
					if flagDryRun {
						fmt.Fprintf(out, "- prune %s\n", name)
					} else if derr := client.Workflows().Delete(cmd.Context(), remote[name].ID.String()); derr != nil {
						return fmt.Errorf("pruning %q: %w", name, derr)
					}
					pruned++
				}
			}
			verb := "applied"
			if flagDryRun {
				verb = "plan"
			}
			fmt.Fprintf(out, "%s: %d created, %d updated, %d unchanged, %d pruned, %d skipped\n", verb, created, updated, unchanged, pruned, skipped)
			return nil
		},
	}
	cmd.Flags().StringVarP(&dir, "dir", "d", "", "directory of workflow files (required)")
	cmd.Flags().BoolVar(&prune, "prune", false, "delete instance workflows not present in the directory")
	cmd.Flags().BoolVar(&activate, "activate", false, "activate newly created workflows")
	return cmd
}

// --- diff ---

func workflowDiffCmd() *cobra.Command {
	var toProfile, file string
	cmd := &cobra.Command{
		Use:   "diff <id> [--to <profile> | --file <path>]",
		Short: "Diff a workflow against another instance or a local file",
		Long: "Show a unified diff of a workflow's writable content (read-only fields are\n" +
			"ignored). Compare the active instance's workflow against the same id on\n" +
			"another --profile, or against a local --file.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getReadClient(cmd)
			if err != nil {
				return err
			}
			var left, right *api.Workflow
			var leftLabel, rightLabel string

			if file != "" {
				raw, rerr := os.ReadFile(file) //nolint:gosec // user path
				if rerr != nil {
					return rerr
				}
				right, err = wffile.Decode(raw, wffile.FormatFromPath(file))
				if err != nil {
					return err
				}
				rightLabel = file
				// left = remote workflow: by id arg, else by name match
				if len(args) == 1 {
					left, err = client.Workflows().Get(cmd.Context(), args[0], nil)
				} else {
					left, err = findWorkflowByName(cmd.Context(), client, right.Name)
					if left == nil && err == nil {
						return fmt.Errorf("no remote workflow named %q to diff against", right.Name)
					}
				}
				if err != nil {
					return err
				}
				leftLabel = "instance:" + left.ID.String()
			} else if toProfile != "" {
				if len(args) != 1 {
					return fmt.Errorf("a workflow <id> is required with --to")
				}
				left, err = client.Workflows().Get(cmd.Context(), args[0], nil)
				if err != nil {
					return err
				}
				dst, derr := clientForProfile(cmd, toProfile, false)
				if derr != nil {
					return derr
				}
				right, err = findWorkflowByName(cmd.Context(), dst, left.Name)
				if err != nil {
					return err
				}
				if right == nil {
					return fmt.Errorf("profile %q has no workflow named %q", toProfile, left.Name)
				}
				leftLabel, rightLabel = "active:"+left.ID.String(), toProfile+":"+right.ID.String()
			} else {
				return fmt.Errorf("provide --to <profile> or --file <path>")
			}

			diff := difflib.UnifiedDiff{
				A: difflib.SplitLines(canonical(left)), B: difflib.SplitLines(canonical(right)),
				FromFile: leftLabel, ToFile: rightLabel, Context: 3,
			}
			text, _ := difflib.GetUnifiedDiffString(diff)
			if text == "" {
				if !flagQuiet {
					fmt.Fprintln(cmd.ErrOrStderr(), "no differences")
				}
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), text)
			return nil
		},
	}
	cmd.Flags().StringVar(&toProfile, "to", "", "compare against the same workflow name on another profile")
	cmd.Flags().StringVar(&file, "file", "", "compare against a local workflow file")
	return cmd
}

// --- small shared helpers ---

func sortedKeys(m map[string]*api.Workflow) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func jsonOutput() bool {
	if flagJQ != "" {
		return true
	}
	f, _ := outputFormat()
	return string(f) == "json"
}

func symbolFor(s wflint.Severity) string {
	if s == wflint.Error {
		return "✗"
	}
	return "⚠"
}
