---
title: n8nctl data-tables upsert-rows
---

## n8nctl data-tables upsert-rows

Insert or update rows (body: {filter, data})

```
n8nctl data-tables upsert-rows <tableId> [flags]
```

### Options

```
      --data string   inline JSON body
      --file string   read JSON body from a file ('-' for stdin)
  -h, --help          help for upsert-rows
      --stdin         read JSON body from stdin
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

