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

</div>

`n8nctl` is a fast, scriptable command-line client for the
[n8n public REST API](https://docs.n8n.io/api/). It manages workflows,
executions, credentials, tags, variables, projects, users, the audit report and
the source-control integration on any n8n instance, self-hosted or Cloud, over
HTTPS with an API key.

It is built in Go as a single static binary. There is no Node runtime to install,
no `npm` tree to resolve, and nothing to keep updated except one file on your
`PATH`.

> Unofficial. Not affiliated with n8n. Talks to the public API at
> `<your-host>/api/v1` using the `X-N8N-API-KEY` header.

## Why n8nctl vs the official `@n8n/cli`

n8n ships an official `@n8n/cli` (Node, currently beta). `n8nctl` is a different
tool with a different shape, and the differences matter if you operate more than
one instance or script against n8n in CI:

- **Multi-instance is first class.** Define named profiles, one per instance, and
  switch with `--profile`, `N8NCTL_PROFILE`, or `n8nctl config use <name>`. The
  official `@n8n/cli` targets a single instance at a time; running it against a
  homelab box, an n8n Cloud tenant, and a client's server means juggling
  environment variables by hand. With `n8nctl` those are three named profiles you
  flip between in one word.
- **Secrets live in your OS keyring.** API keys go into the macOS Keychain, the
  GNOME/KDE Secret Service, or the Windows Credential Manager (service
  `n8nctl-cli`, account = profile name). They are never written to the config
  file in your home directory. The official CLI keeps the key in a plaintext file.
- **A single binary, no Node.** Install one file, or `go install`, or `brew
  install`. Nothing to `npm install`, no lockfile, no global Node version to
  keep happy on your CI runners.
- **Real output formats.** `table` (default), `json`, `yaml`, and `csv`, with
  `--columns` to pick fields. Pipe `csv` into a spreadsheet, `json` into `jq`.
- **Dry-run before you touch anything.** `--dry-run` prints the equivalent
  `curl` and sends no request, so destructive commands are easy to review.

- **Fleet operations beyond CRUD.** `workflows sync` promotes a workflow across
  instances, `backup`/`restore` snapshot an instance to git-friendly JSON, and
  `workflows search` finds workflows by node type, credential, or webhook path.
- **Faster to invoke.** A single Go binary starts in ~6 ms versus ~150 ms for the
  Node-based official CLI — invisible for one command, but real in loops and CI.

`n8nctl` covers every command the official CLI exposes (data tables, package
import/export, `--jq`, an `id-only` output mode, a `skills install` command) and
adds the above. The official `@n8n/cli` is still the right pick if you want the
**first-party** tool, work with a single instance, or already live in Node — it
is maintained by the n8n team and tracks new endpoints first.

See the full side-by-side, including performance benchmarks, in the
[comparison guide](https://jjuanrivvera.github.io/n8n-cli/comparison/).

If you only ever touch one instance from one laptop, the official CLI may be all
you need. If you run several instances, want secrets out of plaintext, or script
n8n from machines without Node, that is what `n8nctl` is for.

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
archives and the Homebrew/Scoop installs all ship **shell completions**
(bash/zsh/fish). Releases are **SBOM-attested and the checksums are signed** with
[cosign](https://github.com/sigstore/cosign) (keyless); see
[RELEASING.md](RELEASING.md) for verification.

## Quickstart

You need an n8n API key first. In the n8n UI, open **Settings → n8n API** and
create one. Copy the key and the instance URL.

```bash
# Friendliest first run: names a profile, captures the base URL and API key
# (stored in your OS keyring), verifies connectivity, and writes the config.
n8nctl init

# List the first page of workflows as a table
n8nctl workflows list

# Get one workflow as JSON
n8nctl workflows get 42 -o json

# Check that auth works against the active instance
n8nctl auth status
```

`init` prompts for a profile name, the instance base URL, and the API key (typed
without echo). If you prefer to script it, pass the values as flags:

```bash
n8nctl init --profile homelab --base-url https://n8n.lan/api/v1 --api-key "$KEY"
```

## Beyond the API

`n8nctl` is more than a thin wrapper over the REST endpoints. A few commands
compose the API into operations the n8n UI cannot do at all — cross-instance
promotion, git-friendly snapshots, and a graph search across every workflow.

### Promote a workflow between instances — `workflows sync`

n8n's own Git Source Control is an Enterprise feature. `workflows sync` gives
Community users a dev → staging → prod promotion path over the plain API: it
reads a workflow from one profile and writes it to another, stripping read-only
fields (id, active state, version) and carrying over nodes, connections, and
settings.

```bash
# Push a workflow from dev to prod, overwriting the one with the same name,
# and activate it on arrival
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate

# Default --from is the active profile; default is to create a new workflow
n8nctl --profile staging workflows sync 2tUt1wbLX592XDdX --to prod
```

Flags: `--to <profile>` (required), `--from <profile>` (default: active
profile), `--update-by-name` (overwrite a destination workflow with the same
name instead of creating a new one), `--activate`.

> **Caveat.** Credentials are referenced by id and are **not** copied.
> Create matching credentials on the destination first (`n8nctl credentials`);
> the synced nodes will resolve them by id.

### Snapshot and restore an instance — `backup` / `restore`

`backup` exports the active instance to a directory of pretty-printed JSON, one
file per workflow plus `tags.json`, `variables.json`, a credentials **inventory**
(metadata only), and a `manifest`. Commit that directory to git and you have
versioned, diffable instance state. `restore` re-applies a backup directory to
an instance.

```bash
# Snapshot prod into a directory you can commit to git
n8nctl --profile prod backup --out ./backups/prod

# Restore that snapshot into staging, overwriting by name and activating
n8nctl --profile staging restore --in ./backups/prod --update-by-name --activate
```

`backup` flags: `--out <dir>` (required). `restore` flags: `--in <dir>`
(required), `--update-by-name`, `--activate`.

> **Caveat.** Credential **secrets** are write-only in the n8n API and are
> never exported — the backup records credential metadata only. On restore,
> referenced credentials must already exist on the target instance.

### Find workflows by what is inside them — `workflows search`

Scan every workflow's node graph and report the ones that match. This answers
questions the UI cannot: which workflows use the Slack node, reference a given
credential, or own the `/orders` webhook.

```bash
# Which workflows use the Slack node?
n8nctl workflows search --node slack

# Which reference a specific credential (by id or name)?
n8nctl workflows search --credential githubApi -o json

# Who owns the /orders webhook path?
n8nctl workflows search --webhook /orders

# Match workflow names with a regular expression
n8nctl workflows search --name '^prod-'
```

Flags: `--node <type>` (substring match on node type), `--credential <id|name>`,
`--webhook <path>`, `--name <regex>`.

See [docs › Beyond the API](https://jjuanrivvera.github.io/n8n-cli/beyond-api/)
for worked examples and the roadmap of further beyond-API features.

## Multi-instance and profiles

This is the core reason `n8nctl` exists. A profile is a named instance: its base
URL, its output preferences, and a pointer to a keyring entry holding its API
key. The config file lives at `~/.n8nctl-cli/config.yaml`; keys never do.

```bash
# Create three instances
n8nctl init --profile homelab --base-url https://n8n.lan/api/v1
n8nctl init --profile cloud   --base-url https://acme.app.n8n.cloud/api/v1
n8nctl init --profile client  --base-url https://n8n.client.com/api/v1

# Run a command against a specific instance with --profile
n8nctl --profile cloud workflows list
n8nctl --profile client executions list --status error

# Or pick a default and drop the flag
n8nctl config use homelab
n8nctl workflows list            # now runs against homelab

# Or scope it to a single command / shell with the env var
N8NCTL_PROFILE=cloud n8nctl workflows list

# See what is configured
n8nctl config list-profiles
n8nctl config view              # resolved config, secrets redacted
```

Precedence is consistent everywhere: a `--profile` flag wins over
`N8NCTL_PROFILE`, which wins over the `default_profile` in the config file.

See the [multi-instance deep dive](docs/profiles.md) for the full `config.yaml`
layout.

## Output formats

```bash
n8nctl workflows list                                   # table (default)
n8nctl workflows list -o json | jq '.[].name'           # json into jq
n8nctl executions list -o yaml                           # yaml
n8nctl workflows list -o csv --columns id,name,active    # csv with chosen columns
```

`-o`/`--output` accepts `table`, `json`, `yaml`, or `csv`. `--columns` selects
which fields appear in `table` and `csv` output. Set a default once with
`n8nctl config set output_format json` or the `N8NCTL_OUTPUT` env var.

## Dry-run

Any command that would send a request honors `--dry-run`. Instead of calling the
API it prints the equivalent `curl`, with the API key redacted unless you add
`--show-token`:

```bash
n8nctl workflows delete 42 --dry-run
n8nctl credentials create --file cred.json --dry-run --show-token
```

## Common workflows

### Create a workflow from an exported JSON file

n8n exports a workflow as a JSON document. Feed it straight in:

```bash
n8nctl workflows create --file workflow.json

# Or from stdin
cat workflow.json | n8nctl workflows create --file -

# Or build the body inline (values parsed as JSON when valid)
n8nctl workflows create --set name="Nightly sync" --set 'settings={}'
```

A workflow body needs `name`, `nodes`, `connections`, and `settings`. After
creation, activate it:

```bash
n8nctl workflows activate 42
```

### Filter executions by status

```bash
# Just the failures, every page
n8nctl executions list --status error --all

# Scope to one workflow, include the full run data
n8nctl executions list --workflow 42 --include-data true

# Retry or stop a specific execution
n8nctl executions retry 1011
n8nctl executions stop 1012
```

Executions are read-only apart from `retry`, `stop`, and `delete`; n8n creates
them by running workflows.

### Create a credential after inspecting its schema

Credential secrets are write-only: the API accepts them on create or update but
never returns them. Discover a type's required fields first, then create:

```bash
# What fields does a GitHub credential need?
n8nctl credentials schema githubApi

# Create it
n8nctl credentials create \
  --set name='My GitHub' \
  --set type=githubApi \
  --set data='{"accessToken":"ghp_..."}'
```

### Drop to the raw API when a flag does not exist

```bash
# PATH is relative to <base-url>; the /api/v1 prefix is added for you
n8nctl api GET /workflows --query limit=5
n8nctl api POST /tags --data '{"name":"Prod"}'
n8nctl api DELETE /executions/42 --dry-run
```

### Save a shortcut as an alias

```bash
n8nctl alias set failures 'executions list --status error --all'
n8nctl failures --profile cloud
```

## Configuration and environment

Config lives at `~/.n8nctl-cli/config.yaml`. API keys live in the OS keyring
(service `n8nctl-cli`, account = profile name), never in the file. Every value
can be overridden by an environment variable; flags override both.

| Env var | Meaning |
| --- | --- |
| `N8NCTL_PROFILE` | Active profile (instance) name |
| `N8NCTL_BASE_URL` | Instance base URL (`<host>/api/v1`) |
| `N8NCTL_API_KEY` | API key (skips the keyring lookup) |
| `N8NCTL_OUTPUT` | Default output format (`table`/`json`/`yaml`/`csv`) |
| `N8NCTL_RPS` | Client-side rate limit, requests per second |
| `N8NCTL_LOG_LEVEL` | Log level (`debug`/`info`/`warn`/`error`) |
| `N8NCTL_CONFIG` | Override the config file path |
| `NO_COLOR` | Disable colored output |

Global flags available on every command:

`--profile`, `-o/--output`, `--base-url`, `--api-key`, `--rps`, `--dry-run`,
`--show-token`, `-v/--verbose`, `--no-color`, `-q/--quiet`, `--columns`.

## Commands

Top-level resources, each with `list`/`get`/`create`/`update`/`delete` plus
resource-specific actions where the API supports them:

`workflows` (also `activate`, `deactivate`, `archive`, `unarchive`, `transfer`,
`tags`, plus the beyond-API `sync` and `search`), `executions` (`retry`, `stop`;
read-only otherwise), `credentials` (`schema`, `transfer`), `tags`, `variables`,
`projects` (member management), `users` (`invite`, `change-role`; instance owner
only), `audit`, and `source-control` (`pull`).

Beyond-API and meta commands: `backup`, `restore`, `auth`, `config`, `init`,
`doctor`, `completion`, `alias`, `api`, `version`.

Run `n8nctl <command> --help` for the actions and flags of any command, or browse
the full [command reference](docs/commands/index.md).

## Install the skill / plugin

`n8nctl` ships an **agent skill** that teaches AI coding agents (Claude Code,
Cursor, Codex, Gemini CLI, …) how to drive it — commands, flags, the
`--dry-run` safety rule, and the beyond-API operations. Install it as a native
Claude Code plugin:

```text
/plugin marketplace add jjuanrivvera/n8n-cli
/plugin install n8nctl-cli@n8nctl
```

The skill wraps this binary, so install the CLI (above) and authenticate first.
To wire up shell completion at the same time:

```bash
# bash / zsh / fish / powershell
source <(n8nctl completion bash)
n8nctl completion zsh > "${fpath[1]}/_n8nctl"
```

## Documentation

The [documentation site](https://jjuanrivvera.github.io/n8n-cli/) covers:

- [Getting started](https://jjuanrivvera.github.io/n8n-cli/getting-started/)
- [Multi-instance and profiles](https://jjuanrivvera.github.io/n8n-cli/profiles/)
- [Authentication](https://jjuanrivvera.github.io/n8n-cli/authentication/)
- [Output and filtering](https://jjuanrivvera.github.io/n8n-cli/output/)
- [Beyond the API](https://jjuanrivvera.github.io/n8n-cli/beyond-api/)
- [Command reference](https://jjuanrivvera.github.io/n8n-cli/commands/)

## Development

```bash
make dev            # fmt + vet + build
make test           # run tests
make lint           # golangci-lint
make check          # full local quality gate
make docs-serve     # preview the docs site
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the architecture and contribution
workflow.

## License

MIT — see [LICENSE](LICENSE).
</content>
</invoke>
