# n8nctl

Control any n8n instance from your terminal. One static binary, many instances.

`n8nctl` is a fast, scriptable command-line client for the
[n8n public REST API](https://docs.n8n.io/api/). It manages workflows,
executions, credentials, tags, variables, projects, users, data tables, the
audit report, and the source-control integration on any n8n instance —
self-hosted or Cloud — over HTTPS with an API key. Beyond plain CRUD it promotes
workflows across instances, snapshots an instance to git, applies a directory of
workflow files like GitOps, runs as an MCP server for AI agents, and gates
workflow writes behind a linter. It is built in Go as a single static binary:
no Node runtime to install and nothing to keep updated except one file on your
`PATH`.

!!! note "Unofficial"
    Not affiliated with n8n. Talks to the public API at `<your-host>/api/v1`
    using the `X-N8N-API-KEY` header.

## Install

```bash
brew install jjuanrivvera/n8n-cli/n8nctl-cli   # macOS/Linux
```

Or `go install github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest`, or download a
prebuilt binary from the [latest release](https://github.com/jjuanrivvera/n8n-cli/releases/latest).
Then run the first-time setup:

```bash
n8nctl init             # name a profile, capture the base URL and API key
n8nctl workflows list   # list workflows on the active instance
```

## Features at a glance

**Resource management** — full CRUD over every API resource.

- Workflows, executions, credentials, projects, users, tags, variables, and
  data tables, with resource-specific actions (activate, retry, transfer,
  members, rows, and more). Plus the audit report, package import/export, and
  source-control pull. → [Features](features.md)

**Fleet operations** — work across instances and to disk.

- `workflows sync` promotes a workflow between instances; `backup`/`restore`
  snapshot an instance for git versioning; `workflows search` finds workflows by
  node type, credential, webhook path, or name. → [Beyond the API](beyond-api.md)

**Workflows as code** — GitOps for workflow definitions.

- `workflows apply --dir` reconciles a directory of files into an instance
  (`--prune`, `--dry-run`); `lint` runs static checks; `diff` compares against
  another instance or a file; `convert` moves between JSON and YAML.
  → [Workflows as Code](workflows-as-code.md)

**AI agents** — a safe tool surface for coding agents.

- `mcp` runs the CLI as a Model Context Protocol server; `agent guard` blocks
  destructive operations; `proxy` lint-gates every workflow write; `skills
  install` ships the agent skill. → [MCP server](mcp.md) ·
  [Agent guard](agent-guard.md) · [Lint-enforcing proxy](proxy.md)

**Output and scripting** — built for pipelines.

- `table`/`json`/`yaml`/`csv`/`id` output, `--columns`, `--jq` (gojq) filtering,
  `--dry-run` that prints the equivalent `curl`, the raw `api` escape hatch, and
  cursor pagination with `--all`. → [Output and filtering](output.md)

**Multi-instance and security** — many instances, secrets in the keyring.

- Named profiles switched with `--profile` or `config use`, API keys stored in
  your OS keyring, and configuration precedence of flag > env > config > default.
  → [Multi-instance and profiles](profiles.md) · [Authentication](authentication.md)

**Resilience** — predictable under load.

- Idempotent-only retry with backoff, adaptive client-side rate limiting, clean
  Ctrl-C cancellation, typed API errors with hints, and tolerant JSON parsing of
  n8n's loose encodings. → [Features](features.md#resilience)

## Where to next

- [Features](features.md) — the complete capability reference in one page.
- [Getting started](getting-started.md) — install, authenticate, first commands.
- [Recipes](recipes.md) — copy-pasteable solutions to common tasks.
- [Multi-instance and profiles](profiles.md) — drive many instances from one binary.
- [Output and filtering](output.md) — formats, columns, jq, and pagination.
- [Command reference](commands/index.md) — every command, flag, and argument.
- [vs. other n8n CLIs](comparison.md) — how `n8nctl` differs from `@n8n/cli`.
