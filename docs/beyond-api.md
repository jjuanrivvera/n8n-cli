# Beyond the API

`n8nctl` is more than a thin wrapper over the n8n REST endpoints. A handful of
commands compose the public API into operations the n8n UI cannot do at all —
cross-instance promotion, git-friendly snapshots, and a graph search across every
workflow. This page documents them with worked examples and lists the roadmap of
further beyond-API ideas.

Every command here honors the global flags, including `--dry-run`, `--profile`,
and `-o json|yaml|csv`. Preview anything destructive with `--dry-run` first.

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

`backup` exports the active instance to a directory of pretty-printed JSON for
git-based versioning. It writes one file per workflow plus `tags.json`,
`variables.json`, a credentials **inventory** (metadata only), and a `manifest`.
Commit that directory and you have versioned, diffable instance state. `restore`
re-applies a backup directory to an instance.

```bash
# Snapshot prod into a directory you can commit to git
n8nctl --profile prod backup --out ./backups/prod

# A typical git-versioning loop
n8nctl --profile prod backup --out ./n8n-state
git -C ./n8n-state add -A && git -C ./n8n-state commit -m "n8n snapshot $(date -u +%F)"

# Restore that snapshot into staging, overwriting by name and activating
n8nctl --profile staging restore --in ./backups/prod --update-by-name --activate
```

| Command | Flag | Meaning |
|---|---|---|
| `backup` | `--out <dir>` | Output directory (**required**) |
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

## Roadmap / proposed

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
- **Cross-instance diff — `workflows diff`** *(proposed)*. Compare the same
  workflow on two profiles (e.g. dev vs prod) and print a structured diff of
  nodes, connections, and settings before a `workflows sync`, so a promotion is
  reviewable rather than blind.
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
