---
title: n8nctl mcp
---

## n8nctl mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help   help for mcp
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
* [n8nctl mcp claude](n8nctl_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [n8nctl mcp cursor](n8nctl_mcp_cursor.md)	 - Manage Cursor MCP servers
* [n8nctl mcp start](n8nctl_mcp_start.md)	 - Start the MCP server
* [n8nctl mcp stream](n8nctl_mcp_stream.md)	 - Stream the MCP server over HTTP
* [n8nctl mcp tools](n8nctl_mcp_tools.md)	 - Export tools as JSON
* [n8nctl mcp vscode](n8nctl_mcp_vscode.md)	 - Manage VSCode MCP servers

