package commands

import "github.com/jjuanrivvera/n8n-cli/internal/api"

func init() {
	registerResource(resourceSpec[api.Variable]{
		Use:     "variables",
		Aliases: []string{"variable", "var", "vars"},
		Short:   "Manage instance variables",
		Long: "Create, list, update and delete variables. The API has no get-by-id endpoint,\n" +
			"so `get <id>` is served by matching id or key within the full list.\n\n" +
			"  n8nctl variables create --set key=API_BASE --set value=https://api.example.com",
		New:     func(c *api.Client) *api.Resource[api.Variable] { return c.Variables() },
		Columns: []string{"id", "key", "value", "type"},
		ListFilters: []listFilter{
			{Flag: "project", Query: "projectId", Usage: "filter by project id"},
			{Flag: "state", Query: "state", Usage: "filter by state", Values: []string{"empty"}},
		},
		NoGet:    true, // no GET /variables/{id}; resolve via list
		IDFields: []string{"id", "key"},
	})
}
