---
title: n8nctl templates
---

## n8nctl templates

Browse and deploy workflows from the n8n template gallery

### Synopsis

Search the public n8n template gallery (api.n8n.io), inspect a template's
workflow definition, and deploy one straight into the active instance.

### Options

```
  -h, --help   help for templates
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

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl templates deploy](n8nctl_templates_deploy.md)	 - Create a workflow on the active instance from a template
* [n8nctl templates get](n8nctl_templates_get.md)	 - Print a template's workflow definition
* [n8nctl templates search](n8nctl_templates_search.md)	 - Search the template gallery

