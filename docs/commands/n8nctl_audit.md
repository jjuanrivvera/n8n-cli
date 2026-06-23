---
title: n8nctl audit
---

## n8nctl audit

Generate a security audit of the instance

### Synopsis

Run n8n's built-in security audit and print the report.

  n8nctl audit
  n8nctl audit --categories credentials,nodes --days 30 -o json

```
n8nctl audit [flags]
```

### Options

```
      --categories strings   restrict to categories: credentials,database,nodes,filesystem,instance
      --days int             days of inactivity before a workflow is flagged as abandoned
  -h, --help                 help for audit
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

