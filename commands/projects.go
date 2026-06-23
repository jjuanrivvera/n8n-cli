package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.Project]{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "Manage projects and their members",
		Long:    "Projects are an n8n Enterprise feature. Create with --set name='My Project'.",
		New:     func(c *api.Client) *api.Resource[api.Project] { return c.Projects() },
		Columns: []string{"id", "name", "type"},
		Extra:   projectExtra,
	})
}

func projectExtra(parent *cobra.Command, _ resourceSpec[api.Project]) {
	// members <projectId>
	parent.AddCommand(&cobra.Command{
		Use:   "members <projectId>",
		Short: "List the members of a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			members, _, err := client.ListProjectMembers(cmd.Context(), args[0], api.ListParams{})
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return render(cmd, members)
		},
	})

	// add-member <projectId> --user <id> --role <role>
	var addUser, addRole string
	addMember := &cobra.Command{
		Use:   "add-member <projectId> --user <userId> --role <role>",
		Short: "Add a user to a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if addUser == "" || addRole == "" {
				return fmt.Errorf("--user and --role are required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			rel := []map[string]string{{"userId": addUser, "role": addRole}}
			if err := client.AddProjectMembers(cmd.Context(), args[0], rel); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "added user %s to project %s as %s\n", addUser, args[0], addRole)
			}
			return nil
		},
	}
	addMember.Flags().StringVar(&addUser, "user", "", "user id to add (required)")
	addMember.Flags().StringVar(&addRole, "role", "", "project role, e.g. project:viewer|project:editor|project:admin (required)")
	parent.AddCommand(addMember)

	// set-member-role <projectId> <userId> --role <role>
	var newRole string
	setRole := &cobra.Command{
		Use:   "set-member-role <projectId> <userId> --role <role>",
		Short: "Change a project member's role",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if newRole == "" {
				return fmt.Errorf("--role is required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := client.ChangeProjectMemberRole(cmd.Context(), args[0], args[1], newRole); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "updated role of user %s in project %s to %s\n", args[1], args[0], newRole)
			}
			return nil
		},
	}
	setRole.Flags().StringVar(&newRole, "role", "", "new project role (required)")
	parent.AddCommand(setRole)

	// remove-member <projectId> <userId>
	parent.AddCommand(&cobra.Command{
		Use:   "remove-member <projectId> <userId>",
		Short: "Remove a user from a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			if err := client.RemoveProjectMember(cmd.Context(), args[0], args[1]); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if !flagQuiet {
				fmt.Fprintf(cmd.OutOrStdout(), "removed user %s from project %s\n", args[1], args[0])
			}
			return nil
		},
	})
}
