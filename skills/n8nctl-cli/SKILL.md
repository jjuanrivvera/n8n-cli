---
name: n8nctl-cli
description: Manage n8n (https://n8n.io) from the terminal with the `n8nctl` CLI - workflows, executions, credentials, tags, variables, projects, users, audit, and source control. Use this whenever the user wants to list/activate/transfer workflows, inspect or retry executions, create credentials, set variables, manage projects and members, invite users, run a security audit, or pull from Git - on a single instance or across MANY instances (self-hosted and Cloud) via named profiles. n8nctl is one static binary that talks to the n8n public REST API (`<host>/api/v1`, `X-N8N-API-KEY` header) with table/json/yaml/csv output. Detect the connected instance with `n8nctl auth status` before any write.
version: 0.1.0
homepage: https://github.com/jjuanrivvera/n8n-cli
license: MIT
allowed-tools: Bash(n8nctl:*)
metadata: {"openclaw":{"category":"automation","emoji":"🔁","requires":{"bins":["n8nctl"],"env":["N8NCTL_API_KEY"]},"install":[{"kind":"brew","formula":"jjuanrivvera/n8n-cli/n8nctl-cli","bins":["n8nctl"]},{"kind":"go","package":"github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest","bins":["n8nctl"]}]}}
---

# n8nctl CLI

Drive the [n8n](https://n8n.io) public REST API through the `n8nctl`
command-line tool. This skill teaches you how and when to use it.

`n8nctl` is a **single static Go binary** - no Node runtime, Homebrew- and
`go install`-able, with API keys stored in your **OS keyring** and
**multi-instance named profiles** so one CLI drives every instance you own.
That is the reason to reach for it over the official `@n8n/cli` (a Node, beta,
single-instance tool that keeps the key in plaintext) or raw `curl`.

## Prerequisites

- The `n8nctl` binary must be on `PATH`. Check with `n8nctl version`. If missing,
  install it: `brew install jjuanrivvera/n8n-cli/n8nctl-cli` or
  `go install github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest`.
- An n8n API key. In the n8n UI go to **Settings > n8n API > Create an API key**,
  copy it, and store it with `n8nctl auth login` (or `n8nctl init` for a guided
  first run). The key never needs to be pasted into a shell command.
- Confirm the setup with `n8nctl auth status` (alias `whoami`) or `n8nctl doctor`.

## Prefer the CLI over raw curl or MCP

- **No hand-built auth.** The CLI sets `X-N8N-API-KEY`, normalizes the base URL
  (adds the trailing `/api/v1` for you), retries on `429`/transient errors, and
  rate-limits client-side. You don't reconstruct any of that per request.
- **Keyring, not plaintext.** The key lives in the OS keyring per profile, so it
  never lands in shell history, a curl flag, or a config file in the repo.
- **Structured output.** `-o json|yaml|csv` plus `--columns` gives clean,
  script-friendly data; raw curl gives you a blob to re-parse.
- **Multi-instance.** One command, every instance, via `--profile`. curl makes
  you juggle URLs and keys by hand.
- **Escape hatch when needed.** Anything the typed commands don't cover is one
  `n8nctl api <METHOD> <PATH>` away - still authenticated, still rate-limited,
  still dry-runnable. Prefer it over reaching for curl.

## Auth & config setup (multi-instance is the point)

`n8nctl` keeps **named profiles**, one per n8n instance. Each profile has a base
URL (stored in `~/.n8nctl-cli/config.yaml`) and an API key (stored in the OS
keyring, never the file). Switch instances with `--profile <name>` on any
command, or change the default with `n8nctl config use <name>`.

```bash
n8nctl init                                 # guided first-run setup of a profile
n8nctl auth login                           # store/verify the key for the active profile
n8nctl auth login --base-url https://n8n.example.com   # set the URL too
n8nctl config list-profiles                 # see every instance and which is active
n8nctl config use homelab                   # switch the default instance
n8nctl workflows list --profile cloud       # one-off against another instance
```

A full multi-instance walkthrough (homelab + cloud + a client) is in
[references/auth-and-config.md](references/auth-and-config.md).

## Golden rules (read before acting)

1. **Preview writes with `--dry-run`.** For any create/update/delete or action
   (activate, transfer, retry, stop, invite, pull), run it once with `--dry-run`
   first - it prints the equivalent curl and sends nothing - then run for real
   after the user confirms. The key is redacted unless you pass `--show-token`.
2. **Inspect a credential's schema before creating it.** Run
   `n8nctl credentials schema <credentialTypeName>` (e.g. `githubApi`,
   `httpHeaderAuth`) to see exactly which `data` fields the type needs, then
   build the body. Never guess credential field names.
3. **Parse with JSON.** Add `-o json` and pipe to `jq` when you need to read
   values; the default `table` output is for humans.
4. **Never print the API key.** Don't echo `N8NCTL_API_KEY`, don't pass
   `--show-token` unless the user explicitly needs the curl with the key, and
   don't read it back out of the keyring. `n8nctl config view` already redacts it.
5. **Confirm destructive actions** (`delete`, `stop`) with the user; `delete`
   prompts unless `-y` is passed.
6. **Know which instance you're on.** Before a write, confirm the active profile
   with `n8nctl auth status` so you don't act on the wrong n8n.

## Workflow: auth → discover → act → verify

```bash
n8nctl auth status                 # 1. which instance am I on, is the key valid?
n8nctl <resource> --help           # 2. discover a resource's actions & filters
n8nctl <resource> list -o json     #    inspect real data and ids
# 3. act (always --dry-run first for writes)
n8nctl workflows activate 42 --dry-run
n8nctl workflows activate 42
n8nctl workflows get 42 -o json    # 4. verify the result
```

## Command cheatsheet

Global flags on every command: `-o/--output table|json|yaml|csv`, `--columns`,
`--profile`, `--base-url`, `--api-key`, `--rps`, `--dry-run`, `--show-token`,
`-v/--verbose`, `--no-color`, `-q/--quiet`. Write bodies on create/update via
`--data '<json>'`, `--file <path|->`, or repeated `--set key=value` (value parsed
as JSON when possible). `list` adds `--limit`, `--cursor`, `--all`,
`--max-pages`, `--param key=value`. `delete` prompts unless `-y`.

| Resource / command | What it does | Example |
|---|---|---|
| `workflows list` | List workflows (filters `--active`, `--name`, `--tags`, `--project`) | `n8nctl workflows list --active true -o json` |
| `workflows get <id>` | Inspect one workflow | `n8nctl workflows get 42` |
| `workflows create` | Create from JSON (needs name, nodes, connections, settings) | `n8nctl workflows create --file workflow.json` |
| `workflows update <id>` | Update a workflow | `n8nctl workflows update 42 --set active=true` |
| `workflows activate\|deactivate <id>` | Toggle a workflow on/off | `n8nctl workflows activate 42` |
| `workflows archive\|unarchive <id>` | Archive or restore | `n8nctl workflows archive 42` |
| `workflows transfer <id> --project <p>` | Move to another project | `n8nctl workflows transfer 42 --project 7` |
| `workflows tags <id> [--set id1,id2]` | Get or replace a workflow's tags | `n8nctl workflows tags 42 --set 3,8` |
| `workflows sync <id> --to <profile>` | Promote a workflow to another instance (dev→prod); credentials NOT copied | `n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate` |
| `workflows search [--node\|--credential\|--webhook\|--name]` | Scan all workflows' node graphs for matches (impossible in the UI) | `n8nctl workflows search --node slack` |
| `executions list` | List runs (filters `--status`, `--workflow`, `--project`, `--include-data`) | `n8nctl executions list --status error --workflow 42` |
| `executions get <id>` | Inspect one run (`--include-data` for payloads) | `n8nctl executions get 9001 --include-data` |
| `executions retry <id>` | Re-run a failed execution (`--load-workflow`) | `n8nctl executions retry 9001` |
| `executions stop <id>` | Stop a running execution | `n8nctl executions stop 9001` |
| `executions delete <id>` | Delete an execution record | `n8nctl executions delete 9001 -y` |
| `credentials schema <type>` | Show a credential type's required fields | `n8nctl credentials schema githubApi` |
| `credentials create` | Create a credential | `n8nctl credentials create --set name='GH' --set type=githubApi --set 'data={"accessToken":"…"}'` |
| `credentials transfer <id> --project <p>` | Move a credential to a project | `n8nctl credentials transfer 5 --project 7` |
| `tags ...` | CRUD for workflow tags | `n8nctl tags create --set name=Production` |
| `variables ...` | CRUD for instance variables (get-by-id matches id or key) | `n8nctl variables create --set key=API_BASE --set value=https://api.example.com` |
| `projects ...` | CRUD for projects (Enterprise) | `n8nctl projects create --set name='Billing'` |
| `projects members\|add-member\|set-member-role\|remove-member <id>` | Manage members | `n8nctl projects add-member 7 --user 12 --role project:editor` |
| `users list\|get` | List/inspect users | `n8nctl users list --include-role true` |
| `users invite --email a@x.com --role global:member` | Invite users | `n8nctl users invite --email a@x.com --email b@y.com` |
| `users change-role <id> --role global:admin` | Change a user's global role | `n8nctl users change-role 3 --role global:admin` |
| `audit` | Run the built-in security audit (`--categories`, `--days`) | `n8nctl audit --categories credentials,nodes -o json` |
| `source-control pull` | Pull from the connected Git repo (`--force`, `--variables`) | `n8nctl source-control pull --dry-run` |
| `backup --out <dir>` | Snapshot the instance (workflows + tags + variables + credential metadata + manifest) to JSON for git | `n8nctl --profile prod backup --out ./backups/prod` |
| `restore --in <dir>` | Re-apply a backup directory (`--update-by-name`, `--activate`); credential secrets are NOT in the backup | `n8nctl --profile staging restore --in ./backups/prod --update-by-name` |
| `api <METHOD> <PATH>` | Raw authenticated request (escape hatch) | `n8nctl api GET /workflows -q limit=5` |
| `auth login\|logout\|status` | Manage the active profile's key | `n8nctl auth status` |
| `config path\|view\|set\|use\|list-profiles` | Inspect/edit config and profiles | `n8nctl config use cloud` |
| `init` · `doctor` · `version` · `completion` · `alias` | Setup, diagnostics, version, shell completion, command shortcuts | `n8nctl doctor` |

### Beyond the API (standout operations)

These commands compose the REST API into things the n8n UI cannot do — reach for
them when the user wants cross-instance promotion, git-versioned snapshots, or a
graph-wide search:

- **`workflows sync <id> --to <profile>`** promotes a workflow between instances
  (dev → staging → prod) over the plain API — a Community-tier substitute for
  Enterprise Git Source Control. Read-only fields are stripped; `--update-by-name`
  overwrites by name, `--activate` turns it on. **Credentials are referenced by id
  and are NOT copied** — ensure matching credentials exist on the destination.
- **`backup --out <dir>` / `restore --in <dir>`** snapshot an instance to pretty
  JSON (workflows + tags + variables + credential metadata + manifest) for git
  versioning, and re-apply it. **Credential secrets are never exported** (the API
  is write-only for them); referenced credentials must already exist on restore.
- **`workflows search`** scans every workflow's node graph: `--node <type>`,
  `--credential <id|name>`, `--webhook <path>`, `--name <regex>`. Use it to answer
  "which workflows use the Slack node / reference credential X / own /orders".

Deeper, per-resource examples are in
[references/n8n-commands.md](references/n8n-commands.md); output formats,
columns, filtering and pagination are in
[references/output-and-filtering.md](references/output-and-filtering.md).

## Troubleshooting

- **`401 Unauthorized`** - the API key is missing, wrong, or revoked. Re-run
  `n8nctl auth login` for the active profile, or check `n8nctl auth status`.
  Verify the key still exists under **Settings > n8n API** in the n8n UI.
- **`403 Forbidden`** - the key is valid but lacks scope, or the feature is
  Enterprise-only (projects, variables, source control, users). Use an
  owner/admin key and confirm the instance's license covers the resource.
- **Base URL** - pass the host with or without `/api/v1`; the CLI appends
  `/api/v1` automatically (`https://n8n.example.com` → `…/api/v1`). A
  non-HTTPS base URL triggers a clear-text warning because the key is sent in
  the header.
- **Wrong instance** - every command takes `--profile <name>`; the default is
  set by `n8nctl config use <name>` (or `N8NCTL_PROFILE`). Confirm with
  `n8nctl auth status` before writing.
- **Rate limiting** - the client backs off on `429` automatically; tune the
  client-side cap with `--rps` or config `requests_per_second`.

## More

A condensed command reference, an auth/profiles deep dive, and an
output/filtering guide ship alongside this skill in `references/`.
