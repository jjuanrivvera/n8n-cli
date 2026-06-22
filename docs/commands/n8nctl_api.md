---
title: n8nctl api
---

## n8nctl api

Make a raw authenticated API request (escape hatch)

### Synopsis

Call any n8n API endpoint directly. PATH is relative to the instance base URL
(the leading /api/v1 is added automatically).

  n8nctl api GET /workflows -q limit=5
  n8nctl api POST /tags -d '{"name":"Prod"}'
  n8nctl api DELETE /executions/42 --dry-run

```
n8nctl api <METHOD> <PATH> [flags]
```

### Options

```
  -d, --data string         request body as inline JSON
      --file string         request body from a file ('-' for stdin)
  -h, --help                help for api
      --query stringArray   query parameter key=value (repeatable)
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
  -o, --output string     output format: table|json|yaml|csv [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API

