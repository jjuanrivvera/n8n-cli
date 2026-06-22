---
title: n8nctl auth login
---

## n8nctl auth login

Store and verify an API key for the active profile

### Synopsis

Stores the API key in your OS keyring and verifies it against the instance.
Get a key from n8n > Settings > n8n API.

```
n8nctl auth login [flags]
```

### Options

```
      --api-key string    API key (otherwise prompted without echo)
      --base-url string   instance base URL to store for this profile
  -h, --help              help for login
```

### Options inherited from parent commands

```
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --jq string         apply a jq program to the result (e.g. '.[].id'); implies JSON input
      --no-color          disable colored output [env: NO_COLOR]
      --no-header         hide the table header row
  -o, --output string     output format: table|json|yaml|csv [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl auth](n8nctl_auth.md)	 - Authenticate against an n8n instance

