package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/wflint"
)

// The `nodes` command explores the embedded catalog of n8n node definitions
// (the same data the lint rules validate against). It is fully local — no API
// call — so it works offline and needs no profile.

func init() {
	nodes := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"node"},
		Short:   "Explore the catalog of n8n node types (offline)",
		Long: "Browse the built-in catalog of n8n node definitions (n8n-nodes-base plus the\n" +
			"langchain nodes). Useful for finding the exact `type` string for a node and the\n" +
			"parameters it accepts. The catalog is embedded, so this works offline.",
	}

	list := &cobra.Command{
		Use:     "list",
		Short:   "List all node types",
		Args:    cobra.NoArgs,
		Example: "  n8nctl nodes list\n  n8nctl nodes list -o json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return render(cmd, wflint.Nodes(), "type", "displayName")
		},
	}

	var limit int
	search := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search node types by type or display name",
		Args:    cobra.ExactArgs(1),
		Example: "  n8nctl nodes search slack\n  n8nctl nodes search \"http request\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			q := strings.ToLower(args[0])
			var hits []wflint.NodeInfo
			for _, n := range wflint.Nodes() {
				if strings.Contains(strings.ToLower(n.Type), q) || strings.Contains(strings.ToLower(n.DisplayName), q) {
					hits = append(hits, n)
					if limit > 0 && len(hits) >= limit {
						break
					}
				}
			}
			if len(hits) == 0 {
				return fmt.Errorf("no node types match %q", args[0])
			}
			return render(cmd, hits, "type", "displayName")
		},
	}
	search.Flags().IntVar(&limit, "limit", 0, "max results (0 = all)")

	show := &cobra.Command{
		Use:     "show <type>",
		Short:   "Show a node type's display name and parameters",
		Args:    cobra.ExactArgs(1),
		Example: "  n8nctl nodes show n8n-nodes-base.slack",
		RunE: func(cmd *cobra.Command, args []string) error {
			n, ok := wflint.Node(args[0])
			if !ok {
				if s, sok := wflint.SuggestNodeType(args[0]); sok {
					return fmt.Errorf("unknown node type %q — did you mean %q?", args[0], s)
				}
				return fmt.Errorf("unknown node type %q (try `n8nctl nodes search`)", args[0])
			}
			return render(cmd, n)
		},
	}

	nodes.AddCommand(readOnlyHints(list), readOnlyHints(search), readOnlyHints(show))
	rootCmd.AddCommand(readOnlyHints(nodes))
}
