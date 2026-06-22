# Security Policy

`n8nctl` talks to the n8n workflow automation API and handles **API keys and
credentials**, so we take security seriously.

## Supported versions

Only the latest released `vX.Y.Z` receives security fixes. Please reproduce on
the latest release (or `main`) before reporting.

## Reporting a vulnerability

**Do not open a public issue, PR, or discussion for security problems.**

Report privately through GitHub's **[Private vulnerability reporting](https://github.com/jjuanrivvera/n8n-cli/security/advisories/new)**
(repo → **Security** → **Report a vulnerability**). If that is unavailable,
contact the maintainer privately via GitHub ([@jjuanrivvera99](https://github.com/jjuanrivvera99)).

Please include:

- a description of the issue and its impact,
- steps to reproduce (a minimal command sequence or PoC),
- affected version (`n8nctl version`) and OS,
- any logs — **with API keys/credentials redacted**.

## What to expect

- Acknowledgement within **5 business days**.
- An initial assessment and severity within **10 business days**.
- A fix released as promptly as the severity warrants, with credit in the
  release notes (unless you prefer to remain anonymous). We follow coordinated
  disclosure: please give us a reasonable window before publishing details.

## Handling credentials safely

- **API keys live in the OS keyring.** `n8nctl auth login` stores the n8n API
  key in the operating-system keyring (Keychain on macOS, Secret Service /
  libsecret on Linux, Credential Manager on Windows) — never in the config file
  in plaintext. Only the instance URL and non-secret profile data are written to
  `~/.n8n-cli/config.yaml`.
- **Keys are never logged or committed.** The CLI redacts the `Authorization` /
  `X-N8N-API-KEY` header in `--dry-run` curl previews and verbose output. Do not
  pass any flag that reveals the raw key in shared logs.
- **When sharing output or a `--dry-run` curl for a bug report, never include a
  real API key.** Treat the key like a password: keep it out of shell history,
  CI logs, and committed files. Prefer the `N8N_API_KEY` environment variable or
  the keyring over config-in-repo.
