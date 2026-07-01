---
title: n8nctl packages import
---

## n8nctl packages import

Import a .n8np package into a project

```
n8nctl packages import --file <file.n8np> --conflict-policy <policy> [flags]
```

### Options

```
      --conflict-policy string            workflow conflict policy (required), e.g. fail|new-version
      --credential-matching-mode string   credential matching mode (id-only)
      --credential-missing-mode string    credential missing mode
      --file string                       path to the .n8np package (required)
      --folder string                     destination folder id
  -h, --help                              help for import
      --project string                    destination project id (default: personal project)
      --workflow-id-policy string         workflow id policy, e.g. new
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

* [n8nctl packages](n8nctl_packages.md)	 - Export and import workflows as .n8np packages (beta)

