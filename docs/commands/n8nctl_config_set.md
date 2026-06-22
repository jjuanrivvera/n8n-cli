---
title: n8nctl config set
---

## n8nctl config set

Set a config value

### Synopsis

Set a global setting or a field on the active profile.

Global keys:  output_format, requests_per_second, log_level
Profile keys: base_url, description

```
n8nctl config set <key> <value> [flags]
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
      --no-color          disable colored output [env: NO_COLOR]
  -o, --output string     output format: table|json|yaml|csv [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl config](n8nctl_config.md)	 - Inspect and edit configuration and profiles

