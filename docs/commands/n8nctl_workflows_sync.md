---
title: n8nctl workflows sync
---

## n8nctl workflows sync

Promote a workflow to another instance (profile)

### Synopsis

Copy a workflow from one instance to another over the API. Read-only fields
(id, active state, version) are stripped; nodes, connections and settings are
carried over. By default a new workflow is created on the destination; use
--update-by-name to overwrite an existing workflow with the same name.

  n8nctl workflows sync 2tUt1wbLX592XDdX --from dev --to prod --update-by-name --activate

Credentials are referenced by id and are NOT copied — create matching
credentials on the destination first (see `n8nctl credentials`).

```
n8nctl workflows sync <id> --to <profile> [flags]
```

### Options

```
      --activate         activate the workflow on the destination after syncing
      --from string      source profile (default: active profile)
  -h, --help             help for sync
      --to string        destination profile (required)
      --update-by-name   overwrite an existing destination workflow with the same name
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --jq string         apply a jq program to the result (e.g. '.[].id'); implies JSON input
      --no-color          disable colored output [env: NO_COLOR]
      --no-header         hide the table header row
  -o, --output string     output format: table|json|yaml|csv|id [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl workflows](n8nctl_workflows.md)	 - Manage workflows

