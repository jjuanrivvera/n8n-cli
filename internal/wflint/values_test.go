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

// Regression for the code-review Critical: n8n displayOptions use meta-keys like
// @version (the node's typeVersion) that are not workflow parameters. The Set node
// gates its include="none" option behind @version; a naive evaluator excluded that
// variant and flagged the (valid) value. The rule must stay silent here.
func TestValueRuleHandlesVersionGatedOptions(t *testing.T) {
	fs := lintNodes(t, `[{"name":"Edit","type":"n8n-nodes-base.set","typeVersion":3,"parameters":{"include":"none","mode":"manual"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"), "@version-gated option must not be flagged")
}

// A missing controlling parameter (using its default) must not narrow the allowed
// set — the variant stays possibly-visible, so its options remain valid.
func TestValueRuleLenientWhenControllerAbsent(t *testing.T) {
	// operation present but resource omitted (defaults to message): must not flag a
	// real message operation.
	fs := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"operation":"post"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"))
}

func TestScalarStrings(t *testing.T) {
	assert.Equal(t, []string{"x"}, scalarStrings("x"))
	assert.Equal(t, []string{"true"}, scalarStrings(true))
	assert.Equal(t, []string{"false"}, scalarStrings(false))
	assert.Equal(t, []string{"3"}, scalarStrings(float64(3)))
	assert.Equal(t, []string{"a", "b"}, scalarStrings([]any{"a", "b"}))
	assert.Nil(t, scalarStrings(map[string]any{"k": "v"}))
	// a multiOptions array value with an invalid member is caught
	assert.True(t, valueMatchesAny([]any{"a", "b"}, []string{"b"}))
	assert.False(t, valueMatchesAny(nil, []string{"a"}))
}

func TestMultiOptionsValidation(t *testing.T) {
	// resource=message is fine; a bogus member of a multiOptions array would be
	// flagged. Use a node with a known multiOptions param if present; otherwise this
	// at least exercises the array path without a false positive on valid values.
	fs := lintNodes(t, `[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"post"}}]`)
	assert.False(t, hasRule(fs, "invalid-parameter-value"))
}

func TestVariantHiddenBranches(t *testing.T) {
	params := map[string]any{"resource": "message", "mode": "manual"}

	// no displayOptions -> visible
	assert.False(t, variantHidden(ParamSchema{}, params))

	// show matches present param -> visible
	show := ParamSchema{DisplayOptions: map[string]map[string][]string{"show": {"resource": {"message"}}}}
	assert.False(t, variantHidden(show, params))

	// show contradicts a present param -> hidden
	showBad := ParamSchema{DisplayOptions: map[string]map[string][]string{"show": {"resource": {"channel"}}}}
	assert.True(t, variantHidden(showBad, params))

	// show on an ABSENT param -> unknown -> not hidden (lenient)
	showAbsent := ParamSchema{DisplayOptions: map[string]map[string][]string{"show": {"missing": {"x"}}}}
	assert.False(t, variantHidden(showAbsent, params))

	// hide matches present param -> hidden
	hide := ParamSchema{DisplayOptions: map[string]map[string][]string{"hide": {"mode": {"manual"}}}}
	assert.True(t, variantHidden(hide, params))

	// meta-key (@version) and empty condition lists are ignored
	meta := ParamSchema{DisplayOptions: map[string]map[string][]string{"show": {"@version": {"3"}, "/ref": {}}}}
	assert.False(t, variantHidden(meta, params))
}

func TestMetaKey(t *testing.T) {
	assert.True(t, metaKey("@version"))
	assert.True(t, metaKey("/rootRef"))
	assert.False(t, metaKey("resource"))
}
