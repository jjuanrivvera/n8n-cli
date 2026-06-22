package commands

import "github.com/jjuanrivvera/n8n-cli/internal/api"

func init() {
	registerResource(resourceSpec[api.Tag]{
		Use:     "tags",
		Aliases: []string{"tag"},
		Short:   "Manage workflow tags",
		Long:    "Create, list, update and delete tags. Create with --set name=Production.",
		New:     func(c *api.Client) *api.Resource[api.Tag] { return c.Tags() },
		Columns: []string{"id", "name", "createdAt", "updatedAt"},
	})
}
