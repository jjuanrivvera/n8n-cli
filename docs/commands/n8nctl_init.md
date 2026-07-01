---
title: n8nctl init
---

## n8nctl init

Interactive first-run setup for an instance/profile

### Synopsis

Walks you through naming an instance (profile), setting its base URL, capturing
an API key (stored in your OS keyring), verifying connectivity, and writing config.

```
n8nctl init [flags]
```

### Options

```
      --api-key string    API key (otherwise prompted without echo)
      --base-url string   instance base URL
  -h, --help              help for init
      --instance string   instance (profile) name to create/update
```

### Options inherited from parent commands

```
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
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

