---
title: n8nctl restore
---

## n8nctl restore

Recreate workflows from a backup directory

### Synopsis

Apply the workflows in a backup directory to the active instance. By default
each workflow is created new; --update-by-name overwrites an existing workflow
with the same name. Credentials are referenced by id and must already exist.

  n8nctl --profile staging restore --in ./n8n-backup --update-by-name

```
n8nctl restore --in <dir> [flags]
```

### Options

```
      --activate         activate each restored workflow
  -h, --help             help for restore
      --in string        backup directory to restore from (required)
      --update-by-name   overwrite existing workflows with the same name
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

