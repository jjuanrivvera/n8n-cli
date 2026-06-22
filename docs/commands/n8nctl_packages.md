---
title: n8nctl packages
---

## n8nctl packages

Export and import workflows as .n8np packages (beta)

### Synopsis

Bundle workflows into a portable .n8np archive and import them elsewhere.
This is a beta n8n feature, disabled by default; the API returns 404 unless
the instance sets N8N_PUBLIC_API_PACKAGES_ENABLED=true.

### Options

```
  -h, --help   help for packages
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
* [n8nctl packages export](n8nctl_packages_export.md)	 - Export workflows as a .n8np package
* [n8nctl packages import](n8nctl_packages_import.md)	 - Import a .n8np package into a project

