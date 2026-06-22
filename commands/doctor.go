package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/auth"
	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

func init() {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration, credentials, and connectivity",
		Long:  "Runs a series of checks and exits non-zero if any fail.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, c, err := activeProfile()
			if err != nil {
				return err
			}
			resolved := c.Resolve(profile)

			type check struct {
				Name   string `json:"name"`
				OK     bool   `json:"ok"`
				Detail string `json:"detail"`
			}
			var checks []check
			add := func(name string, ok bool, detail string) {
				checks = append(checks, check{Name: name, OK: ok, Detail: detail})
			}

			// 1. Config file
			if _, statErr := os.Stat(config.DefaultPath()); statErr == nil {
				add("config file", true, config.DefaultPath())
			} else {
				add("config file", false, "not found — run `n8nctl init`")
			}

			// 2. Base URL
			add("base url", resolved.BaseURL != "", resolved.BaseURL)

			// 3. API key resolvable
			keyPresent := flagAPIKey != "" || resolved.APIKey != "" || auth.Lookup(profile) != ""
			add("api key", keyPresent, keySource(profile, resolved.APIKey))

			// 4. Connectivity + auth (live call)
			authOK := false
			if resolved.BaseURL != "" && keyPresent {
				client, cerr := getAPIClient(cmd)
				if cerr != nil {
					add("api auth", false, cerr.Error())
				} else if _, _, lerr := client.Workflows().List(context.Background(), api.ListParams{Limit: 1}); lerr != nil {
					add("api auth", false, lerr.Error())
				} else {
					authOK = true
					add("api auth", true, "verified against "+client.BaseURL())
				}
			} else {
				add("api auth", false, "skipped (missing base url or key)")
			}

			allOK := true
			for _, ch := range checks {
				mark := "✓"
				if !ch.OK {
					mark = "✗"
					allOK = false
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %-12s %s\n", mark, ch.Name, ch.Detail)
			}

			if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
				_ = render(cmd, map[string]any{"profile": profile, "ok": allOK, "checks": checks})
			}
			if !allOK {
				return fmt.Errorf("doctor: one or more checks failed")
			}
			_ = authOK
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "also emit results as JSON")
	rootCmd.AddCommand(cmd)
}

func keySource(profile, envKey string) string {
	switch {
	case flagAPIKey != "":
		return "from --api-key flag"
	case envKey != "":
		return "from N8NCTL_API_KEY env"
	case auth.Lookup(profile) != "":
		return "from OS keyring"
	default:
		return "not found — run `n8nctl auth login`"
	}
}
