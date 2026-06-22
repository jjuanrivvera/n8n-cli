# Multi-instance and profiles

Most people who automate n8n end up with more than one instance: a homelab box
for tinkering, an n8n Cloud tenant for production, and maybe a client's server.
The official `@n8n/cli` targets one instance at a time, so juggling three means
swapping environment variables by hand and hoping you exported the right one.

`n8nctl` makes instances first-class. Each instance is a named **profile**: its
base URL, its output preferences, and a pointer to a keyring entry holding its
API key. You switch between them in one word and never copy a secret around.

## The config file

Config lives at `~/.n8nctl-cli/config.yaml`. API keys are not in it; they are in
your OS keyring. A three-instance setup looks like this:

```yaml
default_profile: homelab

profiles:
  homelab:
    base_url: https://n8n.lan/api/v1
    description: Self-hosted box on the LAN
  cloud:
    base_url: https://acme.app.n8n.cloud/api/v1
    description: Production n8n Cloud tenant
  client:
    base_url: https://n8n.client.com/api/v1
    description: Client instance

settings:
  output_format: table
  requests_per_second: 0
  log_level: warn

aliases:
  failures: executions list --status error --all
```

`default_profile` is the instance used when you do not pass `--profile`. The
per-instance API keys are stored separately under keyring service `n8nctl-cli`
with the account set to the profile name (`homelab`, `cloud`, `client`).

## Creating instances

```bash
n8nctl init --profile homelab --base-url https://n8n.lan/api/v1
n8nctl init --profile cloud   --base-url https://acme.app.n8n.cloud/api/v1
n8nctl init --profile client  --base-url https://n8n.client.com/api/v1
```

Each `init` captures that instance's API key into the keyring and verifies it.

## Switching between instances

There are three ways to choose an instance, in order of precedence:

```bash
# 1. The --profile flag wins, scoped to a single command
n8nctl --profile cloud workflows list
n8nctl --profile client executions list --status error

# 2. The N8NCTL_PROFILE env var, scoped to a command or shell session
N8NCTL_PROFILE=cloud n8nctl workflows list
export N8NCTL_PROFILE=client   # for the rest of the session

# 3. The default_profile in config, used when neither of the above is set
n8nctl config use homelab      # change the default
n8nctl workflows list          # now runs against homelab
```

A `--profile` flag always beats `N8NCTL_PROFILE`, which always beats the
`default_profile` in config. This holds for every command.

## Inspecting what is configured

```bash
n8nctl config list-profiles    # the names and base URLs you have
n8nctl config view             # the fully resolved config, secrets redacted
n8nctl config path             # where the file lives
```

## Editing profiles

```bash
# Change a profile field (base_url, description)
n8nctl --profile client config set base_url https://new.client.com/api/v1

# Change a global setting (output_format, requests_per_second, log_level)
n8nctl config set output_format json

# Re-capture or rotate an instance's API key
n8nctl --profile cloud auth login
```

## Why this shape

Keeping base URLs in a readable file and secrets in the keyring means you can
commit a sanitized config, share it across machines, or diff it, without ever
exposing a key. Naming instances rather than exporting environment variables per
command removes the most common operational mistake: running a destructive
command against the wrong server. With `n8nctl` the target is written into the
command (`--profile client`), where you can see it before you press enter, or
review it with `--dry-run`.
