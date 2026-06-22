package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	sc := &cobra.Command{
		Use:     "source-control",
		Aliases: []string{"sc"},
		Short:   "Interact with the Source Control (Git) integration",
	}

	var force bool
	var variablesJSON string
	pull := &cobra.Command{
		Use:   "pull",
		Short: "Pull changes from the connected remote repository",
		Long: "Requires the licensed Source Control feature connected to a repository.\n" +
			"Use --force to discard local changes on conflict.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			var vars map[string]any
			if variablesJSON != "" {
				if err := json.Unmarshal([]byte(variablesJSON), &vars); err != nil {
					return fmt.Errorf("parsing --variables JSON: %w", err)
				}
			}
			res, err := client.SourceControlPull(context.Background(), force, vars)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, res)
		},
	}
	pull.Flags().BoolVar(&force, "force", false, "discard local changes / resolve conflicts in favor of remote")
	pull.Flags().StringVar(&variablesJSON, "variables", "", "JSON object of variable overrides to apply during pull")
	sc.AddCommand(pull)
	rootCmd.AddCommand(sc)
}
