---
title: n8nctl users list
---

## n8nctl users list

List users

```
n8nctl users list [flags]
```

### Options

```
      --all                   fetch every page (auto-paginate)
      --cursor string         pagination cursor from a previous response
  -h, --help                  help for list
      --include-role string   include each user's role (true/false)
      --limit int             max items to return (page size, capped at 250)
      --max-pages int         page cap for --all (0 = safety default)
      --param stringArray     extra query parameter key=value (repeatable)
      --project string        filter by project id
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

* [n8nctl users](n8nctl_users.md)	 - Manage users (instance owner only)

