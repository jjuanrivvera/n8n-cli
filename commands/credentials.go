package commands

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.Credential]{
		Use:     "credentials",
		Aliases: []string{"credential", "cred", "creds"},
		Short:   "Manage credentials",
		Long: "Create, inspect, update, delete and transfer credentials. Secret values are\n" +
			"write-only: they are sent on create/update but never returned by the API.\n\n" +
			"Discover a type's required fields first:\n" +
			"  n8nctl credentials schema githubApi\n" +
			"  n8nctl credentials create --set name='My GH' --set type=githubApi --set data='{\"accessToken\":\"...\"}'",
		New:     func(c *api.Client) *api.Resource[api.Credential] { return c.Credentials() },
		Columns: []string{"id", "name", "type", "createdAt", "updatedAt"},
		Extra:   credentialExtra,
	})
}

func credentialExtra(parent *cobra.Command, _ resourceSpec[api.Credential]) {
	// schema <credentialTypeName> — show the JSON schema of a credential type.
	parent.AddCommand(readOnlyHints(&cobra.Command{
		Use:   "schema <credentialTypeName>",
		Short: "Show the field schema for a credential type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			raw, err := client.CredentialSchema(cmd.Context(), args[0])
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return render(cmd, raw)
		},
	}))

	// transfer <id> --project <projectId>
	parent.AddCommand(buildTransferCmd("credential", func(ctx context.Context, c *api.Client, id, projectID string) error {
		return c.TransferCredential(ctx, id, projectID)
	}))
}
