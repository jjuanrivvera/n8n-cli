package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.User]{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Manage users (instance owner only)",
		Long: "List and inspect users, invite new ones, change roles, and delete users.\n" +
			"n8n invites users rather than creating them directly, so use `users invite`.",
		New:     func(c *api.Client) *api.Resource[api.User] { return c.Users() },
		Columns: []string{"id", "email", "firstName", "lastName", "role", "isPending"},
		ListFilters: []listFilter{
			{Flag: "project", Query: "projectId", Usage: "filter by project id"},
			{Flag: "include-role", Query: "includeRole", Usage: "include each user's role (true/false)", Values: []string{"true", "false"}},
		},
		NoCreate: true, // creation is an invite with a bespoke payload
		NoUpdate: true, // no PUT /users/{id}; role changes go through `users change-role`
		Extra:    userExtra,
	})
}

func userExtra(parent *cobra.Command, _ resourceSpec[api.User]) {
	// invite --email a@x.com --email b@y.com [--role global:member]
	var emails []string
	var role string
	invite := &cobra.Command{
		Use:   "invite --email <addr> [--email <addr> ...] [--role <role>]",
		Short: "Invite one or more users by email",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if len(emails) == 0 {
				return fmt.Errorf("at least one --email is required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			invites := make([]map[string]string, 0, len(emails))
			for _, e := range emails {
				m := map[string]string{"email": e}
				if role != "" {
					m["role"] = role
				}
				invites = append(invites, m)
			}
			res, err := client.InviteUsers(context.Background(), invites)
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
	invite.Flags().StringArrayVar(&emails, "email", nil, "email to invite (repeatable)")
	invite.Flags().StringVar(&role, "role", "", "global role: global:admin|global:member (default member)")
	parent.AddCommand(invite)

	// change-role <id> --role global:admin
	var newRole string
	changeRole := &cobra.Command{
		Use:   "change-role <id> --role <role>",
		Short: "Change a user's global role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if newRole == "" {
				return fmt.Errorf("--role is required (global:admin or global:member)")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := client.ChangeUserRole(context.Background(), args[0], newRole); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "changed role of user %s to %s\n", args[0], newRole)
			}
			return nil
		},
	}
	changeRole.Flags().StringVar(&newRole, "role", "", "new global role (required)")
	parent.AddCommand(changeRole)
}
