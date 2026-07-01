---
title: n8nctl workflows convert
---

## n8nctl workflows convert

Convert workflow files between JSON and YAML (local)

### Synopsis

Convert workflow definition files between JSON and YAML on disk. With
--externalize, long code fields (jsCode, query, jsonBody, ...) are split into
sibling files for cleaner review.

```
n8nctl workflows convert <file...> --to json|yaml [flags]
```

### Options

```
      --externalize int   externalize code fields longer than N lines (0 = off)
  -h, --help              help for convert
      --out string        output directory (default: alongside each input)
      --to string         target format: json or yaml (required)
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

* [n8nctl workflows](n8nctl_workflows.md)	 - Manage workflows

