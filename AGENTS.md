# AGENTS.md

Guidance for AI agents (Claude Code, Cursor, Copilot, …) working in this repo.

## What this is

`n8n-cli` is a Go CLI (`n8nctl`) for the n8n public REST API
(`<host>/api/v1`), built with Cobra. It is a portable **single static binary**:
no Node runtime, Homebrew- and `go install`-able, with API keys in the OS
keyring and **multi-instance named profiles**. The architecture is a generic
typed client plus one thin file per resource.

## Commands

```bash
make build        # build to bin/n8nctl
make dev          # fmt + vet + build
make test         # go test ./...
make lint         # golangci-lint
make check        # fmt + vet + lint + test (the full local gate)
make docs-gen     # regenerate docs/commands from the cobra tree
make run ARGS="workflows list"   # build and run
go run ./cmd/n8nctl <args>
```

Run `make check` before proposing a change.

## Architecture

```
cmd/n8nctl/main.go        entry point
commands/
  root.go                 global flags, getAPIClient(), render()
  generic.go              generic CRUD command builders (resourceSpec[T], registerResource)
  <resource>.go           one per resource; self-registers via init()
internal/
  api/
    client.go             auth (X-N8N-API-KEY), base-URL /api/v1 normalize, retries, rate limit, dry-run
    resource.go           generic Resource[T]: List/Get/Create/Update/Delete/Action
    pagination.go         cursor pagination ({data, nextCursor} envelope)
    types.go              ID, Int, Bool, StringOrSlice flexible JSON types
    <resource>.go         one per resource: struct(s) + Client accessor
  config/                 YAML profiles + env overrides
  auth/                   OS keyring token storage
  output/                 table/json/yaml/csv rendering
  version/                build metadata (ldflags)
tools/gendocs/            command reference generator
```

## Key patterns

- **Resource = api type + accessor + register.** Adding a resource is three
  small additions and **zero edits to shared code**:
  1. an API struct in `internal/api/<resource>.go`,
  2. a `Client` accessor - `func (c *Client) Foos() *Resource[Foo] {
     return NewResource[Foo](c, "foos") }`,
  3. a `registerResource(resourceSpec[api.Foo]{…})` call in
     `commands/<resource>.go`'s `init()`.
  The generic core builds `list/get/create/update/delete` from the spec; custom
  verbs (activate, retry, transfer, add-member) attach via the spec's `Extra`
  hook.
- **Generic core.** `internal/api/resource.go` and `commands/generic.go` provide
  CRUD, pagination, and body parsing; resources only declare types, columns,
  list filters, capability flags (`NoCreate`/`NoUpdate`/`NoDelete`/`NoGet`), and
  custom actions.
- **Flexible JSON types.** n8n's API is loose about encodings, so use the
  tolerant types in `internal/api/types.go`: `api.ID` (string, but accepts a
  number/null), `api.Int` (accepts numeric strings), `api.Bool` (accepts
  `"true"`/`1`/…), `api.StringOrSlice` (one string or an array). Unknown JSON
  fields are ignored, so structs need not be exhaustive.
- **Auth.** Header `X-N8N-API-KEY`. Keys live in the OS keyring per profile;
  base URLs in `~/.n8nctl-cli/config.yaml`. The base URL is normalized - the
  trailing `/api/v1` is appended automatically. Env overrides: `N8NCTL_PROFILE`,
  `N8NCTL_BASE_URL`, `N8NCTL_API_KEY`, `N8NCTL_OUTPUT`, `N8NCTL_RPS`,
  `N8NCTL_LOG_LEVEL`, `N8NCTL_CONFIG`.
- **Multi-instance.** Profiles are the headline design choice: one binary drives
  many instances. Every command takes `--profile`; the default is set by
  `n8nctl config use <name>`. Keys are keyed by profile name in the keyring, so
  switching instances never crosses credentials.
- **Pagination.** n8n uses opaque cursor pagination (`{data, nextCursor}`).
  `--all` walks every page via `Resource.ListAll`; `--cursor` continues from a
  token; `--max-pages` caps `--all`.
- **Dry-run.** `--dry-run` prints the equivalent curl (key redacted unless
  `--show-token`) and sends no request.

## Config / auth / output layers

- **config** - `internal/config` loads `~/.n8nctl-cli/config.yaml` (or
  `$XDG_CONFIG_HOME/n8nctl-cli/`, or `N8NCTL_CONFIG`), resolves the active
  profile, and merges env overrides. Precedence everywhere is
  **flag > env > file > default**.
- **auth** - `internal/auth` stores/reads the per-profile API key in the OS
  keyring; the config file holds only non-secret fields.
- **output** - `internal/output` renders any value as `table|json|yaml|csv`,
  honoring `--columns`, `--no-color`, and TTY detection.

## Testing

Service tests spin up an `httptest.NewServer` and point a client at it. Use
`require` for fatal checks, `assert` otherwise. Run `make test` (or
`make test-race` for the race detector).

## Conventions

- Comments explain WHY, not WHAT.
- `gofmt -s` clean; pass `golangci-lint`.
- Never commit credentials. Keys belong in the keyring or env, never in code,
  config-in-repo, or commit messages.
- n8n public API reference: https://docs.n8n.io/api/api-reference/
