# Workflows as Code

n8n's UI is great for building a workflow once. It is poor at the next part:
keeping a fleet of workflows under version control, reviewing changes before they
ship, and promoting the same set of workflows from a dev box to staging to
production. Those are the problems a GitOps workflow solves, and they are what
this group of commands is for.

The model is simple. A directory of workflow files is the **desired state**.
`n8nctl` reconciles that directory against a live instance: it creates what is
missing, updates what changed, leaves untouched what already matches, and (on
request) deletes what no longer belongs. Because every command takes
`--profile`, the same directory promotes cleanly across instances.

This page covers the commands that make up the loop —
[`workflows apply`](#apply-reconcile-a-directory-into-an-instance),
[`workflows lint`](#lint-static-checks-on-workflow-files),
[`workflows autofix`](#autofix-repair-what-lint-detects),
[`workflows breaking-changes`](#breaking-changes-spot-outdated-nodes-before-an-upgrade),
[`workflows convert`](#convert-between-json-and-yaml), and
[`workflows diff`](#diff-a-workflow-against-another-source) — plus the YAML and
code-externalization support shared with [`backup`](#backups-as-yaml-with-externalized-code).

For the single-workflow promotion command and the graph-wide search, see
[Beyond the API](beyond-api.md).

## The GitOps loop

```
  ┌──────────┐   git    ┌──────────┐   CI    ┌──────────┐  apply   ┌──────────┐
  │  backup  │ ───────▶ │   edit   │ ──────▶ │   lint   │ ───────▶ │  target  │
  │  export  │  commit  │ in a PR  │  check  │ in CI    │  --prune │ instance │
  └──────────┘          └──────────┘         └──────────┘          └──────────┘
```

1. **Seed the repo.** Snapshot a live instance to a directory with `backup`
   (YAML and externalized code make the diffs readable), or hand-author workflow
   files.
2. **Edit in git.** Change workflow files in a branch and open a pull request.
   The diff is a real, reviewable text diff — not a screenshot of a canvas.
3. **Lint in CI.** Run `workflows lint --dir ./workflows` in the pipeline. It
   exits non-zero on errors, so a malformed workflow fails the build before it
   ever reaches an instance.
4. **Preview with a dry run.** Run `workflows apply --dir ./workflows --dry-run`
   to see exactly what would be created, updated, or pruned.
5. **Apply to the target.** Run the same command without `--dry-run`. With
   `--prune` the instance is brought to exactly the directory's state.

The headline is the last two steps repeated per environment with a different
`--profile` — the same directory, promoted across instances. That section is
[below](#multi-instance-promotion).

## apply: reconcile a directory into an instance

`workflows apply` treats `--dir` as the desired state for the active instance.
Files may be JSON or YAML. Workflows are matched to instance workflows **by
name**:

- a file with a name not present on the instance is **created**;
- a file whose name exists but whose content differs is **updated**;
- a file whose content already matches the instance is **skipped** (counted as
  *unchanged*), decided by a canonical compare of the writable fields, so
  cosmetic differences like key ordering or read-only fields never trigger a
  spurious update;
- with `--prune`, an instance workflow whose name is **not** in the directory is
  **deleted**.

```bash
# Always preview first. --dry-run is a global flag.
n8nctl workflows apply --dir ./workflows --dry-run

# Apply for real: create and update, but never delete
n8nctl workflows apply --dir ./workflows

# Full reconcile: also prune instance workflows absent from the directory
n8nctl workflows apply --dir ./workflows --prune

# Activate anything newly created on arrival
n8nctl workflows apply --dir ./workflows --activate
```

The output lists each planned change and ends with a one-line summary. A dry run
labels the summary `plan:`; a real run labels it `applied:`.

```text
+ create order-intake
~ update slack-alerts
- prune legacy-cron
plan: 1 created, 1 updated, 4 unchanged, 1 pruned
```

`+ create`, `~ update`, and `- prune` mark the three kinds of change; unchanged
workflows are summarized in the count rather than listed line by line.

| Flag | Meaning |
|---|---|
| `--dir`, `-d <dir>` | Directory of workflow files, the desired state (**required**) |
| `--prune` | Delete instance workflows whose name is not present in the directory |
| `--activate` | Activate workflows created by this run |
| `--dry-run` (global) | Print the plan and send no write requests |

!!! warning "Preview before pruning"
    `--prune` deletes instance workflows. Run with `--dry-run` first and read the
    `- prune` lines. A name typo in a filename, or a workflow created directly in
    the UI, will show up as a prune candidate.

!!! note "Credentials are referenced by id, not copied"
    Like every cross-instance operation in `n8nctl`, `apply` carries node
    definitions and their credential **references** (by id), not the credential
    secrets. Matching credentials must already exist on the target instance.
    See the credentials note under [Beyond the API](beyond-api.md).

### Multi-instance promotion

This is the part the single-instance tools cannot do. Because the instance is
selected by `--profile`, the *same* directory is the desired state for every
environment. A promotion pipeline is the same two commands aimed at different
profiles:

```bash
# Promote dev → staging → prod from one source of truth.

# 1. Reconcile staging (create/update only)
n8nctl --profile staging workflows apply --dir ./workflows

# 2. Validate on staging, then bring prod to an exact match — pruning drift
n8nctl --profile prod workflows apply --dir ./workflows --prune
```

The official `@n8n/cli` and the community GitOps tools built around it target one
instance at a time: the instance URL and key come from environment variables, so
driving three environments means swapping those variables between runs and hoping
the right one is exported. `n8nctl` keeps each instance as a named profile with
its key in the OS keyring, so the environment is a single flag and switching them
never crosses credentials. n8n's own Git-based Source Control, which would cover
this, is an Enterprise feature; `apply` gives Community-tier users a promotion
path over the plain public API.

A typical CI job promoting to production on merge:

```yaml
# .github/workflows/promote.yml (sketch)
- name: Lint workflows
  run: n8nctl workflows lint --dir ./workflows

- name: Preview prod reconcile
  run: n8nctl --profile prod workflows apply --dir ./workflows --prune --dry-run

- name: Apply to prod
  if: github.ref == 'refs/heads/main'
  run: n8nctl --profile prod workflows apply --dir ./workflows --prune
```

The base URL and key for each profile come from the environment in CI
(`N8NCTL_BASE_URL`, `N8NCTL_API_KEY`, or a mounted config), so no secret is
committed.

## lint: static checks on workflow files

`workflows lint` runs static checks over workflow files — or, with `--remote`,
over the workflows live on the instance. It **exits non-zero when any error-level
finding is present**, which is what makes it useful as a CI gate. Warnings do not
fail the run.

```bash
# Lint every workflow file in a directory
n8nctl workflows lint --dir ./workflows

# Lint specific files (repeatable)
n8nctl workflows lint -f order-intake.json -f slack-alerts.yaml

# Lint the live instance instead of files
n8nctl workflows lint --remote

# Machine-readable output for a CI annotation step
n8nctl workflows lint --dir ./workflows -o json

# See the rules and their canonical basis
n8nctl workflows lint --list-rules

# Turn off a rule you don't want enforced
n8nctl workflows lint --dir ./workflows --disable-rule expression-prefix
```

The default output is one line per finding, prefixed `✗` for errors and `⚠` for
warnings:

```text
✗ broken · Missing: connection target references a missing node (connection-reference)
✗ broken · Webhook: webhook/formTrigger node is missing webhookId (webhook-id-required)
⚠ broken · Webhook: parameter "value" looks like an expression but is missing the leading '=' (expression-prefix)
⚠ broken · Orphan: node is not connected to any other node (orphaned-node)
Error: lint found 2 error(s)
```

`-o json` emits a structured array, one object per workflow with its findings —
the shape a CI step parses to post annotations:

```json
[
  {
    "workflow": "./workflows/broken.json",
    "findings": [
      {
        "rule": "connection-reference",
        "severity": "error",
        "node": "Missing",
        "message": "connection target references a missing node"
      },
      {
        "rule": "expression-prefix",
        "severity": "warning",
        "node": "Webhook",
        "message": "parameter \"value\" looks like an expression but is missing the leading '='"
      }
    ]
  }
]
```

| Flag | Meaning |
|---|---|
| `--dir <dir>` | Lint every workflow file in a directory |
| `--file`, `-f <file>` | Lint specific files (repeatable) |
| `--remote` | Lint the live workflows on the instance instead of files |
| `--list-rules` | Print the rules and their canonical basis, then exit |
| `--disable-rule <r>` | Disable one or more rules (comma-separated) |
| `-o json` (global) | Emit findings as JSON instead of the text report |

!!! note "Output format"
    The text report is the default. `-o json` produces the structured form above.
    `--output` is the global format flag (`table\|json\|yaml\|csv`); `json` is the
    machine-readable lint format. There is no separate `text` value — omit
    `--output` for the human report.

### Linting provenance

n8n does not ship an official workflow linter, so these rules are `n8nctl`'s own.
To keep them honest rather than arbitrary, each rule is grounded in a documented
n8n behavior, in the workflow's own data model, or in n8n's real node definitions.
`--list-rules` prints the basis for each:

| Rule | Severity | Basis |
|---|---|---|
| `required-fields` | error | n8n public-API OpenAPI workflow schema (`name`, `nodes`, `connections`, `settings` are required) |
| `connection-reference` | error | The workflow connection graph model — a connection must point at a node that exists |
| `orphaned-node` | warning | The workflow connection graph model — a node disconnected from the graph is usually a mistake |
| `webhook-id-required` | error | n8n webhook registration behavior — webhook and form-trigger nodes need a `webhookId` to register |
| `expression-prefix` | warning | n8n expression syntax — a string is only evaluated as an expression when it starts with `=`, so a `{{ }}` string without the prefix is a literal |
| `unknown-node-type` | error | Embedded catalog of n8n's real node definitions — a node type from a known package (`n8n-nodes-base`, langchain) that is not in the catalog is almost always a typo; community/custom nodes are skipped |
| `unknown-parameter` | warning | Embedded node catalog — a parameter the node type does not define |
| `invalid-parameter-value` | warning | Embedded node catalog (option values plus `displayOptions` visibility) — an `options`/`multiOptions` parameter set to a value not in the node's current options |

The last three are **node-schema-aware**: they validate against a catalog of n8n's
real node definitions (`n8n-nodes-base` plus the langchain nodes, 560+ types),
generated by `make gen-node-schemas` and embedded at build time. They catch
typo'd node types (with a "did you mean …?" suggestion), parameters a node does
not define, and option values a node does not allow.

`invalid-parameter-value` validates parameter *values*, not just names. For an
`options` or `multiOptions` parameter, it checks the chosen value against the set
of values the node actually allows. It is **`displayOptions`-aware**: many nodes
expose different option sets depending on another parameter (a Slack node's
`operation` options differ by `resource`), so the rule validates against the
option set that is *active* given the node's other parameters. A Slack node with
`resource: "message"` and `operation: "psot"` is flagged with a suggestion:

```text
⚠ typo · Slack: parameter "operation" = "psot" is not an allowed value — did you mean "post"? (invalid-parameter-value)
```

It is a **warning**, not an error, so it surfaces the issue without failing the
lint gate: a flagged value may be a typo (see the "did you mean …?" hint) or a
legacy value that an older node version allowed and a newer one removed — which
`workflows breaking-changes` reports separately. The rule is conservative: it
skips dynamic option lists, expression values, and any parameter whose active
option set it cannot resolve, so it does not false-positive on a valid workflow.

The rule is deliberately conservative, so it never false-positives. It skips
dynamic option lists (values n8n loads at runtime), expression values
(`={{ … }}`), and unknown or community nodes. When it cannot prove a value is
wrong, it stays silent.

## autofix: repair what lint detects

`workflows lint` reports mistakes; `workflows autofix` repairs the ones a machine
can fix safely. It is the natural follow-up to a failing lint run: rather than
hand-editing each file, let autofix apply the mechanical corrections, then lint
again to confirm only judgment-level findings remain.

It fixes three classes of mistake, all mechanical:

- **Typo'd node types.** A node `type` from a known package (`n8n-nodes-base`,
  langchain) that is not in the embedded catalog is corrected to the nearest real
  type — the same catalog the `unknown-node-type` lint rule checks against, so
  `n8n-nodes-base.slak` becomes `n8n-nodes-base.slack`.
- **Expressions missing the leading `=`.** A string that looks like an expression
  (`{{ … }}`) but lacks the leading `=` is a literal, not an expression. autofix
  adds the prefix, clearing the `expression-prefix` finding.
- **Missing webhook ids.** A `webhook` or `formTrigger` node without a `webhookId`
  cannot register. autofix generates one, clearing `webhook-id-required`.

By default it **reports** what it would change and writes nothing. Pass `--write`
to apply the fixes in place.

```bash
# Report the fixes for a directory (nothing is written)
n8nctl workflows autofix --dir ./workflows

# Apply them in place
n8nctl workflows autofix --dir ./workflows --write

# Fix specific files (repeatable)
n8nctl workflows autofix -f order-intake.json -f slack-alerts.yaml --write

# The repair → verify loop
n8nctl workflows autofix --dir ./workflows --write
n8nctl workflows lint --dir ./workflows
```

| Flag | Meaning |
|---|---|
| `--dir <dir>` | Fix every workflow file in a directory |
| `--file`, `-f <file>` | Fix specific files (repeatable) |
| `--write` | Write the fixes back (default: report only) |

autofix only touches mechanical mistakes with a single safe correction. Findings
that need a judgment call — an orphaned node, a connection pointing at a node that
no longer exists, an unknown parameter — are left for a human; run `lint` after
autofix to see what remains.

## breaking-changes: spot outdated nodes before an upgrade

`workflows breaking-changes` compares each workflow's nodes against the embedded
node catalog and reports the ones pinned to a `typeVersion` older than the latest
known version of that node type. For each outdated node it also lists any
parameters the workflow uses that the latest schema no longer defines — the
rename and removal candidates that an n8n upgrade is most likely to break.

It is **informational and exits 0**. Unlike `lint`, it is not a gate: it answers
"what in this set is likely to need attention when I move to a newer n8n?" so you
can review those nodes before, not after, an upgrade.

Inputs mirror `lint`. Scan a directory, specific files, the live instance, or a
single live workflow by id:

```bash
# Scan a directory of workflow files
n8nctl workflows breaking-changes --dir ./workflows

# Scan specific files (repeatable)
n8nctl workflows breaking-changes -f order-intake.json -f slack-alerts.yaml

# Scan every workflow live on the instance
n8nctl workflows breaking-changes --remote

# Scan a single live workflow by id
n8nctl workflows breaking-changes 42
```

Each line names the workflow, the node, and the version drift; when the latest
schema has dropped a parameter the node still uses, that parameter is called out:

```text
order-intake · Slack (n8n-nodes-base.slack): typeVersion 1 → latest 2
slack-alerts · Webhook (n8n-nodes-base.webhook): typeVersion 1 → latest 2; params not in latest schema: value
2 node(s) on an outdated typeVersion
```

A `typeVersion` lower than the catalog's latest is the drift signal: the node
runs, but it is not on the current schema, so an upgrade may change its behavior
or migrate its parameters. A parameter listed under "params not in latest schema"
is one the latest version of that node no longer declares — usually because it was
renamed or removed — and is the first thing to check when reconciling the node to
the new version. Community and custom nodes are skipped, since the catalog has no
authoritative latest version for them.

| Flag | Meaning |
|---|---|
| `--dir <dir>` | Scan every workflow file in a directory |
| `--file`, `-f <file>` | Scan specific files (repeatable) |
| `--remote` | Scan the live workflows on the instance |
| `<id>` | Scan a single live workflow by id |

Aliased as `breaking`.

## convert: between JSON and YAML

`workflows convert` rewrites workflow files between JSON and YAML on disk. n8n
exports JSON; YAML is friendlier to review in a pull request, and a workflow's
embedded code reviews better as a real `.js`/`.py`/`.sql` file than as an escaped
one-line string. `convert` handles both.

```bash
# JSON → YAML, written alongside the input (good.json → good.yaml)
n8nctl workflows convert good.json --to yaml

# YAML → JSON, into a separate output directory
n8nctl workflows convert *.yaml --to json --out ./json

# YAML and split out any code field longer than 5 lines
n8nctl workflows convert code.json --to yaml --externalize 5 --out ./out
```

Each conversion prints what it wrote:

```text
converted code.json -> ./out/code.yaml
```

| Flag | Meaning |
|---|---|
| `--to json\|yaml` | Target format (**required**) |
| `--out <dir>` | Output directory (default: alongside each input file) |
| `--externalize <N>` | Split node code fields longer than N lines into sibling files (0 = off) |

### Code externalization

`--externalize N` pulls long code values out of the workflow and into separate
files. It applies to the node fields that commonly hold code or large literals —
`jsCode`, `pythonCode`, `query` / `sqlQuery`, `jsonBody`, and `content` — when the
value exceeds N lines. The value in the workflow file is replaced with a
`{$ref: <relative-path>}` marker, and the code lands under
`_subfiles/<stem>/<Node>-<field>.<ext>`, where `<stem>` is the workflow file's
base name.

Converting a Code node workflow with `--externalize 3` produces:

```text
out/
├── code.yaml
└── _subfiles/
    └── code/
        └── Code-jsCode.js
```

The `code.yaml` now references the code instead of inlining it:

```yaml
connections: {}
name: code-wf
nodes:
    - name: Code
      parameters:
        jsCode:
            $ref: _subfiles/code/Code-jsCode.js
      position:
        - 0
        - 0
      type: n8n-nodes-base.code
      typeVersion: 1
settings: {}
```

…and `_subfiles/code/Code-jsCode.js` holds the real source:

```js
const a = 1;
const b = 2;
const c = 3;
const d = 4;
const e = 5;
return [{json:{sum:a+b+c+d+e}}];
```

The `$ref` markers are **re-inlined automatically on read**. `apply`, `lint`, and
`restore` all resolve them back into the workflow before sending it, so a
directory of externalized files is a valid desired state. The benefit is review
quality: a one-line change to a Code node shows up as a one-line diff in a real
`.js` file, with syntax highlighting, instead of a change buried inside an escaped
JSON string.

## diff: a workflow against another source

`workflows diff` prints a unified diff of a workflow's **writable** content.
Read-only fields (id, active state, version, timestamps) are ignored, so the diff
shows only what a promotion would actually change. Compare against the same
workflow by name on another profile, or against a local file:

```bash
# Compare a workflow on the active instance against the same name on prod
n8nctl workflows diff 2tUt1wbLX592XDdX --to prod

# Compare it against a local file (the version in your repo)
n8nctl workflows diff 2tUt1wbLX592XDdX --file ./workflows/order-intake.json
```

| Flag | Meaning |
|---|---|
| `--to <profile>` | Compare against the same workflow name on another profile |
| `--file <path>` | Compare against a local workflow file |

This is the review step before a [`workflows sync`](beyond-api.md#promote-a-workflow-between-instances-workflows-sync)
or an `apply`: see precisely how dev differs from prod before promoting, so a
promotion is reviewable rather than blind. Because the comparison runs over the
same canonical, writable-only form `apply` uses to decide *unchanged*, an empty
diff means `apply` would skip the workflow.

## Backups as YAML with externalized code

[`backup`](beyond-api.md#snapshot-and-restore-an-instance-backup-restore)
snapshots an instance to a directory you can commit. The same `--format` and
`--externalize` options that `convert` exposes apply here, so a backup can be
written directly in the git-friendly form:

```bash
# Snapshot prod as YAML with code split into sibling files
n8nctl --profile prod backup --out ./backups/prod --format yaml --externalize 5
```

| Flag | Meaning |
|---|---|
| `--format json\|yaml` | Workflow file format in the backup (default `json`) |
| `--externalize <N>` | Split code fields longer than N lines into sibling files (0 = off) |

`restore` reads either format and re-inlines any externalized `$ref` fields before
sending each workflow, so a YAML-and-externalized backup round-trips cleanly.

A YAML backup with externalized code is the recommended seed for the GitOps loop:
commit it, edit the workflow files (and the broken-out code files) in pull
requests, lint in CI, and reconcile with `apply`.

```bash
# Seed a repo from prod, then drive it as code
n8nctl --profile prod backup --out ./n8n-state --format yaml --externalize 5
git -C ./n8n-state init && git -C ./n8n-state add -A
git -C ./n8n-state commit -m "n8n snapshot $(date -u +%F)"
```

!!! warning "Credential secrets are never in a backup"
    A backup records credential metadata only (id, name, type), never the stored
    secrets — the n8n API is write-only for them. `restore` and `apply` reference
    credentials by id, so matching credentials must already exist on the target
    instance.

## See also

- [Beyond the API](beyond-api.md) — `workflows sync`, `backup` / `restore`, and
  `workflows search`.
- [Multi-instance and profiles](profiles.md) — how named profiles make the
  multi-instance promotion above a single flag.
- [vs. other n8n CLIs](comparison.md) — how `n8nctl`, the official `@n8n/cli`,
  `ubie-oss`, and `edenreich` compare by lane.
