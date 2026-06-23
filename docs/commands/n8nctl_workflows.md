---
title: n8nctl workflows
---

## n8nctl workflows

Manage workflows

### Synopsis

Create, list, inspect, update, delete, activate and transfer n8n workflows.

Create from a JSON file exported by n8n:
  n8nctl workflows create --file workflow.json
A workflow body requires name, nodes, connections and settings.

### Options

```
  -h, --help   help for workflows
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
* [n8nctl workflows activate](n8nctl_workflows_activate.md)	 - Activate a workflow
* [n8nctl workflows apply](n8nctl_workflows_apply.md)	 - Reconcile a directory of workflow files into the instance (GitOps)
* [n8nctl workflows archive](n8nctl_workflows_archive.md)	 - Archive a workflow
* [n8nctl workflows convert](n8nctl_workflows_convert.md)	 - Convert workflow files between JSON and YAML (local)
* [n8nctl workflows create](n8nctl_workflows_create.md)	 - Create a workflow
* [n8nctl workflows deactivate](n8nctl_workflows_deactivate.md)	 - Deactivate a workflow
* [n8nctl workflows delete](n8nctl_workflows_delete.md)	 - Delete a workflow
* [n8nctl workflows diff](n8nctl_workflows_diff.md)	 - Diff a workflow against another instance or a local file
* [n8nctl workflows get](n8nctl_workflows_get.md)	 - Get a single workflow by id
* [n8nctl workflows lint](n8nctl_workflows_lint.md)	 - Lint workflow definitions for common mistakes
* [n8nctl workflows list](n8nctl_workflows_list.md)	 - List workflows
* [n8nctl workflows search](n8nctl_workflows_search.md)	 - Find workflows by node type, credential, webhook path, or name
* [n8nctl workflows sync](n8nctl_workflows_sync.md)	 - Promote a workflow to another instance (profile)
* [n8nctl workflows tags](n8nctl_workflows_tags.md)	 - Get or replace a workflow's tags
* [n8nctl workflows transfer](n8nctl_workflows_transfer.md)	 - Transfer a workflow to another project
* [n8nctl workflows unarchive](n8nctl_workflows_unarchive.md)	 - Restore an archived workflow
* [n8nctl workflows update](n8nctl_workflows_update.md)	 - Update a workflow

