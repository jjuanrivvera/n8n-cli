---
name: n8nctl-cli
description: Manage n8n (https://n8n.io) from the terminal with the `n8nctl` CLI - workflows, executions, credentials, tags, variables, projects, users, audit, and source control. Use this whenever the user wants to list/activate/transfer workflows, inspect or retry executions, create credentials, set variables, manage projects and members, invite users, run a security audit, or pull from Git - on a single instance or across MANY instances (self-hosted and Cloud) via named profiles. n8nctl is one static binary that talks to the n8n public REST API (`<host>/api/v1`, `X-N8N-API-KEY` header) with table/json/yaml/csv output. Detect the connected instance with `n8nctl auth status` before any write.
version: 0.5.0
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
| `workflows apply --dir <dir>` | GitOps reconcile: create/update/skip workflows from a dir (`--prune` deletes absent, `--activate`); always `--dry-run` first | `n8nctl workflows apply --dir ./workflows --dry-run` |
| `workflows lint [--dir\|-f\|--remote]` | Static checks (8 grounded rules incl. node-schema typo/param/value checks — `invalid-parameter-value` flags an options value the node disallows); exits non-zero on errors, so it gates CI (`--list-rules`, `--disable-rule`, `-o json`) | `n8nctl workflows lint --dir ./workflows` |
| `workflows autofix [--dir\|-f]` | Repair mechanical mistakes lint detects: typo'd node types (vs catalog), expressions missing leading `=`, missing webhook ids. Report-only unless `--write` | `n8nctl workflows autofix --dir ./workflows --write` |
| `workflows breaking-changes [--dir\|-f\|--remote\|<id>]` | Report nodes pinned to an older `typeVersion` than the catalog's latest, plus params the latest schema dropped. Informational, exits 0. Alias `breaking` | `n8nctl workflows breaking-changes --dir ./workflows` |
| `workflows bulk activate\|deactivate --tag <name>` | Flip every workflow carrying a tag in one call (maintenance windows); previews then needs `--yes` (or `--dry-run`) | `n8nctl workflows bulk deactivate --tag prod --yes` |
| `workflows convert <file…> --to json\|yaml` | Convert workflow files between JSON/YAML; `--externalize N` splits long code fields to sibling files | `n8nctl workflows convert wf.json --to yaml --externalize 5` |
| `workflows diff <id> [--to <profile>\|--file <path>]` | Unified diff of a workflow's writable content vs another profile or a local file | `n8nctl workflows diff 2tUt1wbLX592XDdX --to prod` |
| `workflows search [--node\|--credential\|--webhook\|--name]` | Scan all workflows' node graphs for matches (impossible in the UI) | `n8nctl workflows search --node slack` |
| `nodes list\|search <query>\|show <type>` | Browse the embedded n8n node catalog OFFLINE (no API call) — find the exact `type` string and a node's parameter names; same catalog powers lint/autofix | `n8nctl nodes show n8n-nodes-base.webhook` |
| `templates search <query>\|get <id>\|deploy <id>` | Browse the public n8n template gallery (api.n8n.io, NO key); `get` prints a definition (pipe to a file), `deploy` creates it on the ACTIVE instance (credentials NOT included; `--name`, `--activate`, honors `--dry-run`) | `n8nctl templates search slack --limit 5` |
| `executions list` | List runs (filters `--status`, `--workflow`, `--project`, `--include-data`) | `n8nctl executions list --status error --workflow 42` |
| `executions get <id>` | Inspect one run (`--include-data` for payloads) | `n8nctl executions get 9001 --include-data` |
| `executions retry <id>` | Re-run a failed execution (`--load-workflow`) | `n8nctl executions retry 9001` |
| `executions stop <id>` | Stop a running execution | `n8nctl executions stop 9001` |
| `executions delete <id>` | Delete an execution record | `n8nctl executions delete 9001 -y` |
| `executions prune [--older-than\|--status\|--workflow]` | Bulk-delete executions by age and/or status to reclaim DB space; previews the count, needs `--yes` (or `--dry-run`). `--older-than` takes `30d`/`720h`/`90m` | `n8nctl executions prune --older-than 30d --status error --yes` |
| `executions watch [--status\|--workflow\|--interval]` | Live-tail new executions, coloring failures, until Ctrl-C | `n8nctl executions watch --status error --interval 10s` |
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
| `stats [--recent N]` | One-shot instance summary: total/active/inactive/archived workflows + status mix of recent executions; degrades gracefully on Community-edition 403s | `n8nctl stats --recent 100 -o json` |
| `source-control pull` | Pull from the connected Git repo (`--force`, `--variables`) | `n8nctl source-control pull --dry-run` |
| `data-tables ...` | CRUD for data tables + rows (`rows`, `add-rows`, `update-rows`, `upsert-rows`, `delete-rows`) | `n8nctl data-tables rows <id> --filter '{"type":"and","filters":[]}'` |
| `packages export\|import` | Bundle/restore workflows as a `.n8np` archive (beta; needs `N8N_PUBLIC_API_PACKAGES_ENABLED`) | `n8nctl packages export --workflow 42 --out wf.n8np` |
| `skills install\|path\|print` | Install this skill into an AI agent (claude/cursor/windsurf/codex/gemini/copilot/opencode) | `n8nctl skills install --global` |
| `mcp start\|stream\|tools` | Run n8nctl as an MCP server (stdio / HTTP / export tool list) so an agent drives n8n via 73 annotated tools | `n8nctl mcp start` |
| `mcp claude\|cursor\|vscode enable\|disable\|list` | Wire the MCP server into a host's config | `n8nctl mcp claude enable` |
| `agent guard --host <h>` | Generate host-level rules that block destructive n8n ops for an agent (`--all-writes`, `--write`) | `n8nctl agent guard --host claude-code` |
| `proxy` | Local reverse proxy that lint-gates workflow writes (422 on lint errors) before they reach n8n; reads pass through, key injected from keyring (`--disable-rule`, `--block-destructive`, `--reject-duplicate-names`) | `n8nctl proxy --listen 127.0.0.1:8099` |
| `backup --out <dir>` | Snapshot the instance (workflows + tags + variables + credential metadata + manifest) for git; `--format yaml` + `--externalize N` make it git-friendlier | `n8nctl --profile prod backup --out ./backups/prod --format yaml --externalize 5` |
| `restore --in <dir>` | Re-apply a backup directory (reads JSON or YAML, re-inlines `$ref` code; `--update-by-name`, `--activate`); credential secrets are NOT in the backup | `n8nctl --profile staging restore --in ./backups/prod --update-by-name` |
| `api <METHOD> <PATH>` | Raw authenticated request (escape hatch) | `n8nctl api GET /workflows -q limit=5` |
| `auth login\|logout\|status` | Manage the active profile's key | `n8nctl auth status` |
| `config path\|view\|set\|use\|list-profiles` | Inspect/edit config and profiles | `n8nctl config use cloud` |
| `init` · `doctor` · `version` · `completion` · `alias` | Setup, diagnostics, version, shell completion, command shortcuts | `n8nctl doctor` |
| `login` / `logout` · `config set-url\|set-api-key\|show` | Aliases matching the official CLI's surface | `n8nctl config set-url https://n8n.lan` |

**Output flags (every command):** `-o table\|json\|yaml\|csv\|id`, `--columns a,b`,
`--no-header`, and `--jq '<program>'` for full jq filtering. Prefer `-o json`
(or `--jq`) when you need to parse output programmatically; use `-o id` to pipe
ids into `xargs`.

### Beyond the API (standout operations)

These commands compose the REST API into things the n8n UI cannot do — reach for
them when the user wants cross-instance promotion, git-versioned snapshots, a
graph-wide search, or a full **workflows-as-code / GitOps** loop:

- **`workflows apply --dir <dir>`** is the GitOps reconcile: a directory of
  workflow files (JSON or YAML) is the desired state. It creates new workflows,
  updates changed ones (matched by name), skips unchanged ones (canonical
  compare), and with `--prune` deletes instance workflows absent from the dir;
  `--activate` turns on the newly created. **Always preview with `--dry-run`**
  (especially with `--prune`). The standout is multi-instance promotion — the same
  dir across profiles: `n8nctl --profile staging workflows apply --dir ./workflows`
  then `n8nctl --profile prod workflows apply --dir ./workflows --prune`. The
  official single-instance tools can't do this.
- **`workflows lint`** runs 9 grounded static rules over files (`--dir`/`-f`) or
  live workflows (`--remote`): required-fields, connection-reference,
  orphaned-node, webhook-id-required, expression-prefix, plus three
  node-schema-aware rules that validate against the embedded n8n node catalog —
  unknown-node-type (typo detection), unknown-parameter (a param the node does not
  define), and **invalid-parameter-value** (an `options`/`multiOptions` value the
  node disallows, e.g. Slack `operation: "psot"` → "did you mean post?"). The
  value check is `displayOptions`-aware (it validates against the option set active
  for the node's other parameters) and conservative — it skips dynamic option
  lists, expression values (`={{ }}`), and unknown/community nodes, so it never
  false-positives. It **exits non-zero on errors**, so it gates CI. `--list-rules`
  shows each rule's canonical basis; `-o json` is machine-readable; `--disable-rule`
  turns a rule off.
- **`workflows breaking-changes`** (alias `breaking`) compares each workflow's
  nodes against the embedded catalog and reports those pinned to an older
  `typeVersion` than the latest known one, plus any parameters they use that the
  latest schema no longer defines (rename/removal candidates). Inputs mirror lint:
  `--dir`, `-f`, `--remote`, or a single live `<id>`. It is **informational and
  exits 0** — an upgrade-readiness report, not a CI gate. Run it before bumping an
  instance to a newer n8n to see which nodes will need attention.
- **`workflows autofix`** is the repair counterpart to lint: it fixes the
  mechanical mistakes lint reports — typo'd node types (corrected against the same
  embedded catalog), expression strings missing the leading `=`, and webhook/
  form-trigger nodes missing a `webhookId`. It reports by default and writes only
  with `--write`. Run autofix, then lint again to see the judgment-level findings
  that remain. Operates on files (`--dir` / `-f`).
- **`workflows convert <file…> --to json|yaml`** converts workflow files between
  JSON and YAML locally. `--externalize N` splits node code fields longer than N
  lines (jsCode, pythonCode, query/sqlQuery, jsonBody, content) into sibling files
  under `_subfiles/<stem>/`, replaced by a `{$ref: …}` marker re-inlined on read.
- **`workflows diff <id>`** prints a unified diff of a workflow's writable content
  (read-only fields ignored) against the same name on another `--profile` or a
  local `--file` — the review step before a sync or apply.
- **`workflows sync <id> --to <profile>`** promotes a single workflow between
  instances (dev → staging → prod) over the plain API — a Community-tier substitute
  for Enterprise Git Source Control. Read-only fields are stripped;
  `--update-by-name` overwrites by name, `--activate` turns it on. **Credentials
  are referenced by id and are NOT copied** — ensure matching credentials exist on
  the destination.
- **`backup --out <dir>` / `restore --in <dir>`** snapshot an instance (workflows +
  tags + variables + credential metadata + manifest) for git versioning, and
  re-apply it. `backup --format yaml --externalize N` makes the snapshot
  git-friendlier; `restore` reads either format and re-inlines `$ref` code.
  **Credential secrets are never exported** (the API is write-only for them);
  referenced credentials must already exist on restore.
- **`workflows search`** scans every workflow's node graph: `--node <type>`,
  `--credential <id|name>`, `--webhook <path>`, `--name <regex>`. Use it to answer
  "which workflows use the Slack node / reference credential X / own /orders".
- **`templates search|get|deploy`** browses the public n8n template gallery
  (api.n8n.io). `search` and `get` hit the gallery and **need no API key** (no
  profile, no instance); `get <id>` prints a definition you can pipe to a file and
  lint. `deploy <id>` is the only write: it creates the template as a new workflow
  on the **active instance** (`--name`, `--activate`, honors `--dry-run`).
  **Credentials are NOT included** — the workflow references credential types but
  holds no secrets, so connect them before activating. A safe pattern is to deploy
  into a dev profile, adapt, then promote with `apply`/`sync`.

The full GitOps loop (backup → edit in git → lint in CI → apply with dry-run →
apply with prune to a target) and the multi-instance promotion angle are
documented in the project's `docs/workflows-as-code.md`.

### MCP & agent safety

`n8nctl` can run as an MCP server so an AI agent (Claude Code/Desktop, Cursor,
VS Code) drives n8n through typed tools instead of shelling out — and it ships an
`agent guard` that fences those operations.

- **`mcp start`** runs an MCP server over stdio (use `mcp stream --host H --port N`
  for HTTP, `mcp tools` to export the catalog to `mcp-tools.json`). It auto-exposes
  the CLI as **85 MCP tools** named with an `n8n` prefix (`n8n_workflows_list`,
  `n8n_workflows_create`, `n8n_workflows_delete`, `n8n_data-tables_delete-rows`).
  Each tool replays the matching cobra command, reusing the same keyring auth,
  active profile, and `--dry-run`. Tools carry annotations — **read-only**
  (list/get/search/lint/diff/schema/members/backup/audit), **write**
  (create/update/activate/transfer/restore/sync/…), **destructive**
  (delete/delete-rows) — so hosts gate writes automatically.
- **Wire it in** per host: `n8nctl mcp claude enable` (also `cursor`, `vscode`),
  with `disable`/`list` to match. The server uses **whatever profile is active at
  startup**; `--profile`/`--base-url` and the secret flags (`--api-key`,
  `--show-token`) are never exposed to the model, and setup commands (`auth`,
  `config`, `alias`, `init`, `skills`, `agent`, `doctor`) are excluded.
- **`agent guard --host <claude-code|codex|opencode>`** generates host-level
  safety config (derived from the live command tree, so it stays correct across
  upgrades): hard-block `delete`/`delete-rows`, make ordinary writes require
  approval, leave reads free. `--all-writes` blocks writes too; `--write` installs
  the files (never overwriting existing ones), else it prints for review. For
  Claude Code it emits a `.claude/hooks/n8nctl-guard.sh` PreToolUse hook +
  `.claude/settings.json` deny/ask rules; Codex gets a read-only-sandbox
  `~/.codex/config.toml`; OpenCode gets `opencode.json` rules. The guard is
  excluded from the MCP surface so an agent can't disable its own rails. **MCP-only
  operation is the strongest guarantee** — the MCP-tool block is hard, the Bash
  hook is best-effort (defeats quote/backslash tricks, not variable indirection).
- **`proxy`** complements the guard at the API boundary: it fronts the instance
  and rejects any workflow create/update that fails lint with a `422` (reads pass
  through). Point a client (or agent) at it so bad definitions can't land,
  regardless of who pushes them. `--reject-duplicate-names` adds a second gate:
  reject creating a workflow whose name already exists on the instance.
  `agent guard` blocks *destructive* ops; `proxy` blocks *low-quality* ones.

```bash
n8nctl mcp start                          # expose n8n to an agent over stdio
n8nctl mcp claude enable                  # wire the server into Claude Desktop
n8nctl agent guard --host claude-code     # print the safety config for review
n8nctl agent guard --host claude-code --write   # install it (won't overwrite)
n8nctl proxy                              # lint-gate every workflow write (127.0.0.1:8099)
n8nctl proxy --reject-duplicate-names     # also reject a create whose name already exists
```

See `docs/mcp.md` and `docs/agent-guard.md` for the full setup, manual JSON
config, and a worked list-then-create example.

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
