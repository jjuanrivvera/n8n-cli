# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/) and this project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.5.3] - 2026-07-11

### Added
- One-line install script for macOS/Linux (checksum-verified):
  `curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/n8n-cli/main/install.sh | sh`.

### Security
- Build with Go 1.25.12 to clear GO-2026-5856 (privacy leak in `crypto/tls`
  Encrypted Client Hello).

## [0.5.2] - 2026-07-02

### Fixed
- **`agent guard` now generates a PreToolUse enforcement hook** (Bash + MCP
  matchers); previously it emitted only permission rules, which are bypassed by
  path-invoked binaries, obfuscation, and command chaining. The hook anchors
  blocked subcommands, matches `./bin/n8nctl` and `/usr/local/bin/n8nctl` while
  ignoring a different binary ending in `n8nctl`, and blocks `api
  DELETE/PUT/POST/PATCH` at the method position.
- **`packages import` is now gated.** It was added without annotations and so
  classified as a local/utility command and silently allowed, yet it creates
  workflows and credentials on the instance; it now requires approval, locked by
  a test asserting every API command is annotated.
- `agent guard` now emits every cobra-alias spelling (`wf delete`, `exec prune`,
  …), replaces the non-functional regex MCP permission rules with exact tool
  names, and its no-jq hook fallback no longer fails open.

## [0.5.1] - 2026-06-23

### Fixed
- Corrected stale counts in the embedded agent skill and MCP docs: the lint rule
  set is 8 (including `invalid-parameter-value`) and the MCP server now exposes 85
  tools (after the v0.4.0/0.5.0 commands). No behavior change.

## [0.5.0] - 2026-06-23

### Added
- **`invalid-parameter-value` lint rule** — validates `options`/`multiOptions`
  parameter *values* (not just names) against the node's allowed set, resolved
  through each node's `displayOptions` so the active option set is used (a Slack
  node's `operation` options differ by `resource`). A warning, with a "did you
  mean …?" hint; conservative — it skips dynamic option lists, expression values,
  and any parameter whose active options it cannot resolve, so it does not
  false-positive on a valid workflow.
- **`workflows breaking-changes`** — reports nodes pinned to an older
  `typeVersion` than the catalog's latest, plus parameters they use that the
  catalog does not recognize for the node. Works on files, a directory, a live
  id, or `--remote`. Informational (exits 0).
- **`templates search | get | deploy`** — browse and deploy workflows from the
  public n8n template gallery (`api.n8n.io`, no API key). `deploy` creates the
  workflow on the active instance (credentials not included) and honors
  `--dry-run` / `--activate`.

### Changed
- The embedded node catalog now carries each parameter's full schema (type,
  option values, required, `displayOptions`) and each node's latest typeVersion,
  and merges n8n's per-version node entries so a parameter that exists only in a
  newer version is no longer mis-flagged by `unknown-parameter`.
- `comparison.md`: corrected — `@n8n/cli` has no `package shared` command (its
  package topic is export/import, which `n8nctl` already matches); `n8nctl` now
  matches the workflow-intelligence feature set (autofix, breaking-changes,
  nodes, templates).

## [0.4.0] - 2026-06-23

### Added
- **`nodes` explorer** — `nodes list` / `search <query>` / `show <type>` browse the
  embedded catalog of n8n node definitions offline (no API call), to find a node's
  exact `type` string and its parameters.
- **`workflows autofix`** — auto-repair the mechanical mistakes the linter detects:
  correct typo'd node types against the catalog (e.g. `…slak` → `…slack`), add the
  leading `=` to expression strings, and generate a webhookId for webhook/form
  nodes that lack one. Report-only by default; `--write` applies. Works on `-f`
  files or a `--dir`.
- **`executions prune`** — bulk-delete executions by `--older-than 30d` and/or
  `--status`, to reclaim database space; previews the count, `--yes` skips the
  confirmation. Hard-blocked by `agent guard` (irreversible).
- **`executions watch`** — live-tail new executions, coloring failures, until
  Ctrl-C. Excluded from the MCP surface (a blocking poll).
- **`workflows bulk activate|deactivate --tag <name>`** — flip every workflow
  carrying a tag in one command (maintenance windows); dry-run + confirm.
- **`stats`** — one-shot instance summary (workflow counts + recent-execution
  status mix), degrading gracefully on Community-edition 403s.
- **`proxy --reject-duplicate-names`** — also reject creating a workflow whose
  name already exists on the instance.
- cobra `Example:` blocks inherited by every generic resource command, and a
  scheduled workflow that refreshes the embedded node catalog against the latest
  n8n packages.

### Changed
- The embedded node catalog now keeps only top-level parameter names (cleaner
  `nodes show` output and a smaller embed); the `unknown-parameter` rule is
  unchanged in behavior.
- A git pre-commit hook (`make setup-hooks`) runs gofmt/vet/lint/tests and is now
  installed by default, skipping the Go gate on docs-only commits.

## [0.3.0] - 2026-06-23

### Added
- **Node-schema-aware linting**: two new `workflows lint` rules grounded in an
  embedded catalog of n8n's real node definitions (n8n-nodes-base + langchain,
  560+ node types) — `unknown-node-type` (a node type from a known package that
  isn't real, with a "did you mean …?" suggestion; community/custom nodes are
  skipped) and `unknown-parameter` (a parameter the node type doesn't define).
  The catalog is generated by `make gen-node-schemas` and embedded at build time;
  these rules also apply through `workflows lint --remote` and the `proxy` gate.
- **Lint-enforcing proxy**: `n8nctl proxy` runs a local reverse proxy in front of
  the active instance that **lints every workflow create/update and rejects errors
  with HTTP 422** before they reach n8n — so a workflow with lint errors can never
  land, regardless of who pushes it (a human, a script, or an AI agent). Reads pass
  through; the active profile's API key is injected from the keyring so clients
  forward without the secret. `--disable-rule` tunes the rules and
  `--block-destructive` also rejects workflow DELETEs. (Inspired by the server-side
  enforcement proxy in `ubie-oss/n8n-cli`.)
- **MCP server**: `n8nctl mcp` runs the CLI as a Model Context Protocol server so
  AI agents (Claude Code/Desktop, Cursor, VS Code) drive any n8n instance.
  `mcp start` (stdio), `mcp stream` (HTTP), `mcp tools` (export the catalog), and
  installers `mcp claude` / `mcp cursor` / `mcp vscode` (each `enable`/`disable`/
  `list`). It auto-exposes the command tree as **73 MCP tools** prefixed `n8n_`
  (e.g. `n8n_workflows_list`), each replaying the matching cobra command with the
  same keyring auth, active profile, and `--dry-run`. Tools carry read-only /
  write / destructive annotations so hosts gate writes; setup and secret commands
  (`auth`, `config`, `alias`, `init`, `skills`, `agent`, `doctor`) and the
  `--api-key` / `--show-token` / `--profile` / `--base-url` flags are never
  exposed. Built on `github.com/njayp/ophis` (wrapping the official
  `modelcontextprotocol/go-sdk`).
- **Agent guard**: `n8nctl agent guard --host <claude-code|codex|opencode>`
  generates host-level safety config — derived from the live command tree and MCP
  annotations — that hard-blocks destructive operations (`delete`, `delete-rows`),
  makes ordinary writes require approval, and lets reads run free. `--all-writes`
  blocks writes too; `--write` installs the files without overwriting existing
  ones. Emits a `.claude/hooks` PreToolUse hook + `.claude/settings.json` for
  Claude Code, a read-only-sandbox `~/.codex/config.toml` for Codex, and
  `opencode.json` rules for OpenCode.

### Security
- Bump `modelcontextprotocol/go-sdk` (transitive, via the MCP server) to v1.4.1,
  resolving GO-2026-4569/4770/4773. Externalized-file path confinement now also
  rejects volume-rooted paths on Windows.
- **Path traversal in externalized code files.** A crafted workflow file's `$ref`
  could point outside its directory (e.g. `../../../etc/passwd`), making `restore`,
  `workflows apply`, `lint --dir`, and `diff` read arbitrary local files — and on
  apply/restore upload them to the configured instance. Externalized-file loading
  is now confined to the workflow's directory (absolute and escaping paths are
  refused). The externalization marker also changed from `$ref` to the namespaced
  `$n8nctl_file` so it cannot collide with a legitimate `$ref` parameter
  (re-externalize any 0.2.0 backups).

### Fixed
- `workflows apply` now detects duplicate workflow names on the instance and skips
  them, instead of silently updating/pruning an arbitrary one (n8n does not enforce
  unique names). The plan/summary reports skipped workflows.
- Multipart uploads now share the adaptive rate limiter's 429 throttle/recovery,
  matching regular requests.
- Retry backoff no longer dereferences a nil jitter source (dead nil-guard), and
  context cancellation during a retry backoff now returns immediately.
- `lint` `expression-prefix` recurses into nested parameters and only flags genuine
  n8n expressions, removing false positives on plain `{{ }}` text.
- `--all` truncation warning prints the real page cap; `--externalize N` now means
  "longer than N lines" as documented; table output measures width by rune, not byte.

## [0.2.0]

### Added
- **Data tables**: `data-tables` list/get/create/update/delete plus
  `rows`/`add-rows`/`update-rows`/`upsert-rows`/`delete-rows`.
- **Packages (beta)**: `packages export` / `packages import` (.n8np), including a
  multipart upload path with `--dry-run` curl support.
- **`--jq`** global flag backed by a full jq engine (gojq).
- **`id` / `id-only` output format** and **`--no-header`** for table output.
- **`skills install` / `path` / `print`** — install the bundled agent skill into
  Claude, Cursor, Windsurf, Codex, Gemini, Copilot, or opencode (also installable
  with `npx skills add jjuanrivvera/n8n-cli`).
- Command-surface parity with the official CLI: top-level `login` / `logout`,
  `config set-url` / `set-api-key` / `show`, and a distinct exit code (2) for auth
  failures.
- HTTP request/response tracing at `-v`/debug level.
- A `comparison` docs page (n8nctl vs. the official `@n8n/cli`) with benchmarks.
- **Workflows as code / GitOps**:
  - `workflows apply --dir <dir>` — declarative reconcile of a directory of
    workflow files (JSON/YAML) into an instance: create, update, skip-unchanged,
    and `--prune` to delete drift, with `--dry-run` preview and `--activate`.
    Combine with profiles to promote the same directory across instances.
  - `workflows lint` — static checks over files (`--dir`/`-f`) or live workflows
    (`--remote`) with 5 grounded rules (`--list-rules`); exits non-zero on errors
    as a CI gate, `--disable-rule` and `-o json` supported.
  - `workflows convert <file…> --to json|yaml` — convert workflow files between
    JSON and YAML, with `--externalize <N>` to split long node code fields into
    sibling `$ref` files.
  - `workflows diff <id>` — unified diff of a workflow's writable content against
    another `--profile` or a local `--file`.
  - YAML and code externalization in `backup` (`--format json|yaml`,
    `--externalize <N>`); `restore` reads either format and re-inlines `$ref` code.
  - A `workflows-as-code` docs page documenting the GitOps loop.

## [0.1.0]

### Added
- Initial release of `n8nctl`, a command-line interface for the n8n workflow
  automation API.

[Unreleased]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jjuanrivvera/n8n-cli/releases/tag/v0.1.0
