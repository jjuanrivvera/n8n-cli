# Agent guard

`n8nctl agent guard` generates host-level safety config so an AI agent driving
`n8nctl` — through Bash or through the [MCP server](mcp.md) — cannot run
destructive n8n operations. It produces permission rules and hooks for the host
you name and either prints them for review or installs them.

```bash
n8nctl agent guard --host claude-code
n8nctl agent guard --host codex
n8nctl agent guard --host opencode --all-writes
n8nctl agent guard --host claude-code --write
```

## What it protects against

Left unguarded, an agent that can call `n8nctl` can delete workflows, drop data
table rows, or remove credentials — all irreversible over the API. The guard
fences those operations at the host level, before the agent's tool call ever
reaches the CLI.

## Default posture vs `--all-writes`

The default posture splits operations into three tiers:

- **Hard-block** the irreversible operations: `delete`, `delete-rows`, and
  `prune`, plus the destructive raw-api methods (`n8nctl api
  DELETE/PUT/POST/PATCH`). The agent cannot run these at all.
- **Require approval** for ordinary writes: `create`, `update`, `activate`,
  `deactivate`, `archive`, `transfer`, `restore`, `sync`, `apply`, `retry`,
  `stop`, `packages import`, member changes, and the rest. These pause for a
  human to confirm.
- **Allow** reads (`list`, `get`, `search`, `lint`, `diff`, `schema`, `members`,
  `audit`, `backup`) to run freely.

Pass `--all-writes` to fold the write tier into the hard-block set: with it,
every state-changing operation is blocked outright and only reads run.

The operation list is **derived from the live command tree and the MCP tool
annotations**, not hardcoded. Regenerate the config after upgrading `n8nctl` and
any new actions are classified and covered automatically.

## Review and install

The guard prints its output for review by default. Pass `--write` to install the
files. Installation **never overwrites an existing file** — if a target already
exists, the guard leaves it untouched so it cannot clobber config you have
already customized.

## Hosts

### Claude Code

For `--host claude-code` the guard emits two files:

- `.claude/hooks/n8nctl-guard.sh` — a `PreToolUse` hook that hard-blocks the
  MCP-tool branch and best-effort-blocks the Bash branch.
- `.claude/settings.json` — deny/ask permission rules plus the hook wiring.

The hook matches blocked operations by **exact subcommand path at the command
position**, not by bare verbs anywhere in the line. That means
`n8nctl workflows create --set name=delete-old` is allowed (the blocked word is
in an argument), while `n8nctl workflows delete 42` is denied — including after
`;`, `|`, `&&`, an `env` prefix, a newline continuation, or when the binary is
invoked by path (`./bin/n8nctl`, `/usr/local/bin/n8nctl`). A different binary
that merely ends in `n8nctl` is not matched. The blocked set enumerates every
built-in alias spelling (`wf delete`, `exec prune`, `dt delete-rows`, …), and
quotes/backslashes are stripped first so `de""lete`-style obfuscation cannot
slip past. The MCP branch is an exact set-membership check on the tool basename,
so it covers any MCP server name and cannot false-match a near-miss tool name.

The hook (excerpt):

```bash
blocked_cmds=(
  'workflows delete'
  'workflow delete'
  'wf delete'
  'executions prune'
  ...
)
blocked_tools=(
  'n8n_workflows_delete'
  ...
)
...
case "$tool" in
  Bash)
    raw_cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
    cleaned=$(deobfuscate "$raw_cmd")
    if bash_is_blocked "$cleaned"; then
      deny "n8nctl agent guard: irreversible operation blocked."
    fi
    if api_is_blocked "$cleaned"; then
      deny "n8nctl agent guard: destructive raw-api call blocked."
    fi
    ;;
  mcp__*)
    base="${tool##*__}"
    for t in "${blocked_tools[@]}"; do
      if [ "$base" = "$t" ]; then
        deny "n8nctl agent guard: irreversible MCP tool blocked (${tool})."
      fi
    done
    ;;
esac
```

The permission rules in `.claude/settings.json` are belt-and-suspenders: exact
per-command `deny` rules for the destructive operations (one per alias
spelling), `ask` for the writes, plus exact MCP tool names:

```json
{
  "permissions": {
    "ask": [
      "Bash(n8nctl workflows create:*)",
      "Bash(n8nctl workflows update:*)",
      "Bash(n8nctl packages import:*)",
      "mcp__n8nctl__n8n_workflows_create"
    ],
    "deny": [
      "Bash(n8nctl workflows delete:*)",
      "Bash(n8nctl wf delete:*)",
      "Bash(n8nctl api DELETE:*)",
      "Bash(n8nctl api POST:*)",
      "mcp__n8nctl__n8n_workflows_delete"
    ]
  }
}
```

### Codex

For `--host codex` the guard emits `~/.codex/config.toml`. Codex has no
per-command deny hook, so the read-only sandbox is the hard block and the
destructive MCP tools (which `n8nctl` annotates) already require approval:

```toml
sandbox_mode    = "read-only"
approval_policy = "untrusted"
```

Under the default posture, writes pause for approval and reads run free; switch
`sandbox_mode` to `workspace-write` only if you accept unattended writes. With
`--all-writes`, the read-only sandbox means no `n8nctl` write can run without an
explicit approval.

### OpenCode

For `--host opencode` the guard emits `opencode.json` permission rules. `deny` is
a hard block and `ask` prompts; the specific `n8nctl` patterns take precedence
over the `*` catch-all, which allows everything else:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "bash": {
      "*": "allow",
      "n8nctl workflows create": "ask",
      "n8nctl workflows update": "ask",
      "n8nctl workflows delete": "deny",
      "n8nctl wf delete": "deny",
      "n8nctl api DELETE": "deny",
      "n8nctl data-tables delete-rows": "deny"
    },
    "n8n_workflows_create": "ask",
    "n8n_workflows_update": "ask",
    "n8n_workflows_delete": "deny",
    "n8n_data-tables_delete-rows": "deny"
  }
}
```

## MCP-only is the strongest guarantee

The guard fences two surfaces, and they are not equally strong:

- **The MCP-tool branch is a hard block.** MCP tool names are structured and
  cannot be obfuscated, so the guard matches them exactly (by tool basename,
  covering any server name) and denies the destructive ones outright.
- **The Bash branch is best-effort.** It defeats quote/backslash obfuscation,
  flattens newlines so a split verb cannot slip past the match, covers
  path-invoked binaries and every built-in alias spelling — but it cannot
  defeat variable indirection (`a=delete; n8nctl workflows $a 42`), shell
  aliases, or user-defined `n8nctl alias` expansions.

The strongest configuration is therefore to run the agent **MCP-only** (no Bash
access to `n8nctl`) — or in a read-only sandbox — combined with the guard. That
way the only operations available are the [MCP tools](mcp.md), with the
destructive ones hard-blocked. Because `agent guard` is itself excluded from the
MCP surface, an agent cannot disable its own rails.

## Known limitations

- **The `n8nctl api` escape hatch.** The guard blocks
  `n8nctl api DELETE/PUT/POST/PATCH` at the method position on the Bash surface
  (a `GET` whose path merely contains "delete" is not blocked), but it cannot
  enumerate arbitrary path arguments, and the `n8n_api` MCP tool cannot be
  classified by verb. Treat raw-api access with the same care as Bash access.
- **Conservative false positives.** De-obfuscation strips quotes before
  matching, so a quoted blocked string at a command position — e.g.
  `rg "n8nctl workflows delete" src/` — is denied. This errs on the safe side;
  unquoted or regex-style search patterns are unaffected.
- **Variable indirection, shell aliases, and `n8nctl alias`.** These rewrite
  the command line outside the hook's view. MCP-only operation or a read-only
  sandbox is the hard guarantee.

## See also

- [MCP server](mcp.md) — the tool surface the guard fences, its naming, and the
  read-only/write/destructive annotations the guard derives its rules from.
