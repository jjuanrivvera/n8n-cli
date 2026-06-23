# Getting started

## Install

```bash
# Homebrew (macOS/Linux)
brew install jjuanrivvera/n8n-cli/n8nctl-cli

# From source (needs Go)
go install github.com/jjuanrivvera/n8n-cli/cmd/n8nctl@latest
```

Prebuilt binaries for macOS, Linux, and Windows are attached to each
[release](https://github.com/jjuanrivvera/n8n-cli/releases/latest). Download the
archive for your platform, put the `n8nctl` binary on your `PATH`, and confirm
it runs:

```bash
n8nctl version
```

## Get an API key

`n8nctl` authenticates with an n8n API key over the `X-N8N-API-KEY` header. In
the n8n UI, open **Settings → n8n API** and create a key. Copy both the key and
your instance URL; you will need the API base URL in the form
`<your-host>/api/v1` (for example `https://n8n.lan/api/v1`).

See [Authentication](authentication.md) for where the key is stored and how to
override it.

## Run init

`init` is the friendliest first run. It names a profile, captures the base URL
and API key, stores the key in your OS keyring, verifies connectivity, and writes
the config file.

```bash
n8nctl init
```

It prompts for a profile name, the base URL, and the API key (typed without
echo). To script it, pass the values as flags instead:

```bash
n8nctl init --profile homelab --base-url https://n8n.lan/api/v1 --api-key "$KEY"
```

## First commands

```bash
# List the first page of workflows as a table
n8nctl workflows list

# Get one workflow as JSON
n8nctl workflows get 42 -o json

# Confirm auth works against the active instance
n8nctl auth status

# Diagnose config, credentials, and connectivity if something is off
n8nctl doctor
```

## Where things live

- Config file: `~/.n8nctl-cli/config.yaml` (base URLs, defaults, aliases). No
  secrets.
- API keys: your OS keyring (service `n8nctl-cli`, account = profile name).

## Next steps

- See everything `n8nctl` can do: [Features](features.md).
- Solve a common task: [Recipes](recipes.md).
- Run more than one instance: [Multi-instance and profiles](profiles.md).
- Shape the output: [Output and filtering](output.md).
- Browse every command and flag: [Command reference](commands/index.md).
