package commands

import (
	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	var days int
	var categories []string
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Generate a security audit of the instance",
		Long: "Run n8n's built-in security audit and print the report.\n\n" +
			"  n8nctl audit\n" +
			"  n8nctl audit --categories credentials,nodes --days 30 -o json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			report, err := client.GenerateAudit(cmd.Context(), api.AuditOptions{
				DaysAbandonedWorkflow: days,
				Categories:            categories,
			})
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, report)
		},
	}
	cmd.Flags().IntVar(&days, "days", 0, "days of inactivity before a workflow is flagged as abandoned")
	cmd.Flags().StringSliceVar(&categories, "categories", nil,
		"restrict to categories: credentials,database,nodes,filesystem,instance")
	_ = cmd.RegisterFlagCompletionFunc("categories",
		fixedCompletions([]string{"credentials", "database", "nodes", "filesystem", "instance"}))
	rootCmd.AddCommand(cmd)
}
