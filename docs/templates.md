# Template gallery

n8n publishes a public gallery of community workflow templates at
[n8n.io/workflows](https://n8n.io/workflows), served by `api.n8n.io`. `n8nctl`
lets you search that gallery, inspect a template's definition, and deploy one
straight into an instance without leaving the terminal.

Two endpoints are in play, and the distinction matters:

- **The gallery (`api.n8n.io`)** is a public, read-only catalog. `search` and
  `get` talk to it and **need no API key**, no profile, and no instance — they
  work before you have configured `n8nctl` at all.
- **The active instance (`<host>/api/v1`)** is where `deploy` writes. It creates
  a workflow on the instance selected by your active profile (or `--profile`),
  using the key in your keyring, exactly like `workflows create`.

## search: find a template

`templates search <query>` queries the gallery and returns matching templates
with their id, name, and view count. Use the id with `get` or `deploy`.

```bash
# Search by keyword
n8nctl templates search slack

# Multi-word query, capped to the top results
n8nctl templates search "google sheets" --limit 5
```

```text
ID     NAME                                                          TOTALVIEWS
12345  Notify Slack when a Google Sheets row is added                52
10174  Post daily standup reminders to Slack                         36
```

| Flag | Meaning |
|---|---|
| `--limit <N>` | Maximum number of results (default 20) |

`-o json` returns the full result objects for scripting.

## get: print a template's definition

`templates get <id>` prints the template's workflow definition. It writes the
definition to stdout, so pipe it to a file to save it, or to `jq` to inspect it.

```bash
# Inspect the definition
n8nctl templates get 1750 -o json

# Save it as a workflow file you can edit and lint
n8nctl templates get 1750 -o json > slack-bot.json
```

A saved definition is a plain workflow file: edit it, run `workflows lint` over
it, or feed it through `workflows apply` like any other file in a workflows-as-code
directory.

## deploy: create the template on an instance

`templates deploy <id>` fetches the gallery template and creates it as a new
workflow on the active instance. This is the one command in the group that
writes, so it honors `--dry-run` and uses the active profile's stored key.

```bash
# Deploy into the active instance under the template's own name
n8nctl templates deploy 1750

# Give the new workflow a name of your own
n8nctl templates deploy 1750 --name "My Slack bot"

# Deploy into a specific instance and activate it on arrival
n8nctl --profile dev templates deploy 1750 --activate

# Preview the create call without sending it
n8nctl templates deploy 1750 --dry-run
```

| Flag | Meaning |
|---|---|
| `--name <name>` | Name for the new workflow (default: the template's name) |
| `--activate` | Activate the workflow after creating it |
| `--dry-run` (global) | Print the equivalent create call and send no request |

!!! warning "Credentials are not included"
    A gallery template carries node definitions, not secrets. `deploy` creates the
    workflow with its credential **references** but no credential values, so the
    new workflow cannot run until you open it and connect the credentials each
    node needs. Deploy first, then wire credentials, then activate.

!!! tip "Deploy into a dev instance first"
    Because `deploy` targets whichever profile is active, the safe pattern is to
    deploy into a throwaway or dev profile, inspect and adapt the workflow there,
    and only then promote it with [`workflows sync`](beyond-api.md) or
    [`workflows apply`](workflows-as-code.md).

## See also

- [Node catalog](features.md#node-catalog) — browse the embedded node definitions
  a template's nodes are built from.
- [Workflows as Code](workflows-as-code.md) — lint, convert, and reconcile a saved
  template into an instance.
- [Command reference](commands/n8nctl_templates.md) — every flag and argument.
