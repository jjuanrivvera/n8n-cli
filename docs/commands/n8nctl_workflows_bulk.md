---
title: n8nctl workflows bulk
---

## n8nctl workflows bulk

Bulk activate/deactivate workflows by tag

### Synopsis

Flip every workflow carrying a tag in one command — useful for maintenance
windows (deactivate the `prod` set, do the work, reactivate). Always previews;
pass --yes to skip the confirmation or --dry-run to only list.

### Options

```
  -h, --help   help for bulk
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

* [n8nctl workflows](n8nctl_workflows.md)	 - Manage workflows
* [n8nctl workflows bulk activate](n8nctl_workflows_bulk_activate.md)	 - activate every workflow carrying a tag
* [n8nctl workflows bulk deactivate](n8nctl_workflows_bulk_deactivate.md)	 - deactivate every workflow carrying a tag

