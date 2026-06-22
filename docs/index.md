# n8nctl

Control any n8n instance from your terminal. One static binary, many instances.

`n8nctl` is a fast, scriptable command-line client for the
[n8n public REST API](https://docs.n8n.io/api/). It manages workflows,
executions, credentials, tags, variables, projects, users, the audit report and
the source-control integration on any n8n instance, self-hosted or Cloud, over
HTTPS with an API key.

It is built in Go as a single static binary. There is no Node runtime to install
and nothing to keep updated except one file on your `PATH`.

!!! note "Unofficial"
    Not affiliated with n8n. Talks to the public API at `<your-host>/api/v1`
    using the `X-N8N-API-KEY` header.

## Why n8nctl

n8n ships an official `@n8n/cli` (Node, currently beta). `n8nctl` is a different
tool, and the differences matter if you operate more than one instance or script
against n8n in CI:

- **Multi-instance is first class.** Named profiles, one per instance, switched
  with `--profile`, `N8NCTL_PROFILE`, or `n8nctl config use <name>`. The official
  CLI targets a single instance at a time.
- **Secrets live in your OS keyring**, not in a plaintext config file.
- **A single binary, no Node.** One file to install, nothing to `npm install`.
- **Real output formats.** `table`, `json`, `yaml`, `csv`, with `--columns`.
- **Dry-run** prints the equivalent `curl` and sends no request.
- **Production-grade client** with retries, client-side rate limiting, and
  cursor pagination that walks every page with `--all`.

## Install

```bash
# Homebrew (macOS/Linux)
brew install jjuanrivvera/n8n-cli/n8nctl-cli

# From source
go install github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest
```

Prebuilt binaries are attached to each
[release](https://github.com/jjuanrivvera/n8n-cli/releases/latest).

## First steps

```bash
n8nctl init             # name a profile, capture the base URL and API key
n8nctl workflows list   # list workflows on the active instance
```

Continue with:

- [Getting started](getting-started.md)
- [Multi-instance and profiles](profiles.md)
- [Authentication](authentication.md)
- [Output and filtering](output.md)
- [Command reference](commands/index.md)
