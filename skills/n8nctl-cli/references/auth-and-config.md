# n8nctl - auth & config deep dive

How `n8nctl` finds an instance, authenticates, and switches between many of them.
Authoritative docs: https://github.com/jjuanrivvera/n8n-cli

## 1. Get an API key from n8n

1. Open your n8n instance UI.
2. Go to **Settings > n8n API**.
3. Click **Create an API key**, give it a label, and copy the value (you see it
   once).

The key is sent on every request as the `X-N8N-API-KEY` header against
`<base-url>/api/v1`. Anyone with the key has the permissions of the user who
created it, so treat it like a password - store it in the keyring, never paste
it into a command or commit it.

## 2. Store the key: keyring vs config file

`n8nctl` separates the two halves of a connection:

- **Base URL** → lives in the config file `~/.n8nctl-cli/config.yaml`, per
  profile. Not a secret.
- **API key** → lives in your **OS keyring** (macOS Keychain, Linux Secret
  Service / libsecret, Windows Credential Manager), keyed by profile name. The
  config file does **not** hold the key by default, and `n8nctl config view`
  redacts it if you ever put one there.

Capture both with either of:

```bash
n8nctl init                          # guided: prompts for profile, base URL, key, then verifies
n8nctl auth login                    # store/verify the key for the active profile (prompts, no echo)
n8nctl auth login --base-url https://n8n.example.com --api-key "$KEY"  # non-interactive
```

`auth login` verifies the key against the instance before saving, so a bad key
fails fast. Remove a stored key with `n8nctl auth logout`. Check state with:

```bash
n8nctl auth status        # or: n8nctl whoami
# -> {"profile":"homelab","base_url":"https://n8n.lan/api/v1","key_present":true,"valid":true}
n8nctl doctor             # config file? base URL? key resolvable? auth works?
```

## 3. Environment variables

Every config value has an env override (`N8NCTL_` prefix):

| Env var | Overrides |
|---|---|
| `N8NCTL_PROFILE` | active profile name |
| `N8NCTL_BASE_URL` | base URL for the active profile |
| `N8NCTL_API_KEY` | the API key (skips the keyring) |
| `N8NCTL_OUTPUT` | output format (`table\|json\|yaml\|csv`) |
| `N8NCTL_RPS` | client-side rate limit (requests/sec) |
| `N8NCTL_LOG_LEVEL` | log level (`debug\|info\|warn\|error`) |
| `N8NCTL_CONFIG` | path to the config file |
| `NO_COLOR` | disable colored output (standard) |

`N8NCTL_API_KEY` is the right choice for CI and containers where no keyring
exists: set it in the environment (from a secret store) and skip `auth login`.

## 4. Profiles & multi-instance (the worked example)

A **profile** is one n8n instance: a name, a base URL (config file), and a key
(keyring). This is `n8nctl`'s core strength over the single-instance official
CLI - one binary drives every instance you own.

Set up three instances - a homelab box, an n8n Cloud tenant, and a client's
server:

```bash
# homelab (self-hosted)
n8nctl init --profile homelab --base-url https://n8n.lan

# n8n Cloud
n8nctl auth login --profile cloud --base-url https://yourco.app.n8n.cloud

# a client's instance
n8nctl auth login --profile acme --base-url https://automation.acme.com
```

List them and pick a default:

```bash
n8nctl config list-profiles      # alias: n8nctl config profiles
# profile   active  base_url                          has_key  description
# homelab   *       https://n8n.lan/api/v1            true
# cloud             https://yourco.app.n8n.cloud/...  true
# acme              https://automation.acme.com/...   true

n8nctl config use cloud          # switch the default instance
```

Use a different instance for a single command without changing the default:

```bash
n8nctl workflows list --profile acme
n8nctl executions list --status error --profile homelab
N8NCTL_PROFILE=cloud n8nctl workflows list      # via env, for a whole shell session
```

Each profile carries its own key in the keyring, so switching instances never
mixes credentials. Re-run `n8nctl auth status` after switching to confirm you're
pointed at the instance you think you are before any write.

## 5. The config file

Path resolution: `N8NCTL_CONFIG` if set, else
`$XDG_CONFIG_HOME/n8nctl-cli/config.yaml` when `XDG_CONFIG_HOME` is set, else
`~/.n8nctl-cli/config.yaml`. Inspect it with:

```bash
n8nctl config path               # print the file path
n8nctl config view               # resolved config, secrets redacted
```

Shape:

```yaml
default_profile: homelab
profiles:
  homelab:
    base_url: https://n8n.lan/api/v1
    description: home lab box
  cloud:
    base_url: https://yourco.app.n8n.cloud/api/v1
settings:
  output_format: table
  requests_per_second: 5
  log_level: warn
aliases:
  errors: executions list --status error
```

Edit it with `n8nctl config set`:

```bash
n8nctl config set output_format json          # global setting
n8nctl config set requests_per_second 3       # (alias key: rps)
n8nctl config set base_url https://new.host    # active profile's base URL
```

> A profile may optionally hold an `api_key` field in the file, but it is **not
> recommended** - the keyring is the secure default and `config view` redacts the
> field. Prefer `auth login` or `N8NCTL_API_KEY`.

## 6. Precedence: flag > env > file > default

For every value, `n8nctl` resolves in this order - the first that is set wins:

1. **Command-line flag** - `--base-url`, `--api-key`, `--profile`, `--output`,
   `--rps`.
2. **Environment variable** - `N8NCTL_BASE_URL`, `N8NCTL_API_KEY`,
   `N8NCTL_PROFILE`, `N8NCTL_OUTPUT`, `N8NCTL_RPS`.
3. **Config file** - the active profile's fields and the `settings:` block.
4. **Built-in default** - e.g. `table` output, the default rate limit.

The API key has one extra step: flag `--api-key` > `N8NCTL_API_KEY` > the
profile's keyring entry. This lets you override a single command's instance or
key without touching stored config.
