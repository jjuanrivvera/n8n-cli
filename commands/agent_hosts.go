package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// guardMCPServer is the default MCP server name Claude Code registers for this
// CLI (`n8nctl mcp claude enable` derives it from the executable name). The
// exact permission rules use it; the hook's MCP branch matches on the tool
// basename instead, so a server registered under a different name is still
// covered by the hook.
const guardMCPServer = "n8nctl"

// --- Claude Code ---

func emitClaudeCode(cmd *cobra.Command, g guardPlan, write bool) error {
	blocked := g.blocked()
	asked := g.asked()

	// Exact, non-regex deny rules — one entry per blocked command spelling
	// (canonical path plus every cobra alias combination, so "n8nctl wf delete"
	// is covered too). Claude matches permission rules as literal prefix
	// patterns, so "Bash(n8nctl workflows delete:*)" matches exactly that
	// subcommand prefix and cannot false-match an argument that happens to
	// contain "delete".
	deny := []string{}
	for _, gc := range blocked {
		for _, sp := range gc.spellings {
			deny = append(deny, fmt.Sprintf("Bash(%s %s:*)", guardBinary, sp))
		}
	}
	// Hard-block the raw-api escape for destructive HTTP methods. The method is
	// the first positional argument ("n8nctl api DELETE /path"), so these exact
	// prefix patterns match the real syntax and cannot false-match a GET whose
	// PATH contains "delete".
	for _, m := range guardAPIMethods {
		deny = append(deny, fmt.Sprintf("Bash(%s api %s:*)", guardBinary, m))
	}
	// Exact MCP tool names — no regex, no glob.
	for _, gc := range blocked {
		deny = append(deny, "mcp__"+guardMCPServer+"__"+gc.tool)
	}

	ask := []string{}
	for _, gc := range asked {
		for _, sp := range gc.spellings {
			ask = append(ask, fmt.Sprintf("Bash(%s %s:*)", guardBinary, sp))
		}
	}
	for _, gc := range asked {
		ask = append(ask, "mcp__"+guardMCPServer+"__"+gc.tool)
	}

	hookPath := "${CLAUDE_PROJECT_DIR}/.claude/hooks/n8nctl-guard.sh"
	hookEntry := func(matcher string) map[string]any {
		return map[string]any{
			"matcher": matcher,
			"hooks":   []any{map[string]any{"type": "command", "command": hookPath}},
		}
	}
	settings := map[string]any{
		"permissions": map[string]any{"deny": deny, "ask": ask},
		"hooks": map[string]any{
			// Hook matchers ARE regexes (unlike permission rules), so the MCP
			// matcher covers any server name that carries the n8n tool prefix.
			"PreToolUse": []any{hookEntry("Bash"), hookEntry("mcp__.*" + guardMCPPrefix + ".*")},
		},
	}
	settingsJSON, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "# Claude Code agent-safety config. The MCP-tool branch of the hook is a hard")
	fmt.Fprintln(out, "# block (structured tool names can't be obfuscated); the Bash branch is")
	fmt.Fprintln(out, "# best-effort — it defeats quote/backslash tricks but not variable")
	fmt.Fprintln(out, "# indirection or shell aliases, so MCP-only operation (or a read-only")
	fmt.Fprintln(out, "# sandbox) is the strongest guarantee. The permission rules are")
	fmt.Fprintln(out, "# belt-and-suspenders. The 'n8nctl api' escape hatch is partially covered")
	fmt.Fprintln(out, "# (DELETE/PUT/POST/PATCH method patterns) but its MCP tool (n8n_api) cannot")
	fmt.Fprintln(out, "# be classified by verb — treat it with the same care. Regenerate after")
	fmt.Fprintln(out, "# upgrading n8nctl.")
	fmt.Fprintln(out)
	if err := writeOrPrint(cmd, write, claudeHookPath(), claudeHookScript(blocked), 0o755); err != nil { //nolint:gosec // executable hook
		return err
	}
	fmt.Fprintln(out)
	return writeOrPrint(cmd, write, claudeSettingsPath(), string(settingsJSON)+"\n", 0o644)
}

func claudeHookPath() string     { return filepath.Join(".claude", "hooks", "n8nctl-guard.sh") }
func claudeSettingsPath() string { return filepath.Join(".claude", "settings.json") }

// guardAPIMethods are the write-capable HTTP methods of the `n8nctl api`
// escape hatch, blocked at the method position on the Bash surface.
var guardAPIMethods = []string{"DELETE", "PUT", "POST", "PATCH"}

// claudeHookScript generates the PreToolUse hook script. It uses exact
// command-path anchored matching rather than bare-verb grep so that arguments
// containing a blocked word do not cause false positives
// (e.g. `n8nctl workflows create --set name=delete-old` is allowed; only
// "n8nctl workflows delete" at the command position is blocked).
func claudeHookScript(blocked []guardCmd) string {
	// Ordered, deduplicated cli-path list (every alias spelling) for the Bash
	// branch and exact tool-basename list for the MCP branch.
	seenCLI := map[string]bool{}
	seenTool := map[string]bool{}
	var cliPaths []string
	var toolNames []string
	for _, gc := range blocked {
		for _, sp := range gc.spellings {
			if !seenCLI[sp] {
				seenCLI[sp] = true
				cliPaths = append(cliPaths, sp)
			}
		}
		if !seenTool[gc.tool] {
			seenTool[gc.tool] = true
			toolNames = append(toolNames, gc.tool)
		}
	}

	var cliArray strings.Builder
	cliArray.WriteString("blocked_cmds=(\n")
	for _, p := range cliPaths {
		safe := strings.ReplaceAll(p, "'", "'\\''")
		cliArray.WriteString("  '" + safe + "'\n")
	}
	cliArray.WriteString(")")

	var toolArray strings.Builder
	toolArray.WriteString("blocked_tools=(\n")
	for _, t := range toolNames {
		safe := strings.ReplaceAll(t, "'", "'\\''")
		toolArray.WriteString("  '" + safe + "'\n")
	}
	toolArray.WriteString(")")

	return `#!/usr/bin/env bash
# n8nctl agent guard — blocks irreversible n8n operations on the Bash and MCP
# surfaces. Generated by ` + "`n8nctl agent guard`" + `; regenerate after
# upgrading n8nctl so new actions are covered.
#
# MATCHING STRATEGY: commands are matched by exact SUBCOMMAND PATH at the
# command position, not by bare verbs anywhere in the line. This prevents:
#   - false positives: "n8nctl workflows create --set name=delete-old" is NOT
#     blocked because "workflows create" is not in the blocked set.
#   - false negatives: a blocked path is still caught after ;, |, &&, env
#     prefixes, or newline continuations.
#
# For each blocked cli-path P, we check whether the CLEANED command string
# matches the anchored ERE:
#   (^|[;&|([:space:]]+)([^[:space:]]*/)?n8nctl[[:space:]]+<P>([[:space:];&|)]|$)
# which ensures n8nctl+P appears at a command position. The optional
# ([^[:space:]]*/)? prefix also catches path-invoked binaries like
# ./bin/n8nctl or /usr/local/bin/n8nctl, while a different binary that merely
# ends in "n8nctl" (e.g. myn8nctl) is NOT blocked. The blocked set includes
# every built-in cobra alias spelling (wf delete, exec prune, dt delete-rows,
# ...), so alias spellings cannot bypass the match.
#
# The MCP branch is an exact set-membership check on the tool basename (the
# part after the final "__"), so it covers any MCP server name and cannot
# false-match a near-miss tool like n8n_workflows_delete2.
#
# The "n8nctl api" escape hatch is caught by method-position matching:
#   n8nctl api (DELETE|PUT|POST|PATCH) at word boundary; a GET whose PATH
#   contains "delete" is NOT blocked.
#
# De-obfuscation: quotes (\042 / \047) and backslash (\134) are stripped from
# the raw command string before pattern matching to defeat trivial tricks like
#   n8nctl workflows de""lete 42
# Variable indirection (a=delete; n8nctl workflows $a 42), shell aliases, and
# user-defined ` + "`n8nctl alias`" + ` expansions are NOT defeated — use MCP-only mode
# or a read-only sandbox for a hard guarantee.

# --- blocked command paths (Bash surface) ---
` + cliArray.String() + `

# --- blocked MCP tool basenames (exact) ---
` + toolArray.String() + `

input=$(cat)

# deny_raw emits a denial with a FIXED reason via printf (no jq needed).
deny_raw() {
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"%s"}}\n' "$1"
  exit 0
}

# deobfuscate strips quote chars (\042=" \047=' \134=\) and collapses newlines
# to a single line so obfuscation tricks can't split a token across lines.
deobfuscate() {
  printf '%s' "$1" | tr -d '\042\047\134' | tr '\n' ' '
}

# bash_is_blocked returns 0 (true) if the cleaned command contains a blocked
# n8nctl subcommand at the command position.
bash_is_blocked() {
  local cleaned="$1"
  local p
  for p in "${blocked_cmds[@]}"; do
    local pat
    pat=$(printf '%s' "$p" | sed 's/ /[[:space:]]+/g')
    if printf '%s' "$cleaned" | grep -qiE "(^|[;&|([:space:]]+)([^[:space:]]*/)?` + guardBinary + `[[:space:]]+${pat}([[:space:];&|)]|\$)"; then
      return 0
    fi
  done
  return 1
}

# api_is_blocked returns 0 (true) if the command is a destructive raw-api call.
# Matches n8nctl api (DELETE|PUT|POST|PATCH) at the method position only.
api_is_blocked() {
  local cleaned="$1"
  printf '%s' "$cleaned" | grep -qiE "(^|[;&|([:space:]]+)([^[:space:]]*/)?` + guardBinary + `[[:space:]]+api[[:space:]]+(DELETE|PUT|POST|PATCH)([[:space:];&|)/]|\$)"
}

# Without jq we cannot isolate the tool_name/command fields. Fail safe: apply
# the same de-obfuscation and anchored path matching on the raw payload rather
# than a loose binary+verb scan that would flag any Bash line mentioning them
# (e.g. "cat workflows_delete.go"). JSON punctuation (:,{}[]) is translated to
# spaces first — otherwise "command":"n8nctl …" would hide the binary behind a
# colon and the anchored command-position match would fail open.
if ! command -v jq >/dev/null 2>&1; then
  flat=$(printf '%s' "$input" | tr '\n:,{}[]' '       ')
  cleaned=$(deobfuscate "$flat")
  if bash_is_blocked "$cleaned"; then
    deny_raw "n8nctl agent guard: irreversible operation blocked (jq unavailable; raw match)."
  fi
  if api_is_blocked "$cleaned"; then
    deny_raw "n8nctl agent guard: destructive raw-api call blocked (jq unavailable; raw match)."
  fi
  # Blocked MCP tool names are structured strings; a raw fixed-string scan is
  # conservative (an unrelated payload mentioning one is denied) but fail-safe.
  for t in "${blocked_tools[@]}"; do
    if printf '%s' "$flat" | grep -qF "$t"; then
      deny_raw "n8nctl agent guard: irreversible MCP tool blocked (jq unavailable; raw match)."
    fi
  done
  exit 0
fi

# deny emits the denial with a jq-escaped reason so an interpolated value can
# never break out of the JSON string.
deny() {
  jq -c -n --arg r "$1" \
    '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$r}}'
  exit 0
}

tool=$(printf '%s' "$input" | jq -r '.tool_name // empty')
case "$tool" in
  Bash)
    raw_cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
    cleaned=$(deobfuscate "$raw_cmd")
    if bash_is_blocked "$cleaned"; then
      deny "n8nctl agent guard: irreversible operation blocked."
    fi
    if api_is_blocked "$cleaned"; then
      deny "n8nctl agent guard: destructive raw-api call blocked. Use dedicated commands or MCP-only mode."
    fi
    ;;
  mcp__*)
    # Exact set-membership on the tool basename — no substring or regex match.
    base="${tool##*__}"
    for t in "${blocked_tools[@]}"; do
      if [ "$base" = "$t" ]; then
        deny "n8nctl agent guard: irreversible MCP tool blocked (${tool})."
      fi
    done
    ;;
esac
exit 0
`
}

// --- Codex ---

func emitCodex(cmd *cobra.Command, g guardPlan, write bool) error {
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "# Codex agent-safety config (~/.codex/config.toml).")
	fmt.Fprintln(out, "# Codex has no per-command deny hook: read-only sandbox is the hard block,")
	fmt.Fprintln(out, "# and destructive MCP tools (which n8nctl annotates) already require approval.")
	fmt.Fprintln(out)
	mode := "read-only"
	policy := "untrusted"
	note := "# read-only sandbox: no n8nctl write can run without an explicit approval."
	if !g.allWrites {
		note = "# read-only sandbox + untrusted: writes pause for approval; reads run free.\n" +
			"# Switch sandbox_mode to \"workspace-write\" only if you accept unattended writes."
	}
	content := fmt.Sprintf("sandbox_mode    = %q\napproval_policy = %q\n\n%s\n", mode, policy, note)
	return writeOrPrint(cmd, write, codexConfigPath(write), content, 0o644)
}

func codexConfigPath(write bool) string {
	if write {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".codex", "config.toml")
		}
	}
	return "~/.codex/config.toml"
}

// --- OpenCode ---

func emitOpenCode(cmd *cobra.Command, g guardPlan, write bool) error {
	blocked := g.blocked()
	asked := g.asked()

	// Exact per-command rules — no verb wildcards. "n8nctl * delete*" style
	// patterns cannot match top-level commands (nothing fills the middle "*")
	// and can shadow unrelated paths; one exact rule per command avoids both.
	bash := map[string]any{"*": "allow"}
	perm := map[string]any{}
	for _, gc := range blocked {
		for _, sp := range gc.spellings {
			bash[guardBinary+" "+sp] = "deny"
		}
		perm[gc.tool] = "deny"
	}
	// Block destructive raw-api methods at the method position.
	for _, m := range guardAPIMethods {
		bash[guardBinary+" api "+m] = "deny"
	}
	for _, gc := range asked {
		for _, sp := range gc.spellings {
			bash[guardBinary+" "+sp] = "ask"
		}
		perm[gc.tool] = "ask"
	}
	perm["bash"] = bash

	cfg := map[string]any{
		"$schema":    "https://opencode.ai/config.json",
		"permission": perm,
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "# OpenCode agent-safety config (opencode.json). `deny` is a hard block;")
	fmt.Fprintln(out, "# `ask` prompts. Exact per-command rules take precedence over the `*`")
	fmt.Fprintln(out, "# catch-all, which allows everything else. Destructive raw-api HTTP methods")
	fmt.Fprintln(out, "# (DELETE/PUT/POST/PATCH) are also blocked on the Bash surface.")
	fmt.Fprintln(out)
	return writeOrPrint(cmd, write, "opencode.json", string(b)+"\n", 0o644)
}
