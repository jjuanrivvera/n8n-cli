package commands

import "github.com/njayp/ophis"

// mcpExcludedCommands are local setup and meta commands that are not n8n
// operations, so they stay out of the agent-facing MCP tool surface (an AI agent
// has no business calling `n8nctl auth login`, editing config, or — crucially —
// running `n8nctl agent guard` to disable its own safety rails). mcp, help, and
// completion are already excluded by ophis itself.
//
// Matching is by path substring (ExcludeCmdsContaining) so a whole subtree is
// dropped by its parent name. This can only ever remove a tool, never leak one;
// TestMCPExcludesSetupCommands guards the surface regardless.
var mcpExcludedCommands = []string{"agent", "auth", "config", "alias", "init", "skills", "doctor"}

// init registers `n8nctl mcp`, which exposes the CLI's n8n operations as a Model
// Context Protocol server so AI agents can drive any n8n instance. Each generated
// tool carries read-only/destructive annotations (set in buildResourceCmd) that
// MCP hosts honor to gate writes; `n8nctl agent guard` generates host-level rules
// for stronger enforcement.
func init() {
	rootCmd.AddCommand(ophis.Command(&ophis.Config{
		ToolNamePrefix: "n8n",
		Selectors: []ophis.Selector{
			{
				CmdSelector: ophis.ExcludeCmdsContaining(mcpExcludedCommands...),
				// Never surface secret-bearing or instance-targeting flags to the
				// model: the MCP server uses whatever profile is active at startup.
				InheritedFlagSelector: ophis.ExcludeFlags("show-token", "profile", "api-key", "base-url"),
			},
		},
	}))
}
