# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/) and this project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Security
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

[Unreleased]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/jjuanrivvera/n8n-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/jjuanrivvera/n8n-cli/releases/tag/v0.1.0
