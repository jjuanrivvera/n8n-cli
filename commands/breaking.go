package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/wffile"
	"github.com/jjuanrivvera/n8n-cli/internal/wflint"
)

// breakingRow is one outdated node, flattened with its workflow for rendering.
type breakingRow struct {
	Workflow       string   `json:"workflow"`
	Node           string   `json:"node"`
	Type           string   `json:"type"`
	CurrentVersion int      `json:"currentVersion"`
	LatestVersion  int      `json:"latestVersion"`
	UnknownParams  []string `json:"unknownParams,omitempty"`
}

// workflowBreakingChangesCmd reports nodes pinned to an older typeVersion than the
// catalog's latest — the upgrade-risk signal — plus parameters those nodes use that
// the catalog does not recognize for any version of the node.
func workflowBreakingChangesCmd() *cobra.Command {
	var dir string
	var files []string
	var remote bool
	cmd := &cobra.Command{
		Use:     "breaking-changes [--dir <dir> | -f <file>... | --remote | <id>]",
		Aliases: []string{"breaking"},
		Short:   "Find nodes pinned to an outdated typeVersion (upgrade risk)",
		Long: "Compare each workflow's nodes against the embedded node catalog and report\n" +
			"those pinned to an older typeVersion than the latest known one, along with any\n" +
			"parameters they use that the catalog does not recognize for the node (renamed,\n" +
			"removed, or typos). Community/custom nodes are skipped. Informational — exits 0.",
		Args: cobra.MaximumNArgs(1),
		Example: "  n8nctl workflows breaking-changes --dir ./workflows\n" +
			"  n8nctl workflows breaking-changes 42\n" +
			"  n8nctl workflows breaking-changes --remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			workflows := map[string]*api.Workflow{}
			switch {
			case len(args) == 1:
				client, err := getReadClient(cmd)
				if err != nil {
					return err
				}
				wf, err := client.Workflows().Get(cmd.Context(), args[0], nil)
				if err != nil {
					if api.IsDryRun(err) {
						return nil
					}
					return err
				}
				workflows[args[0]] = wf
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
				return fmt.Errorf("provide --dir, -f/--file, --remote, or a workflow id")
			}

			var rows []breakingRow
			for _, k := range sortedKeys(workflows) {
				for _, vi := range wflint.BreakingChanges(workflows[k]) {
					rows = append(rows, breakingRow{
						Workflow: k, Node: vi.Node, Type: vi.Type,
						CurrentVersion: vi.CurrentVersion, LatestVersion: vi.LatestVersion,
						UnknownParams: vi.UnknownParams,
					})
				}
			}

			if jsonOutput() {
				return render(cmd, rows)
			}
			if len(rows) == 0 {
				if !flagQuiet {
					fmt.Fprintln(cmd.OutOrStdout(), "no outdated nodes found")
				}
				return nil
			}
			for _, r := range rows {
				line := fmt.Sprintf("%s · %s (%s): typeVersion %d → latest %d",
					r.Workflow, r.Node, r.Type, r.CurrentVersion, r.LatestVersion)
				if len(r.UnknownParams) > 0 {
					line += "; params not in catalog: " + strings.Join(r.UnknownParams, ", ")
				}
				fmt.Fprintln(cmd.OutOrStdout(), line)
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "%d node(s) on an outdated typeVersion\n", len(rows))
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "scan all workflow files in a directory")
	cmd.Flags().StringSliceVarP(&files, "file", "f", nil, "scan specific files")
	cmd.Flags().BoolVar(&remote, "remote", false, "scan live workflows from the instance")
	return cmd
}
