---
title: n8nctl update
---

## n8nctl update

Update n8nctl to the latest GitHub release

### Synopsis

Download the latest release from GitHub, verify it against checksums.txt,
and atomically replace the running binary.

A dev build (installed via "go install" or built from source) is never
self-updated; use your package manager or rebuild instead.

```
n8nctl update [flags]
```

### Options

```
  -h, --help   help for update
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --instance string   n8n instance to use: a named profile [env: N8NCTL_INSTANCE, N8NCTL_PROFILE]
      --jq string         apply a jq program to the result (e.g. '.[].id'); implies JSON input
      --no-color          disable colored output [env: NO_COLOR]
      --no-header         hide the table header row
  -o, --output string     output format: table|json|yaml|csv|id [env: N8NCTL_OUTPUT]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl update check](n8nctl_update_check.md)	 - Check for a newer release without installing it

