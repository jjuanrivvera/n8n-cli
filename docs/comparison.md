# n8nctl vs other n8n CLIs

Several command-line clients exist for n8n, built by different people for
different jobs. This page maps the landscape honestly and shows where each tool
fits — including where the others are a better choice than `n8nctl`.

_As of June 2026: `n8nctl` 0.3.0, `@n8n/cli` 0.8.0, `ubie-oss/n8n-cli` 2.2.1,
`edenreich/n8n-cli` 0.7.1. All four talk to the same n8n public API, so the CRUD
surface converges over time; the real differences are in focus, runtime, and the
layers each tool adds on top._

## The cast

- **`@n8n/cli`** (binary `n8n-cli`) — the **official, first-party** CLI from the
  n8n team. Node.js/oclif, distributed on npm, currently **beta** ("for
  experimenting, local development, and personal projects"). The reference client
  for the public API; tracks new endpoints first.
- **`n8nctl`** (this project) — an independent, single-binary **Go** client
  focused on **multi-instance operation, workflows-as-code, and driving n8n from
  AI agents**.
- **`ubie-oss/n8n-cli`** — an independent **Bun** binary focused on **workflow
  authoring quality**: deep linting, node-schema inspection, formatting, data-flow
  tracing, and a server-side lint-enforcement proxy.
- **`edenreich/n8n-cli`** — an independent **Go** binary focused on one job:
  **GitOps sync** of workflows from a directory to an instance.

These tools are not all trying to be the same thing. The useful question is not
"which is best" but "which lane is yours."

## At a glance

| | n8nctl | @n8n/cli (official) | ubie-oss | edenreich |
|---|---|---|---|---|
| Vendor | 3rd-party | **First-party (n8n team)** | 3rd-party | 3rd-party |
| Runtime | **Go static binary** | Node.js + npm | Bun binary | Go binary |
| Primary lane | multi-instance **ops + GitOps + agent mgmt** | first-party general client | **authoring + quality** | **GitOps sync** |
| CRUD breadth | full + data-tables + packages | full + data-tables + packages (`+ shared`) | full + data-tables + `node-schema` | **workflows only** |
| Multi-instance | **profiles + keyring** | single, plaintext config | single, env vars | single, `.env` |
| Apply from files | **`apply`** (+ prune / dry-run, **cross-instance**) | — | **`apply`** (+ `--from-git-changes`) | **`sync`** (+ prune) |
| Lint | 5 schema-grounded rules | — | **6 rules incl. `node-params` (schema-aware) + `node-schema` dump** | — |
| Other authoring | convert, diff | — | **convert, fmt, trace, `proxy`** | refresh |
| Management MCP | **`mcp` (73 tools)** | — | — | — |
| Agent safety | **`agent guard`** | — | `proxy` (server-side lint gate) | — |
| backup / restore / search | **yes** | — | — | — |
| Secrets | **OS keyring** | plaintext config | plaintext env | plaintext `.env` |
| Resilience (retry / rate-limit) | **adaptive** | single fetch | basic | basic |
| Distribution | brew / scoop / deb-rpm-apk + cosign + SBOM | npm | source / cross-compile | curl / `go install` |

## Where each tool fits

### First-party API client — `@n8n/cli`

Built and maintained by the n8n team alongside the API, so it tracks new endpoints
and changes immediately and is the tool the n8n docs point to. It has thoughtful
scripting defaults — it **auto-selects JSON when stdout is piped** and
**auto-paginates lists by default** — and `npx @n8n/cli` is zero-friction inside a
Node project. It is single-instance (one `{url, apiKey}` in a plaintext config,
no keyring), issues a single `fetch` with no retry or rate limiting, and is
labelled not-for-production. **Use it** if you want the first-party tool, work
with one instance, already live in Node, or need brand-new endpoints on day one.
It is the only tool here with the `package shared` command.

### Workflow authoring & quality — `ubie-oss/n8n-cli`

The deepest authoring tool of the four, and genuinely ahead of `n8nctl` on
workflow *quality*. Its `lint` ships rules including **`node-params`, which
validates node parameters**, backed by a **`node-schema`** command that dumps the
real node-type schemas — per-node validation that `n8nctl`'s five structural rules
do not attempt. `fmt` tidies node layout, `trace` analyses data-flow cardinality,
and **`proxy`** is a distinctive idea: an HTTP proxy in front of the n8n API that
enforces lint **server-side**, so any push that fails lint — from a human or an AI
agent — is rejected with a 422, making quality structural rather than a convention.
It is single-instance (env vars, no keyring) and has no MCP or agent-management
layer. **Use it** if your priority is keeping a team's workflow definitions clean,
schema-valid, and consistently formatted.

### GitOps sync — `edenreich/n8n-cli`

Does one job cleanly. `workflows sync --directory --prune --dry-run` reconciles a
folder of workflow JSON into an instance; `refresh` pulls the other way; plus
`activate`/`deactivate`. It is workflow-only (no executions, credentials, tags,
projects, …), single-instance, and reads the key from a plaintext `.env`. **Use
it** if all you need is "push this folder of workflows to this instance in CI" with
the least surface to learn.

### Multi-instance operations, GitOps & agent management — `n8nctl`

This project's lane. It is the only one of the four that combines:

- **Multi-instance profiles + OS keyring.** Named profiles, one per instance,
  switched with `--profile` / `N8NCTL_PROFILE` / `config use`; keys in the macOS
  Keychain / Linux Secret Service / Windows Credential Manager. The others target a
  single instance with the key in a plaintext file or env var.
- **Production resilience.** Exponential backoff with jitter, idempotency-aware
  retries (never retries POST/PATCH), adaptive rate limiting, and 429 handling.
- **Workflows as code.** `workflows apply --dir` reconciles a directory into an
  instance (create / update / skip-unchanged / `--prune`, with a `--dry-run`
  plan); `workflows lint` (schema-grounded rules as a CI gate), `diff`, and
  `convert` (JSON↔YAML with long code fields externalised to sibling files). Both
  ubie and edenreich do declarative apply too; `n8nctl`'s differentiator is that,
  combined with profiles, the **same directory promotes across instances**.
- **Manage instances from an AI agent.** `n8nctl mcp` runs the CLI as a Model
  Context Protocol server (73 annotated tools, reusing keyring auth and the active
  profile) and `n8nctl agent guard` generates host rules that hard-block
  destructive operations. See [the layer note](#a-note-on-mcp-and-agents).
- **Fleet operations:** `backup`/`restore` (git-friendly snapshots), `workflows
  search` (find by node type / credential / webhook path), cross-instance `sync`.
- **Richer output** (table / json / yaml / csv, `--columns`, `--jq` via full
  gojq, `-o id`), `--dry-run` that prints a redacted `curl`, and a signed,
  multi-channel distribution (Homebrew/Scoop/deb/rpm/apk + cosign + SBOM).

**Use it** if you operate **several instances**, want **no Node** and a signed
single binary, value **keyring** security and **production resilience**, keep
**workflows as code** across environments, drive n8n from an **AI agent**, or want
**backup / promote / search** across a fleet.

## Where the others are genuinely better than n8nctl

- **`@n8n/cli`** — first-party and tracks new endpoints first (it already has
  `package shared`, which `n8nctl` lacks); auto-JSON-on-pipe and auto-pagination
  are nicer scripting defaults; npm-native for Node projects.
- **`ubie-oss/n8n-cli`** — schema-aware linting (`node-params` + `node-schema`),
  `fmt`, `trace`, and the server-side `proxy` enforcement gateway make it the
  better tool for authoring quality. `n8nctl`'s lint roadmap tracks node-schema
  validation; the lint-enforcing proxy is a genuinely different idea worth knowing.
- **`edenreich/n8n-cli`** — if you want *only* GitOps sync and nothing else, it is
  smaller and has less to learn.

## A note on MCP and agents

`n8nctl`'s MCP server operates at a different layer from n8n's own MCP support, and
is complementary to it. **n8n the platform is already MCP-native at the workflow
layer** — the **MCP Server Trigger** node turns a workflow into an MCP server
(agents call your workflows as tools) and the **MCP Client Tool** node lets a
workflow consume external MCP tools. That is the *data plane*: automations exposed
as tools. `n8nctl mcp` is the *control plane*: it exposes the n8n **management
API** (list / create / activate / delete workflows, manage credentials, projects,
executions, **across instances**) so an agent can *operate and administer* a fleet.
None of the other CLIs expose their management surface over MCP.

## Performance

A single API call is **network-bound**: per-request time is dominated by the
instance's latency, not the CLI. The CLI's own processing is in the microseconds:

| Operation (50-record page, no network) | n8nctl |
|---|---|
| Decode a list page | ~50 µs |
| Render a table | ~150 µs |
| Render JSON | ~80 µs |
| Apply a `--jq` filter (full gojq) | ~90 µs |

The one place a CLI's own cost is visible is **process startup**, paid on every
invocation. The two Go binaries (`n8nctl`, edenreich) and the Bun binary (ubie)
start in single-digit milliseconds; the Node-based official CLI pays the Node +
oclif boot:

| | Startup (warm, `--version`) | Footprint |
|---|---|---|
| **n8nctl** | **~6–8 ms** | one 14 MB static binary, no runtime |
| @n8n/cli | ~150–160 ms | 5.6 MB package **+ a Node.js runtime** |

For a single command you will not notice it next to a network round-trip. It only
adds up when the CLI is invoked **many times** — a loop over 200 ids, an `xargs`
pipeline, or a CI job — where ~150 ms each becomes ~30 s of overhead versus ~1.5 s
for a fast-starting binary.

_Measured on Apple Silicon with `hyperfine --warmup 3 --runs 40`; Node startup has
high variance. Reproduce with `go test -bench=. ./internal/...` and
`hyperfine -N 'n8nctl version' 'n8n-cli --version'`._

## Which should you use?

- **`@n8n/cli`** — the official tool; single instance, Node-native, newest endpoints.
- **`ubie-oss/n8n-cli`** — authoring quality: schema-aware lint, formatting, and
  server-side lint enforcement.
- **`edenreich/n8n-cli`** — minimal GitOps sync, nothing else.
- **`n8nctl`** — multiple instances, no Node + a signed binary, keyring security,
  production resilience, workflows-as-code across environments, fleet
  backup/promote/search, and agent management (MCP + guard).

Many teams will reasonably use more than one — for example, `ubie` to lint and
author workflows, and `n8nctl` to operate the fleet, promote across instances, and
back it up.
