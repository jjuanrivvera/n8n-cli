---
title: n8nctl executions prune
---

## n8nctl executions prune

Bulk-delete executions by age and/or status

### Synopsis

Delete execution records older than a cutoff and/or matching a status, to
reclaim database space. Always previews the count first; pass --yes to skip the
confirmation, or --dry-run to only count.

```
n8nctl executions prune [flags]
```

### Examples

```
  n8nctl executions prune --older-than 30d
  n8nctl executions prune --older-than 7d --status error --yes
```

### Options

```
  -h, --help                help for prune
      --older-than string   delete executions older than this (e.g. 30d, 720h, 90m)
      --status string       only delete this status (error, success, ...)
      --workflow string     only delete executions of this workflow id
  -y, --yes                 skip the confirmation prompt
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

* [n8nctl executions](n8nctl_executions.md)	 - Inspect and control workflow executions

