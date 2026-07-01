package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/auth"
	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

func init() {
	cfgCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and edit configuration and profiles",
		Long:  "Manage the config file and named instance profiles. Secrets are redacted in `view`.",
	}
	cfgCmd.AddCommand(
		configPathCmd(),
		configViewCmd(),
		configSetCmd(),
		configSetURLCmd(),
		configSetAPIKeyCmd(),
		configUseCmd(),
		configListProfilesCmd(),
	)
	rootCmd.AddCommand(cfgCmd)
}

// configSetURLCmd sets the active profile's base URL (familiar shorthand for
// `config set base_url`, matching the official CLI's `config set-url`).
func configSetURLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-url <url>",
		Short: "Set the active profile's instance URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			p := c.Profile(c.ActiveProfileName(instanceOverride()))
			p.BaseURL = args[0]
			c.SetProfile(p)
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set base_url = %s\n", args[0])
			return nil
		},
	}
}

// configSetAPIKeyCmd stores an API key in the OS keyring for the active profile
// (matching the official CLI's `config set-api-key`; unlike `auth login` it does
// not verify the key against the instance).
func configSetAPIKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-api-key <key>",
		Short: "Store an API key in the keyring for the active profile (no verification)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, _, err := activeProfile()
			if err != nil {
				return err
			}
			if err := auth.Set(profile, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "stored API key for profile %q in the keyring\n", profile)
			return nil
		},
	}
}

func configPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), config.DefaultPath())
			return nil
		},
	}
}

func configViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "view",
		Aliases: []string{"show"},
		Short:   "Show the resolved configuration (secrets redacted)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			profiles := map[string]any{}
			for name, p := range c.Profiles {
				key := "(none)"
				if p.APIKey != "" {
					key = "(in config file — move it to the keyring with `auth login`)"
				} else if auth.Lookup(name) != "" {
					key = "(in OS keyring)"
				}
				profiles[name] = map[string]any{
					"base_url":    p.BaseURL,
					"description": p.Description,
					"api_key":     key,
				}
			}
			view := map[string]any{
				"path":            c.Path(),
				"default_profile": c.ActiveProfileName(instanceOverride()),
				"settings": map[string]any{
					"output_format":       c.Settings.OutputFormat,
					"requests_per_second": c.Settings.RequestsPerSecond,
					"log_level":           c.Settings.LogLevel,
				},
				"profiles": profiles,
				"aliases":  c.Aliases,
			}
			return render(cmd, view)
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Long: "Set a global setting or a field on the active profile.\n\n" +
			"Global keys:  output_format, requests_per_second, log_level\n" +
			"Profile keys: base_url, description",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			c, err := loadConfig()
			if err != nil {
				return err
			}
			switch key {
			case "output_format":
				c.Settings.OutputFormat = value
			case "log_level":
				c.Settings.LogLevel = value
			case "requests_per_second", "rps":
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return fmt.Errorf("invalid number for %s: %w", key, err)
				}
				c.Settings.RequestsPerSecond = f
			case "base_url", "description":
				p := c.Profile(c.ActiveProfileName(instanceOverride()))
				if key == "base_url" {
					p.BaseURL = value
				} else {
					p.Description = value
				}
				c.SetProfile(p)
			default:
				return fmt.Errorf("unknown key %q (see `config set --help`)", key)
			}
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "set %s = %s\n", key, value)
			return nil
		},
	}
}

func configUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Switch the default profile (active instance)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			c.DefaultProfile = args[0]
			c.Profile(args[0]) // ensure it exists
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "now using profile %q\n", args[0])
			return nil
		},
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return profileNames(), cobra.ShellCompDirectiveNoFileComp
		},
	}
}

func configListProfilesCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list-profiles",
		Aliases: []string{"profiles"},
		Short:   "List configured profiles (instances)",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}
			active := c.ActiveProfileName(instanceOverride())
			type row struct {
				Profile     string `json:"profile"`
				Active      bool   `json:"active"`
				BaseURL     string `json:"base_url"`
				HasKey      bool   `json:"has_key"`
				Description string `json:"description,omitempty"`
			}
			rows := make([]row, 0, len(c.Profiles))
			for name, p := range c.Profiles {
				rows = append(rows, row{
					Profile:     name,
					Active:      name == active,
					BaseURL:     p.BaseURL,
					HasKey:      p.APIKey != "" || auth.Lookup(name) != "",
					Description: p.Description,
				})
			}
			if len(rows) == 0 {
				fmt.Fprintln(cmd.ErrOrStderr(), "no profiles yet — run `n8nctl init`")
			}
			return render(cmd, rows)
		},
	}
}

func profileNames() []string {
	c, err := loadConfig()
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(c.Profiles))
	for n := range c.Profiles {
		names = append(names, n)
	}
	return names
}
