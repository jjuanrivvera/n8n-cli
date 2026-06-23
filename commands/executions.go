package commands

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.Execution]{
		Use:     "executions",
		Aliases: []string{"execution", "exec"},
		Short:   "Inspect and control workflow executions",
		Long:    "Executions are read-only with retry/stop actions — n8n creates them by running workflows.",
		New:     func(c *api.Client) *api.Resource[api.Execution] { return c.Executions() },
		Columns: []string{"id", "workflowId", "status", "mode", "finished", "startedAt", "stoppedAt"},
		ListFilters: []listFilter{
			{Flag: "status", Query: "status", Usage: "filter by status",
				Values: []string{"canceled", "crashed", "error", "new", "running", "success", "unknown", "waiting"}},
			{Flag: "workflow", Query: "workflowId", Usage: "filter by workflow id"},
			{Flag: "project", Query: "projectId", Usage: "filter by project id"},
			{Flag: "include-data", Query: "includeData", Usage: "include full execution data (true/false)", Values: []string{"true", "false"}},
		},
		NoCreate: true,
		NoUpdate: true,
		Extra:    executionExtra,
	})
}

func executionExtra(parent *cobra.Command, _ resourceSpec[api.Execution]) {
	// get supports --include-data, so override nothing but add a richer get is unnecessary;
	// instead expose retry and stop.
	var loadWorkflow bool
	retry := &cobra.Command{
		Use:   "retry <id>",
		Short: "Retry a failed execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			ex, err := client.RetryExecution(cmd.Context(), args[0], loadWorkflow)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, ex)
		},
	}
	retry.Flags().BoolVar(&loadWorkflow, "load-workflow", false, "re-load the current workflow definition instead of the one used originally")
	parent.AddCommand(retry)

	parent.AddCommand(&cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a running execution",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			ex, err := client.StopExecution(cmd.Context(), args[0])
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return render(cmd, ex)
		},
	})

	// Replace the generic get so it can pass ?includeData=true.
	var includeData bool
	get := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a single execution by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			q := url.Values{}
			if includeData {
				q.Set("includeData", "true")
			}
			ex, err := client.Executions().Get(cmd.Context(), args[0], q)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return render(cmd, ex)
		},
	}
	get.Flags().BoolVar(&includeData, "include-data", false, "include full execution data in the response")
	// Remove the generic get (added by buildResourceCmd) and use this one.
	for _, c := range parent.Commands() {
		if c.Name() == "get" {
			parent.RemoveCommand(c)
		}
	}
	parent.AddCommand(readOnlyHints(get))

	parent.AddCommand(destructiveHints(executionPruneCmd()), readOnlyHints(executionWatchCmd()))
}
