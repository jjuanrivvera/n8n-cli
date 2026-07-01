# Decisions

Recorded assumptions and deliberate trade-offs the deterministic gate checks each run
(cliwright GOAL.md §11). One line per decision; keep the rationale terse and factual.

## API coverage

- **coverage-waiver: shipping the stable public-API surface first; 60/80 (75%) of the
  enumerated n8n public API v1 operations are wrapped.** The 20 uncovered operations are the
  newest / enterprise-gated surfaces most self-hosted instances do not expose:
  data-table columns (create/list/update/delete), folders (5), insights summary, discover,
  community-packages (install/uninstall/update/list-installed), execution tags (get/update)
  and stop-many, `getWorkflowVersion`, and `testCredential`. Enumerated from the OpenAPI path
  specs on n8n `master` (2026-06-30). These are deferred until they stabilise across the LTS
  line; the 60 wrapped operations are the surface every supported instance exposes.

## Multi-instance selector flag

- The selector flag is `--instance` (an n8nctl profile IS one n8n instance). `--profile`
  remains a hidden, still-working alias for back-compat; `N8NCTL_INSTANCE` and the legacy
  `N8NCTL_PROFILE` env vars both resolve it.

## Gate

- `make verify` is the deterministic gate (build/vet/lint/test + spec-completeness +
  coverage); it carries no LLM judge, so it is CI- and token-safe. `make judge` runs the
  subjective LLM gate; `make accept` = `verify` + `judge` is the build-acceptance gate.
