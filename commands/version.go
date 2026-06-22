package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

func init() {
	var jsonOut, check bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if jsonOut {
				info := map[string]string{
					"version":   version.Version,
					"commit":    version.Commit,
					"buildDate": version.BuildDate,
					"go":        runtime.Version(),
					"platform":  runtime.GOOS + "/" + runtime.GOARCH,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(info); err != nil {
					return err
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), version.Info())
			}
			if check {
				return checkLatest(cmd)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "print as JSON")
	cmd.Flags().BoolVar(&check, "check", false, "check for a newer release on GitHub")
	rootCmd.AddCommand(cmd)
}

// releaseURL is the GitHub "latest release" endpoint; overridable in tests.
var releaseURL = "https://api.github.com/repos/jjuanrivvera/n8n-cli/releases/latest"

// checkLatest compares the current version against the latest GitHub release.
func checkLatest(cmd *cobra.Command) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("checking latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(cmd.ErrOrStderr(), "could not determine latest release (HTTP %d)\n", resp.StatusCode)
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return err
	}
	latest := rel.TagName
	switch {
	case latest == "":
		fmt.Fprintln(cmd.ErrOrStderr(), "no releases found")
	case "v"+version.Version == latest || version.Version == latest:
		fmt.Fprintln(cmd.OutOrStdout(), "you are on the latest version")
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "a newer version is available: %s (you have %s)\n", latest, version.Version)
	}
	return nil
}
