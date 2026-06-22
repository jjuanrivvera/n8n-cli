---
title: n8nctl variables
---

## n8nctl variables

Manage instance variables

### Synopsis

Create, list, update and delete variables. The API has no get-by-id endpoint,
so `get <id>` is served by matching id or key within the full list.

  n8nctl variables create --set key=API_BASE --set value=https://api.example.com

### Options

```
  -h, --help   help for variables
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
* [n8nctl variables create](n8nctl_variables_create.md)	 - Create a variable
* [n8nctl variables delete](n8nctl_variables_delete.md)	 - Delete a variable
* [n8nctl variables get](n8nctl_variables_get.md)	 - Get a single variable by id
* [n8nctl variables list](n8nctl_variables_list.md)	 - List variables
* [n8nctl variables update](n8nctl_variables_update.md)	 - Update a variable

