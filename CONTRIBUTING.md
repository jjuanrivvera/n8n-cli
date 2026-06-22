# Contributing to n8n-cli

Thanks for your interest in improving `n8nctl`!

## Development setup

```bash
git clone https://github.com/jjuanrivvera/n8n-cli
cd n8n-cli
make setup-hooks   # install the pre-commit hook (.githooks)
make dev           # fmt + vet + build
make check         # fmt + vet + lint + test (the full local gate)
```

Requires Go 1.25+ (the version is pinned in `go.mod` and read by CI via
`go-version-file`). Linting uses `golangci-lint` **v2** (`go install
github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`).

Run `make check` before pushing; CI must be green.

## Architecture

See the [AGENTS.md](AGENTS.md) architecture section. The key idea: each n8n
resource is a thin typed wrapper over a generic core. The generic client
(`internal/api`) handles HTTP, the `X-N8N-API-KEY` header, base-URL
normalization (the trailing `/api/v1` is appended for you), retries, client-side
rate limiting, cursor pagination, dry-run, and output. Resources only declare
their types, columns, list filters, and any custom verbs.

### Adding a resource

Adding a resource is **three small additions and zero edits to shared code**:

1. `internal/api/<resource>.go` — the typed struct(s) and a `Client` accessor:

   ```go
   package api

   type Widget struct {
       ID   ID     `json:"id,omitempty"`
       Name string `json:"name,omitempty"`
   }

   func (c *Client) Widgets() *Resource[Widget] {
       return NewResource[Widget](c, "widgets")
   }
   ```

2. `commands/<resource>.go` — register it (it self-attaches via `init()`):

   ```go
   package commands

   import "github.com/jjuanrivvera/n8n-cli/internal/api"

   func init() {
       registerResource(resourceSpec[api.Widget]{
           Use:     "widgets",
           Short:   "Manage widgets",
           New:     func(c *api.Client) *api.Resource[api.Widget] { return c.Widgets() },
           Columns: []string{"id", "name"},
       })
   }
   ```

   The generic core builds `list/get/create/update/delete` from the spec.
   Capability flags (`NoCreate`/`NoUpdate`/`NoDelete`/`NoGet`) drop verbs the API
   does not support; custom verbs (activate, retry, transfer, add-member) attach
   through the spec's `Extra` hook.

3. `internal/api/<resource>_test.go` — an httptest-based service test. Test what
   is **unique** to the resource — special field types, custom actions, odd
   response shapes — not the generic CRUD plumbing, which is already covered once
   by the `Resource[T]` tests. A List/Get happy-path pair adds volume, not signal.

Use the tolerant JSON types in `internal/api/types.go` (`api.ID`, `api.Int`,
`api.Bool`, `api.StringOrSlice`) — n8n's API is loose about encodings. Unknown
JSON fields are ignored, so structs need not be exhaustive.

If you change the command tree (new resource, flag, or help text), regenerate the
command reference with `make docs-gen` and commit the result.

## Commits & branches

- **Branch from `develop`** (not `main`) and open PRs against `develop`. Use a
  type-prefixed branch name matching the change: `feature/...`, `fix/...`,
  `docs/...`, `test/...`, `chore/...`. Releases are tagged on `main`.
- Write [Conventional Commits](https://www.conventionalcommits.org/), e.g.
  `feat(workflows): add sync --update-by-name` or `fix(output): rune-safe
  truncation`. The [CHANGELOG](CHANGELOG.md) follows
  [Keep a Changelog](https://keepachangelog.com/), and GoReleaser groups the
  release notes by commit type.

## Tests

- Service tests spin up an `httptest.NewServer` and point a client at it.
- Prefer `require` for fatal assertions, `assert` for the rest.
- Keep coverage healthy — CI gates total coverage at **80%**; new code should
  ship tests. Run `make cover-check` locally to check the gate.
- **Test failure paths, not just happy paths.** Every parse of external state
  (API bodies, config files, backups) needs a test with corrupt input; every
  batch operation (sync, backup, restore) needs a partial-failure test asserting
  counts and a non-zero exit. Coverage measures execution, not assertion quality
  — a swallowed error can be "covered" and still hide a bug.

## Reporting bugs & security issues

- Bugs and feature requests: open an issue (templates guide the details we need).
- **Security vulnerabilities**: do **not** open a public issue — see
  [SECURITY.md](SECURITY.md).
- By participating you agree to our [Code of Conduct](CODE_OF_CONDUCT.md).
</content>
