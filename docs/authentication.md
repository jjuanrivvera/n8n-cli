# Authentication

`n8nctl` talks to the n8n public API with an API key sent in the
`X-N8N-API-KEY` header on every request. There is no OAuth flow and no
username/password; the key is the only credential.

## Get an API key

In the n8n UI, open **Settings → n8n API** and create a key. n8n shows the key
once, so copy it immediately. You also need your instance's API base URL in the
form `<your-host>/api/v1`, for example `https://n8n.lan/api/v1`.

## Store it in the keyring

The recommended path captures the key into your OS keyring, where it never
touches the config file:

```bash
# As part of first-run setup
n8nctl init

# Or sign in to the active profile directly
n8nctl auth login
```

`auth login` prompts for the key without echoing it, verifies it against the
instance, and stores it. Keys are kept in:

- macOS Keychain
- the GNOME/KDE Secret Service (libsecret) on Linux
- the Windows Credential Manager

The keyring entry uses service `n8nctl-cli` with the account set to the profile
name, so each instance has its own key under its own profile.

## Verify and remove

```bash
n8nctl auth status     # shows the active profile and whether its key works
n8nctl auth logout     # removes the stored key for the active profile
```

## Override with an environment variable

For CI and scripts, set the key (and base URL) in the environment instead of the
keyring:

```bash
export N8NCTL_API_KEY="your-api-key"
export N8NCTL_BASE_URL="https://n8n.lan/api/v1"

n8nctl workflows list
```

When `N8NCTL_API_KEY` is set it is used directly and the keyring is not consulted.
A `--api-key` flag overrides the environment for a single command. The resolution
order for the key is:

1. `--api-key` flag
2. `N8NCTL_API_KEY` environment variable
3. the OS keyring entry for the active profile

The base URL resolves similarly: `--base-url` flag, then `N8NCTL_BASE_URL`, then
the profile's `base_url` in config.

## A note on HTTPS

The API key is a bearer-style secret sent on every request. If you point a
profile at an `http://` URL, `n8nctl` warns you that the key will travel in clear
text. Use HTTPS for anything beyond a local test instance.

## Troubleshooting

```bash
n8nctl doctor          # checks config, the stored key, and connectivity
n8nctl auth status     # confirms the key authenticates against /workflows
```

If `auth status` fails, confirm the base URL ends in `/api/v1`, that the key is
still valid in **Settings → n8n API**, and that you are pointed at the intended
[profile](profiles.md).
