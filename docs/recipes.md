# Recipes

Short, copy-pasteable solutions to common tasks. Each example targets the active
profile; add `--profile <name>` to run against a specific instance. Preview any
write with `--dry-run` to see the equivalent `curl` first.

## Create a workflow from a file, stdin, or inline

`workflows create` accepts a request body three ways: a file, stdin, or
repeatable `--set key=value` pairs (values are parsed as JSON when possible).

```bash
# From a file on disk
n8nctl workflows create --file my-workflow.json

# From stdin (use '-' as the file)
cat my-workflow.json | n8nctl workflows create --file -

# Inline JSON
n8nctl workflows create --data '{"name":"My workflow","nodes":[],"connections":{}}'

# Field by field with --set (repeatable; values parsed as JSON when valid)
n8nctl workflows create \
  --set name="My workflow" \
  --set nodes='[]' \
  --set connections='{}'
```

`credentials`, `tags`, `variables`, and `projects` `create` the same way.

## Filter executions by status

```bash
# Only failed executions
n8nctl executions list --status error

# Failures for one workflow, every page
n8nctl executions list --status error --workflow 42 --all

# Just the ids of failed runs, one per line
n8nctl executions list --status error -o id

# Retry every failed execution of a workflow
n8nctl executions list --status error --workflow 42 -o id \
  | xargs -n1 n8nctl executions retry
```

## Create a credential after inspecting its schema

Look up the type's required fields, then create the credential.

```bash
# What fields does a Slack API credential need?
n8nctl credentials schema slackApi

# Create it with the fields the schema listed
n8nctl credentials create \
  --set name="Slack (prod)" \
  --set type=slackApi \
  --set 'data={"accessToken":"xoxb-..."}'
```

## Promote a workflow across instances

Copy a workflow from one profile to another. Read-only fields are stripped;
credentials are referenced by id and are not copied, so create matching
credentials on the destination first.

```bash
# Push from dev to prod, overwriting the same-named workflow, and activate it
n8nctl workflows sync 2tUt1wbLX592XDdX \
  --from dev --to prod --update-by-name --activate

# Preview the calls without sending them
n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --dry-run
```

See [Beyond the API](beyond-api.md) for the full sync model.

## Back up and restore an instance

```bash
# Snapshot the active instance to a directory you can commit to git
n8nctl backup --out ./n8n-backup

# Git-friendlier: YAML with code fields split into sibling files
n8nctl --profile prod backup --out ./backups/prod --format yaml --externalize 5

# Restore into staging, overwriting by name and activating each workflow
n8nctl --profile staging restore --in ./n8n-backup --update-by-name --activate
```

Credentials are inventoried as metadata only — secrets are never exported, and
must already exist on the target before restore.

## Lint and apply a directory in CI

Treat a directory of workflow files as the desired state. Lint first, then
reconcile. Both exit non-zero on failure, so they gate a pipeline cleanly.

```bash
# Fail the build if any workflow file has structural errors
n8nctl workflows lint --dir ./workflows

# Preview the reconcile (create/update/prune) before touching prod
n8nctl --profile prod workflows apply --dir ./workflows --prune --dry-run

# Apply for real: create new, update existing, delete drift
n8nctl --profile prod workflows apply --dir ./workflows --prune
```

See [Workflows as Code](workflows-as-code.md) for a full CI example.

## Find a node's type and parameters

The lint and autofix rules need the exact node `type` string, and hand-authoring
a workflow file needs the parameter names a node accepts. Both come from the
embedded catalog, offline.

```bash
# Find the type string for a node by display name
n8nctl nodes search slack

# List the parameters the Webhook node accepts
n8nctl nodes show n8n-nodes-base.webhook

# Pull just the parameter names as JSON for scripting
n8nctl nodes show n8n-nodes-base.slack -o json --jq '.params'
```

## Auto-fix a directory of workflows before committing

`workflows autofix` repairs the mechanical mistakes `workflows lint` reports:
typo'd node types, expression strings missing the leading `=`, and webhook nodes
without a `webhookId`. Report first, then write.

```bash
# See what would change (report only — nothing is written)
n8nctl workflows autofix --dir ./workflows

# Apply the fixes in place
n8nctl workflows autofix --dir ./workflows --write

# Lint again to confirm the remaining findings are not mechanical
n8nctl workflows lint --dir ./workflows
```

## Prune old failed executions in CI

Reclaim database space by deleting stale execution records. Always count first;
`--yes` skips the confirmation so it runs unattended.

```bash
# Count what would be deleted, without deleting (good for a CI report step)
n8nctl executions prune --older-than 30d --status error --dry-run

# Delete failed executions older than 30 days, no prompt
n8nctl executions prune --older-than 30d --status error --yes

# Scope to a single workflow
n8nctl executions prune --older-than 7d --workflow 42 --yes
```

## Bulk-deactivate a tag for a maintenance window

Flip every workflow carrying a tag in one command — deactivate the set, do the
maintenance, reactivate it.

```bash
# Preview the affected workflows
n8nctl workflows bulk deactivate --tag prod --dry-run

# Deactivate them
n8nctl workflows bulk deactivate --tag prod --yes

# …run the maintenance, then bring them back
n8nctl workflows bulk activate --tag prod --yes
```

## Drop to the raw API

When a capability is not yet a first-class command, call any endpoint directly.
The leading `/api/v1` is added for you.

```bash
n8nctl api GET /workflows --query limit=5
n8nctl api POST /tags -d '{"name":"Prod"}'
n8nctl api DELETE /executions/42 --dry-run
```

## Save an alias

Aliases are shortcuts expanded before parsing. Define one for a command you run
often.

```bash
# Define an alias for "list failed executions as ids"
n8nctl alias set failures 'executions list --status error -o id'

# Use it
n8nctl failures

# Inspect and remove
n8nctl alias list
n8nctl alias remove failures
```

## Pipe with jq and ids

Use `-o json --jq` for structured filtering, or `-o id` to feed ids into
`xargs`.

```bash
# Names of every active workflow
n8nctl workflows list --all -o json --jq '.[] | select(.active) | .name'

# Tag ids matching a name
n8nctl tags list -o json --jq '.[] | select(.name=="prod") | .id'

# Archive every workflow tagged "legacy"
n8nctl workflows list --tags legacy -o id \
  | xargs -n1 n8nctl workflows archive
```

## Where to next

- [Features](features.md) — the full capability reference.
- [Output and filtering](output.md) — formats, columns, jq, and pagination.
- [Command reference](commands/index.md) — every command and flag.
