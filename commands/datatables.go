package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	registerResource(resourceSpec[api.DataTable]{
		Use:     "data-tables",
		Aliases: []string{"data-table", "datatable", "dt"},
		Short:   "Manage data tables and their rows",
		Long: "Create, list, inspect, update and delete data tables, and manage their rows.\n" +
			"Data tables may be unlicensed on some editions (the API returns 403).\n\n" +
			"  n8nctl data-tables create --set name=orders --set 'columns=[{\"name\":\"sku\",\"type\":\"string\"}]'\n" +
			"  n8nctl data-tables rows <id> --filter '{\"type\":\"and\",\"filters\":[]}'",
		New:     func(c *api.Client) *api.Resource[api.DataTable] { return c.DataTables() },
		Columns: []string{"id", "name", "projectId", "updatedAt"},
		ListFilters: []listFilter{
			{Flag: "project", Query: "projectId", Usage: "filter by project id"},
		},
		Extra: dataTableExtra,
	})
}

func dataTableExtra(parent *cobra.Command, _ resourceSpec[api.DataTable]) {
	// rows <tableId>
	var search, filter string
	var limit int
	rows := &cobra.Command{
		Use:   "rows <tableId>",
		Short: "List rows in a data table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			params := api.ListParams{Limit: limit, Extra: map[string][]string{}}
			if search != "" {
				params.Extra.Set("search", search)
			}
			if filter != "" {
				params.Extra.Set("filter", filter)
			}
			raw, err := client.ListDataTableRows(cmd.Context(), args[0], params)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}
			return renderRaw(cmd, raw)
		},
	}
	rows.Flags().StringVar(&search, "search", "", "full-text search across string columns")
	rows.Flags().StringVar(&filter, "filter", "", "filter as a JSON string")
	rows.Flags().IntVar(&limit, "limit", 0, "max rows to return")
	parent.AddCommand(rows)

	// add-rows <tableId> --data/--file/--stdin (a JSON array of row objects)
	parent.AddCommand(rowMutationCmd("add-rows", "Add rows (body: a JSON array of row objects)",
		func(ctx context.Context, c *api.Client, id string, body json.RawMessage) (json.RawMessage, error) {
			return c.AddDataTableRows(ctx, id, body)
		}))

	// update-rows / upsert-rows <tableId> --data/--file/--stdin (body: {filter, data})
	parent.AddCommand(rowMutationCmd("update-rows", "Update rows matching a filter (body: {filter, data})",
		func(ctx context.Context, c *api.Client, id string, body json.RawMessage) (json.RawMessage, error) {
			return c.UpdateDataTableRows(ctx, id, body)
		}))
	parent.AddCommand(rowMutationCmd("upsert-rows", "Insert or update rows (body: {filter, data})",
		func(ctx context.Context, c *api.Client, id string, body json.RawMessage) (json.RawMessage, error) {
			return c.UpsertDataTableRows(ctx, id, body)
		}))

	// delete-rows <tableId> --filter <json>
	var delFilter string
	del := &cobra.Command{
		Use:   "delete-rows <tableId> --filter <json>",
		Short: "Delete rows matching a filter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if delFilter == "" {
				return fmt.Errorf("--filter is required")
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			raw, err := client.DeleteDataTableRows(cmd.Context(), args[0], delFilter)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return renderRaw(cmd, raw)
		},
	}
	del.Flags().StringVar(&delFilter, "filter", "", "filter as a JSON string (required)")
	parent.AddCommand(del)
}

// rowMutationCmd builds a row-writing subcommand that reads a raw JSON body from
// --data, --file, or --stdin and passes it to fn.
func rowMutationCmd(use, short string, fn func(context.Context, *api.Client, string, json.RawMessage) (json.RawMessage, error)) *cobra.Command {
	var data, file string
	var stdin bool
	cmd := &cobra.Command{
		Use:   use + " <tableId>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := readRawBody(data, file, stdin)
			if err != nil {
				return err
			}
			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			out, err := fn(cmd.Context(), client, args[0], body)
			if err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			return renderRaw(cmd, out)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "inline JSON body")
	cmd.Flags().StringVar(&file, "file", "", "read JSON body from a file ('-' for stdin)")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "read JSON body from stdin")
	return cmd
}

// readRawBody returns a JSON body from --data, --file, or --stdin.
func readRawBody(data, file string, stdin bool) (json.RawMessage, error) {
	switch {
	case stdin:
		b, err := io.ReadAll(os.Stdin)
		return json.RawMessage(b), err
	case file != "":
		b, err := readFileOrStdin(file)
		return json.RawMessage(b), err
	case data != "":
		return json.RawMessage(data), nil
	default:
		return nil, fmt.Errorf("provide --data, --file, or --stdin")
	}
}

// renderRaw decodes a raw response into a generic value and renders it. List
// envelopes ({data, nextCursor}) are unwrapped to their data array for table view.
func renderRaw(cmd *cobra.Command, raw json.RawMessage) error {
	var env struct {
		Data       json.RawMessage `json:"data"`
		NextCursor string          `json:"nextCursor"`
	}
	if err := json.Unmarshal(raw, &env); err == nil && env.Data != nil {
		if env.NextCursor != "" && !flagQuiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "more rows available (nextCursor %s)\n", env.NextCursor)
		}
		return render(cmd, env.Data)
	}
	return render(cmd, raw)
}
