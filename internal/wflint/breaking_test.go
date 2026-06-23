package wflint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func TestBreakingChanges(t *testing.T) {
	wf := &api.Workflow{Nodes: json.RawMessage(`[
		{"name":"Old","type":"n8n-nodes-base.httpRequest","typeVersion":1,"parameters":{"url":"x","goneParam":"y"}},
		{"name":"Cur","type":"n8n-nodes-base.httpRequest","typeVersion":4,"parameters":{"url":"x"}},
		{"name":"Custom","type":"n8n-nodes-community.thing","typeVersion":1,"parameters":{}}
	]`)}
	issues := BreakingChanges(wf)
	require.Len(t, issues, 1, "only the outdated catalogued node is reported")
	assert.Equal(t, "Old", issues[0].Node)
	assert.Equal(t, 1, issues[0].CurrentVersion)
	assert.GreaterOrEqual(t, issues[0].LatestVersion, 4)
	assert.Contains(t, issues[0].UnknownParams, "goneParam")
	assert.NotContains(t, issues[0].UnknownParams, "url", "a still-valid param is not 'removed'")
}

func TestBreakingChangesEmptyWhenCurrent(t *testing.T) {
	wf := &api.Workflow{Nodes: json.RawMessage(`[{"name":"N","type":"n8n-nodes-base.set","typeVersion":3,"parameters":{}}]`)}
	// set's latest is 3.x -> int 3, so typeVersion 3 is current
	assert.Empty(t, BreakingChanges(wf))
}
