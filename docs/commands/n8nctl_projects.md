---
title: n8nctl projects
---

## n8nctl projects

Manage projects and their members

### Synopsis

Projects are an n8n Enterprise feature. Create with --set name='My Project'.

### Options

```
  -h, --help   help for projects
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
  -o, --output string     output format: table|json|yaml|csv [env: N8NCTL_OUTPUT]
      --profile string    config profile (instance) to use [env: N8NCTL_PROFILE]
  -q, --quiet             suppress non-essential chatter
      --rps float         client-side rate limit in requests/sec (0 = use config/default)
      --show-token        do not redact the API key in --dry-run output
  -v, --verbose           verbose (debug) logging to stderr
```

### SEE ALSO

* [n8nctl](n8nctl.md)	 - Control any n8n instance from the terminal via its public API
* [n8nctl projects add-member](n8nctl_projects_add-member.md)	 - Add a user to a project
* [n8nctl projects create](n8nctl_projects_create.md)	 - Create a project
* [n8nctl projects delete](n8nctl_projects_delete.md)	 - Delete a project
* [n8nctl projects get](n8nctl_projects_get.md)	 - Get a single project by id
* [n8nctl projects list](n8nctl_projects_list.md)	 - List projects
* [n8nctl projects members](n8nctl_projects_members.md)	 - List the members of a project
* [n8nctl projects remove-member](n8nctl_projects_remove-member.md)	 - Remove a user from a project
* [n8nctl projects set-member-role](n8nctl_projects_set-member-role.md)	 - Change a project member's role
* [n8nctl projects update](n8nctl_projects_update.md)	 - Update a project

