// Command n8nctl is a portable, single-binary client for the n8n public REST API.
package main

import (
	"fmt"
	"os"

	"github.com/jjuanrivvera/n8n-cli/commands"
	"github.com/jjuanrivvera/n8n-cli/internal/version"
)

func main() {
	root := commands.RootCmd()
	root.Version = version.Short()
	root.SetVersionTemplate("n8nctl {{.Version}}\n")

	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
