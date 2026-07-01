---
title: n8nctl workflows lint
---

## n8nctl workflows lint

Lint workflow definitions for common mistakes

### Synopsis

Static checks over workflow files (or live workflows with --remote):
required fields, dangling connections, orphaned nodes, missing webhookId,
and expression strings missing the leading '='. Exits non-zero on errors.

```
n8nctl workflows lint [--dir <dir> | -f <file>... | --remote] [flags]
```

### Options

```
      --dir string             lint all workflow files in a directory
      --disable-rule strings   rules to disable (comma-separated)
  -f, --file strings           lint specific files
  -h, --help                   help for lint
      --list-rules             list available rules and exit
      --remote                 lint live workflows from the instance
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

