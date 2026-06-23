package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreakingChangesCommand(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	dir := t.TempDir()
	f := dir + "/wf.json"
	require.NoError(t, os.WriteFile(f, []byte(
		`{"name":"t","nodes":[{"name":"H","type":"n8n-nodes-base.httpRequest","typeVersion":1,"parameters":{"url":"x"}}],"connections":{},"settings":{}}`), 0o600))
	out, _, err := run(t, "workflows", "breaking-changes", "-f", f)
	require.NoError(t, err)
	assert.Contains(t, out, "typeVersion 1")
	assert.Contains(t, out, "latest")

	// a workflow with no outdated nodes
	g := dir + "/ok.json"
	require.NoError(t, os.WriteFile(g, []byte(
		`{"name":"t","nodes":[{"name":"S","type":"n8n-nodes-base.set","typeVersion":3,"parameters":{}}],"connections":{},"settings":{}}`), 0o600))
	out, _, err = run(t, "workflows", "breaking-changes", "-f", g)
	require.NoError(t, err)
	assert.Contains(t, out, "no outdated nodes")
}

func TestLintValueRuleEndToEnd(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	dir := t.TempDir()
	f := dir + "/wf.json"
	require.NoError(t, os.WriteFile(f, []byte(
		`{"name":"t","nodes":[{"name":"S","type":"n8n-nodes-base.slack","parameters":{"resource":"message","operation":"nope"}}],"connections":{},"settings":{}}`), 0o600))
	_, _, err := run(t, "workflows", "lint", "-f", f)
	require.Error(t, err) // lint exits non-zero on the value error
}
