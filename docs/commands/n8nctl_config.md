---
title: n8nctl config
---

## n8nctl config

Inspect and edit configuration and profiles

### Synopsis

Manage the config file and named instance profiles. Secrets are redacted in `view`.

### Options

```
  -h, --help   help for config
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

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl config list-profiles](n8nctl_config_list-profiles.md)	 - List configured profiles (instances)
* [n8nctl config path](n8nctl_config_path.md)	 - Print the config file path
* [n8nctl config set](n8nctl_config_set.md)	 - Set a config value
* [n8nctl config use](n8nctl_config_use.md)	 - Switch the default profile (active instance)
* [n8nctl config view](n8nctl_config_view.md)	 - Show the resolved configuration (secrets redacted)

