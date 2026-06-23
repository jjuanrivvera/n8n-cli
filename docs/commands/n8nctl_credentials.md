---
title: n8nctl credentials
---

## n8nctl credentials

Manage credentials

### Synopsis

Create, inspect, update, delete and transfer credentials. Secret values are
write-only: they are sent on create/update but never returned by the API.

Discover a type's required fields first:
  n8nctl credentials schema githubApi
  n8nctl credentials create --set name='My GH' --set type=githubApi --set data='{"accessToken":"..."}'

### Options

```
  -h, --help   help for credentials
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
* [n8nctl credentials create](n8nctl_credentials_create.md)	 - Create a credential
* [n8nctl credentials delete](n8nctl_credentials_delete.md)	 - Delete a credential
* [n8nctl credentials get](n8nctl_credentials_get.md)	 - Get a single credential by id
* [n8nctl credentials list](n8nctl_credentials_list.md)	 - List credentials
* [n8nctl credentials schema](n8nctl_credentials_schema.md)	 - Show the field schema for a credential type
* [n8nctl credentials transfer](n8nctl_credentials_transfer.md)	 - Transfer a credential to another project
* [n8nctl credentials update](n8nctl_credentials_update.md)	 - Update a credential

