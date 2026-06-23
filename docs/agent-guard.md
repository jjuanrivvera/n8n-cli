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

- **Hard-block** the irreversible operations: `delete` and `delete-rows`. The
  agent cannot run these at all.
- **Require approval** for ordinary writes: `create`, `update`, `activate`,
  `deactivate`, `archive`, `transfer`, `restore`, `sync`, `apply`, `retry`,
  `stop`, member changes, and the rest. These pause for a human to confirm.
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

The hook (excerpt):

```bash
verbs='(delete|delete-rows)'
...
tool=$(printf '%s' "$input" | jq -r '.tool_name // empty')
case "$tool" in
  Bash)
    cmd=$(printf '%s' "$input" | jq -r '.tool_input.command // empty')
    # Strip quotes/backslashes to defeat trivial obfuscation, flatten newlines.
    stripped=$(printf '%s' "$cmd" | tr -d '\042\047\134')
    if printf '%s\n%s' "$cmd" "$stripped" | tr '\n' ' ' | grep -qiE "\bn8nctl\b.*\b${verbs}\b"; then
      deny "n8nctl agent guard: irreversible operation blocked (${verbs})."
    fi
    ;;
  mcp__*n8n*)
    if printf '%s' "$tool" | grep -qiE "_${verbs}$"; then
      deny "n8nctl agent guard: irreversible MCP tool blocked (${tool})."
    fi
    ;;
esac
```

The permission rules in `.claude/settings.json` are belt-and-suspenders: `deny`
for the destructive operations, `ask` for the writes, across both the Bash and
MCP surfaces:

```json
{
  "permissions": {
    "ask": [
      "Bash(n8nctl * create:*)",
      "Bash(n8nctl * update:*)",
      "Bash(n8nctl * activate:*)",
      "mcp__.*n8n.*_(activate|create|update|…)"
    ],
    "deny": [
      "Bash(n8nctl * delete:*)",
      "Bash(n8nctl * delete-rows:*)",
      "mcp__.*n8n.*_(delete|delete-rows)"
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
      "n8nctl * create*": "ask",
      "n8nctl * update*": "ask",
      "n8nctl * delete*": "deny",
      "n8nctl * delete-rows*": "deny"
    },
    "n8n_*_create": "ask",
    "n8n_*_update": "ask",
    "n8n_*_delete": "deny",
    "n8n_*_delete-rows": "deny"
  }
}
```

## MCP-only is the strongest guarantee

The guard fences two surfaces, and they are not equally strong:

- **The MCP-tool branch is a hard block.** MCP tool names are structured and
  cannot be obfuscated, so the guard matches them exactly and denies the
  destructive ones outright.
- **The Bash branch is best-effort.** It defeats quote/backslash obfuscation and
  flattens newlines so a split verb cannot slip past the match, but it cannot
  defeat variable indirection or shell aliases.

The strongest configuration is therefore to run the agent **MCP-only** (no Bash
access to `n8nctl`) — or in a read-only sandbox — combined with the guard. That
way the only operations available are the [MCP tools](mcp.md), with the
destructive ones hard-blocked. Because `agent guard` is itself excluded from the
MCP surface, an agent cannot disable its own rails.

## See also

- [MCP server](mcp.md) — the tool surface the guard fences, its naming, and the
  read-only/write/destructive annotations the guard derives its rules from.
