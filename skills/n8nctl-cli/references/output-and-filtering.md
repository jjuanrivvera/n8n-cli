# n8nctl - output, columns, filtering & pagination

How to shape what `n8nctl` prints and how to fetch exactly the records you want.
Authoritative docs: https://github.com/jjuanrivvera/n8n-cli

## Output formats

Every command takes `-o/--output` with one of `table`, `json`, `yaml`, `csv`.
The default is `table` (human-readable). Set a session default with
`N8NCTL_OUTPUT` or a persistent one with `n8nctl config set output_format json`.

```bash
n8nctl workflows list                 # table (default)
n8nctl workflows list -o json         # machine-readable, for jq
n8nctl workflows get 42 -o yaml       # full object as YAML
n8nctl executions list -o csv         # spreadsheet-friendly
```

- **table** - curated columns, aligned, colorized on a TTY (disable with
  `--no-color` or `NO_COLOR`).
- **json** - the full object/array exactly as the API returns it; the right
  choice whenever you pipe to `jq` or another tool.
- **yaml** - same data as JSON, easier to eyeball for nested structures.
- **csv** - flat rows; pair with `--columns` to control the header order.

## Columns (`--columns`)

`table` and `csv` show a curated set of columns per resource. Override with
`--columns a,b,c` to pick exactly which fields appear, in order:

```bash
n8nctl workflows list --columns id,name,active
n8nctl executions list --columns id,workflowId,status,startedAt -o csv
n8nctl credentials list --columns id,name,type
```

Column names are the JSON field names from the API (use `-o json` once to see
what's available). `--columns` is ignored for `json`/`yaml`, which always emit
the complete object.

## Filtering with `--param` and typed filters

Most resources expose a few **typed filter flags** on `list` that map to API
query parameters:

| Resource | Typed filters |
|---|---|
| `workflows` | `--active true\|false`, `--name <substr>`, `--tags a,b`, `--project <id>` |
| `executions` | `--status <s>`, `--workflow <id>`, `--project <id>`, `--include-data true\|false` |
| `variables` | `--project <id>`, `--state empty` |
| `users` | `--project <id>`, `--include-role true\|false` |

```bash
n8nctl workflows list --active true --tags Production
n8nctl executions list --status error --workflow 42
```

For anything the typed flags don't cover, `--param key=value` is the escape
hatch - it sets a raw query parameter and is repeatable:

```bash
n8nctl workflows list --param active=true --param tags=Prod
n8nctl executions list --param status=error --param limit=10
```

`--status` accepts: `canceled`, `crashed`, `error`, `new`, `running`, `success`,
`unknown`, `waiting`.

## Pagination: `--all`, `--limit`, `--cursor`

n8n uses **cursor pagination**: a page of results plus a `nextCursor` for the
next page. `n8nctl` surfaces three controls on every `list`:

- `--limit N` - page size (capped at 250).
- `--cursor STR` - continue from a cursor returned by a previous call.
- `--all` - auto-paginate: walk every page and return the combined list.
- `--max-pages N` - safety cap for `--all` (0 = built-in default).

By default a single `list` returns one page; if more exist, `n8nctl` prints a
hint to stderr telling you the cursor to continue with (or to use `--all`):

```bash
n8nctl executions list --limit 100
# -> more results available - re-run with --cursor <token> (or --all)

n8nctl executions list --limit 100 --cursor <token>   # next page

n8nctl workflows list --all                # every workflow, all pages
n8nctl executions list --status error --all --max-pages 20
```

When `--all` hits `--max-pages` before exhausting results, it warns on stderr
that the output was truncated - narrow your filters or raise `--max-pages`.

## Piping to jq

`-o json` is designed for `jq`. The hint lines above go to **stderr**, so they
never pollute the JSON on stdout.

```bash
# ids of all active workflows
n8nctl workflows list --active true --all -o json | jq -r '.[].id'

# count failed executions for one workflow
n8nctl executions list --status error --workflow 42 --all -o json | jq 'length'

# map workflow id -> name
n8nctl workflows list --all -o json | jq -r '.[] | "\(.id)\t\(.name)"'

# pull one field out of a single object
n8nctl workflows get 42 -o json | jq -r '.active'
```

## CSV export

Combine `--all`, `-o csv`, and `--columns` to export a resource to a file:

```bash
n8nctl workflows list --all -o csv --columns id,name,active,updatedAt > workflows.csv
n8nctl executions list --status error --all -o csv \
  --columns id,workflowId,status,startedAt > failed-executions.csv
n8nctl credentials list --all -o csv --columns id,name,type > credentials.csv
```

The first row is the header (the `--columns` order); each subsequent row is one
record. For nested data, prefer `-o json` and post-process with `jq`.
