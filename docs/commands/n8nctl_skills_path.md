---
title: n8nctl skills path
---

## n8nctl skills path

Print where the skill would be installed

```
n8nctl skills path [flags]
```

### Options

```
      --agent string   Target agent: claude, codex, copilot, cursor, gemini, opencode, windsurf (default "claude")
  -g, --global         User-level path
  -h, --help           help for path
```

### Options inherited from parent commands

```
      --api-key string    override the API key (prefer keyring via 'auth login')
      --base-url string   override the instance base URL (e.g. https://host/api/v1)
      --columns strings   comma-separated columns for table/csv output
      --dry-run           print the equivalent curl and send no request
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

* [n8nctl skills](n8nctl_skills.md)	 - Install this CLI's AI-agent skill into Claude, Cursor, and other agents

