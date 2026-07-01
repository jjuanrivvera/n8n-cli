---
title: n8nctl source-control pull
---

## n8nctl source-control pull

Pull changes from the connected remote repository

### Synopsis

Requires the licensed Source Control feature connected to a repository.
Use --force to discard local changes on conflict.

```
n8nctl source-control pull [flags]
```

### Options

```
      --force              discard local changes / resolve conflicts in favor of remote
  -h, --help               help for pull
      --variables string   JSON object of variable overrides to apply during pull
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

* [n8nctl source-control](n8nctl_source-control.md)	 - Interact with the Source Control (Git) integration

