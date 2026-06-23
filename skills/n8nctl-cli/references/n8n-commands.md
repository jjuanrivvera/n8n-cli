# n8nctl - command cheatsheet

Condensed, per-resource reference loaded on demand by the `n8nctl-cli` skill.
Authoritative docs: https://github.com/jjuanrivvera/n8n-cli

## Global flags (any command)

| Flag | Meaning |
|---|---|
| `-o, --output table\|json\|yaml\|csv` | Output format (default table) |
| `--columns a,b,c` | Columns for table/csv |
| `--profile NAME` | Instance profile (env `N8NCTL_PROFILE`) |
| `--base-url URL` | Override the instance base URL (`/api/v1` auto-added) |
| `--api-key KEY` | Override the API key (prefer the keyring via `auth login`) |
| `--rps N` | Client-side rate limit, requests/sec |
| `--dry-run` | Print the equivalent curl and send nothing |
| `--show-token` | Don't redact the key in `--dry-run` |
| `-v, --verbose` | Debug logging to stderr |
| `--no-color` | Disable colored output |
| `-q, --quiet` | Suppress non-essential chatter |

## List flags (every `<resource> list`)

| Flag | Meaning |
|---|---|
| `--limit N` | Page size (capped at 250) |
| `--cursor STR` | Pagination cursor from a previous response |
| `--all` | Fetch every page (auto-paginate) |
| `--max-pages N` | Page cap for `--all` (0 = safety default) |
| `--param key=value` | Any raw n8n query parameter (repeatable) |

## Write bodies (every `create` / `update`)

Three interchangeable ways, combinable:

```bash
n8nctl <res> create --file body.json        # from a file ('-' for stdin)
n8nctl <res> create --data '{"name":"x"}'    # inline JSON
n8nctl <res> create --set name=x --set active=true   # flat key=value (value parsed as JSON when possible)
```

`--set value` is JSON-aware: `--set count=5` → number, `--set active=true` →
bool, `--set 'data={"k":1}'` → object, `--set 'tags=["a","b"]'` → array, anything
else stays a string. `delete` prompts unless `-y/--yes`.

---

## workflows  (aliases: workflow, wf)

Columns: `id, name, active, isArchived, triggerCount, updatedAt`.
List filters: `--active true|false`, `--name <substr>`, `--tags a,b`,
`--project <id>`.

```bash
n8nctl workflows list --active true
n8nctl workflows list --tags Production --project 7 -o json
n8nctl workflows get 42

# create from an n8n export (body needs name, nodes, connections, settings)
n8nctl workflows create --file workflow.json
n8nctl workflows create --file -            # from stdin

n8nctl workflows update 42 --set active=true
n8nctl workflows delete 42 -y

# lifecycle actions (simple POSTs)
n8nctl workflows activate 42
n8nctl workflows deactivate 42
n8nctl workflows archive 42
n8nctl workflows unarchive 42

# move to another project (Enterprise)
n8nctl workflows transfer 42 --project 7

# tags: get, or replace with a comma-separated id list (empty clears)
n8nctl workflows tags 42
n8nctl workflows tags 42 --set 3,8
n8nctl workflows tags 42 --set ""

# --- beyond the API (see the "Beyond the API" section below) ---

# promote a workflow to another instance (dev -> staging -> prod)
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate

# workflows as code / GitOps (see the "Workflows as code" section below)
n8nctl workflows apply --dir ./workflows --dry-run        # preview a reconcile
n8nctl workflows apply --dir ./workflows --prune          # reconcile + delete drift
n8nctl workflows lint --dir ./workflows                   # static checks (CI gate)
n8nctl workflows convert wf.json --to yaml --externalize 5
n8nctl workflows diff 2tUt1wbLX592XDdX --to prod          # diff vs another profile

# search every workflow's node graph (impossible in the UI)
n8nctl workflows search --node slack
n8nctl workflows search --credential githubApi -o json
n8nctl workflows search --webhook /orders
n8nctl workflows search --name '^prod-'
```

## executions  (aliases: execution, exec)

Read-only plus retry/stop/delete - n8n creates executions by running workflows.
Columns: `id, workflowId, status, mode, finished, startedAt, stoppedAt`.
List filters: `--status`, `--workflow <id>`, `--project <id>`,
`--include-data true|false`. `--status` values: `canceled`, `crashed`, `error`,
`new`, `running`, `success`, `unknown`, `waiting`.

```bash
n8nctl executions list --status error
n8nctl executions list --status success --workflow 42 -o json
n8nctl executions get 9001                 # summary
n8nctl executions get 9001 --include-data  # full run payload

n8nctl executions retry 9001               # re-run a failed execution
n8nctl executions retry 9001 --load-workflow   # use the current workflow definition
n8nctl executions stop 9001                # stop a running execution
n8nctl executions delete 9001 -y
```

## credentials  (aliases: credential, cred, creds)

Columns: `id, name, type, createdAt, updatedAt`. **Always read the type's schema
before creating** so you know which `data` fields are required.

```bash
# 1. discover the shape of a credential type
n8nctl credentials schema githubApi
n8nctl credentials schema httpHeaderAuth -o json

# 2. create with name + type + data (data shape comes from the schema)
n8nctl credentials create \
  --set name='GitHub (CI)' \
  --set type=githubApi \
  --set 'data={"accessToken":"ghp_…"}'

n8nctl credentials list
n8nctl credentials get 5
n8nctl credentials update 5 --set name='GitHub (renamed)'
n8nctl credentials delete 5 -y

# move a credential to a project (Enterprise)
n8nctl credentials transfer 5 --project 7
```

> n8n does not return credential secrets on `get`/`list` - you see metadata, not
> the stored values.

## tags  (alias: tag)

Plain CRUD for workflow tags. Columns: `id, name, createdAt, updatedAt`.

```bash
n8nctl tags list
n8nctl tags create --set name=Production
n8nctl tags get 3
n8nctl tags update 3 --set name=Prod
n8nctl tags delete 3 -y
```

## variables  (aliases: variable, var, vars)

Instance variables. The API has **no get-by-id endpoint**, so `get <id>` is
served by matching `id` *or* `key` within the full list. Columns:
`id, key, value, type`. List filter: `--project <id>`, `--state empty`.

```bash
n8nctl variables list
n8nctl variables create --set key=API_BASE --set value=https://api.example.com
n8nctl variables get API_BASE            # match by key…
n8nctl variables get 12                  # …or by id
n8nctl variables update 12 --set value=https://api.new.example.com
n8nctl variables delete 12 -y
```

## projects  (aliases: project, proj)

Projects and their members (n8n Enterprise). Columns: `id, name, type`.

```bash
n8nctl projects list
n8nctl projects create --set name='Billing'
n8nctl projects get 7
n8nctl projects update 7 --set name='Billing & Ops'
n8nctl projects delete 7 -y

# members
n8nctl projects members 7
n8nctl projects add-member 7 --user 12 --role project:editor
n8nctl projects set-member-role 7 12 --role project:admin
n8nctl projects remove-member 7 12
```

Member roles are `project:viewer`, `project:editor`, `project:admin`.

## users  (alias: user)

Instance-owner only. Columns:
`id, email, firstName, lastName, role, isPending`. List filters:
`--project <id>`, `--include-role true|false`.

```bash
n8nctl users list --include-role true
n8nctl users get 3

# invite one or more users (repeat --email); role defaults to global:member
n8nctl users invite --email new@acme.com
n8nctl users invite --email a@x.com --email b@y.com --role global:admin

n8nctl users change-role 3 --role global:admin   # global:admin | global:member
n8nctl users delete 3 -y
```

## audit

Run n8n's built-in security audit and print the report.

```bash
n8nctl audit
n8nctl audit -o json
n8nctl audit --categories credentials,nodes --days 30
```

`--categories` restricts to any of: `credentials`, `database`, `nodes`,
`filesystem`, `instance`. `--days` sets the inactivity window before a workflow
counts as abandoned.

## source-control  (alias: sc)

Git integration (Enterprise). `pull` applies the connected repo's state.

```bash
n8nctl source-control pull --dry-run
n8nctl source-control pull
n8nctl source-control pull --force        # resolve conflicts in favor of remote
n8nctl source-control pull --variables '{"ENV":"prod"}'   # variable overrides during pull
```

## api - raw request (escape hatch)

Call any endpoint the typed commands don't cover. `PATH` is relative to the
instance base; the leading `/api/v1` is added automatically. Still
authenticated, rate-limited, and `--dry-run`-able.

```bash
n8nctl api GET /workflows -q limit=5
n8nctl api GET /executions -q status=error -q limit=10
n8nctl api POST /tags -d '{"name":"Prod"}'
n8nctl api POST /workflows --file workflow.json
n8nctl api DELETE /executions/9001 --dry-run
```

Flags: `-d/--data '<json>'`, `--file <path|->`, `-q/--query key=value`
(repeatable). The method arg is upper-cased for you.

## data-tables  (aliases: data-table, dt)

Standard CRUD plus row operations. Rows are filtered with an n8n filter object.

```bash
n8nctl data-tables list
n8nctl data-tables create --set name=orders --set 'columns=[{"name":"sku","type":"string"}]'
n8nctl data-tables rows <id> --filter '{"type":"and","filters":[]}' --limit 50
n8nctl data-tables add-rows <id> --data '[{"sku":"A-1"}]'        # or --file rows.json / --stdin
n8nctl data-tables update-rows <id> --data '{"filter":{...},"data":{"sku":"A-2"}}'
n8nctl data-tables upsert-rows <id> --data '{"filter":{...},"data":{...}}'
n8nctl data-tables delete-rows <id> --filter '{"type":"and","filters":[...]}'
```

Data tables may be unlicensed on some editions (the API returns 403).

## packages - export / import (.n8np, beta)

Bundle workflows into a portable `.n8np` archive and import them elsewhere. Beta;
disabled unless the instance sets `N8N_PUBLIC_API_PACKAGES_ENABLED=true` (else 404).

```bash
n8nctl packages export --workflow 42 --workflow 43 --out bundle.n8np
n8nctl packages import --file bundle.n8np --conflict-policy fail --project <id>
```

`import` flags: `--conflict-policy` (required), `--project`, `--folder`,
`--workflow-id-policy`, `--credential-matching-mode`, `--credential-missing-mode`.

## skills - install this skill into an agent

```bash
n8nctl skills install                 # ./.claude/skills (this project)
n8nctl skills install --global        # ~/.claude/skills
n8nctl skills install --agent cursor --global
n8nctl skills path --agent windsurf   # print where it would install
```

Agents: claude, cursor, windsurf, codex, gemini, copilot, opencode. Or install
across every agent at once with `npx skills add jjuanrivvera/n8n-cli`.

## Beyond the API

These commands compose the public API into operations the n8n UI cannot do.
They honor the global flags, including `--dry-run` and `--profile`.

### workflows sync — promote a workflow between instances

Read a workflow from one profile and write it to another (dev → staging → prod).
A Community-tier substitute for Enterprise Git Source Control. Read-only fields
(`id`, active state, version) are stripped; nodes, connections and settings are
carried over.

```bash
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod                # create new on prod
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate
n8nctl --profile staging workflows sync 2tUt1wbLX592XDdX --to prod        # --from defaults to active profile
```

Flags: `--to <profile>` (required), `--from <profile>` (default: active
profile), `--update-by-name` (overwrite a destination workflow with the same
name), `--activate`.

> **Credentials are referenced by id and are NOT copied.** Create matching
> credentials on the destination first (`n8nctl credentials`); the synced nodes
> resolve them by id.

### backup / restore — snapshot an instance for git

`backup` exports the active instance to a directory: one file per workflow plus
`tags.json`, `variables.json`, a credentials inventory (metadata only) and a
manifest. `restore` re-applies a backup directory. Workflow files default to JSON;
`--format yaml` and `--externalize N` make the snapshot git-friendlier (YAML +
long code fields split into sibling files). `restore` reads either format and
re-inlines externalized `$ref` code automatically.

```bash
n8nctl --profile prod backup --out ./backups/prod
n8nctl --profile prod backup --out ./n8n-state --format yaml --externalize 5
n8nctl --profile staging restore --in ./backups/prod --update-by-name --activate
```

Flags: `backup --out <dir>` (required), `--format json|yaml` (default json),
`--externalize <N>` (0 = off); `restore --in <dir>` (required),
`--update-by-name`, `--activate`.

> **Credential secrets are never exported** (the n8n API is write-only for them);
> the backup holds credential metadata only. On restore, referenced credentials
> must already exist on the target instance.

### workflows search — scan node graphs

Report workflows that match a node type, credential reference, webhook path, or
name regex — questions the UI cannot answer.

```bash
n8nctl workflows search --node slack                       # node type substring
n8nctl workflows search --credential githubApi -o json     # by credential id or name
n8nctl workflows search --webhook /orders                  # by webhook path
n8nctl workflows search --name '^prod-'                    # by name regex
```

Flags: `--node <type>`, `--credential <id|name>`, `--webhook <path>`,
`--name <regex>`. Read-only.

## Workflows as code (GitOps)

A directory of workflow files (JSON or YAML) is the **desired state**; `apply`
reconciles it into an instance. The loop is: `backup` (seed) → edit in git → `lint`
in CI → `apply --dry-run` (preview) → `apply --prune` (reconcile a target).
Full guide: the project's `docs/workflows-as-code.md`.

### workflows apply — reconcile a directory into an instance

Treat `--dir` as the desired state for the active instance. Workflows match by
**name**: missing names are created, changed ones updated, already-matching ones
skipped (canonical compare of writable fields), and with `--prune`, instance
workflows whose name is absent from the dir are deleted. `--activate` turns on the
newly created. **Always preview with `--dry-run`**, especially with `--prune`.

```bash
n8nctl workflows apply --dir ./workflows --dry-run   # plan: N created, N updated, N unchanged, N pruned
n8nctl workflows apply --dir ./workflows             # create + update only
n8nctl workflows apply --dir ./workflows --prune     # also delete drift
n8nctl workflows apply --dir ./workflows --activate  # activate newly created

# Multi-instance promotion — the same dir across profiles (single-instance tools can't):
n8nctl --profile staging workflows apply --dir ./workflows
n8nctl --profile prod    workflows apply --dir ./workflows --prune
```

Output is one line per change (`+ create <name>`, `~ update <name>`,
`- prune <name>`) then a summary; a dry run labels it `plan:`, a real run
`applied:`. Flags: `--dir/-d <dir>` (required), `--prune`, `--activate`, plus the
global `--dry-run`.

> **Credentials are referenced by id, not copied.** Matching credentials must
> already exist on the target instance (same as `sync` / `restore`).

### workflows lint — static checks (CI gate)

Static checks over files (`--dir` / `-f`, repeatable) or live workflows
(`--remote`). **Exits non-zero on errors**, so it gates CI. Default output is a
text report (`✗` error, `⚠` warning); `-o json` is machine-readable.

```bash
n8nctl workflows lint --dir ./workflows
n8nctl workflows lint -f a.json -f b.yaml
n8nctl workflows lint --remote                       # lint the live instance
n8nctl workflows lint --dir ./workflows -o json      # structured findings
n8nctl workflows lint --list-rules                   # rules + canonical basis
n8nctl workflows lint --dir ./workflows --disable-rule expression-prefix
```

The 5 rules and their grounding (shown by `--list-rules`):

| Rule | Severity | Basis |
|---|---|---|
| `required-fields` | error | n8n public-API OpenAPI workflow schema (name, nodes, connections, settings) |
| `connection-reference` | error | workflow connection graph model (connection must target an existing node) |
| `orphaned-node` | warning | workflow connection graph model (node disconnected from the graph) |
| `webhook-id-required` | error | n8n webhook registration behavior (webhook/formTrigger need a `webhookId`) |
| `expression-prefix` | warning | n8n expression syntax (a `{{ }}` string is only evaluated if it starts with `=`) |

> Honest scope: there is **no official n8n linter**; these are `n8nctl`'s rules.
> They are structural/graph-level. **Node-schema param validation** (validating a
> node's parameters against that node type's schema) is **planned, not yet
> implemented**.

### workflows convert — JSON ↔ YAML (+ code externalization)

Convert workflow files between JSON and YAML on disk. `--externalize N` splits node
code fields longer than N lines (`jsCode`, `pythonCode`, `query`/`sqlQuery`,
`jsonBody`, `content`) into sibling files under `_subfiles/<stem>/`, replacing the
value with a `{$ref: <relpath>}` marker that `apply`/`lint`/`restore` re-inline on
read.

```bash
n8nctl workflows convert good.json --to yaml                     # good.json -> good.yaml (alongside)
n8nctl workflows convert *.yaml --to json --out ./json           # into a separate dir
n8nctl workflows convert code.json --to yaml --externalize 5 --out ./out
# -> ./out/code.yaml + ./out/_subfiles/code/Code-jsCode.js (jsCode replaced by {$ref: …})
```

Flags: `--to json|yaml` (required), `--out <dir>` (default: alongside input),
`--externalize <N>` (0 = off).

### workflows diff — review before promoting

Unified diff of a workflow's **writable** content (read-only fields ignored)
against the same name on another `--profile`, or a local `--file`. An empty diff
means `apply` would skip the workflow.

```bash
n8nctl workflows diff 2tUt1wbLX592XDdX --to prod                 # vs another profile
n8nctl workflows diff 2tUt1wbLX592XDdX --file ./workflows/intake.json
```

Flags: `--to <profile>` or `--file <path>`.

## MCP server & agent safety

Run `n8nctl` as an MCP server so an AI host (Claude Code/Desktop, Cursor, VS Code)
drives n8n through typed tools, and fence those operations with `agent guard`.

```bash
n8nctl mcp start                       # MCP server over stdio (what a host spawns)
n8nctl mcp stream --host 127.0.0.1 --port 8080   # same tools over HTTP
n8nctl mcp tools                       # export the tool catalog to mcp-tools.json
n8nctl mcp claude enable               # wire into Claude Desktop (also: list | disable)
n8nctl mcp cursor enable               # wire into Cursor
n8nctl mcp vscode enable               # wire into VS Code
```

The server auto-exposes the CLI as **73 MCP tools** named with an `n8n` prefix
(`n8n_workflows_list`, `n8n_workflows_create`, `n8n_workflows_delete`,
`n8n_data-tables_delete-rows`). Each tool replays the matching cobra command,
reusing the same keyring auth, active profile, and `--dry-run`. Tools are
annotated **read-only** (list/get/search/lint/diff/schema/members/backup/audit),
**write** (create/update/activate/transfer/restore/sync/…), or **destructive**
(delete/delete-rows), so a host can gate writes automatically.

The server uses **whatever profile is active at startup**. `--profile`/`--base-url`
and the secret flags (`--api-key`, `--show-token`) are never exposed to the model;
setup commands (`auth`, `config`, `alias`, `init`, `skills`, `agent`, `doctor`) are
excluded. `mcp ... enable` accepts `--server-name`, `--config-path`, and
`--env KEY=value`; set `N8NCTL_PROFILE` in the host config to pin a server to one
instance. Built on `github.com/njayp/ophis` (wrapping the official
`modelcontextprotocol/go-sdk`).

```bash
n8nctl agent guard --host claude-code           # print safety config for review
n8nctl agent guard --host codex                 # ~/.codex/config.toml (read-only sandbox)
n8nctl agent guard --host opencode --all-writes # opencode.json; block writes too
n8nctl agent guard --host claude-code --write   # install (never overwrites existing)
```

`agent guard` generates host-level rules **derived from the live command tree** (so
they stay correct across upgrades): hard-block `delete`/`delete-rows`, make ordinary
writes require approval, leave reads free. `--all-writes` blocks writes too; `--write`
installs the files (never overwriting), else prints for review. Claude Code gets a
`.claude/hooks/n8nctl-guard.sh` PreToolUse hook + `.claude/settings.json` deny/ask
rules; Codex gets a read-only-sandbox `~/.codex/config.toml`; OpenCode gets
`opencode.json` rules. The guard is excluded from the MCP surface so an agent can't
disable its own rails. **MCP-only operation is the strongest guarantee** — the
MCP-tool branch is a hard block; the Bash hook is best-effort (defeats quote/backslash
obfuscation, not variable indirection). Full setup: `docs/mcp.md`,
`docs/agent-guard.md`.

## Meta commands

```bash
n8nctl version [--json] [--check]      # version, commit, build date; --check looks for a newer release
n8nctl doctor [--json]                 # config / base URL / key / connectivity checks (non-zero exit on failure)
n8nctl init [--profile N] [--base-url U] [--api-key K]   # guided first-run setup
n8nctl auth login | logout | status    # manage the active profile's key (status alias: whoami)
n8nctl config path | view | set <k> <v> | use <profile> | list-profiles
n8nctl alias set <name> <expansion…> | list | remove <name>   # command shortcuts (can't shadow built-ins)
n8nctl completion bash | zsh | fish | powershell
```

## Notes

- **Preview writes** with `--dry-run` before running for real; the key is
  redacted in the printed curl unless `--show-token`.
- **Enterprise gating:** projects, variables, users management, and source
  control require the matching n8n license; expect `403` otherwise.
- **Ids are strings.** n8n ids are returned/printed as strings even when numeric;
  pass them as written.
- **Credentials never expose secrets** on read; `schema <type>` tells you what a
  type needs on create.
