package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/update"
	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update n8nctl to the latest GitHub release",
		Long: `Download the latest release from GitHub, verify it against checksums.txt,
and atomically replace the running binary.

A dev build (installed via "go install" or built from source) is never
self-updated; use your package manager or rebuild instead.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			u := update.NewUpdater(version.Version)
			res := u.CheckAndUpdate(ctx)
			if res.Error != nil {
				return res.Error
			}
			if res.Updated {
				fmt.Fprintf(cmd.OutOrStdout(), "Updated %s → %s. Restart to use the new version.\n", res.FromVersion, res.ToVersion)
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "Already on the latest version.")
			}
			return nil
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check for a newer release without installing it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			u := update.NewUpdater(version.Version)
			rel, err := u.GetLatestRelease(ctx)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Current: %s\n", version.Version)
			fmt.Fprintf(out, "Latest:  %s\n", rel.TagName)
			switch {
			case version.Version == "dev" || version.Version == "":
				fmt.Fprintln(out, "Development build; not comparing against releases.")
			case rel.TagName == "":
				fmt.Fprintln(out, "No releases found.")
			case strings.TrimPrefix(rel.TagName, "v") == strings.TrimPrefix(version.Version, "v"):
				fmt.Fprintln(out, "You are on the latest version.")
			default:
				fmt.Fprintln(out, "A newer version is available. Run `n8nctl update` to install it.")
			}
			return nil
		},
	}
	cmd.AddCommand(checkCmd)

	rootCmd.AddCommand(cmd)
}
