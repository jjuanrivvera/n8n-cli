---
title: n8nctl workflows search
---

## n8nctl workflows search

Find workflows by node type, credential, webhook path, or name

### Synopsis

Scan all workflows and report those matching a filter:
  --node <type>        substring match on node type (e.g. slack, httpRequest)
  --credential <id|nm> workflows referencing a credential by id or name
  --webhook <path>     workflows with a webhook node on that path
  --name <regex>       workflow name matches a regular expression

  n8nctl workflows search --node slack
  n8nctl workflows search --credential githubApi -o json
  n8nctl workflows search --webhook /orders

```
n8nctl workflows search [flags]
```

### Options

```
      --credential string   match workflows referencing this credential id or name
  -h, --help                help for search
      --name string         match workflow name against a regular expression
      --node string         match a node type substring (e.g. slack, httpRequest)
      --webhook string      match workflows with a webhook node on this path
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

