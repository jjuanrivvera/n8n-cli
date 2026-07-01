---
title: n8nctl workflows apply
---

## n8nctl workflows apply

Reconcile a directory of workflow files into the instance (GitOps)

### Synopsis

Treat a directory of workflow files (JSON/YAML) as the desired state and
apply it: create new workflows, update existing ones (matched by name), and
with --prune, delete instance workflows not present in the directory.

Combine with profiles to promote the same desired state across instances:
  n8nctl --profile staging workflows apply --dir ./workflows
  n8nctl --profile prod    workflows apply --dir ./workflows --prune

Workflows are matched by name (the only stable handle the API exposes), so
renaming a file creates a new workflow and, with --prune, deletes the old one.
Duplicate names on the instance are skipped to avoid acting on the wrong one.
Reconcile covers name, nodes, connections and settings; runtime-only fields
(pinData, meta) are not managed.

Always preview with --dry-run first, especially with --prune.

```
n8nctl workflows apply --dir <dir> [flags]
```

### Options

```
      --activate     activate newly created workflows
  -d, --dir string   directory of workflow files (required)
  -h, --help         help for apply
      --prune        delete instance workflows not present in the directory
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --instance string   n8n instance to use: a named profile [env: N8NCTL_INSTANCE, N8NCTL_PROFILE]
      --jq string         apply a jq program to the result (e.g. '.[].id'); implies JSON input
      --no-color          disable colored output [env: NO_COLOR]
      --no-header         hide the table header row
  -o, --output string     output format: table|json|yaml|csv|id [env: N8NCTL_OUTPUT]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl workflows](n8nctl_workflows.md)	 - Manage workflows

