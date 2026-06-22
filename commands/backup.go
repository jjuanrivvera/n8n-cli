package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// n8n stores workflows, tags, and variables in its database, with no built-in
// export-to-disk on the Community edition. backup/restore turn an instance into a
// directory of pretty-printed JSON that lives happily in git, so workflows can be
// versioned, diffed, code-reviewed, and re-applied to another instance.

func init() {
	rootCmd.AddCommand(backupCmd(), restoreCmd())
}

func backupCmd() *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "backup --out <dir>",
		Short: "Export workflows, tags, and variables to a directory of JSON",
		Long: "Snapshot the active instance to disk for git-based versioning and backup.\n" +
			"Writes one file per workflow plus tags.json, variables.json, a credentials\n" +
			"inventory (metadata only — secrets are never exported), and a manifest.\n\n" +
			"  n8nctl backup --out ./n8n-backup\n" +
			"  n8nctl --profile prod backup --out ./backups/prod",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if outDir == "" {
				return fmt.Errorf("--out <dir> is required")
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

			workflows, err := client.Workflows().ListAll(context.Background(), api.ListParams{}, 0)
			if err != nil {
				return err
			}
			for i := range workflows {
				wf := &workflows[i]
				full, gerr := client.Workflows().Get(context.Background(), wf.ID.String(), nil)
				if gerr != nil {
					return fmt.Errorf("fetching workflow %s: %w", wf.ID, gerr)
				}
				fname := slugify(wf.Name) + "." + wf.ID.String() + ".json"
				if err := writeJSON(filepath.Join(wfDir, fname), full); err != nil {
					return err
				}
			}

			tags, err := client.Tags().ListAll(context.Background(), api.ListParams{}, 0)
			if err != nil {
				return err
			}
			if err := writeJSON(filepath.Join(outDir, "tags.json"), tags); err != nil {
				return err
			}

			variables, err := client.Variables().ListAll(context.Background(), api.ListParams{}, 0)
			if err != nil {
				return err
			}
			if err := writeJSON(filepath.Join(outDir, "variables.json"), variables); err != nil {
				return err
			}

			// Credential inventory: metadata only. Secrets are write-only in the API.
			creds, cerr := client.Credentials().ListAll(context.Background(), api.ListParams{}, 0)
			if cerr == nil {
				_ = writeJSON(filepath.Join(outDir, "credentials.inventory.json"), creds)
			}

			manifest := map[string]any{
				"profile":    profile,
				"baseUrl":    client.BaseURL(),
				"exportedAt": time.Now().UTC().Format(time.RFC3339),
				"counts": map[string]int{
					"workflows": len(workflows),
					"tags":      len(tags),
					"variables": len(variables),
				},
				"note": "credentials.inventory.json holds metadata only; secret values are not exported by the n8n API",
			}
			if err := writeJSON(filepath.Join(outDir, "manifest.json"), manifest); err != nil {
				return err
			}

			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "backed up %d workflows, %d tags, %d variables to %s\n",
					len(workflows), len(tags), len(variables), outDir)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&outDir, "out", "", "output directory (required)")
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

			var created, updated int
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
					continue
				}
				raw, rerr := os.ReadFile(filepath.Join(wfDir, e.Name())) //nolint:gosec // path within user-supplied backup dir
				if rerr != nil {
					return rerr
				}
				var wf api.Workflow
				if jerr := json.Unmarshal(raw, &wf); jerr != nil {
					return fmt.Errorf("parsing %s: %w", e.Name(), jerr)
				}
				body := workflowCreateBody(&wf)

				var result *api.Workflow
				if updateByName {
					existing, ferr := findWorkflowByName(client, wf.Name)
					if ferr != nil {
						return ferr
					}
					if existing != nil {
						result, err = client.Workflows().Update(context.Background(), existing.ID.String(), body)
						updated++
					} else {
						result, err = client.Workflows().Create(context.Background(), body)
						created++
					}
				} else {
					result, err = client.Workflows().Create(context.Background(), body)
					created++
				}
				if err != nil {
					if api.IsDryRun(err) {
						continue
					}
					return fmt.Errorf("restoring %q: %w", wf.Name, err)
				}
				if activate && result != nil && result.ID != "" {
					_, _ = client.ActivateWorkflow(context.Background(), result.ID.String())
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
