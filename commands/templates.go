package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// The `templates` command browses the public n8n template gallery (api.n8n.io)
// and deploys a template into the active instance. Search/get are pure gallery
// reads (no instance, no key); deploy creates a workflow on the active profile.

func init() {
	templates := &cobra.Command{
		Use:     "templates",
		Aliases: []string{"template"},
		Short:   "Browse and deploy workflows from the n8n template gallery",
		Long: "Search the public n8n template gallery (api.n8n.io), inspect a template's\n" +
			"workflow definition, and deploy one straight into the active instance.",
	}

	var searchLimit int
	search := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search the template gallery",
		Args:    cobra.ExactArgs(1),
		Example: "  n8nctl templates search slack\n  n8nctl templates search \"google sheets\" --limit 5",
		RunE: func(cmd *cobra.Command, args []string) error {
			hits, err := api.NewTemplateAPI().Search(cmd.Context(), args[0], searchLimit)
			if err != nil {
				return err
			}
			if len(hits) == 0 {
				return fmt.Errorf("no templates match %q", args[0])
			}
			return render(cmd, hits, "id", "name", "totalViews")
		},
	}
	search.Flags().IntVar(&searchLimit, "limit", 20, "max results")

	get := &cobra.Command{
		Use:     "get <id>",
		Short:   "Print a template's workflow definition",
		Args:    cobra.ExactArgs(1),
		Example: "  n8nctl templates get 1750 -o json",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := api.NewTemplateAPI().Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			// The definition is the workflow JSON; print it directly so it can be
			// piped to a file and applied.
			var pretty any
			if uerr := json.Unmarshal(d.Definition, &pretty); uerr != nil {
				return uerr
			}
			return render(cmd, pretty)
		},
	}

	var deployName string
	var activate bool
	deploy := &cobra.Command{
		Use:   "deploy <id>",
		Short: "Create a workflow on the active instance from a template",
		Long: "Fetch a gallery template and create it as a new workflow on the active\n" +
			"instance. Credentials are NOT included — open the workflow and connect them\n" +
			"afterwards. Honors --dry-run.",
		Args:    cobra.ExactArgs(1),
		Example: "  n8nctl templates deploy 1750 --name \"My Slack bot\"\n  n8nctl --profile dev templates deploy 1750 --activate",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := api.NewTemplateAPI().Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			body, err := templateCreateBody(d, deployName)
			if err != nil {
				return err
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			wf, err := client.Workflows().Create(cmd.Context(), body)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if activate {
				if _, aerr := client.ActivateWorkflow(cmd.Context(), wf.ID.String()); aerr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "created but failed to activate: %v\n", aerr)
				}
			}
			return render(cmd, wf)
		},
	}
	deploy.Flags().StringVar(&deployName, "name", "", "name for the new workflow (default: the template's name)")
	deploy.Flags().BoolVar(&activate, "activate", false, "activate the workflow after creating it")

	templates.AddCommand(readOnlyHints(search), readOnlyHints(get), writeHints(deploy))
	rootCmd.AddCommand(readOnlyHints(templates))
}

// templateCreateBody turns a gallery template into a clean workflow create body:
// name + nodes + connections + settings only (the gallery definition may carry
// extra fields the create endpoint rejects).
func templateCreateBody(d *api.TemplateDetail, name string) (map[string]any, error) {
	var def map[string]any
	if err := json.Unmarshal(d.Definition, &def); err != nil {
		return nil, fmt.Errorf("template definition is not an object: %w", err)
	}
	if name == "" {
		name = d.Name
	}
	if name == "" {
		name = "Imported template"
	}
	nodes, err := wantArray(def["nodes"], "nodes")
	if err != nil {
		return nil, err
	}
	connections, err := wantObject(def["connections"], "connections")
	if err != nil {
		return nil, err
	}
	settings, err := wantObject(def["settings"], "settings")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"name":        name,
		"nodes":       nodes,
		"connections": connections,
		"settings":    settings,
	}, nil
}

// wantArray returns v as a JSON array, an empty array if absent, or an error if it
// is present but the wrong shape — so a malformed template fails clearly here
// rather than as an opaque 422 from the instance.
func wantArray(v any, field string) ([]any, error) {
	if v == nil {
		return []any{}, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("template %q is not an array", field)
	}
	return arr, nil
}

func wantObject(v any, field string) (map[string]any, error) {
	if v == nil {
		return map[string]any{}, nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("template %q is not an object", field)
	}
	return obj, nil
}
