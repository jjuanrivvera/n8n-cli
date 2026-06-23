package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/auth"
)

func init() {
	var profileName, baseURL, apiKey string
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"setup"},
		Short:   "Interactive first-run setup for an instance/profile",
		Long: "Walks you through naming a profile, setting its base URL, capturing an API key\n" +
			"(stored in your OS keyring), verifying connectivity, and writing the config.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := loadConfig()
			if err != nil {
				return err
			}

			if profileName == "" {
				profileName = prompt(cmd, "Profile name [default]: ")
			}
			if profileName == "" {
				profileName = "default"
			}

			p := c.Profile(profileName)
			if baseURL == "" {
				def := p.BaseURL
				label := "Instance base URL (e.g. https://n8n.example.com): "
				if def != "" {
					label = fmt.Sprintf("Instance base URL [%s]: ", def)
				}
				baseURL = prompt(cmd, label)
				if baseURL == "" {
					baseURL = def
				}
			}
			if baseURL == "" {
				return fmt.Errorf("a base URL is required")
			}
			p.BaseURL = baseURL

			if apiKey == "" {
				apiKey, err = promptSecret(cmd, "n8n API key (Settings > n8n API): ")
				if err != nil {
					return err
				}
			}
			apiKey = strings.TrimSpace(apiKey)
			if apiKey == "" {
				return fmt.Errorf("an API key is required")
			}

			client := api.New(
				api.WithBaseURL(p.BaseURL),
				api.WithAPIKey(apiKey),
				api.WithLogger(newLogger("warn")),
			)
			fmt.Fprintf(cmd.ErrOrStderr(), "verifying against %s ...\n", client.BaseURL())
			if _, _, err := client.Workflows().List(cmd.Context(), api.ListParams{Limit: 1}); err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			if err := auth.Set(profileName, apiKey); err != nil {
				return fmt.Errorf("storing API key in keyring: %w", err)
			}
			c.SetProfile(p)
			if c.DefaultProfile == "" {
				c.DefaultProfile = profileName
			}
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"✓ profile %q ready (%s). Try: n8nctl workflows list\n", profileName, client.BaseURL())
			return nil
		},
	}
	cmd.Flags().StringVar(&profileName, "profile", "", "profile name to create/update")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "instance base URL")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (otherwise prompted without echo)")
	rootCmd.AddCommand(cmd)
}
