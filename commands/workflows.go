package commands

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.Workflow]{
		Use:     "workflows",
		Aliases: []string{"workflow", "wf"},
		Short:   "Manage workflows",
		Long: "Create, list, inspect, update, delete, activate and transfer n8n workflows.\n\n" +
			"Create from a JSON file exported by n8n:\n" +
			"  n8nctl workflows create --file workflow.json\n" +
			"A workflow body requires name, nodes, connections and settings.",
		New:     func(c *api.Client) *api.Resource[api.Workflow] { return c.Workflows() },
		Columns: []string{"id", "name", "active", "isArchived", "triggerCount", "updatedAt"},
		ListFilters: []listFilter{
			{Flag: "active", Query: "active", Usage: "filter by active state", Values: []string{"true", "false"}},
			{Flag: "name", Query: "name", Usage: "filter by (substring of) name"},
			{Flag: "tags", Query: "tags", Usage: "filter by comma-separated tag names"},
			{Flag: "project", Query: "projectId", Usage: "filter by project id"},
		},
		Extra: workflowExtra,
	})
}

func workflowExtra(parent *cobra.Command, sp resourceSpec[api.Workflow]) {
	// activate / deactivate / archive / unarchive — simple POST actions.
	for _, a := range []struct{ use, short string }{
		{"activate", "Activate a workflow"},
		{"deactivate", "Deactivate a workflow"},
		{"archive", "Archive a workflow"},
		{"unarchive", "Restore an archived workflow"},
	} {
		action := a.use
		parent.AddCommand(&cobra.Command{
			Use:   action + " <id>",
			Short: a.short,
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				client, err := getAPIClient(cmd)
				if err != nil {
					return err
				}
				var wf *api.Workflow
				switch action {
				case "activate":
					wf, err = client.ActivateWorkflow(cmd.Context(), args[0])
				case "deactivate":
					wf, err = client.DeactivateWorkflow(cmd.Context(), args[0])
				case "archive":
					wf, err = client.ArchiveWorkflow(cmd.Context(), args[0])
				case "unarchive":
					wf, err = client.UnarchiveWorkflow(cmd.Context(), args[0])
				}
				if err != nil {
					if api.IsDryRun(err) {
						dryRunNotice(cmd)
						return nil
					}
					return err
				}
				return render(cmd, wf)
			},
		})
	}

	// transfer <id> --project <projectId>
	parent.AddCommand(buildTransferCmd("workflow", func(ctx context.Context, c *api.Client, id, projectID string) error {
		return c.TransferWorkflow(ctx, id, projectID)
	}))

	// tags <id> [--set id1,id2] — get or replace a workflow's tags.
	var setTags []string
	tagsCmd := &cobra.Command{
		Use:   "tags <id> [--set <tagId,...>]",
		Short: "Get or replace a workflow's tags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("set") {
				ids := splitNonEmpty(setTags)
				tags, err := client.SetWorkflowTags(cmd.Context(), args[0], ids)
				if err != nil {
					if api.IsDryRun(err) {
						dryRunNotice(cmd)
						return nil
					}
					return err
				}
				return render(cmd, tags)
			}
			tags, err := client.GetWorkflowTags(cmd.Context(), args[0])
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return render(cmd, tags)
		},
	}
	tagsCmd.Flags().StringSliceVar(&setTags, "set", nil, "replace tags with these tag ids (comma-separated; empty clears)")
	parent.AddCommand(tagsCmd)

	// Beyond-the-API workflow features.
	addWorkflowSyncCmd(parent)
	addWorkflowSearchCmd(parent)
	// Workflows-as-code (GitOps): convert, lint, apply, diff, autofix.
	addWorkflowGitopsCmds(parent)
	parent.AddCommand(writeHints(workflowBulkCmd()))
}

// workflowCreateBody builds a create/update payload from a fetched workflow,
// keeping only the writable structural fields (read-only fields like id, active,
// versionId, and tags are dropped). Missing structural fields are defaulted so
// the n8n "create" contract (name, nodes, connections, settings required) holds.
func workflowCreateBody(wf *api.Workflow) map[string]any {
	body := map[string]any{
		"name":        wf.Name,
		"nodes":       defaultRaw(wf.Nodes, "[]"),
		"connections": defaultRaw(wf.Connections, "{}"),
		"settings":    defaultRaw(wf.Settings, "{}"),
	}
	if len(wf.StaticData) > 0 && string(wf.StaticData) != "null" {
		body["staticData"] = wf.StaticData
	}
	return body
}

// defaultRaw returns r, or a fallback literal when r is empty/null, so the n8n
// create contract (nodes/connections/settings required) is always satisfied.
func defaultRaw(r json.RawMessage, fallback string) json.RawMessage {
	if len(r) == 0 || string(r) == "null" {
		return json.RawMessage(fallback)
	}
	return r
}

func splitNonEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
