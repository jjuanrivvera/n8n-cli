package wflint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func lintNodes(t *testing.T, nodesJSON string) []Finding {
	t.Helper()
	wf := &api.Workflow{
		Name:        "t",
		Nodes:       json.RawMessage(nodesJSON),
		Connections: json.RawMessage(`{}`),
	}
	return Lint(wf, map[string]bool{"orphaned-node": true})
}

func hasRule(fs []Finding, rule string) bool {
	for _, f := range fs {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func TestInvalidParameterValue(t *testing.T) {
	// a bad option value, resolved against the active resource variant -> flagged
	fs := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"psot"}}]`)
	require.True(t, hasRule(fs, "invalid-parameter-value"), "psot should be flagged")
	var msg string
	for _, f := range fs {
		if f.Rule == "invalid-parameter-value" {
			msg = f.Message
		}
	}
	assert.Contains(t, msg, "did you mean \"post\"")
}

func TestValidParameterValueIsQuiet(t *testing.T) {
	// the correct value for the active variant -> no finding
	fs := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"post"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"))
}

func TestValueRuleSkipsExpressions(t *testing.T) {
	fs := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"={{ $json.op }}"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"), "expression values must not be validated")
}

func TestValueRuleRespectsDisplayOptions(t *testing.T) {
	// "post" is valid for resource=message but NOT for resource=channel; the rule
	// must validate against the variant active for the node's actual resource.
	bad := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"channel","operation":"post"}}]`)
	assert.True(t, hasRule(bad, "invalid-parameter-value"), "post is not a channel operation")
}

func TestValueRuleSkipsUnknownAndDynamic(t *testing.T) {
	// unknown node -> no value validation (no schema)
	fs := lintNodes(t, `[{"name":"X","type":"n8n-nodes-base.notARealNode","parameters":{"foo":"bar"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"))
	// a free-text/string parameter (not options) -> never flagged
	fs = lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"post","text":"anything at all"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"))
}
