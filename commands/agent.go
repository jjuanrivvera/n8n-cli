package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// guardBinary is the CLI name used in Bash permission patterns; guardMCPPrefix is
// the ophis ToolNamePrefix (see mcp.go) used in MCP tool-name patterns. They
// differ — the binary is `n8nctl` but its MCP tools are prefixed `n8n`.
const (
	guardBinary    = "n8nctl"
	guardMCPPrefix = "n8n"
)

// irreversibleVerbs are the resource actions that cannot be undone. n8n has no
// fiscal operations, so deletion is the only irreversible verb; everything else
// that mutates (create/update/activate/transfer/…) is a reversible write.
// `n8nctl agent guard` hard-blocks irreversible verbs by default and makes
// ordinary writes require approval.
var irreversibleVerbs = map[string]bool{
	"delete": true,
}

// isIrreversibleVerb reports whether a subcommand name is an irreversible action,
// handling compound names like "delete-rows" / "remove-member".
func isIrreversibleVerb(verb string) bool {
	for _, tok := range strings.Split(verb, "-") {
		if irreversibleVerbs[tok] {
			return true
		}
	}
	return false
}

// guardCmd is one API operation the guard config targets.
type guardCmd struct {
	cli  string // CLI path without the root, e.g. "workflows delete"
	tool string // MCP tool name, e.g. "n8n_workflows_delete"
	verb string // last path segment, e.g. "delete"
}

func init() {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Helpers for running n8nctl under an AI agent",
		Long:  "Helpers for running n8nctl under an AI agent (Claude Code, Codex, OpenCode, …).",
	}
	agentCmd.AddCommand(newAgentGuardCmd())
	rootCmd.AddCommand(readOnlyHints(agentCmd))
}

func newAgentGuardCmd() *cobra.Command {
	var host string
	var allWrites bool
	var write bool

	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Generate agent-safety config that blocks destructive n8n operations",
		Long: `guard generates the permission rules and hooks that stop an AI agent from
running destructive n8n operations, derived from the live command tree (and the
MCP tool annotations) so the list is always complete and stays correct across
upgrades.

By default it hard-blocks deletion and makes ordinary writes (create, update,
activate, transfer, retry, …) require approval; read operations stay allowed.
Pass --all-writes to block writes too.

Because the MCP server uses whatever profile is active at startup (the --profile
flag is not exposed to the model), an agent cannot switch instances on its own.

Output is printed for review by default; pass --write to install it.`,
		Args: cobra.NoArgs,
		Example: "  n8nctl agent guard --host claude-code\n" +
			"  n8nctl agent guard --host codex\n" +
			"  n8nctl agent guard --host opencode --all-writes\n" +
			"  n8nctl agent guard --host claude-code --write",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, writes, irreversible := classifyAPICommands(rootCmd)
			g := guardPlan{
				irreversible: irreversible,
				writes:       writes,
				allWrites:    allWrites,
			}
			switch host {
			case "claude-code", "claude":
				return emitClaudeCode(cmd, g, write)
			case "codex":
				return emitCodex(cmd, g, write)
			case "opencode":
				return emitOpenCode(cmd, g, write)
			case "":
				return fmt.Errorf("--host is required (claude-code, codex, or opencode)")
			default:
				return fmt.Errorf("unknown host %q (use claude-code, codex, or opencode)", host)
			}
		},
	}
	cmd.Flags().StringVar(&host, "host", "", "Target agent host: claude-code, codex, opencode")
	cmd.Flags().BoolVar(&allWrites, "all-writes", false, "Also block create/update/activate/… (default: those require approval)")
	cmd.Flags().BoolVar(&write, "write", false, "Write the config/hook files instead of printing them")
	_ = cmd.RegisterFlagCompletionFunc("host", fixedCompletions([]string{"claude-code", "codex", "opencode"}))
	return cmd
}

// guardPlan is the classified set of operations the generators turn into config.
type guardPlan struct {
	irreversible []guardCmd
	writes       []guardCmd
	allWrites    bool // fold writes into the hard-blocked set
}

// blocked returns the operations to hard-block (irreversible, plus writes when
// --all-writes); asked returns the ones that only need approval.
func (g guardPlan) blocked() []guardCmd {
	if g.allWrites {
		return append(append([]guardCmd{}, g.irreversible...), g.writes...)
	}
	return g.irreversible
}

func (g guardPlan) asked() []guardCmd {
	if g.allWrites {
		return nil
	}
	return g.writes
}

// blockedVerbs returns the distinct verbs in the hard-block set (for regexes).
func (g guardPlan) blockedVerbs() []string {
	return distinctVerbs(g.blocked())
}

// classifyAPICommands walks the command tree and buckets the operations that hit
// the n8n API (those carry the openWorldHint annotation, which excludes local
// utility commands like auth/config/agent). Read operations carry readOnlyHint;
// the irreversible verbs are hard-blocked; the rest are writes.
func classifyAPICommands(root *cobra.Command) (read, writes, irreversible []guardCmd) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if sub.Runnable() && !sub.Hidden && sub.Name() != "help" {
				cli := strings.TrimPrefix(sub.CommandPath(), root.Name()+" ")
				gc := guardCmd{
					cli:  cli,
					tool: guardMCPPrefix + "_" + strings.ReplaceAll(cli, " ", "_"),
					verb: sub.Name(),
				}
				switch {
				case sub.Annotations["openWorldHint"] != "true":
					// Local/utility command (not an API operation) — never gated.
				case sub.Annotations["readOnlyHint"] == "true":
					read = append(read, gc)
				case isIrreversibleVerb(sub.Name()):
					irreversible = append(irreversible, gc)
				default:
					writes = append(writes, gc)
				}
			}
			walk(sub)
		}
	}
	walk(root)
	sortGuard(read)
	sortGuard(writes)
	sortGuard(irreversible)
	return read, writes, irreversible
}

func sortGuard(cs []guardCmd) {
	sort.Slice(cs, func(i, j int) bool { return cs[i].tool < cs[j].tool })
}

func distinctVerbs(cs []guardCmd) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range cs {
		if !seen[c.verb] {
			seen[c.verb] = true
			out = append(out, c.verb)
		}
	}
	sort.Strings(out)
	return out
}

// writeOrPrint either writes content to path (creating parent dirs) when write is
// set and the file does not already exist, or prints it to the command's output
// with a header. It never overwrites an existing file.
func writeOrPrint(cmd *cobra.Command, write bool, path, content string, perm os.FileMode) error {
	out := cmd.OutOrStdout()
	if !write {
		fmt.Fprintf(out, "# ----- %s -----\n%s\n", path, content)
		return nil
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(out, "# %s already exists — review and merge manually:\n%s\n", path, content)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec // agent config dir, not secret
		return err
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return err
	}
	fmt.Fprintf(out, "wrote %s\n", path)
	return nil
}
