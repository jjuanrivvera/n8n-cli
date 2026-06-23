# n8nctl vs. the official n8n CLI

There are two command-line clients for the n8n public API:

- **`@n8n/cli`** (binary `n8n-cli`) — the **official, first-party** CLI from the
  n8n team. Node.js/oclif, distributed on npm. Currently labelled **beta**
  ("use it only for experimenting, local development, and personal projects").
- **`n8nctl`** (this project) — an independent, single-binary Go client focused
  on multi-instance use, production resilience, and fleet operations.

The official CLI is a capable, well-designed tool maintained by the people who
build the API itself; several ideas here (a built-in `--jq`, an `id-only` output
mode, a `skill install` command) originated there, and `n8nctl` matches or
extends them. For a single instance inside a Node project, the official tool is
an excellent choice; `n8nctl` is aimed at operating several instances from one
static binary, treating workflows as code, and driving n8n from AI agents safely.

_Comparison as of `@n8n/cli` 0.8.0 and `n8nctl` 0.3.0 (June 2026). Both track the
same public API, so the CRUD surface converges over time; the differences are in
multi-instance operation, workflows-as-code, agent tooling, and runtime model._

## At a glance

| | n8nctl | @n8n/cli (official) |
|---|---|---|
| Vendor | Independent (third-party) | **First-party (n8n team)** |
| Runtime | **Single static Go binary, no deps** | Node.js + npm |
| Install | Homebrew, Scoop, `go install`, deb/rpm/apk, raw binary | `npm i -g @n8n/cli` / `npx` |
| Maturity | New, but versioned + released | **Official**, but **beta / not for production** |
| Tracks API changes | Community pace | **Day one (same team)** |
| Multi-instance | **Named profiles + keyring** | Single instance, plaintext config |
| Workflows as code | **apply / lint / diff / convert (GitOps)** | — |
| Manage instances via MCP | **MCP server (mgmt API) + agent guard** | — |

_The "manage via MCP" row is the **management/control plane** — an agent
administering instances. n8n the platform separately has first-party MCP at the
**workflow** layer (MCP Server Trigger / MCP Client Tool nodes); the two are
complementary, not competing — see [below](#where-n8nctl-is-different)._

## Command coverage

As of this writing, **n8nctl covers every command the official CLI exposes**,
plus more. Parity across the shared surface:

| Area | Both have | n8nctl also adds |
|---|---|---|
| workflows | list, get, create, update, delete, activate, deactivate, transfer, tags | **archive/unarchive, sync (cross-instance), search, apply/lint/diff/convert (GitOps)** |
| executions | list, get, retry, stop, delete | — |
| credentials | create, get, list, schema, transfer | **update** |
| tags / variables | full CRUD | — |
| projects | create, get, list, update, delete, members, add/remove-member | **set-member-role** |
| users | get, list | **invite, change-role, delete** |
| data-tables | list, get, create, delete, rows, add/update/upsert/delete-rows | update |
| packages (beta) | export, import (official also has `shared`) | — |
| audit | generate | **--days / --categories options** |
| source-control | pull | — |
| backup / restore | — | **whole-instance snapshot to git-friendly JSON/YAML** |
| agent access (mgmt) | — | **`mcp` server over the management API (73 tools) + `agent guard`** |
| config / auth | config set-url/set-api-key/show, login, logout | **profiles: use, list-profiles, view** |
| skill | install (Claude/Cursor/Windsurf) | **+ codex/gemini/copilot/opencode, path, print** |

The official CLI's one command `n8nctl` does not have is `package shared` (sharing a
workflow package); everything else in the shared CRUD surface, n8nctl matches or extends.

## Where n8nctl is different

These are not in the official CLI today:

- **Multi-instance profiles.** Named profiles with `config use` / `--profile` /
  `N8NCTL_PROFILE`. The official CLI stores a single `{url, apiKey}`; switching
  instances means overwriting it.
- **OS keyring for secrets.** Keys live in the macOS Keychain / Linux Secret
  Service / Windows Credential Manager, per profile. The official CLI writes the
  key in plaintext to `~/.n8n-cli/config.json` (mode 0600).
- **Production resilience.** Exponential backoff with jitter, idempotency-aware
  retries (never retries POST/PATCH), adaptive rate limiting, and 429 handling.
  The official client issues a single `fetch` with no retry or rate limiting.
- **`--dry-run`** prints a copy-pasteable, secret-redacted `curl` and sends
  nothing.
- **Richer output:** table / json / **yaml** / **csv** with `--columns`
  (official: table / json / id-only).
- **Fleet operations:** `workflows sync` (promote dev→prod across instances),
  `backup` / `restore` (git-friendly snapshots), `workflows search`
  (find by node type / credential / webhook path).
- **Workflows as code (GitOps).** `workflows apply --dir` reconciles a directory of
  workflow files into an instance (create / update / skip-unchanged / `--prune`,
  with a `--dry-run` plan); `workflows lint` runs schema-grounded checks as a CI
  gate; `workflows diff` and `workflows convert` (JSON↔YAML, with long code fields
  externalized to sibling files) round out the loop. Combined with profiles, the
  same directory promotes across instances.
- **Manage instances from an AI agent (MCP).** `n8nctl mcp` runs the CLI as a
  Model Context Protocol server — it auto-exposes the command tree as 73 tools
  (each tagged read-only / write / destructive), reusing the same keyring auth and
  active profile, and installs config for Claude/Cursor/VS Code. `n8nctl agent
  guard` then generates host-level safety rules (Claude Code / Codex / OpenCode)
  that hard-block destructive operations and gate writes.

    This operates at a different layer than n8n's own MCP support, and is
    complementary to it. **n8n the platform is already MCP-native at the workflow
    layer** — the **MCP Server Trigger** node turns a workflow into an MCP server
    (agents call your workflows as tools), and the **MCP Client Tool** node lets a
    workflow consume external MCP tools. That is the *data plane*: automations as
    tools. `n8nctl`'s MCP is the *control plane*: it exposes the n8n **management
    API** (list / create / activate / delete workflows, manage credentials,
    projects, executions, across instances) so an agent can **operate and
    administer** your n8n fleet. The official `@n8n/cli` does not expose its
    management surface over MCP, and given n8n's first-party workflow-level MCP it
    likely never will — which is exactly the gap `n8nctl` fills.
- **More meta tooling:** `doctor`, `init` wizard, `alias`, raw `api` escape hatch.
- A **full jq** engine (gojq) behind `--jq`, vs the official's simpler path
  filter; **7 skill targets** vs 3.

## Where the official CLI is genuinely better

Reasons to prefer `@n8n/cli`:

- **First-party.** Built and maintained by the n8n team alongside the API, so it
  tracks new endpoints and changes immediately. A third-party CLI can lag — for
  example, `@n8n/cli` 0.8.0 added `package shared` (workflow-package sharing),
  which `n8nctl` does not have yet.
- **Thoughtful scripting defaults.** It **auto-selects JSON when stdout is piped**
  (so `n8n-cli workflow list | jq` just works) and **auto-paginates lists by
  default**. n8nctl defaults to a table and uses `--all` to walk pages, which is
  more predictable but less convenient for piping. (Use `n8nctl ... -o json` or
  set `N8NCTL_OUTPUT=json` when scripting.)
- **npm-native.** If you already have Node and a JS toolchain, `npx @n8n/cli` is
  zero-friction and fits naturally in package scripts.
- **Official support and trust.** It is the tool the n8n docs point to.

## Performance

A single API call is **network-bound**: per-request time is dominated by the
instance's latency, not the CLI. The CLI's own processing is in the microseconds:

| Operation (50-record page, no network) | n8nctl |
|---|---|
| Decode a list page | ~50 µs |
| Render a table | ~150 µs |
| Render JSON | ~80 µs |
| Apply a `--jq` filter (full gojq) | ~90 µs |

The one place a CLI's own cost is visible — and where being a single binary
matters — is **process startup**, which you pay on every invocation:

| | Startup (warm, `--version`) | Footprint |
|---|---|---|
| **n8nctl** | **~6–8 ms** | one 14 MB static binary, no runtime |
| @n8n/cli | ~150–160 ms | 5.6 MB package **+ a Node.js runtime** |

That is roughly a **19× faster cold start** (Go binary vs Node + oclif boot).

For a single command you will not notice it — 150 ms of startup disappears next
to a network round-trip. It only adds up when
the CLI is invoked **many times**: a loop over 200 workflow ids, an `xargs`
pipeline, or a CI job spawns the process repeatedly, where ~150 ms each becomes
~30 s of pure overhead versus ~1.5 s for n8nctl. If you invoke the CLI once
interactively, both feel the same.

_Measured on Apple Silicon with `hyperfine --warmup 3 --runs 40`; Node startup
has high variance. Reproduce with `go test -bench=. ./internal/...` and
`hyperfine -N 'n8nctl version' 'n8n-cli --version'`._

## Which should you use?

- **Use `@n8n/cli`** if you want the official tool, work primarily with a single
  instance, already live in Node, or need brand-new endpoints the moment they
  ship.
- **Use `n8nctl`** if you manage **multiple instances**, want **no Node** and a
  signed single binary, value **keyring** security and **production resilience**,
  keep **workflows as code** (apply / lint / diff in CI), drive n8n from an **AI
  agent** (MCP + guard), or want **backup / promote / search** across a fleet — and
  you are comfortable with a community tool that follows the API rather than
  defining it.

Many teams will reasonably use both: the official CLI inside Node projects, and
n8nctl for ops, multi-instance promotion, GitOps, agent access, and backups.
