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
