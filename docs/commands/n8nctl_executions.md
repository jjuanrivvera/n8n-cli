---
title: n8nctl executions
---

## n8nctl executions

Inspect and control workflow executions

### Synopsis

Executions are read-only with retry/stop actions — n8n creates them by running workflows.

### Options

```
  -h, --help   help for executions
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

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl executions delete](n8nctl_executions_delete.md)	 - Delete a execution
* [n8nctl executions get](n8nctl_executions_get.md)	 - Get a single execution by id
* [n8nctl executions list](n8nctl_executions_list.md)	 - List executions
* [n8nctl executions prune](n8nctl_executions_prune.md)	 - Bulk-delete executions by age and/or status
* [n8nctl executions retry](n8nctl_executions_retry.md)	 - Retry a failed execution
* [n8nctl executions stop](n8nctl_executions_stop.md)	 - Stop a running execution
* [n8nctl executions watch](n8nctl_executions_watch.md)	 - Live-tail new executions, highlighting failures

