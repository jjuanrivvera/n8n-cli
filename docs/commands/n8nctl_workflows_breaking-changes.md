---
title: n8nctl workflows breaking-changes
---

## n8nctl workflows breaking-changes

Find nodes pinned to an outdated typeVersion (upgrade risk)

### Synopsis

Compare each workflow's nodes against the embedded node catalog and report
those pinned to an older typeVersion than the latest known one, along with any
parameters they use that the latest schema no longer defines (renamed/removed).
Community/custom nodes are skipped. Informational — exits 0.

```
n8nctl workflows breaking-changes [--dir <dir> | -f <file>... | --remote | <id>] [flags]
```

### Examples

```
  n8nctl workflows breaking-changes --dir ./workflows
  n8nctl workflows breaking-changes 42
  n8nctl workflows breaking-changes --remote
```

### Options

```
      --dir string     scan all workflow files in a directory
  -f, --file strings   scan specific files
  -h, --help           help for breaking-changes
      --remote         scan live workflows from the instance
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

