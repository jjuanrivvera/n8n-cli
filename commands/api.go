package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func init() {
	var (
		data    string
		file    string
		queries []string
	)
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Make a raw authenticated API request (escape hatch)",
		Long: "Call any n8n API endpoint directly. PATH is relative to the instance base URL\n" +
			"(the leading /api/v1 is added automatically).\n\n" +
			"  n8nctl api GET /workflows -q limit=5\n" +
			"  n8nctl api POST /tags -d '{\"name\":\"Prod\"}'\n" +
			"  n8nctl api DELETE /executions/42 --dry-run",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]

			var body any
			switch {
			case file != "":
				raw, err := readFileOrStdin(file)
				if err != nil {
					return err
				}
				body = json.RawMessage(raw)
			case data != "":
				body = json.RawMessage(data)
			}

			q := url.Values{}
			for _, kv := range queries {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid -q %q (want key=value)", kv)
				}
				q.Add(k, v)
			}

			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			var out json.RawMessage
			if err := client.Do(context.Background(), method, path, q, body, &out); err != nil {
				if api.IsDryRun(err) {
					dryRunNotice(cmd)
					return nil
				}
				return err
			}
			if len(out) == 0 {
				if !flagQuiet {
					fmt.Fprintln(cmd.ErrOrStderr(), "(empty response)")
				}
				return nil
			}
			return render(cmd, out)
		},
	}
	cmd.Flags().StringVarP(&data, "data", "d", "", "request body as inline JSON")
	cmd.Flags().StringVar(&file, "file", "", "request body from a file ('-' for stdin)")
	cmd.Flags().StringArrayVar(&queries, "query", nil, "query parameter key=value (repeatable)")
	rootCmd.AddCommand(cmd)
}
