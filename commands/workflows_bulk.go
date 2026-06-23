package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// workflowBulkCmd groups maintenance-window operations that flip many workflows
// at once, selected by tag.
func workflowBulkCmd() *cobra.Command {
	bulk := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk activate/deactivate workflows by tag",
		Long: "Flip every workflow carrying a tag in one command — useful for maintenance\n" +
			"windows (deactivate the `prod` set, do the work, reactivate). Always previews;\n" +
			"pass --yes to skip the confirmation or --dry-run to only list.",
	}
	bulk.AddCommand(writeHints(bulkToggleCmd("activate", true)), writeHints(bulkToggleCmd("deactivate", false)))
	return bulk
}

func bulkToggleCmd(verb string, activate bool) *cobra.Command {
	var tag string
	var yes bool
	cmd := &cobra.Command{
		Use:     verb + " --tag <name>",
		Short:   verb + " every workflow carrying a tag",
		Args:    cobra.NoArgs,
		Example: "  n8nctl workflows bulk " + verb + " --tag prod",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if tag == "" {
				return fmt.Errorf("--tag is required")
			}
			client, err := getReadClient(cmd)
			if err != nil {
				return err
			}
			wfs, err := client.Workflows().ListAll(cmd.Context(),
				api.ListParams{Extra: url.Values{"tags": {tag}}}, 0)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(wfs) == 0 {
				fmt.Fprintf(out, "no workflows carry the tag %q\n", tag)
				return nil
			}
			if flagDryRun {
				fmt.Fprintf(out, "would %s %d workflow(s):\n", verb, len(wfs))
				for i := range wfs {
					fmt.Fprintf(out, "  %s  %s\n", wfs[i].ID, wfs[i].Name)
				}
				return nil
			}
			if !yes && !confirm(cmd, fmt.Sprintf("%s %d workflow(s) tagged %q?", verb, len(wfs), tag)) {
				fmt.Fprintln(out, "aborted")
				return nil
			}
			done, failed := 0, 0
			for i := range wfs {
				id := wfs[i].ID.String()
				var aerr error
				if activate {
					_, aerr = client.ActivateWorkflow(cmd.Context(), id)
				} else {
					_, aerr = client.DeactivateWorkflow(cmd.Context(), id)
				}
				if aerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  %s (%s): %v\n", wfs[i].Name, id, aerr)
					failed++
					continue
				}
				done++
			}
			fmt.Fprintf(out, "%sd %d workflow(s)", verb, done)
			if failed > 0 {
				fmt.Fprintf(out, ", %d failed", failed)
			}
			fmt.Fprintln(out)
			if failed > 0 {
				return fmt.Errorf("%d workflow(s) could not be %sd", failed, verb)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "tag name selecting the workflows (required)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}
