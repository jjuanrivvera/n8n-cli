# Contributing

`n8nctl` is built in Go with a small generic core: a typed API client plus one
thin file per resource. Adding a resource means declaring its type, columns, and
any custom actions; the shared core handles HTTP, auth, retries, rate limiting,
pagination, and output.

The authoritative contributor guide lives in the repository:

- [CONTRIBUTING.md on GitHub](https://github.com/jjuanrivvera/n8n-cli/blob/main/CONTRIBUTING.md)

## Regenerating the command reference

The pages under [Command Reference](commands/index.md) are generated from the
cobra command tree. After changing any command or flag, regenerate them:

```bash
go run ./tools/gendocs
```

This writes Markdown into `docs/commands/`. The generator disables cobra's
auto-generated timestamp footer so the output is reproducible.

## Local docs preview

```bash
pip install mkdocs-material mkdocs-git-revision-date-localized-plugin
mkdocs serve
```
