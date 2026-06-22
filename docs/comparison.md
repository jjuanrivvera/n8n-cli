# n8nctl vs. the official n8n CLI

There are two command-line clients for the n8n public API:

- **`@n8n/cli`** (binary `n8n-cli`) — the **official, first-party** CLI from the
  n8n team. Node.js/oclif, distributed on npm. Currently labelled **beta**
  ("use it only for experimenting, local development, and personal projects").
- **`n8nctl`** (this project) — an independent, single-binary Go client focused
  on multi-instance use, production resilience, and fleet operations.

A side-by-side comparison. The official CLI is a capable, well-designed tool
maintained by the people who build the API itself; several nice ideas here (a
built-in `--jq`, an `id-only` output mode, a `skill install` command) were theirs
first, and we matched or extended them. Pick whichever fits your workflow — and if
you live in a Node project and work with a single instance, the official tool is
an excellent choice.

_Comparison as of `@n8n/cli` 0.7.0 and `n8nctl` 0.2.0 (June 2026). Both track the
same public API, so coverage converges over time._

## At a glance

| | n8nctl | @n8n/cli (official) |
|---|---|---|
| Vendor | Independent (third-party) | **First-party (n8n team)** |
| Runtime | **Single static Go binary, no deps** | Node.js + npm |
| Install | Homebrew, Scoop, `go install`, deb/rpm/apk, raw binary | `npm i -g @n8n/cli` / `npx` |
| Maturity | New, but versioned + released | **Official**, but **beta / not for production** |
| Tracks API changes | Community pace | **Day one (same team)** |

## Command coverage

As of this writing, **n8nctl covers every command the official CLI exposes**,
plus more. Parity across the shared surface:

| Area | Both have | n8nctl also adds |
|---|---|---|
| workflows | list, get, create, update, delete, activate, deactivate, transfer, tags | **archive/unarchive, sync (cross-instance), search** |
| executions | list, get, retry, stop, delete | — |
| credentials | create, get, list, schema, transfer | **update** |
| tags / variables | full CRUD | — |
| projects | create, get, list, update, delete, members, add/remove-member | **set-member-role** |
| users | get, list | **invite, change-role, delete** |
| data-tables | list, get, create, delete, rows, add/update/upsert/delete-rows | update |
| packages (beta) | export, import | — |
| audit | generate | **--days / --categories options** |
| source-control | pull | — |
| config / auth | config set-url/set-api-key/show, login, logout | **profiles: use, list-profiles, view** |
| skill | install (Claude/Cursor/Windsurf) | **+ codex/gemini/copilot/opencode, path, print** |

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
- **More meta tooling:** `doctor`, `init` wizard, `alias`, raw `api` escape hatch.
- A **full jq** engine (gojq) behind `--jq`, vs the official's simpler path
  filter; **7 skill targets** vs 3.

## Where the official CLI is genuinely better

Giving credit where it is due — reasons to prefer `@n8n/cli`:

- **First-party.** Built and maintained by the n8n team alongside the API, so it
  tracks new endpoints and changes immediately. A third-party CLI can lag.
- **Thoughtful scripting defaults.** It **auto-selects JSON when stdout is piped**
  (so `n8n-cli workflow list | jq` just works) and **auto-paginates lists by
  default**. n8nctl defaults to a table and uses `--all` to walk pages, which is
  more predictable but less convenient for piping. (Use `n8nctl ... -o json` or
  set `N8NCTL_OUTPUT=json` when scripting.)
- **npm-native.** If you already have Node and a JS toolchain, `npx @n8n/cli` is
  zero-friction and fits naturally in package scripts.
- **Official support and trust.** It is the tool the n8n docs point to.

## Performance

You were right to be skeptical of benchmarks here: **a single API call is
network-bound**, so per-request time is dominated by your instance's latency, not
the CLI. Our own processing is in the microseconds:

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
  or want **backup / promote / search** across a fleet — and you are comfortable
  with a community tool that follows the API rather than defining it.

Many teams will reasonably use both: the official CLI inside Node projects, and
n8nctl for ops, multi-instance promotion, and backups.
