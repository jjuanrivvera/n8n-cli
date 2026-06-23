package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	rootCmd.AddCommand(readOnlyHints(statsCmd()))
}

type statRow struct {
	Metric string `json:"metric"`
	Value  int    `json:"value"`
}

// statsCmd composes a one-shot instance summary from the workflows and executions
// endpoints: how many workflows exist/are active, and the recent execution mix.
func statsCmd() *cobra.Command {
	var recent int
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "One-shot instance health summary",
		Long:    "Summarize an instance: total/active/archived workflows, and the status mix of the most recent executions.",
		Args:    cobra.NoArgs,
		Example: "  n8nctl stats\n  n8nctl --profile prod stats -o json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			ctx := cmd.Context()

			wfs, err := client.Workflows().ListAll(ctx, api.ListParams{}, 0)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			var active, archived int
			for i := range wfs {
				if bool(wfs[i].Active) {
					active++
				}
				if bool(wfs[i].IsArchived) {
					archived++
				}
			}
			rows := []statRow{
				{"workflows.total", len(wfs)},
				{"workflows.active", active},
				{"workflows.inactive", len(wfs) - active},
				{"workflows.archived", archived},
			}

			// Recent executions are best-effort; on instances where the endpoint is
			// restricted, summarize workflows only rather than failing.
			exs, _, eerr := client.Executions().List(ctx, api.ListParams{Limit: recent})
			if eerr == nil {
				byStatus := map[string]int{}
				for i := range exs {
					byStatus[exs[i].Status]++
				}
				rows = append(rows,
					statRow{fmt.Sprintf("executions.recent (last %d)", recent), len(exs)},
					statRow{"executions.success", byStatus["success"]},
					statRow{"executions.error", byStatus["error"]},
					statRow{"executions.crashed", byStatus["crashed"]},
					statRow{"executions.waiting", byStatus["waiting"]},
				)
			} else if !api.IsForbidden(eerr) && !api.IsDryRun(eerr) {
				return eerr
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "executions summary unavailable: %v\n", eerr)
			}

			return render(cmd, rows, "metric", "value")
		},
	}
	cmd.Flags().IntVar(&recent, "recent", 100, "number of recent executions to summarize")
	return cmd
}
