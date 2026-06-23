package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// addWorkflowSearchCmd adds `workflows search`, which scans every workflow's node
// graph and finds matches by node type, credential, webhook path, or name. This is
// impossible in the n8n UI and tedious via raw API calls: "which workflows use the
// Slack node?", "what still references this credential before I delete it?",
// "which workflow owns webhook path /orders?".
func addWorkflowSearchCmd(parent *cobra.Command) {
	var (
		nameRe     string
		nodeType   string
		credential string
		webhook    string
	)
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find workflows by node type, credential, webhook path, or name",
		Long: "Scan all workflows and report those matching a filter:\n" +
			"  --node <type>        substring match on node type (e.g. slack, httpRequest)\n" +
			"  --credential <id|nm> workflows referencing a credential by id or name\n" +
			"  --webhook <path>     workflows with a webhook node on that path\n" +
			"  --name <regex>       workflow name matches a regular expression\n\n" +
			"  n8nctl workflows search --node slack\n" +
			"  n8nctl workflows search --credential githubApi -o json\n" +
			"  n8nctl workflows search --webhook /orders",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if nameRe == "" && nodeType == "" && credential == "" && webhook == "" {
				return fmt.Errorf("provide at least one filter: --node, --credential, --webhook, or --name")
			}
			var nameMatcher *regexp.Regexp
			if nameRe != "" {
				re, err := regexp.Compile(nameRe)
				if err != nil {
					return fmt.Errorf("invalid --name regex: %w", err)
				}
				nameMatcher = re
			}

			client, err := getAPIClient(cmd)
			if err != nil {
				return err
			}
			workflows, err := client.Workflows().ListAll(cmd.Context(), api.ListParams{}, 0)
			if err != nil {
				if api.IsDryRun(err) {
					return nil
				}
				return err
			}

			type hit struct {
				ID      string   `json:"id"`
				Name    string   `json:"name"`
				Active  bool     `json:"active"`
				Matched []string `json:"matched"`
			}
			var results []hit
			for i := range workflows {
				wf := &workflows[i]
				if nameMatcher != nil && !nameMatcher.MatchString(wf.Name) {
					continue
				}
				matched := matchNodes(wf.Nodes, nodeType, credential, webhook)
				// A pure name search has no node matches but still qualifies.
				if nameMatcher != nil && nodeType == "" && credential == "" && webhook == "" {
					matched = []string{"name"}
				}
				if len(matched) == 0 {
					continue
				}
				results = append(results, hit{
					ID:      wf.ID.String(),
					Name:    wf.Name,
					Active:  wf.Active.Bool(),
					Matched: matched,
				})
			}
			if len(results) == 0 && !flagQuiet {
				fmt.Fprintln(cmd.ErrOrStderr(), "no workflows matched")
			}
			return render(cmd, results)
		},
	}
	cmd.Flags().StringVar(&nameRe, "name", "", "match workflow name against a regular expression")
	cmd.Flags().StringVar(&nodeType, "node", "", "match a node type substring (e.g. slack, httpRequest)")
	cmd.Flags().StringVar(&credential, "credential", "", "match workflows referencing this credential id or name")
	cmd.Flags().StringVar(&webhook, "webhook", "", "match workflows with a webhook node on this path")
	parent.AddCommand(readOnlyHints(cmd))
}

// node mirrors the subset of an n8n node we inspect during search.
type searchNode struct {
	Name        string                     `json:"name"`
	Type        string                     `json:"type"`
	Parameters  map[string]any             `json:"parameters"`
	Credentials map[string]json.RawMessage `json:"credentials"`
}

// matchNodes decodes a workflow's nodes and returns the names of nodes matching
// any active filter. An empty result means no match.
func matchNodes(raw json.RawMessage, nodeType, credential, webhook string) []string {
	if len(raw) == 0 {
		return nil
	}
	var nodes []searchNode
	if err := json.Unmarshal(raw, &nodes); err != nil {
		return nil
	}
	var matched []string
	for _, n := range nodes {
		if nodeType != "" && strings.Contains(strings.ToLower(n.Type), strings.ToLower(nodeType)) {
			matched = append(matched, n.Name+" ("+shortType(n.Type)+")")
			continue
		}
		if credential != "" && nodeReferencesCredential(n, credential) {
			matched = append(matched, n.Name+" [credential]")
			continue
		}
		if webhook != "" && nodeMatchesWebhook(n, webhook) {
			matched = append(matched, n.Name+" [webhook]")
			continue
		}
	}
	return matched
}

// nodeReferencesCredential reports whether a node uses a credential by id or name.
func nodeReferencesCredential(n searchNode, want string) bool {
	for _, ref := range n.Credentials {
		var c struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(ref, &c); err != nil {
			continue
		}
		if c.ID == want || strings.EqualFold(c.Name, want) {
			return true
		}
	}
	return false
}

// nodeMatchesWebhook reports whether a node is a webhook on the given path.
func nodeMatchesWebhook(n searchNode, path string) bool {
	if !strings.Contains(strings.ToLower(n.Type), "webhook") {
		return false
	}
	p, ok := n.Parameters["path"].(string)
	if !ok {
		return false
	}
	return strings.TrimPrefix(p, "/") == strings.TrimPrefix(path, "/")
}

// shortType drops the package prefix from a node type for readable output.
func shortType(t string) string {
	if i := strings.LastIndex(t, "."); i >= 0 {
		return t[i+1:]
	}
	return t
}
