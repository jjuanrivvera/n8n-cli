# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/) and this project adheres to
[Semantic Versioning](https://semver.org/).

## [Unreleased]

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
