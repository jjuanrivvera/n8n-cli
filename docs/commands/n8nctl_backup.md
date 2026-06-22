---
title: n8nctl backup
---

## n8nctl backup

Export workflows, tags, and variables to a directory of JSON

### Synopsis

Snapshot the active instance to disk for git-based versioning and backup.
Writes one file per workflow plus tags.json, variables.json, a credentials
inventory (metadata only — secrets are never exported), and a manifest.

  n8nctl backup --out ./n8n-backup
  n8nctl --profile prod backup --out ./backups/prod

```
n8nctl backup --out <dir> [flags]
```

### Options

```
  -h, --help         help for backup
      --out string   output directory (required)
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

