---
title: n8nctl templates deploy
---

## n8nctl templates deploy

Create a workflow on the active instance from a template

### Synopsis

Fetch a gallery template and create it as a new workflow on the active
instance. Credentials are NOT included — open the workflow and connect them
afterwards. Honors --dry-run.

```
n8nctl templates deploy <id> [flags]
```

### Examples

```
  n8nctl templates deploy 1750 --name "My Slack bot"
  n8nctl --profile dev templates deploy 1750 --activate
```

### Options

```
      --activate      activate the workflow after creating it
  -h, --help          help for deploy
      --name string   name for the new workflow (default: the template's name)
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

* [n8nctl templates](n8nctl_templates.md)	 - Browse and deploy workflows from the n8n template gallery

