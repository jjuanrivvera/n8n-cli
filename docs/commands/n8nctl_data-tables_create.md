---
title: n8nctl data-tables create
---

## n8nctl data-tables create

Create a data-table

```
n8nctl data-tables create [flags]
```

### Options

```
      --data string       inline JSON request body (e.g. '{"name":"x"}')
      --file string       read JSON request body from a file ('-' for stdin)
  -h, --help              help for create
      --set stringArray   set a field key=value (repeatable; value parsed as JSON when possible)
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

* [n8nctl data-tables](n8nctl_data-tables.md)	 - Manage data tables and their rows

