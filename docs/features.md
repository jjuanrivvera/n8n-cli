# Features

Every `n8nctl` capability in one place. Commands are grouped by category; each
table lists a command group's subcommands and what they do. For full flags and
arguments on any command, run `n8nctl <command> --help` or browse the
[command reference](commands/index.md).

Most commands accept the [global flags](#global-flags-everywhere) for output
formatting, profiles, dry-run, and rate limiting.

## Resource management

The CLI manages every resource the n8n public REST API exposes. Each group
supports the standard `list`, `get`, `create`, `update`, and `delete` verbs
unless noted, plus resource-specific actions.

### Workflows — `n8nctl workflows`

| Command | Description |
|---|---|
| `list` | List workflows (filters: `--name`, `--active`, `--tags`, `--project`) |
| `get <id>` | Get a single workflow by id |
| `create` | Create a workflow (`--file`, `--data`, or repeatable `--set key=value`) |
| `update <id>` | Update a workflow |
| `delete <id>` | Delete a workflow |
| `activate <id>` | Activate a workflow |
| `deactivate <id>` | Deactivate a workflow |
| `archive <id>` | Archive a workflow |
| `unarchive <id>` | Restore an archived workflow |
| `transfer <id>` | Transfer a workflow to another project |
| `tags <id>` | Get or replace a workflow's tags (`--set <tagId,...>`) |

Workflows also host the [fleet operations](#fleet-operations) and
[workflows-as-code](#workflows-as-code) verbs below.

### Executions — `n8nctl executions`

| Command | Description |
|---|---|
| `list` | List executions (filters: `--status`, `--workflow`, `--project`, `--include-data`) |
| `get <id>` | Get a single execution by id |
| `retry <id>` | Retry a failed execution |
| `stop <id>` | Stop a running execution |
| `delete <id>` | Delete an execution |

### Credentials — `n8nctl credentials`

| Command | Description |
|---|---|
| `list` | List credentials |
| `get <id>` | Get a single credential by id |
| `create` | Create a credential (`--file`, `--data`, or `--set`) |
| `update <id>` | Update a credential |
| `delete <id>` | Delete a credential |
| `schema <type>` | Show the field schema for a credential type |
| `transfer <id>` | Transfer a credential to another project |

Inspect a type's required fields with `schema` before you `create`. See the
[create a credential recipe](recipes.md#create-a-credential-after-inspecting-its-schema).

### Projects — `n8nctl projects`

| Command | Description |
|---|---|
| `list` | List projects |
| `get <id>` | Get a single project by id |
| `create` | Create a project |
| `update <id>` | Update a project |
| `delete <id>` | Delete a project |
| `members <id>` | List the members of a project |
| `add-member <id>` | Add a user to a project |
| `remove-member <id>` | Remove a user from a project |
| `set-member-role <id>` | Change a project member's role |

### Users — `n8nctl users` (instance owner only)

| Command | Description |
|---|---|
| `list` | List users |
| `get <id>` | Get a single user by id |
| `invite` | Invite one or more users by email |
| `change-role <id>` | Change a user's global role |
| `delete <id>` | Delete a user |

### Tags — `n8nctl tags`

| Command | Description |
|---|---|
| `list` | List tags |
| `get <id>` | Get a single tag by id |
| `create` | Create a tag |
| `update <id>` | Update a tag |
| `delete <id>` | Delete a tag |

### Variables — `n8nctl variables`

| Command | Description |
|---|---|
| `list` | List variables |
| `get <id>` | Get a single variable by id |
| `create` | Create a variable |
| `update <id>` | Update a variable |
| `delete <id>` | Delete a variable |

### Data tables — `n8nctl data-tables`

| Command | Description |
|---|---|
| `list` | List data tables |
| `get <id>` | Get a single data table by id |
| `create` | Create a data table |
| `update <id>` | Update a data table |
| `delete <id>` | Delete a data table |
| `rows <id>` | List rows in a data table |
| `add-rows <id>` | Add rows (body: a JSON array of row objects) |
| `update-rows <id>` | Update rows matching a filter (body: `{filter, data}`) |
| `upsert-rows <id>` | Insert or update rows (body: `{filter, data}`) |
| `delete-rows <id>` | Delete rows matching a filter |

### Packages — `n8nctl packages` (beta)

| Command | Description |
|---|---|
| `export` | Export workflows as a `.n8np` package |
| `import` | Import a `.n8np` package into a project |

### Audit and source control

| Command | Description |
|---|---|
| `audit` | Generate a security audit of the instance |
| `source-control pull` | Pull changes from the connected remote repository |

## Fleet operations

Commands that go beyond a single API call to operate across instances or
snapshot an instance to disk. See [Beyond the API](beyond-api.md) for the full
treatment.

| Command | Description |
|---|---|
| `workflows sync <id> --to <profile>` | Promote a workflow to another instance (profile) |
| `workflows search` | Find workflows by node type, credential, webhook path, or name |
| `backup --out <dir>` | Export workflows, tags, and variables to a directory (JSON or YAML) |
| `restore --in <dir>` | Recreate workflows from a backup directory |

`backup` writes one file per workflow plus `tags.json`, `variables.json`, a
credentials inventory (metadata only — secrets are never exported), and a
manifest. `--externalize N` splits long code fields into sibling files for
cleaner diffs.

## Workflows as code

GitOps for workflow definitions: treat a directory of JSON/YAML files as the
desired state. See [Workflows as Code](workflows-as-code.md).

| Command | Description |
|---|---|
| `workflows apply --dir <dir>` | Reconcile a directory of workflow files into the instance (`--prune`, `--activate`, `--dry-run`) |
| `workflows lint` | Lint workflow definitions for common mistakes (`--dir`, `-f`, `--remote`, `--disable-rule`) |
| `workflows diff <id>` | Diff a workflow against another instance (`--to`) or a local file (`--file`) |
| `workflows convert <file...>` | Convert workflow files between JSON and YAML, optionally `--externalize` |

`apply` matches workflows by name, reconciles nodes, connections and settings,
and with `--prune` deletes instance workflows absent from the directory. Always
preview with `--dry-run` first.

## AI agents

Run `n8nctl` as a tool surface for AI coding agents, and constrain what those
agents are allowed to do.

| Command | Description |
|---|---|
| `mcp start` | Start the MCP server (stdio) so an agent can drive n8n directly |
| `mcp stream` | Stream the MCP server over HTTP |
| `mcp tools` | Export the tool list as JSON (no server started) |
| `mcp claude` / `mcp cursor` / `mcp vscode` | Register, enable, disable, and list the server in each host |
| `agent guard --host <host>` | Generate agent-safety config that blocks destructive n8n operations |
| `proxy` | Run a local n8n API proxy that lint-gates workflow writes |
| `skills install` | Install this CLI's AI-agent skill into Claude, Cursor, and other agents |

The MCP server exposes the safe CLI command tree as typed tools, reusing the
same keyring auth and active profile. `agent guard` derives its allow/deny rules
from that same tree, so the safety list stays correct across upgrades. The
`proxy` enforces linting at the HTTP layer, rejecting invalid workflow writes
with `422` no matter which client sends them. See [MCP server](mcp.md),
[Agent guard](agent-guard.md), and [Lint-enforcing proxy](proxy.md).

## Output and scripting

Everything below is available on (nearly) every command through global flags.
See [Output and filtering](output.md).

| Capability | Flag | Description |
|---|---|---|
| Output format | `-o table\|json\|yaml\|csv\|id` | Pick a renderer; `id` prints bare ids for piping |
| Column selection | `--columns a,b,c` | Choose columns for `table`/`csv` |
| Header control | `--no-header` | Drop the table/csv header row |
| jq filtering | `--jq '<program>'` | Apply a [gojq](https://github.com/itchyny/gojq) program to the JSON result |
| Dry run | `--dry-run` | Print the equivalent `curl` and send no request |
| Pagination | `--all` / `--cursor` / `--max-pages` / `--limit` | Walk every page, resume from a cursor, or cap pages |
| Raw escape hatch | `n8nctl api <METHOD> <PATH>` | Call any endpoint directly (`-d`, `--file`, `--query`) |

## Multi-instance and security

Drive many instances from one binary, with secrets in the OS keyring. See
[Multi-instance and profiles](profiles.md) and [Authentication](authentication.md).

| Command | Description |
|---|---|
| `init` | Interactive first-run setup for an instance/profile |
| `auth login` / `logout` / `status` | Store, remove, and verify the active profile's API key |
| `config use <name>` | Switch the default profile (active instance) |
| `config view` | Show the resolved configuration (secrets redacted) |
| `config set` / `set-url` / `set-api-key` / `path` / `list-profiles` | Edit and inspect configuration and profiles |
| `--profile <name>` | Target a specific instance for one command |

API keys live in the OS keyring keyed by profile name; the config file holds
only non-secret fields. Configuration precedence is **flag > env > config file >
default**.

## Resilience

Built-in behaviors that keep scripting against n8n safe and predictable.

- **Idempotent retry with backoff.** Safe (idempotent) requests are retried with
  exponential backoff; non-idempotent writes are not retried.
- **Adaptive rate limiting.** A client-side limiter (`--rps` or config) paces
  requests; the client backs off on `429`.
- **Cursor pagination.** `--all` walks the `{data, nextCursor}` envelope; capped
  by `--max-pages`.
- **Ctrl-C cancellation.** In-flight requests cancel cleanly on interrupt.
- **Typed API errors.** Failures surface as a typed `APIError` with actionable
  hints rather than a raw status dump.
- **Flexible JSON types.** Tolerant parsing of n8n's loose encodings (numeric
  strings, `"true"`/`1` booleans, single-string-or-array fields).

## Meta and utilities

| Command | Description |
|---|---|
| `version` | Print version, commit, and build date (`--check` for a newer release, `--json`) |
| `doctor` | Diagnose configuration, credentials, and connectivity (`--json`) |
| `completion <shell>` | Generate a completion script for bash, zsh, fish, or powershell |
| `alias set` / `list` / `remove` | Define command shortcuts expanded before parsing |

## Global flags (everywhere)

These apply to nearly every command:

| Flag | Description |
|---|---|
| `--profile <name>` | Config profile (instance) to use (`N8NCTL_PROFILE`) |
| `--base-url <url>` | Override the instance base URL |
| `--api-key <key>` | Override the API key (prefer the keyring via `auth login`) |
| `-o, --output <fmt>` | Output format: `table\|json\|yaml\|csv\|id` (`N8NCTL_OUTPUT`) |
| `--columns <a,b>` | Comma-separated columns for `table`/`csv` |
| `--no-header` | Hide the table header row |
| `--jq <program>` | Apply a jq program to the result |
| `--dry-run` | Print the equivalent `curl` and send no request |
| `--show-token` | Do not redact the API key in `--dry-run` output |
| `--rps <float>` | Client-side rate limit in requests/sec |
| `--no-color` | Disable colored output (`NO_COLOR`) |
| `-q, --quiet` | Suppress non-essential chatter |
| `-v, --verbose` | Verbose (debug) logging to stderr |

## Where to next

- [Recipes](recipes.md) — copy-pasteable solutions to common tasks.
- [Command reference](commands/index.md) — every command, flag, and argument.
- [Beyond the API](beyond-api.md) — sync, backup/restore, and search in depth.
- [Workflows as Code](workflows-as-code.md) — the GitOps loop.
