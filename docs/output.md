# Output and filtering

## Output formats

Every command that returns data accepts `-o`/`--output` with one of four
formats:

```bash
n8nctl workflows list                 # table (default)
n8nctl workflows list -o json         # json
n8nctl workflows list -o yaml         # yaml
n8nctl workflows list -o csv          # csv
```

- `table` is the human default: aligned columns, colored in a TTY.
- `json` is the scripting default: pipe it into `jq`.
- `yaml` is handy for eyeballing nested structures.
- `csv` drops straight into a spreadsheet.

Set a default once so you do not repeat the flag:

```bash
n8nctl config set output_format json
# or, per command / shell
export N8NCTL_OUTPUT=json
```

Color is automatic in a terminal and off when piped. Force it off with
`--no-color` or the `NO_COLOR` environment variable.

## Choosing columns

`--columns` selects which fields appear in `table` and `csv` output, in the
order you list them:

```bash
n8nctl workflows list --columns id,name,active
n8nctl workflows list -o csv --columns id,name,active > workflows.csv
n8nctl executions list -o csv --columns id,status,startedAt
```

## Filtering with flags

List commands expose the filters the n8n API supports as flags. For workflows:

```bash
n8nctl workflows list --active true
n8nctl workflows list --name "sync"            # substring of the name
n8nctl workflows list --tags prod,nightly      # comma-separated tag names
n8nctl workflows list --project <project-id>
```

For executions:

```bash
n8nctl executions list --status error
n8nctl executions list --workflow 42
n8nctl executions list --status success --include-data true
```

## Filtering with --param

When the API accepts a query parameter that `n8nctl` does not expose as a
dedicated flag, pass it through with `--param key=value`. The flag is repeatable:

```bash
n8nctl workflows list --param someFilter=value --param another=thing
```

This is the escape hatch: anything the n8n API understands as a query parameter
on that endpoint can be set here without waiting for a named flag. For arbitrary
endpoints entirely, use the raw API command:

```bash
n8nctl api GET /workflows --query limit=5 --query active=true
```

## Pagination

The n8n API uses cursor pagination. By default a list command returns one page.

```bash
# One page, with a page size
n8nctl workflows list --limit 50

# Walk every page automatically
n8nctl workflows list --all

# Resume from a cursor a previous response gave you
n8nctl workflows list --cursor <cursor>
```

`--limit` sets the page size (capped at 250 by the API). `--all` follows the
cursor until there are no more pages. When auto-paginating, `--max-pages` caps
how many pages `--all` will fetch, which is a useful guard on very large
instances.

## Putting it together

```bash
# Every failed execution across all pages, as a tidy CSV
n8nctl executions list --status error --all -o csv \
  --columns id,workflowId,status,startedAt > failures.csv

# Active workflow names from the cloud instance, into jq
n8nctl --profile cloud workflows list --active true --all -o json \
  | jq -r '.[].name'
```
