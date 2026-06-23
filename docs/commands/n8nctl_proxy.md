---
title: n8nctl proxy
---

## n8nctl proxy

Run a local n8n API proxy that lint-gates workflow writes

### Synopsis

Stand a reverse proxy in front of the active instance that lints every
workflow create/update and rejects failures with HTTP 422 before they reach
n8n. This makes linting structural rather than a convention: anything that
pushes a workflow through the proxy — a human, a script, or an AI agent —
is held to the same rules, so a definition with errors can never land.

Point your n8n client at the proxy as if it were the instance host:
  n8nctl proxy &                       # listens on 127.0.0.1:8099
  export N8N_API_URL=http://127.0.0.1:8099   # for any n8n client
  n8nctl --base-url http://127.0.0.1:8099 workflows create --file wf.json

Reads pass straight through. The proxy injects the active profile's API key
(from your keyring), so the client never needs it. Bind to localhost only
unless you understand that the proxy is an authenticated gateway to n8n.

```
n8nctl proxy [flags]
```

### Options

```
      --block-destructive        also reject workflow DELETE requests
      --disable-rule strings     lint rules to disable (comma-separated)
  -h, --help                     help for proxy
      --listen string            address to listen on (default "127.0.0.1:8099")
      --reject-duplicate-names   reject creating a workflow whose name already exists
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

