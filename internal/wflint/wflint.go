// Package wflint runs static checks over an n8n workflow definition. It is a
// focused subset of the rules a workflows-as-code reviewer wants in CI:
// structural integrity, dangling connections, orphaned nodes, missing webhook
// ids, and expression strings that forgot the leading '='.
package wflint

import (
	"encoding/json"
	"strings"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// Severity of a finding.
type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

// Finding is a single lint result.
type Finding struct {
	Rule     string   `json:"rule"`
	Severity Severity `json:"severity"`
	Node     string   `json:"node,omitempty"`
	Message  string   `json:"message"`
}

// Rules lists the available rules, what they check, and the canonical basis for
// each — these are grounded in n8n's own schema/behavior, not invented. n8n has
// no official workflow linter; deeper node-parameter validation would require
// bundling n8n's per-node schemas (a planned extension).
var Rules = []struct {
	Name  string
	Desc  string
	Basis string
}{
	{"required-fields", "name, nodes and connections must be present", "n8n public-API OpenAPI workflow schema (required: name, nodes, connections, settings)"},
	{"connection-reference", "connections must reference existing nodes", "workflow connection graph model"},
	{"orphaned-node", "nodes should be connected to the graph", "workflow connection graph model"},
	{"webhook-id-required", "webhook/formTrigger nodes need a webhookId", "n8n webhook registration behavior"},
	{"expression-prefix", "expression strings ({{ }}) should start with '='", "n8n expression syntax (the '=' prefix marks an evaluated expression)"},
}

type node struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	WebhookID  string         `json:"webhookId"`
	Parameters map[string]any `json:"parameters"`
}

// Lint returns all findings for a workflow. disabled names are skipped.
func Lint(wf *api.Workflow, disabled map[string]bool) []Finding {
	var out []Finding
	on := func(rule string) bool { return !disabled[rule] }

	// required-fields
	if on("required-fields") {
		if strings.TrimSpace(wf.Name) == "" {
			out = append(out, Finding{"required-fields", Error, "", "workflow has no name"})
		}
		if len(wf.Nodes) == 0 || string(wf.Nodes) == "null" {
			out = append(out, Finding{"required-fields", Error, "", "workflow has no nodes"})
		}
		if len(wf.Connections) == 0 || string(wf.Connections) == "null" {
			out = append(out, Finding{"required-fields", Warning, "", "workflow has no connections"})
		}
	}

	var nodes []node
	_ = json.Unmarshal(wf.Nodes, &nodes)
	names := map[string]bool{}
	for _, n := range nodes {
		names[n.Name] = true
	}

	// connection-reference + collect connected node set
	connected := map[string]bool{}
	var conns map[string]map[string][][]struct {
		Node string `json:"node"`
	}
	_ = json.Unmarshal(wf.Connections, &conns)
	for src, outputs := range conns {
		if on("connection-reference") && !names[src] {
			out = append(out, Finding{"connection-reference", Error, src, "connection source references a missing node"})
		}
		connected[src] = true
		for _, groups := range outputs {
			for _, group := range groups {
				for _, c := range group {
					if on("connection-reference") && !names[c.Node] {
						out = append(out, Finding{"connection-reference", Error, c.Node, "connection target references a missing node"})
					}
					connected[c.Node] = true
				}
			}
		}
	}

	for _, n := range nodes {
		// orphaned-node (only meaningful with >1 node)
		if on("orphaned-node") && len(nodes) > 1 && !connected[n.Name] {
			out = append(out, Finding{"orphaned-node", Warning, n.Name, "node is not connected to any other node"})
		}
		// webhook-id-required
		if on("webhook-id-required") && isWebhookish(n.Type) && n.WebhookID == "" {
			out = append(out, Finding{"webhook-id-required", Error, n.Name, "webhook/formTrigger node is missing webhookId"})
		}
		// expression-prefix: walk all nested string params; only flag strings that
		// look like a genuine n8n expression (contain {{ }} and an n8n token) yet
		// are missing the leading '=' that marks an evaluated expression.
		if on("expression-prefix") {
			for field, v := range n.Parameters {
				walkStrings(v, func(s string) {
					if looksLikeExpression(s) && !strings.HasPrefix(s, "=") {
						out = append(out, Finding{"expression-prefix", Warning, n.Name,
							"parameter \"" + field + "\" contains an n8n expression but is missing the leading '='"})
					}
				})
			}
		}
	}
	return out
}

// walkStrings calls fn for every string value reachable in v (nested maps/slices).
func walkStrings(v any, fn func(string)) {
	switch t := v.(type) {
	case string:
		fn(t)
	case map[string]any:
		for _, vv := range t {
			walkStrings(vv, fn)
		}
	case []any:
		for _, vv := range t {
			walkStrings(vv, fn)
		}
	}
}

// exprTokens are markers of a real n8n expression, used to avoid flagging plain
// strings that merely contain "{{".
var exprTokens = []string{"$json", "$node", "$(", "$vars", "$workflow", "$env",
	"$now", "$today", "$items", "$input", "$prevNode", "$execution", "$runIndex", "$itemIndex"}

// looksLikeExpression reports whether s is plausibly an unprefixed n8n expression.
func looksLikeExpression(s string) bool {
	if !strings.Contains(s, "{{") {
		return false
	}
	for _, tok := range exprTokens {
		if strings.Contains(s, tok) {
			return true
		}
	}
	return false
}

func isWebhookish(t string) bool {
	l := strings.ToLower(t)
	return strings.Contains(l, "webhook") || strings.Contains(l, "formtrigger")
}

// Errors reports how many findings are errors.
func Errors(fs []Finding) int {
	n := 0
	for _, f := range fs {
		if f.Severity == Error {
			n++
		}
	}
	return n
}
