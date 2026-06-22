// Command n8nctl is a portable, single-binary client for the n8n public REST API.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jjuanrivvera/n8n-cli/commands"
	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

func main() {
	root := commands.RootCmd()
	root.Version = version.Short()
	root.SetVersionTemplate("n8nctl {{.Version}}\n")

	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		// Distinct exit code for auth failures (2), matching common CLI convention.
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsUnauthorized() {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
