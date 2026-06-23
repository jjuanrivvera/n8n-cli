# MCP server

`n8nctl mcp` turns the CLI into a [Model Context Protocol](https://modelcontextprotocol.io)
server, so an AI agent — Claude Code, Claude Desktop, Cursor, VS Code, and any
other MCP host — can drive your n8n instance directly. Instead of teaching the
agent to shell out to `n8nctl`, the server exposes every safe CLI command as a
typed MCP tool the host can call.

The server is built on [github.com/njayp/ophis](https://github.com/njayp/ophis),
which wraps the official [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk).
It reuses the same client, OS-keyring auth, named profiles, and `--dry-run`
support as the CLI: each tool call replays the corresponding cobra command.

!!! info "Relation to n8n's built-in MCP"
    n8n the platform is already MCP-native at the **workflow layer**: the **MCP
    Server Trigger** node turns a workflow into an MCP server (agents call your
    workflows as tools), and the **MCP Client Tool** node lets a workflow consume
    external MCP tools. That is the *data plane* — automations exposed as tools.

    `n8nctl mcp` is a different, complementary layer: the **control plane**. It
    exposes the n8n **management API** — list / create / activate / delete
    workflows, manage credentials, projects, executions, **across instances** — so
    an agent can *operate and administer* your n8n fleet, not run a single
    workflow. Use n8n's nodes to expose automations to agents; use `n8nctl mcp` to
    let an agent manage the instances themselves.

## What MCP is

MCP is an open protocol that lets an AI host discover and call tools exposed by
an external server. The host sends the model a list of tools (each with a name,
description, and JSON input schema); when the model decides to use one, the host
calls the server and feeds the result back. `n8nctl mcp` is that server for n8n:
it advertises the CLI's commands as tools and executes them against whatever
instance the active profile points at.

## Starting the server

```bash
# stdio transport (what most hosts launch directly)
n8nctl mcp start

# HTTP transport, for hosts or scripts that connect over the network
n8nctl mcp stream --host 127.0.0.1 --port 8080

# Export the tool list to mcp-tools.json for inspection (no server started)
n8nctl mcp tools
```

`mcp start` speaks MCP over stdio and is the form an MCP host spawns. `mcp stream`
serves the same tools over HTTP for hosts that connect to a running endpoint.
`mcp tools` writes the full tool catalog to `mcp-tools.json` so you can review
the names, schemas, and annotations without wiring up a host.

Both transports accept `--log-level debug|info|warn|error` for troubleshooting.

## Wiring it into a host

Each supported host has an installer subcommand that edits the host's config for
you, with matching `enable` / `disable` / `list` verbs:

```bash
# Claude Desktop
n8nctl mcp claude enable
n8nctl mcp claude list
n8nctl mcp claude disable

# Cursor
n8nctl mcp cursor enable

# VS Code
n8nctl mcp vscode enable
```

`enable` adds an MCP server entry that launches `n8nctl mcp start`; `list` shows
the host's current MCP servers; `disable` removes the entry. `enable` accepts
`--server-name <name>` to override the entry name, `--config-path <path>` to
target a non-default config file, and `--env KEY=value` (repeatable) to inject
environment variables into the spawned server.

### Manual configuration

If your host is not one of the three installers, point it at `n8nctl mcp start`
yourself. The config shape MCP hosts use looks like this:

```json
{
  "mcpServers": {
    "n8n": {
      "command": "n8nctl",
      "args": ["mcp", "start"]
    }
  }
}
```

To pin the server to a specific instance without changing your default profile,
set the profile through the environment in the host config:

```json
{
  "mcpServers": {
    "n8n-prod": {
      "command": "n8nctl",
      "args": ["mcp", "start"],
      "env": { "N8NCTL_PROFILE": "prod" }
    }
  }
}
```

For Claude Code, register the server with the standard MCP wiring (the same
`n8nctl mcp start` command) and the tools appear under the `n8n` prefix.

## Tool naming and annotations

The server auto-exposes the CLI's command tree as **85 MCP tools**, named with an
`n8n` prefix that mirrors the command path. A few examples:

| CLI command | MCP tool |
|---|---|
| `n8nctl workflows list` | `n8n_workflows_list` |
| `n8nctl workflows create` | `n8n_workflows_create` |
| `n8nctl workflows delete` | `n8n_workflows_delete` |
| `n8nctl executions retry` | `n8n_executions_retry` |
| `n8nctl credentials schema` | `n8n_credentials_schema` |
| `n8nctl data-tables delete-rows` | `n8n_data-tables_delete-rows` |

Every tool carries MCP annotations so a host can reason about its risk before
calling it:

- **Read-only** (`readOnlyHint`) — `list`, `get`, `search`, `lint`, `diff`,
  `schema`, `members`, `backup`, `audit`, and the other inspection tools. These
  never modify the instance.
- **Write** — `create`, `update`, `restore`, `sync`, `apply`, and similar tools
  that change state.
- **Destructive** (`destructiveHint`) — `delete` and `delete-rows`, the
  irreversible operations.

MCP hosts that honor these annotations gate writes and destructive calls
automatically, prompting for confirmation (or refusing) before they run. The
[agent guard](agent-guard.md) hardens this further: it generates host-level rules
that hard-block the destructive tools regardless of whether the host honors the
hints.

## What is not exposed

The server deliberately omits setup and secret-management surface so an agent
cannot reconfigure the CLI or read credentials out from under you:

- **Excluded commands:** `auth`, `config`, `alias`, `init`, `skills`, `agent`,
  and `doctor` are never surfaced as tools. (`agent guard` being excluded means
  an agent cannot disable its own safety rails — see the
  [agent guard](agent-guard.md) page.)
- **Excluded flags:** the secret-bearing flags (`--api-key`, `--show-token`) and
  the instance-targeting flags (`--profile`, `--base-url`) are never offered to
  the model. The server operates on **whatever profile is active when it starts**.

## Auth and profile model

The MCP server authenticates exactly the way the CLI does: it reads the active
profile's API key from the OS keyring and its base URL from
`~/.n8nctl-cli/config.yaml`. The key is never passed through a tool argument and
is never visible to the model.

Because `--profile` is not exposed as a tool flag, the model **cannot switch
instances**. The instance is fixed at startup. To target a specific instance,
either run `n8nctl config use <name>` before starting the server, or set
`N8NCTL_PROFILE` in the host config (as shown above) so each registered server
maps to one instance. Run separate server entries — `n8n-dev`, `n8n-prod` — if an
agent needs more than one.

## Worked example: list, then create a workflow

A typical agent session reads the instance first, then writes. With the server
wired into the host, the model calls tools rather than shell commands.

The agent lists workflows by calling `n8n_workflows_list` (a read-only tool, so
the host runs it freely):

```json
{
  "tool": "n8n_workflows_list",
  "arguments": { "flags": { "active": "true", "output": "json" } }
}
```

It inspects the returned ids and names, then creates a new workflow by calling
`n8n_workflows_create` (a write tool, which an MCP host gates for approval):

```json
{
  "tool": "n8n_workflows_create",
  "arguments": {
    "flags": { "set": ["name=Lead intake", "nodes=[]", "connections={}", "settings={}"] }
  }
}
```

The host pauses for confirmation before the create runs. The agent can preview
any tool first by passing `"dry-run": true` in its flags, which prints the
equivalent curl and sends no request — the same `--dry-run` safety used from the
CLI.

## Security note

Exposing an instance to an agent is powerful and worth fencing. Run the
[agent guard](agent-guard.md) to generate host-level rules that hard-block the
destructive tools, require approval for ordinary writes, and let reads run free.
The strongest configuration is MCP-only operation (no Bash access to `n8nctl`)
combined with the guard, so the only operations available to the agent are the
annotated tools, with the destructive ones blocked outright.
