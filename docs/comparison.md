# n8nctl vs other n8n CLIs

Several command-line clients exist for n8n, built by different people for
different jobs. This page maps the landscape honestly and shows where each tool
fits — including where the others are a better choice than `n8nctl`.

_As of June 2026: `n8nctl` 0.3.0, `@n8n/cli` 0.8.0, `ubie-oss/n8n-cli` 2.2.1,
`yigitkonur/n8n-cli` 1.9.3, `edenreich/n8n-cli` 0.7.1. They all talk to the same
n8n public API, so the CRUD surface converges over time; the real differences are
in focus, runtime, and the layers each tool adds on top._

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
- **`yigitkonur/n8n-cli`** — an independent **Node** CLI focused on **workflow
  intelligence**. It bundles n8n's own packages (`n8n-nodes-base`, `n8n-workflow`)
  and a SQLite node catalog to validate and **auto-repair** workflows offline,
  explore the node library, detect cross-version breaking changes, and deploy from
  the n8n template gallery.
- **`edenreich/n8n-cli`** — an independent **Go** binary focused on one job:
  **GitOps sync** of workflows from a directory to an instance.

These tools are not all trying to be the same thing. The useful question is not
"which is best" but "which lane is yours."

## At a glance

| | n8nctl | @n8n/cli | ubie-oss | yigitkonur | edenreich |
|---|---|---|---|---|---|
| Vendor | 3rd-party | **first-party** | 3rd-party | 3rd-party | 3rd-party |
| Runtime | **Go static binary** | Node + npm | Bun binary | Node + npm | Go binary |
| Primary lane | **ops + GitOps + agent mgmt** | first-party general | **authoring quality** | **workflow intelligence** | GitOps sync |
| CRUD breadth | full + data-tables + packages + `nodes` + `templates` | full + data-tables + packages | full + data-tables | full + `nodes` + `templates` | workflows only |
| Multi-instance | **profiles + keyring** | single, plaintext | single, env | multi-profile, plaintext | single, `.env` |
| Apply from files | **`apply`** (prune, cross-instance) | — | `apply` (+ git-changes) | import / export | `sync` (+ prune) |
| Validation | **node-schema lint (type + params + values) + autofix + breaking-changes + `proxy` gate** | — | **node-schema lint + `proxy` gate** | **offline validate + autofix + breaking-changes** | — |
| Templates / node catalog | — | — | `node-schema` dump | **templates + node catalog (FTS5)** | — |
| Agent tooling | **`mcp` + `agent guard`** | — | — | — | — |
| backup / restore / sync / search | **all four** | — | — | export/import + version history | `sync` |
| Secrets | **OS keyring** | plaintext | plaintext | plaintext | plaintext |
| Distribution | **brew / scoop / deb-rpm-apk + cosign + SBOM** | npm | source / cross-compile | npm | curl / `go install` |

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

### Workflow authoring & quality — `ubie-oss/n8n-cli`

The deepest authoring tool, and ahead of `n8nctl` on workflow *quality*. Its
`lint` ships rules including **`node-params`, which validates node parameter
values against the schema**, backed by a **`node-schema`** command that dumps the
real node-type schemas. (`n8nctl` now validates node types and parameter *names*
against an embedded catalog too, but ubie checks parameter *values* and types more
deeply.) `fmt` tidies node layout, `trace` analyses data-flow cardinality,
and **`proxy`** is a distinctive idea ubie pioneered: an HTTP proxy in front of the
n8n API that enforces lint **server-side**, so any push that fails lint — from a
human or an AI agent — is rejected with a 422, making quality structural rather
than a convention. (`n8nctl` has since adopted the same pattern in `n8nctl proxy`;
ubie's remains deeper — more rules, schema-aware validation, and duplicate-name
rejection.) It is single-instance (env vars, no keyring) and has no MCP or
agent-management layer. **Use it** if your priority is keeping a team's workflow
definitions clean, schema-valid, and consistently formatted.

### Workflow intelligence — `yigitkonur/n8n-cli`

The most workflow-aware tool of the five, and the deepest on *correctness*. It
bundles n8n's real packages (`n8n-nodes-base`, `n8n-workflow`) and a SQLite,
FTS5-indexed catalog of every node, which unlocks things no other CLI here does:
**offline `workflows validate`** (check a workflow against the actual node
definitions with no running instance), **`workflows autofix`** (auto-repair
expression prefixes, missing webhook paths, Switch v3 conditions, and node-type
typos, applied confidence-filtered), a **`nodes`** explorer (`list`, `search`,
`show`, and `breaking-changes --from <v> --to <v>` for cross-version
compatibility), **`templates`** search/get with `workflows deploy-template`, plus
surgical partial edits, local version history with rollback, `diff`, `trigger`,
`bulk`, and `health`. `n8nctl` now does node-type and parameter-name validation
against an embedded catalog, but this tool goes further — schema-accurate value
validation **and automatic repair**.

It is single-purpose in the other direction: no MCP, no agent guard, no GitOps
reconcile with prune, no cross-instance sync, no OS keyring (credentials sit in a
plaintext `.n8nrc.json`), and it needs a Node runtime (npm-only, no signed
binary). **Use it** if your priority is validating, repairing, and authoring
correct workflows against real node schemas — it is the strongest tool here for
that.

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
- **Lint enforcement at the boundary.** `n8nctl proxy` (adopting ubie's idea)
  fronts the instance and rejects any workflow create/update that fails lint with a
  422, so bad definitions can't land regardless of who pushes them.
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

- **`@n8n/cli`** — first-party and tracks new endpoints first; auto-JSON-on-pipe
  and auto-pagination are nicer scripting defaults; npm-native for Node projects.
- **`ubie-oss/n8n-cli`** — schema-aware linting (`node-params` + `node-schema`),
  `fmt`, and `trace` make it strong for authoring quality. `n8nctl` adopted
  ubie's server-side lint-enforcement `proxy` idea (`n8nctl proxy`, now with
  duplicate-name rejection); ubie still ships `fmt` and `trace`, which `n8nctl`
  does not.
- **`yigitkonur/n8n-cli`** — a deep validation engine: node-type and parameter
  checks, **autofix**, cross-version **breaking-change** detection, a searchable
  node catalog, and template deployment. `n8nctl` now matches each of these
  (`workflows autofix`, `workflows breaking-changes`, `nodes`, `templates`); the
  two are closely comparable on workflow intelligence.
- `n8nctl` does **node-schema-aware linting** validated against an embedded
  catalog of n8n's real node definitions: `unknown-node-type` and
  `unknown-parameter`, plus `invalid-parameter-value` (option values resolved
  through each node's `displayOptions`). Combined with `autofix`,
  `breaking-changes`, `nodes`, and `templates`, the workflow-intelligence gap
  that `ubie` and `yigitkonur` once held is largely closed; ubie's remaining edge
  is `fmt`/`trace`.
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
invocation. The numbers below are measured (`hyperfine`, warm, 40 runs of
`--help` on an Apple-silicon Mac); the two Go binaries start ~20–40× faster than
anything running a JavaScript runtime:

| CLI | Runtime | Startup (warm) | Footprint |
|---|---|---|---|
| **n8nctl** | Go static | **5.5 ms** | one 15 MB binary, no runtime |
| edenreich | Go static | 4.6 ms | one 11 MB binary, no runtime |
| @n8n/cli | Node | 108 ms | 5.7 MB `node_modules` **+ a Node runtime** |
| yigitkonur | Node | 136 ms | **514 MB `node_modules`** + a Node runtime |
| ubie | Bun compiled | 197 ms | one 59 MB binary (Bun runtime embedded) |

Two findings stand out. A **Bun-compiled binary is not a fast-start binary**: ubie
is the slowest of the five, because the single file still boots the embedded Bun
runtime and its bundled modules on every call. And `yigitkonur` depends on
`isolated-vm` (a native V8 module, for sandboxed expression evaluation) which
**fails to compile on current Node** and pulls a 514 MB `node_modules` — a
portability cost a static Go binary does not have.

For a single command none of this matters next to a network round-trip. It adds
up when the CLI is invoked **many times** — a loop over 200 ids, an `xargs`
pipeline, or a CI job — where ~150 ms each becomes ~30 s of overhead versus ~1 s
for a fast-starting binary.

_Measured on Apple Silicon with `hyperfine --warmup 3 --runs 40`; Node startup has
high variance. Reproduce with `go test -bench=. ./internal/...` and
`hyperfine -N 'n8nctl version' 'n8n-cli --version'`._

## Which should you use?

- **`@n8n/cli`** — the official tool; single instance, Node-native, newest endpoints.
- **`ubie-oss/n8n-cli`** — authoring quality: schema-aware lint, formatting, and
  server-side lint enforcement.
- **`yigitkonur/n8n-cli`** — workflow correctness: offline validation, autofix,
  node catalog, breaking-change detection, and template deployment.
- **`edenreich/n8n-cli`** — minimal GitOps sync, nothing else.
- **`n8nctl`** — multiple instances, no Node + a signed binary, keyring security,
  production resilience, workflows-as-code across environments, fleet
  backup/promote/search, and agent management (MCP + guard).

Many teams will reasonably use more than one — for example, `yigitkonur` or `ubie`
to validate and author workflows, and `n8nctl` to operate the fleet, promote
across instances, manage it from an agent, and back it up.
