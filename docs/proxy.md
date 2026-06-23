# Lint-enforcing proxy

`workflows lint` is only as good as the discipline to run it. `n8nctl proxy`
makes it **structural**: it stands a local reverse proxy in front of your n8n
instance that lints every workflow create and update, and rejects any with lint
errors before they reach n8n. A bad definition cannot land, whether it is pushed
by a person, a CI job, a script, or an AI agent.

The idea is borrowed from the server-side enforcement proxy in
[`ubie-oss/n8n-cli`](https://github.com/ubie-oss/n8n-cli).

## How it works

```
client ──▶ n8nctl proxy ──▶ n8n instance
              │
              ├─ POST /workflows, PUT/PATCH /workflows/{id}
              │     └─ lint the body → 422 if it has errors, otherwise forward
              ├─ GET / reads → forward unchanged
              └─ injects the active profile's API key (X-N8N-API-KEY)
```

- **Writes are gated.** A workflow create (`POST .../workflows`) or update
  (`PUT`/`PATCH .../workflows/{id}`) is linted with the same engine as
  `workflows lint`. If it has any **error**-severity findings, the proxy returns
  `422 Unprocessable Entity` with the findings and never forwards it. Warnings do
  not block. Sub-resource writes (`.../workflows/{id}/tags`, `/activate`, …) are
  not workflow bodies, so they pass through untouched.
- **Reads pass through** unchanged.
- **The key is injected** from your keyring (the active profile), so the client
  pointed at the proxy never needs the API key.

## Usage

Start the proxy (it targets your active profile's instance):

```bash
n8nctl proxy                       # listens on 127.0.0.1:8099
n8nctl --profile prod proxy        # gate the prod instance
```

Point any n8n client at the proxy as if it were the instance host:

```bash
# n8nctl itself
n8nctl --base-url http://127.0.0.1:8099 workflows create --file workflow.json

# any other n8n client / SDK
export N8N_API_URL=http://127.0.0.1:8099
```

A rejected push looks like this:

```json
{
  "message": "n8nctl proxy: workflow rejected by lint",
  "lint": [
    { "rule": "webhook-id-required", "severity": "error",
      "node": "Webhook", "message": "webhook/formTrigger node is missing webhookId" }
  ]
}
```

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `--listen` | `127.0.0.1:8099` | Address to bind. Keep it on localhost unless you understand the security note below. |
| `--disable-rule` | — | Lint rules to skip (comma-separated), e.g. `--disable-rule expression-prefix`. |
| `--block-destructive` | off | Also reject workflow `DELETE` requests with `403`. |
| `--reject-duplicate-names` | off | Reject creating a workflow whose name already exists on the instance. |

The rules and their grounding are the same as `workflows lint`; see
`n8nctl workflows lint --list-rules`.

## Reject duplicate names

`--reject-duplicate-names` adds a second gate alongside the lint check. On a
workflow create (`POST .../workflows`), the proxy first looks up the instance for
an existing workflow with the same name; if one exists, it returns `422` and
never forwards the create. This keeps an `apply`/`sync` that matches by name
unambiguous — two workflows sharing a name make name-based reconciliation pick
the wrong one — and stops a script or agent from silently creating a duplicate.
The check applies to creates only; updates to an existing workflow are
unaffected.

```bash
n8nctl proxy --reject-duplicate-names
```

A rejected create looks like the lint rejection, with a name-collision message in
place of the lint findings.

!!! warning "The proxy is an authenticated gateway"
    Because the proxy injects your API key, anything that can reach its listen
    address can talk to your n8n instance through it. Bind it to `127.0.0.1`
    (the default) on a trusted machine. Do not expose it on a shared network.

## How it relates to `agent guard`

`proxy` and [`agent guard`](agent-guard.md) enforce different things and compose
well:

- **`agent guard`** stops an agent from *running destructive operations* — it
  blocks `delete` at the agent host (Claude Code / Codex / OpenCode).
- **`proxy`** stops *low-quality workflows from landing* — it lints writes at the
  API boundary, for any client.

Run an agent against the proxy (so its writes are lint-gated) and install the
guard (so it cannot delete), and you have both quality and safety enforced
structurally rather than by convention.
