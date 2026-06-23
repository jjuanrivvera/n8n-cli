# Beyond the API

`n8nctl` is more than a thin wrapper over the n8n REST endpoints. A handful of
commands compose the public API into operations the n8n UI cannot do at all —
cross-instance promotion, git-friendly snapshots, and a graph search across every
workflow. This page documents them with worked examples and lists the roadmap of
further beyond-API ideas.

Every command here honors the global flags, including `--dry-run`, `--profile`,
and `-o json|yaml|csv`. Preview anything destructive with `--dry-run` first.

!!! tip "Workflows as code"
    For the full GitOps loop — reconciling a directory of workflow files into an
    instance (`workflows apply`), linting them (`workflows lint`), converting
    between JSON and YAML (`workflows convert`), and diffing a workflow against
    another source (`workflows diff`) — see the dedicated
    [Workflows as Code](workflows-as-code.md) guide. This page covers the
    single-workflow promotion (`workflows sync`), instance snapshots
    (`backup` / `restore`), and the graph search (`workflows search`).

## Promote a workflow between instances — `workflows sync`

n8n's own Git Source Control is an **Enterprise** feature. `workflows sync` gives
Community users a dev → staging → prod promotion path over the plain API. It
reads a workflow from one profile and writes it to another. Read-only fields
(`id`, active state, version) are stripped; nodes, connections, and settings are
carried over. By default a new workflow is created on the destination;
`--update-by-name` overwrites an existing workflow with the same name.

```bash
# Push a workflow from dev to prod, overwriting the one with the same name,
# and activate it on arrival
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate

# --from defaults to the active profile, so this promotes from staging to prod
n8nctl --profile staging workflows sync 2tUt1wbLX592XDdX --to prod

# Preview the calls without sending them
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --dry-run
```

| Flag | Meaning |
|---|---|
| `--to <profile>` | Destination profile (**required**) |
| `--from <profile>` | Source profile (default: the active profile) |
| `--update-by-name` | Overwrite a destination workflow with the same name instead of creating a new one |
| `--activate` | Activate the workflow on the destination after syncing |

!!! warning "Credentials are not copied"
    Credentials are referenced by **id** and are **not** copied between
    instances. Create matching credentials on the destination first (see
    `n8nctl credentials`); the synced nodes will resolve them by id. If the
    destination has no credential with that id, the promoted workflow will need
    its nodes re-pointed before it can run.

## Snapshot and restore an instance — `backup` / `restore`

`backup` exports the active instance to a directory for git-based versioning. It
writes one file per workflow plus `tags.json`, `variables.json`, a credentials
**inventory** (metadata only), and a `manifest`. Commit that directory and you
have versioned, diffable instance state. `restore` re-applies a backup directory
to an instance.

Workflow files default to pretty-printed JSON, but `--format yaml` and
`--externalize <N>` make the backup far git-friendlier: YAML reviews better in a
pull request, and `--externalize` splits long node code fields (`jsCode`,
`pythonCode`, `query`/`sqlQuery`, `jsonBody`, `content`) into sibling `.js`/`.py`/
`.sql` files referenced by a `{$ref: …}` marker. `restore` reads either format and
re-inlines any externalized `$ref` fields automatically. See
[Workflows as Code](workflows-as-code.md#backups-as-yaml-with-externalized-code)
for the externalization layout.

```bash
# Snapshot prod into a directory you can commit to git
n8nctl --profile prod backup --out ./backups/prod

# Git-friendlier: YAML with code split into sibling files
n8nctl --profile prod backup --out ./n8n-state --format yaml --externalize 5
git -C ./n8n-state add -A && git -C ./n8n-state commit -m "n8n snapshot $(date -u +%F)"

# Restore that snapshot into staging, overwriting by name and activating
n8nctl --profile staging restore --in ./backups/prod --update-by-name --activate
```

| Command | Flag | Meaning |
|---|---|---|
| `backup` | `--out <dir>` | Output directory (**required**) |
| `backup` | `--format json\|yaml` | Workflow file format in the backup (default `json`) |
| `backup` | `--externalize <N>` | Split code fields longer than N lines into sibling files (0 = off) |
| `restore` | `--in <dir>` | Backup directory to restore from (**required**) |
| `restore` | `--update-by-name` | Overwrite existing workflows with the same name |
| `restore` | `--activate` | Activate each restored workflow |

!!! warning "Credential secrets are never exported"
    Credential **secrets** are write-only in the n8n API — they are accepted on
    create/update but never returned. A backup therefore records credential
    metadata only (id, name, type), not the stored values. On `restore`,
    referenced credentials must already exist on the target instance.

## Find workflows by what is inside them — `workflows search`

Scan every workflow's node graph and report the ones that match. This answers
questions the UI cannot: which workflows use a given node, reference a specific
credential, or own a webhook path.

```bash
# Which workflows use the Slack node?
n8nctl workflows search --node slack

# Which reference a specific credential (by id or name)? Pipe to jq for ids.
n8nctl workflows search --credential githubApi -o json | jq '.[].name'

# Who owns the /orders webhook path?
n8nctl workflows search --webhook /orders

# Match workflow names with a regular expression
n8nctl workflows search --name '^prod-'
```

| Flag | Meaning |
|---|---|
| `--node <type>` | Substring match on node type (e.g. `slack`, `httpRequest`) |
| `--credential <id\|name>` | Workflows referencing a credential by id or name |
| `--webhook <path>` | Workflows with a webhook node on that path |
| `--name <regex>` | Workflow name matches a regular expression |

`search` reads from the active instance; combine it with `--profile` to audit a
specific one. It is read-only — nothing is modified.

## Roadmap

### Implemented

The "workflows as code" set shipped in **v0.2.0** and is documented in full on the
[Workflows as Code](workflows-as-code.md) page:

- **Declarative reconcile — `workflows apply`** *(implemented)*. Treat a directory
  of workflow files as the desired state: create, update, skip-unchanged, and
  with `--prune` delete instance workflows absent from the directory. The
  multi-instance angle — promoting the same directory across profiles — is the
  Community-tier answer to Enterprise Git Source Control.
- **Static linting — `workflows lint`** *(implemented)*. Five grounded rules over
  workflow files or live workflows (`--remote`), exiting non-zero on errors so it
  works as a CI gate.
- **Format conversion — `workflows convert`** *(implemented)*. Convert workflow
  files between JSON and YAML, with `--externalize` to split long code fields into
  reviewable sibling files.
- **Cross-instance diff — `workflows diff`** *(implemented)*. Compare the same
  workflow on two profiles (e.g. dev vs prod), or against a local file, and print
  a unified diff of writable content before a promotion, so it is reviewable
  rather than blind.

### Proposed

The following are **proposed, not yet implemented**. They are grounded in common
n8n operational pain points and follow the same "compose the public API into
something the UI can't do" pattern. Tracking them here so the direction is clear;
they may change or be dropped.

- **Execution pruning — `executions prune`** *(proposed)*. Bulk-delete old
  execution records by age and/or status to reclaim database space on busy
  self-hosted instances, with a `--dry-run` count first. n8n's UI deletes
  executions one page at a time; the API exposes the filters to do it in bulk.
- **Live failure watch — `executions watch`** *(proposed)*. Poll the executions
  endpoint and stream new `error`/`crashed` runs as they happen, so a terminal
  can tail an instance's failures during a deploy or incident.
- **Node-schema lint validation** *(proposed)*. Extend `workflows lint` to validate
  each node's parameters against that node type's own schema (for example, a
  missing required parameter or an invalid option value), beyond the current
  structural and graph-level rules. This needs per-node-type schemas and is the
  natural next step for the linter.
- **Bulk activate/deactivate by tag — `workflows activate --tag <name>`**
  *(proposed)*. Toggle every workflow carrying a tag in one command, for
  maintenance windows (deactivate the `prod` set, do the work, reactivate),
  previewable with `--dry-run`.
- **Instance health summary — `stats`** *(proposed)*. A one-shot health and
  usage summary for an instance, composed from the insights and audit endpoints:
  workflow/execution counts, recent failure rate, and the audit's risk
  highlights, rendered as a table or `json`.

If one of these would help you, open an issue — concrete use cases drive what
ships next.
</content>
