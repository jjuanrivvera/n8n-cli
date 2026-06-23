---
title: n8nctl data-tables
---

## n8nctl data-tables

Manage data tables and their rows

### Synopsis

Create, list, inspect, update and delete data tables, and manage their rows.
Data tables may be unlicensed on some editions (the API returns 403).

  n8nctl data-tables create --set name=orders --set 'columns=[{"name":"sku","type":"string"}]'
  n8nctl data-tables rows <id> --filter '{"type":"and","filters":[]}'

### Options

```
  -h, --help   help for data-tables
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
* [n8nctl data-tables add-rows](n8nctl_data-tables_add-rows.md)	 - Add rows (body: a JSON array of row objects)
* [n8nctl data-tables create](n8nctl_data-tables_create.md)	 - Create a data-table
* [n8nctl data-tables delete](n8nctl_data-tables_delete.md)	 - Delete a data-table
* [n8nctl data-tables delete-rows](n8nctl_data-tables_delete-rows.md)	 - Delete rows matching a filter
* [n8nctl data-tables get](n8nctl_data-tables_get.md)	 - Get a single data-table by id
* [n8nctl data-tables list](n8nctl_data-tables_list.md)	 - List data-tables
* [n8nctl data-tables rows](n8nctl_data-tables_rows.md)	 - List rows in a data table
* [n8nctl data-tables update](n8nctl_data-tables_update.md)	 - Update a data-table
* [n8nctl data-tables update-rows](n8nctl_data-tables_update-rows.md)	 - Update rows matching a filter (body: {filter, data})
* [n8nctl data-tables upsert-rows](n8nctl_data-tables_upsert-rows.md)	 - Insert or update rows (body: {filter, data})

