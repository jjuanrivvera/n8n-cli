---
title: n8nctl workflows autofix
---

## n8nctl workflows autofix

Auto-repair common workflow mistakes in files

### Synopsis

Apply mechanical fixes to workflow files: correct typo'd node types (against
the embedded node catalog), add the leading '=' to expression strings that are
missing it, and generate a webhookId for webhook/form-trigger nodes that lack one.

By default it reports what it would change; pass --write to apply the fixes.

```
n8nctl workflows autofix [-f <file>... | --dir <dir>] [flags]
```

### Examples

```
  n8nctl workflows autofix --dir ./workflows
  n8nctl workflows autofix -f wf.json --write
```

### Options

```
      --dir string     fix all workflow files in a directory
  -f, --file strings   workflow files to fix
  -h, --help           help for autofix
      --write          write the fixes back (default: report only)
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

