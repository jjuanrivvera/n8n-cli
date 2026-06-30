<div align="center">

# n8nctl

Control any n8n instance from your terminal. One static binary, many instances.

[![CI](https://github.com/jjuanrivvera/n8n-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/jjuanrivvera/n8n-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/jjuanrivvera/n8n-cli)](https://github.com/jjuanrivvera/n8n-cli/releases/latest)
[![codecov](https://codecov.io/gh/jjuanrivvera/n8n-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/jjuanrivvera/n8n-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/jjuanrivvera/n8n-cli)](https://goreportcard.com/report/github.com/jjuanrivvera/n8n-cli)
[![Go Reference](https://pkg.go.dev/badge/github.com/jjuanrivvera/n8n-cli.svg)](https://pkg.go.dev/github.com/jjuanrivvera/n8n-cli)
[![Go version](https://img.shields.io/github/go-mod/go-version/jjuanrivvera/n8n-cli)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

<br>

<img src="assets/demo.gif" alt="n8nctl in action: list, lint, GitOps apply, search, and YAML backup" width="900">

</div>

`n8nctl` is a fast, scriptable command-line client for the
[n8n public REST API](https://docs.n8n.io/api/). It manages workflows,
executions, credentials, tags, variables, projects, users, the audit report and
source control on any n8n instance — self-hosted or Cloud — over HTTPS with an
API key.

Built in Go as a **single static binary**: no Node runtime, no `npm` tree,
nothing to keep updated but one file on your `PATH`. **Multi-instance by design**,
with secrets in your **OS keyring**.

> Unofficial. Not affiliated with n8n. Talks to the public API at
> `<your-host>/api/v1` using the `X-N8N-API-KEY` header.

> 🏭 Part of the [cliwright](https://github.com/jjuanrivvera/cliwright) CLI fleet.

## Why n8nctl

- **Multi-instance, first class.** Named profiles, one per instance — switch with
  `--profile`. API keys live in your **OS keyring**, never in a plaintext file.
- **One static binary, no Node.** `brew install`, `go install`, or a single
  download; starts in ~6 ms, painless on CI runners.
- **Workflows as code.** Declarative `apply` (with `--prune` / `--dry-run`),
  plus `lint`, `diff`, `convert`, and `backup` / `restore` / `sync` across
  instances.
- **Built for AI agents.** An MCP server (`mcp`), an `agent guard` that blocks
  destructive operations, and a lint-enforcing `proxy`.
- **Production-grade.** Idempotent-only retries with backoff, adaptive rate
  limiting, Ctrl-C cancellation, a `--dry-run` that prints the equivalent `curl`,
  and signed releases (cosign + SBOM).

n8n also ships an official `@n8n/cli`, and there are several community CLIs. For
an honest, lane-by-lane breakdown of all of them, see the
[**comparison guide**](https://jjuanrivvera.github.io/n8n-cli/comparison/).

## Install

```bash
# Homebrew (macOS/Linux) — installed as a cask
brew install jjuanrivvera/n8n-cli/n8nctl-cli
# or, tap first then install by name
brew tap jjuanrivvera/n8n-cli && brew install n8nctl-cli

# From source (needs Go 1.25+)
go install github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest

# Or build locally
make build && ./bin/n8nctl --help
```

Prebuilt binaries for macOS, Linux, and Windows (amd64/arm64), plus Linux
`.deb`/`.rpm`/`.apk` packages, are attached to each
[release](https://github.com/jjuanrivvera/n8n-cli/releases/latest). Release
archives and the Homebrew/Scoop installs ship **shell completions**
(bash/zsh/fish). Releases are **SBOM-attested and the checksums are signed** with
[cosign](https://github.com/sigstore/cosign) (keyless); see
[RELEASING.md](RELEASING.md) for verification.

## Quickstart

Create an API key in the n8n UI under **Settings → n8n API**, then:

```bash
# Friendliest first run: names a profile, captures the base URL and API key
# (stored in your OS keyring), verifies connectivity, and writes the config.
n8nctl init

n8nctl workflows list           # list workflows as a table
n8nctl workflows get 42 -o json # one workflow as JSON
n8nctl auth status              # confirm which instance you're on
```

Prefer to script it? `n8nctl init --profile homelab --base-url https://n8n.lan/api/v1 --api-key "$KEY"`.

## Features at a glance

| Area | What you get | Docs |
|---|---|---|
| **Resources** | Full CRUD across workflows, executions, credentials, tags, variables, projects, users, data-tables, packages, audit, and source-control — plus verbs like `activate`, `transfer`, `retry`, `invite`, `schema`. | [Features](https://jjuanrivvera.github.io/n8n-cli/features/) |
| **Multi-instance & secure** | Named profiles, OS keyring for secrets, precedence flag > env > config > default. | [Profiles](https://jjuanrivvera.github.io/n8n-cli/profiles/) |
| **Workflows as code** | `apply --dir` (prune / dry-run), `lint`, `autofix`, `diff`, `convert` (JSON↔YAML + code externalization). | [Workflows as Code](https://jjuanrivvera.github.io/n8n-cli/workflows-as-code/) |
| **Node catalog** | `nodes list / search / show` over an embedded catalog of 560+ real n8n nodes — the same data powering the `unknown-node-type` / `unknown-parameter` lint rules. | [Features](https://jjuanrivvera.github.io/n8n-cli/features/) |
| **Fleet operations** | `sync` (promote across instances), `backup` / `restore`, `search`, `stats`, `executions prune` / `watch`, `workflows bulk --tag`. | [Beyond the API](https://jjuanrivvera.github.io/n8n-cli/beyond-api/) |
| **AI agents** | `mcp` server, `agent guard`, lint-enforcing `proxy`. | [MCP](https://jjuanrivvera.github.io/n8n-cli/mcp/) · [Guard](https://jjuanrivvera.github.io/n8n-cli/agent-guard/) · [Proxy](https://jjuanrivvera.github.io/n8n-cli/proxy/) |
| **Output & scripting** | table / json / yaml / csv / `-o id`, `--jq` (full gojq), `--columns`, `--dry-run`, raw `api`. | [Output](https://jjuanrivvera.github.io/n8n-cli/output/) |
| **Resilience** | Idempotent retry + backoff, adaptive rate limiting, Ctrl-C cancellation, typed errors with hints. | [Features](https://jjuanrivvera.github.io/n8n-cli/features/) |

A short tour:

```bash
n8nctl workflows apply --dir ./workflows --dry-run    # GitOps: preview a reconcile
n8nctl workflows sync 2tUt1wbLX --from dev --to prod   # promote a workflow across instances
n8nctl backup --out ./backup --format yaml            # git-friendly instance snapshot
n8nctl workflows search --node slack -o json          # find workflows by what's inside them
n8nctl mcp start                                      # expose n8n to an AI agent (MCP)
n8nctl proxy                                          # lint-gate every workflow write
```

The [**feature reference**](https://jjuanrivvera.github.io/n8n-cli/features/) lists
every command, and the [**recipes**](https://jjuanrivvera.github.io/n8n-cli/recipes/)
are copy-pasteable tasks.

## Drive it from an AI agent

`n8nctl` ships an **agent skill** that teaches AI coding agents (Claude Code,
Cursor, Codex, Gemini, …) how to use it. Install it as a Claude Code plugin:

```text
/plugin marketplace add jjuanrivvera/n8n-cli
/plugin install n8nctl-cli@n8nctl
```

For programmatic access, `n8nctl mcp` exposes the command tree to any MCP host,
`n8nctl agent guard` writes host rules that block destructive operations, and
`n8nctl proxy` lint-gates every workflow write. See
[MCP server](https://jjuanrivvera.github.io/n8n-cli/mcp/),
[Agent guard](https://jjuanrivvera.github.io/n8n-cli/agent-guard/), and
[Lint-enforcing proxy](https://jjuanrivvera.github.io/n8n-cli/proxy/).

## Documentation

Full documentation lives at **<https://jjuanrivvera.github.io/n8n-cli/>**:

- [Features](https://jjuanrivvera.github.io/n8n-cli/features/) — the complete capability reference
- [Getting started](https://jjuanrivvera.github.io/n8n-cli/getting-started/) · [Recipes](https://jjuanrivvera.github.io/n8n-cli/recipes/)
- [Multi-instance & profiles](https://jjuanrivvera.github.io/n8n-cli/profiles/) · [Authentication](https://jjuanrivvera.github.io/n8n-cli/authentication/) · [Output & filtering](https://jjuanrivvera.github.io/n8n-cli/output/)
- [Beyond the API](https://jjuanrivvera.github.io/n8n-cli/beyond-api/) · [Workflows as code](https://jjuanrivvera.github.io/n8n-cli/workflows-as-code/)
- [MCP server](https://jjuanrivvera.github.io/n8n-cli/mcp/) · [Agent guard](https://jjuanrivvera.github.io/n8n-cli/agent-guard/) · [Lint-enforcing proxy](https://jjuanrivvera.github.io/n8n-cli/proxy/)
- [vs. other n8n CLIs](https://jjuanrivvera.github.io/n8n-cli/comparison/) · [Command reference](https://jjuanrivvera.github.io/n8n-cli/commands/)

## Development

```bash
make dev            # fmt + vet + build
make test           # run tests
make check          # full local quality gate (fmt + vet + lint + test)
make docs-serve     # preview the docs site
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the architecture and contribution
workflow.

## License

MIT — see [LICENSE](LICENSE).
