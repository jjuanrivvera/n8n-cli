package commands

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	aliasCmd := &cobra.Command{
		Use:   "alias",
		Short: "Define command shortcuts expanded before parsing",
		Long: "User aliases expand the first argument before cobra parses it.\n" +
			"They cannot shadow built-in commands.\n\n" +
			"  n8nctl alias set ls 'workflows list --all'\n" +
			"  n8nctl ls",
	}
	aliasCmd.AddCommand(aliasSetCmd(), aliasListCmd(), aliasRemoveCmd())
	rootCmd.AddCommand(aliasCmd)
}

func aliasSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> <expansion...>",
		Short: "Create or update an alias",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if isBuiltinCommand(name) {
				return fmt.Errorf("%q is a built-in command and cannot be aliased", name)
			}
			c, err := loadConfig()
			if err != nil {
				return err
			}
			if c.Aliases == nil {
				c.Aliases = map[string]string{}
			}
			c.Aliases[name] = strings.Join(args[1:], " ")
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "alias %q = %q\n", name, c.Aliases[name])
			return nil
		},
	}
}

func aliasListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List defined aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			type row struct {
				Alias     string `json:"alias"`
				Expansion string `json:"expansion"`
			}
			names := make([]string, 0, len(c.Aliases))
			for n := range c.Aliases {
				names = append(names, n)
			}
			sort.Strings(names)
			rows := make([]row, 0, len(names))
			for _, n := range names {
				rows = append(rows, row{Alias: n, Expansion: c.Aliases[n]})
			}
			return render(cmd, rows)
		},
	}
}

func aliasRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove an alias",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			if _, ok := c.Aliases[args[0]]; !ok {
				return fmt.Errorf("no such alias %q", args[0])
			}
			delete(c.Aliases, args[0])
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed alias %q\n", args[0])
			return nil
		},
	}
}

// isBuiltinCommand reports whether name is a top-level command (so aliases never
// shadow built-ins).
func isBuiltinCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name || slices.Contains(c.Aliases, name) {
			return true
		}
	}
	return false
}

// expandAliases rewrites os.Args[1] when it is a user alias (and not a built-in),
// splicing the expansion in place. Runs before cobra parses anything.
func expandAliases() {
	if len(os.Args) < 2 {
		return
	}
	first := os.Args[1]
	if strings.HasPrefix(first, "-") || isBuiltinCommand(first) {
		return
	}
	c, err := loadConfig()
	if err != nil || c.Aliases == nil {
		return
	}
	expansion, ok := c.Aliases[first]
	if !ok {
		return
	}
	fields := strings.Fields(expansion)
	if len(fields) == 0 {
		return
	}
	newArgs := append([]string{os.Args[0]}, fields...)
	newArgs = append(newArgs, os.Args[2:]...)
	os.Args = newArgs
}
