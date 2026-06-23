// Command n8nctl is a portable, single-binary client for the n8n public REST API.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jjuanrivvera/n8n-cli/commands"
	"github.com/jjuanrivvera/n8n-cli/internal/api"
	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

func main() {
	// Cancel in-flight work (pagination, retry backoff, multi-step loops) on the
	// first Ctrl-C / SIGTERM; a second signal force-kills via the default handler.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := commands.RootCmd()
	root.Version = version.Short()
	root.SetVersionTemplate("n8nctl {{.Version}}\n")

	if err := commands.Execute(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		// Distinct exit code for auth failures (2), matching common CLI convention.
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.IsUnauthorized() {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
