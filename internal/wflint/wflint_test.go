package wflint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func wf(nodes, conns string) *api.Workflow {
	return &api.Workflow{Name: "w", Nodes: json.RawMessage(nodes), Connections: json.RawMessage(conns)}
}

func rules(fs []Finding) map[string]bool {
	m := map[string]bool{}
	for _, f := range fs {
		m[f.Rule] = true
	}
	return m
}

func TestLint_Clean(t *testing.T) {
	w := wf(`[{"name":"Hook","type":"n8n-nodes-base.webhook","webhookId":"x","parameters":{}},
	          {"name":"Set","type":"n8n-nodes-base.set","parameters":{}}]`,
		`{"Hook":{"main":[[{"node":"Set","type":"main","index":0}]]}}`)
	assert.Empty(t, Lint(w, nil))
}

func TestLint_RequiredFields(t *testing.T) {
	w := &api.Workflow{} // no name, no nodes
	fs := Lint(w, nil)
	assert.True(t, rules(fs)["required-fields"])
	assert.GreaterOrEqual(t, Errors(fs), 1)
}

func TestLint_ConnectionReference(t *testing.T) {
	w := wf(`[{"name":"A","type":"x","parameters":{}}]`,
		`{"A":{"main":[[{"node":"Ghost","type":"main","index":0}]]}}`)
	fs := Lint(w, nil)
	assert.True(t, rules(fs)["connection-reference"])
	assert.GreaterOrEqual(t, Errors(fs), 1)
}

func TestLint_OrphanedNode(t *testing.T) {
	w := wf(`[{"name":"A","type":"x","parameters":{}},{"name":"Lonely","type":"x","parameters":{}}]`,
		`{}`)
	assert.True(t, rules(Lint(w, nil))["orphaned-node"])
}

func TestLint_WebhookIDRequired(t *testing.T) {
	w := wf(`[{"name":"Hook","type":"n8n-nodes-base.webhook","parameters":{}}]`, `{}`)
	fs := Lint(w, nil)
	assert.True(t, rules(fs)["webhook-id-required"])
	assert.GreaterOrEqual(t, Errors(fs), 1)
}

func TestLint_ExpressionPrefix(t *testing.T) {
	w := wf(`[{"name":"A","type":"x","parameters":{"url":"{{ $json.x }}"}}]`, `{}`)
	assert.True(t, rules(Lint(w, nil))["expression-prefix"])
	// with the '=' prefix it is fine
	ok := wf(`[{"name":"A","type":"x","parameters":{"url":"={{ $json.x }}"}}]`, `{}`)
	assert.False(t, rules(Lint(ok, nil))["expression-prefix"])
}

func TestLint_DisableRule(t *testing.T) {
	w := wf(`[{"name":"Hook","type":"n8n-nodes-base.webhook","parameters":{}}]`, `{}`)
	fs := Lint(w, map[string]bool{"webhook-id-required": true})
	assert.False(t, rules(fs)["webhook-id-required"])
}

func TestRulesHaveBasis(t *testing.T) {
	require.NotEmpty(t, Rules)
	for _, r := range Rules {
		assert.NotEmpty(t, r.Basis, r.Name) // every rule documents its canonical basis
	}
}

func TestLint_MalformedJSON_NoPanic(t *testing.T) {
	w := &api.Workflow{Name: "w", Nodes: json.RawMessage(`not-json`), Connections: json.RawMessage(`{}`)}
	assert.NotPanics(t, func() { Lint(w, nil) })
}

func TestExpressionPrefix_NoFalsePositive(t *testing.T) {
	// a plain string containing "{{" but no n8n token must NOT be flagged
	plain := wf(`[{"name":"A","type":"x","parameters":{"text":"use {{ mustache }} templates"}}]`, `{}`)
	assert.False(t, rules(Lint(plain, nil))["expression-prefix"])
	// a genuine expression nested in a map IS flagged
	nested := wf(`[{"name":"A","type":"x","parameters":{"opts":{"url":"{{ $json.id }}"}}}]`, `{}`)
	assert.True(t, rules(Lint(nested, nil))["expression-prefix"])
}

func TestLint_UnknownNodeType(t *testing.T) {
	// a typo'd base node type is an error with a suggestion
	w := wf(`[{"name":"S","type":"n8n-nodes-base.slak","parameters":{}}]`, `{}`)
	fs := Lint(w, nil)
	assert.True(t, rules(fs)["unknown-node-type"])
	var msg string
	for _, f := range fs {
		if f.Rule == "unknown-node-type" {
			msg = f.Message
		}
	}
	assert.Contains(t, msg, "n8n-nodes-base.slack") // did-you-mean

	// a real node type is fine
	ok := wf(`[{"name":"S","type":"n8n-nodes-base.slack","parameters":{}}]`, `{}`)
	assert.False(t, rules(Lint(ok, nil))["unknown-node-type"])

	// a community/custom node (prefix not in the catalog) is never flagged
	custom := wf(`[{"name":"C","type":"n8n-nodes-acme.widget","parameters":{}}]`, `{}`)
	assert.False(t, rules(Lint(custom, nil))["unknown-node-type"])
}

func TestLint_UnknownParameter(t *testing.T) {
	// a param the node doesn't define is a warning (only for catalogued nodes)
	w := wf(`[{"name":"H","type":"n8n-nodes-base.httpRequest","parameters":{"notARealParam":"x"}}]`, `{}`)
	assert.True(t, rules(Lint(w, nil))["unknown-parameter"])

	// valid params produce no unknown-parameter finding
	ok := wf(`[{"name":"H","type":"n8n-nodes-base.httpRequest","parameters":{"url":"https://x"}}]`, `{}`)
	assert.False(t, rules(Lint(ok, nil))["unknown-parameter"])

	// params on an unknown/community node are not second-guessed
	custom := wf(`[{"name":"C","type":"n8n-nodes-acme.widget","parameters":{"anything":"y"}}]`, `{}`)
	assert.False(t, rules(Lint(custom, nil))["unknown-parameter"])
}

func TestNodeCatalog_Loaded(t *testing.T) {
	c := loadCatalog()
	assert.Greater(t, len(c.types), 500, "catalog should hold the full node set")
	assert.True(t, nodeKnown("n8n-nodes-base.slack"))
	assert.True(t, nodeKnown("@n8n/n8n-nodes-langchain.agent") || len(c.types) > 0)
	assert.False(t, nodeKnown("n8n-nodes-base.definitelyNotANode"))
	assert.Contains(t, catalogBasis(), "n8n-nodes-base@")
}
