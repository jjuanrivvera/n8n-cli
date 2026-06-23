package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// addWorkflowSyncCmd adds `workflows sync`, which promotes a workflow from one
// instance (profile) to another. This is the multi-instance differentiator: n8n's
// own Git-based Source Control is an Enterprise feature, so Community users have
// no built-in way to promote a workflow dev -> staging -> prod. sync does it over
// the public API, working on any edition.
func addWorkflowSyncCmd(parent *cobra.Command) {
	var (
		from, to     string
		updateByName bool
		activate     bool
	)
	cmd := &cobra.Command{
		Use:   "sync <id> --to <profile>",
		Short: "Promote a workflow to another instance (profile)",
		Long: "Copy a workflow from one instance to another over the API. Read-only fields\n" +
			"(id, active state, version) are stripped; nodes, connections and settings are\n" +
			"carried over. By default a new workflow is created on the destination; use\n" +
			"--update-by-name to overwrite an existing workflow with the same name.\n\n" +
			"  n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate\n\n" +
			"Credentials are referenced by id and are NOT copied — create matching\n" +
			"credentials on the destination first (see `n8nctl credentials`).",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if to == "" {
				return fmt.Errorf("--to <profile> is required")
			}
			source := from
			if source == "" {
				active, _, err := activeProfile()
				if err != nil {
					return err
				}
				source = active
			}
			if source == to {
				return fmt.Errorf("source and destination profiles are the same (%q)", to)
			}

			// Source read is always live; destination honors --dry-run.
			srcClient, err := clientForProfile(cmd, source, false)
			if err != nil {
				return fmt.Errorf("source profile %q: %w", source, err)
			}
			wf, err := srcClient.Workflows().Get(cmd.Context(), args[0], nil)
			if err != nil {
				return fmt.Errorf("reading workflow from %q: %w", source, err)
			}

			dstClient, err := clientForProfile(cmd, to, flagDryRun)
			if err != nil {
				return fmt.Errorf("destination profile %q: %w", to, err)
			}
			body := workflowCreateBody(wf)

			var result *api.Workflow
			if updateByName {
				existing, ferr := findWorkflowByName(cmd.Context(), dstClient, wf.Name)
				if ferr != nil {
					return ferr
				}
				if existing != nil {
					result, err = dstClient.Workflows().Update(cmd.Context(), existing.ID.String(), body)
				} else {
					result, err = dstClient.Workflows().Create(cmd.Context(), body)
				}
			} else {
				result, err = dstClient.Workflows().Create(cmd.Context(), body)
			}
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}

			if activate && result != nil && result.ID != "" {
				if _, err := dstClient.ActivateWorkflow(cmd.Context(), result.ID.String()); err != nil && !api.IsDryRun(err) {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: synced but failed to activate: %v\n", err)
				}
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "synced %q from %q to %q\n", wf.Name, source, to)
			}
			return render(cmd, result)
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "source profile (default: active profile)")
	cmd.Flags().StringVar(&to, "to", "", "destination profile (required)")
	cmd.Flags().BoolVar(&updateByName, "update-by-name", false, "overwrite an existing destination workflow with the same name")
	cmd.Flags().BoolVar(&activate, "activate", false, "activate the workflow on the destination after syncing")
	_ = cmd.RegisterFlagCompletionFunc("to", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return profileNames(), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("from", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return profileNames(), cobra.ShellCompDirectiveNoFileComp
	})
	parent.AddCommand(cmd)
}

// findWorkflowByName returns the first workflow whose name matches, or nil.
func findWorkflowByName(ctx context.Context, client *api.Client, name string) (*api.Workflow, error) {
	items, err := client.Workflows().ListAll(ctx, api.ListParams{}, 0)
	if err != nil {
		if api.IsDryRun(err) {
			return nil, nil
		}
		return nil, err
	}
	for i := range items {
		if items[i].Name == name {
			return &items[i], nil
		}
	}
	return nil, nil
}
