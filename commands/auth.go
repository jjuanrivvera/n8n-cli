package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/auth"
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate against an n8n instance",
		Long:  "Capture, verify, and remove the API key for the active profile (stored in your OS keyring).",
	}
	authCmd.AddCommand(authLoginCmd(), authLogoutCmd(), authStatusCmd())
	rootCmd.AddCommand(authCmd)
}

func authLoginCmd() *cobra.Command {
	var apiKey, baseURL string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store and verify an API key for the active profile",
		Long: "Stores the API key in your OS keyring and verifies it against the instance.\n" +
			"Get a key from n8n > Settings > n8n API.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, c, err := activeProfile()
			if err != nil {
				return err
			}
			p := c.Profile(profile)
			if baseURL != "" {
				p.BaseURL = baseURL
			}
			if flagBaseURL != "" {
				p.BaseURL = flagBaseURL
			}
			if p.BaseURL == "" {
				p.BaseURL = prompt(cmd, "Instance base URL (e.g. https://n8n.example.com): ")
			}
			if p.BaseURL == "" {
				return fmt.Errorf("a base URL is required")
			}

			if apiKey == "" {
				apiKey = flagAPIKey
			}
			if apiKey == "" {
				apiKey, err = promptSecret(cmd, "n8n API key: ")
				if err != nil {
					return err
				}
			}
			apiKey = strings.TrimSpace(apiKey)
			if apiKey == "" {
				return fmt.Errorf("an API key is required")
			}

			// Verify before persisting.
			client := api.New(
				api.WithBaseURL(p.BaseURL),
				api.WithAPIKey(apiKey),
				api.WithLogger(newLogger("warn")),
			)
			if _, _, err := client.Workflows().List(cmd.Context(), api.ListParams{Limit: 1}); err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			if err := auth.Set(profile, apiKey); err != nil {
				return fmt.Errorf("storing API key in keyring: %w", err)
			}
			if c.DefaultProfile == "" {
				c.DefaultProfile = profile
			}
			c.SetProfile(p)
			if err := c.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ authenticated profile %q against %s\n", profile, client.BaseURL())
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (otherwise prompted without echo)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "instance base URL to store for this profile")
	return cmd
}

func authLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored API key for the active profile",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, _, err := activeProfile()
			if err != nil {
				return err
			}
			if err := auth.Delete(profile); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✓ removed API key for profile %q\n", profile)
			return nil
		},
	}
}

func authStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the active profile and whether its API key works",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, c, err := activeProfile()
			if err != nil {
				return err
			}
			resolved := c.Resolve(profile)
			keyPresent := flagAPIKey != "" || resolved.APIKey != "" || auth.Lookup(profile) != ""

			status := map[string]any{
				"profile":     profile,
				"base_url":    resolved.BaseURL,
				"key_present": keyPresent,
				"valid":       false,
			}
			if resolved.BaseURL != "" && keyPresent {
				client, cerr := getAPIClient(cmd)
				if cerr == nil {
					if _, _, lerr := client.Workflows().List(cmd.Context(), api.ListParams{Limit: 1}); lerr == nil {
						status["valid"] = true
					} else {
						status["error"] = lerr.Error()
					}
				}
			}
			return render(cmd, status)
		},
	}
}

func prompt(cmd *cobra.Command, label string) string {
	fmt.Fprint(cmd.ErrOrStderr(), label)
	line, _ := stdinReader().ReadString('\n')
	return strings.TrimSpace(line)
}

func promptSecret(cmd *cobra.Command, label string) (string, error) {
	fmt.Fprint(cmd.ErrOrStderr(), label)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Non-interactive (piped): read a single line from the shared reader.
		line, _ := stdinReader().ReadString('\n')
		return strings.TrimSpace(line), nil
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(cmd.ErrOrStderr())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
