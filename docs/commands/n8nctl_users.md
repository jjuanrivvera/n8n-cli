---
title: n8nctl users
---

## n8nctl users

Manage users (instance owner only)

### Synopsis

List and inspect users, invite new ones, change roles, and delete users.
n8n invites users rather than creating them directly, so use `users invite`.

### Options

```
  -h, --help   help for users
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
* [n8nctl users change-role](n8nctl_users_change-role.md)	 - Change a user's global role
* [n8nctl users delete](n8nctl_users_delete.md)	 - Delete a user
* [n8nctl users get](n8nctl_users_get.md)	 - Get a single user by id
* [n8nctl users invite](n8nctl_users_invite.md)	 - Invite one or more users by email
* [n8nctl users list](n8nctl_users_list.md)	 - List users

