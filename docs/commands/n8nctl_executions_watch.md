---
title: n8nctl executions watch
---

## n8nctl executions watch

Live-tail new executions, highlighting failures

### Synopsis

Poll the executions endpoint and print each new run as it appears, coloring
failures. Runs until interrupted (Ctrl-C).

```
n8nctl executions watch [flags]
```

### Examples

```
  n8nctl executions watch
  n8nctl executions watch --status error --interval 10s
```

### Options

```
  -h, --help                help for watch
      --interval duration   poll interval (default 5s)
      --status string       only watch this status
      --workflow string     only watch executions of this workflow id
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

* [n8nctl executions](n8nctl_executions.md)	 - Inspect and control workflow executions

