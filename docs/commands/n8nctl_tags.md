---
title: n8nctl tags
---

## n8nctl tags

Manage workflow tags

### Synopsis

Create, list, update and delete tags. Create with --set name=Production.

### Options

```
  -h, --help   help for tags
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --no-color          disable colored output [env: NO_COLOR]
  -o, --output string     output format: table|json|yaml|csv [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl tags create](n8nctl_tags_create.md)	 - Create a tag
* [n8nctl tags delete](n8nctl_tags_delete.md)	 - Delete a tag
* [n8nctl tags get](n8nctl_tags_get.md)	 - Get a single tag by id
* [n8nctl tags list](n8nctl_tags_list.md)	 - List tags
* [n8nctl tags update](n8nctl_tags_update.md)	 - Update a tag

