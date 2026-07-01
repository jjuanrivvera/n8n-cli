---
title: n8nctl
---

## n8nctl

Control any n8n instance from the terminal via its public API

### Synopsis

n8nctl is a portable, single-binary client for the n8n public REST API.

It manages workflows, executions, credentials, tags, variables, projects and
users on any n8n instance — self-hosted or Cloud — over HTTPS with an API key.

Multi-instance is first class: define one named profile per instance, store each
instance's API key in your OS keyring, and switch with --instance or
"n8nctl config use <name>".

### Options

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
  -h, --help              help for n8nctl
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

* [n8nctl agent](n8nctl_agent.md)	 - Helpers for running n8nctl under an AI agent
* [n8nctl alias](n8nctl_alias.md)	 - Define command shortcuts expanded before parsing
* [n8nctl api](n8nctl_api.md)	 - Make a raw authenticated API request (escape hatch)
* [n8nctl audit](n8nctl_audit.md)	 - Generate a security audit of the instance
* [n8nctl auth](n8nctl_auth.md)	 - Authenticate against an n8n instance
* [n8nctl backup](n8nctl_backup.md)	 - Export workflows, tags, and variables to a directory (JSON or YAML)
* [n8nctl completion](n8nctl_completion.md)	 - Generate a shell completion script
* [n8nctl config](n8nctl_config.md)	 - Inspect and edit configuration and profiles
* [n8nctl credentials](n8nctl_credentials.md)	 - Manage credentials
* [n8nctl data-tables](n8nctl_data-tables.md)	 - Manage data tables and their rows
* [n8nctl doctor](n8nctl_doctor.md)	 - Diagnose configuration, credentials, and connectivity
* [n8nctl executions](n8nctl_executions.md)	 - Inspect and control workflow executions
* [n8nctl init](n8nctl_init.md)	 - Interactive first-run setup for an instance/profile
* [n8nctl login](n8nctl_login.md)	 - Authenticate the active profile (alias for `auth login`)
* [n8nctl logout](n8nctl_logout.md)	 - Remove the active profile's API key (alias for `auth logout`)
* [n8nctl mcp](n8nctl_mcp.md)	 - MCP server management
* [n8nctl nodes](n8nctl_nodes.md)	 - Explore the catalog of n8n node types (offline)
* [n8nctl packages](n8nctl_packages.md)	 - Export and import workflows as .n8np packages (beta)
* [n8nctl projects](n8nctl_projects.md)	 - Manage projects and their members
* [n8nctl proxy](n8nctl_proxy.md)	 - Run a local n8n API proxy that lint-gates workflow writes
* [n8nctl restore](n8nctl_restore.md)	 - Recreate workflows from a backup directory
* [n8nctl skills](n8nctl_skills.md)	 - Install this CLI's AI-agent skill into Claude, Cursor, and other agents
* [n8nctl source-control](n8nctl_source-control.md)	 - Interact with the Source Control (Git) integration
* [n8nctl stats](n8nctl_stats.md)	 - One-shot instance health summary
* [n8nctl tags](n8nctl_tags.md)	 - Manage workflow tags
* [n8nctl templates](n8nctl_templates.md)	 - Browse and deploy workflows from the n8n template gallery
* [n8nctl users](n8nctl_users.md)	 - Manage users (instance owner only)
* [n8nctl variables](n8nctl_variables.md)	 - Manage instance variables
* [n8nctl version](n8nctl_version.md)	 - Print version, commit, and build date
* [n8nctl workflows](n8nctl_workflows.md)	 - Manage workflows

