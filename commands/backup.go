package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/wffile"
)

// n8n stores workflows, tags, and variables in its database, with no built-in
// export-to-disk on the Community edition. backup/restore turn an instance into a
// directory of pretty-printed JSON that lives happily in git, so workflows can be
// versioned, diffed, code-reviewed, and re-applied to another instance.

func init() {
	rootCmd.AddCommand(readOnlyHints(backupCmd()), writeHints(restoreCmd()))
}

func backupCmd() *cobra.Command {
	var outDir, format string
	var externalize int
	cmd := &cobra.Command{
		Use:   "backup --out <dir>",
		Short: "Export workflows, tags, and variables to a directory (JSON or YAML)",
		Long: "Snapshot the active instance to disk for git-based versioning and backup.\n" +
			"Writes one file per workflow plus tags.json, variables.json, a credentials\n" +
			"inventory (metadata only — secrets are never exported), and a manifest.\n\n" +
			"  n8nctl backup --out ./n8n-backup\n" +
			"  n8nctl --profile prod backup --out ./backups/prod --format yaml --externalize 5",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if outDir == "" {
				return fmt.Errorf("--out <dir> is required")
			}
			wfFormat := wffile.Format(strings.ToLower(format))
			if wfFormat != wffile.JSON && wfFormat != wffile.YAML {
				return fmt.Errorf("--format must be json or yaml")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			profile, _, _ := activeProfile()

			wfDir := filepath.Join(outDir, "workflows")
			if err := os.MkdirAll(wfDir, 0o755); err != nil { //nolint:gosec // backup dir is user-facing, not secret
				return err
			}

			workflows, err := client.Workflows().ListAll(cmd.Context(), api.ListParams{}, 0)
			if err != nil {
				return err
			}
			var failures []string
			savedWorkflows := 0
			for i := range workflows {
				full := &workflows[i]
				// The public-API list endpoint returns full workflow bodies
				// (nodes/connections are required fields in the workflow schema), so
				// we avoid a redundant GET per workflow. Older n8n versions may omit
				// nodes from the list response; fall back to a per-workflow fetch only
				// when they're missing.
				if len(full.Nodes) == 0 || string(full.Nodes) == "null" {
					fetched, gerr := client.Workflows().Get(cmd.Context(), full.ID.String(), nil)
					if gerr != nil {
						// Collect-and-continue: one unreadable workflow must not silently
						// abort the whole backup, nor be reported as a clean success.
						fmt.Fprintf(cmd.ErrOrStderr(), "failed to back up workflow %s: %v\n", full.ID, gerr)
						failures = append(failures, "workflow "+full.ID.String())
						continue
					}
					full = fetched
				}
				stem := slugify(full.Name) + "." + full.ID.String()
				main, subfiles, eerr := wffile.EncodeExternalized(full, wfFormat, stem, externalize)
				if eerr != nil {
					return eerr
				}
				if err := os.WriteFile(filepath.Join(wfDir, stem+"."+string(wfFormat)), main, 0o600); err != nil {
					return err
				}
				for rel, content := range subfiles {
					sub := filepath.Join(wfDir, filepath.FromSlash(rel))
					if err := os.MkdirAll(filepath.Dir(sub), 0o755); err != nil { //nolint:gosec // backup dir
						return err
					}
					if err := os.WriteFile(sub, content, 0o600); err != nil {
						return err
					}
				}
				savedWorkflows++
			}

			// Tags, variables, and credentials are best-effort: on a Community
			// instance some are unlicensed (403). A forbidden/unlicensed section is an
			// expected skip; any other error (network, 5xx) is a real failure that
			// must be surfaced, not silently recorded as a clean skip.
			counts := map[string]int{"workflows": savedWorkflows}
			var skipped []string
			optional := func(name, file string, fetch func() (any, int, error)) error {
				v, n, err := fetch()
				if err != nil {
					if api.IsForbidden(err) {
						fmt.Fprintf(cmd.ErrOrStderr(), "skipping %s (unlicensed or forbidden): %v\n", name, err)
						skipped = append(skipped, name)
						return nil
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "failed to back up %s: %v\n", name, err)
					failures = append(failures, name)
					return nil
				}
				counts[name] = n
				return writeJSON(filepath.Join(outDir, file), v)
			}

			if err := optional("tags", "tags.json", func() (any, int, error) {
				t, e := client.Tags().ListAll(cmd.Context(), api.ListParams{}, 0)
				return t, len(t), e
			}); err != nil {
				return err
			}
			if err := optional("variables", "variables.json", func() (any, int, error) {
				v, e := client.Variables().ListAll(cmd.Context(), api.ListParams{}, 0)
				return v, len(v), e
			}); err != nil {
				return err
			}
			// Credential inventory: metadata only. Secrets are write-only in the API.
			if err := optional("credentials", "credentials.inventory.json", func() (any, int, error) {
				c, e := client.Credentials().ListAll(cmd.Context(), api.ListParams{}, 0)
				return c, len(c), e
			}); err != nil {
				return err
			}

			manifest := map[string]any{
				"profile":    profile,
				"baseUrl":    client.BaseURL(),
				"exportedAt": time.Now().UTC().Format(time.RFC3339),
				"counts":     counts,
				"skipped":    skipped,
				"failures":   failures,
				"note":       "credentials.inventory.json holds metadata only; secret values are not exported by the n8n API",
			}
			if err := writeJSON(filepath.Join(outDir, "manifest.json"), manifest); err != nil {
				return err
			}

			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "backed up %d workflows, %d tags, %d variables to %s\n",
					counts["workflows"], counts["tags"], counts["variables"], outDir)
				if len(skipped) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "skipped (unlicensed or unavailable): %s\n", strings.Join(skipped, ", "))
				}
			}
			// A partial backup must exit non-zero so callers (and CI) never mistake
			// it for a complete snapshot.
			if len(failures) > 0 {
				return fmt.Errorf("backup incomplete: %d section(s) failed: %s", len(failures), strings.Join(failures, ", "))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outDir, "out", "", "output directory (required)")
	cmd.Flags().StringVar(&format, "format", "json", "workflow file format: json or yaml")
	cmd.Flags().IntVar(&externalize, "externalize", 0, "externalize code fields longer than N lines (0 = off)")
	return cmd
}

func restoreCmd() *cobra.Command {
	var inDir string
	var updateByName, activate bool
	cmd := &cobra.Command{
		Use:   "restore --in <dir>",
		Short: "Recreate workflows from a backup directory",
		Long: "Apply the workflows in a backup directory to the active instance. By default\n" +
			"each workflow is created new; --update-by-name overwrites an existing workflow\n" +
			"with the same name. Credentials are referenced by id and must already exist.\n\n" +
			"  n8nctl --profile staging restore --in ./n8n-backup --update-by-name",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if inDir == "" {
				return fmt.Errorf("--in <dir> is required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			wfDir := filepath.Join(inDir, "workflows")
			entries, err := os.ReadDir(wfDir)
			if err != nil {
				return fmt.Errorf("reading %s: %w", wfDir, err)
			}

			loader := wffile.DirLoader(wfDir)
			var created, updated int
			for _, e := range entries {
				if e.IsDir() || !isWorkflowFile(e.Name()) {
					continue
				}
				raw, rerr := os.ReadFile(filepath.Join(wfDir, e.Name())) //nolint:gosec // path within user-supplied backup dir
				if rerr != nil {
					return rerr
				}
				wfp, jerr := wffile.DecodeWithFiles(raw, wffile.FormatFromPath(e.Name()), loader)
				if jerr != nil {
					return fmt.Errorf("parsing %s: %w", e.Name(), jerr)
				}
				wf := *wfp
				body := workflowCreateBody(&wf)

				var result *api.Workflow
				if updateByName {
					existing, ferr := findWorkflowByName(cmd.Context(), client, wf.Name)
					if ferr != nil {
						return ferr
					}
					if existing != nil {
						result, err = client.Workflows().Update(cmd.Context(), existing.ID.String(), body)
						updated++
					} else {
						result, err = client.Workflows().Create(cmd.Context(), body)
						created++
					}
				} else {
					result, err = client.Workflows().Create(cmd.Context(), body)
					created++
				}
				if err != nil {
					if api.IsDryRun(err) {
						continue
					}
					return fmt.Errorf("restoring %q: %w", wf.Name, err)
				}
				if activate && result != nil && result.ID != "" {
					if _, aerr := client.ActivateWorkflow(cmd.Context(), result.ID.String()); aerr != nil && !api.IsDryRun(aerr) {
						fmt.Fprintf(cmd.ErrOrStderr(), "warning: restored %q but failed to activate: %v\n", wf.Name, aerr)
					}
				}
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "restored %d created, %d updated\n", created, updated)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&inDir, "in", "", "backup directory to restore from (required)")
	cmd.Flags().BoolVar(&updateByName, "update-by-name", false, "overwrite existing workflows with the same name")
	cmd.Flags().BoolVar(&activate, "activate", false, "activate each restored workflow")
	return cmd
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

var slugRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// slugify makes a filesystem-friendly stem from a workflow name.
func slugify(name string) string {
	s := slugRe.ReplaceAllString(strings.TrimSpace(name), "-")
	s = strings.Trim(strings.ToLower(s), "-")
	if s == "" {
		return "workflow"
	}
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}
