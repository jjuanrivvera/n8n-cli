---
title: n8nctl projects add-member
---

## n8nctl projects add-member

Add a user to a project

```
n8nctl projects add-member <projectId> --user <userId> --role <role> [flags]
```

### Options

```
  -h, --help          help for add-member
      --role string   project role, e.g. project:viewer|project:editor|project:admin (required)
      --user string   user id to add (required)
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

* [n8nctl projects](n8nctl_projects.md)	 - Manage projects and their members

