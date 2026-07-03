---
title: n8nctl agent guard
---

## n8nctl agent guard

Generate agent-safety config that blocks destructive n8n operations

### Synopsis

guard generates the permission rules and hooks that stop an AI agent from
running destructive n8n operations, derived from the live command tree (and the
MCP tool annotations) so the list is always complete and stays correct across
upgrades.

By default it hard-blocks deletion and makes ordinary writes (create, update,
activate, transfer, retry, …) require approval; read operations stay allowed.
Pass --all-writes to block writes too.

Because the MCP server uses whatever profile is active at startup (the --profile
flag is not exposed to the model), an agent cannot switch instances on its own.

IMPORTANT: the "n8nctl api" escape hatch can issue any HTTP verb. The guard
blocks "n8nctl api DELETE/PUT/POST/PATCH" method patterns on the Bash surface
but cannot enumerate arbitrary path arguments, and the n8n_api MCP tool cannot
be classified by verb. For a hard guarantee, run the agent MCP-only (no Bash
tool) or inside a read-only sandbox.

Output is printed for review by default; pass --write to install it.

```
n8nctl agent guard [flags]
```

### Examples

```
  n8nctl agent guard --host claude-code
  n8nctl agent guard --host codex
  n8nctl agent guard --host opencode --all-writes
  n8nctl agent guard --host claude-code --write
```

### Options

```
      --all-writes    Also block create/update/activate/… (default: those require approval)
  -h, --help          help for guard
      --host string   Target agent host: claude-code, codex, opencode
      --write         Write the config/hook files instead of printing them
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --instance string   n8n instance to use: a named profile [env: N8NCTL_INSTANCE, N8NCTL_PROFILE]
      --jq string         apply a jq program to the result (e.g. '.[].id'); implies JSON input
      --no-color          disable colored output [env: NO_COLOR]
      --no-header         hide the table header row
  -o, --output string     output format: table|json|yaml|csv|id [env: N8NCTL_OUTPUT]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl agent](n8nctl_agent.md)	 - Helpers for running n8nctl under an AI agent

