---
title: n8nctl workflows diff
---

## n8nctl workflows diff

Diff a workflow against another instance or a local file

### Synopsis

Show a unified diff of a workflow's writable content (read-only fields are
ignored). Compare the active instance's workflow against the same id on
another --profile, or against a local --file.

```
n8nctl workflows diff <id> [--to <profile> | --file <path>] [flags]
```

### Options

```
      --file string   compare against a local workflow file
  -h, --help          help for diff
      --to string     compare against the same workflow name on another profile
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

