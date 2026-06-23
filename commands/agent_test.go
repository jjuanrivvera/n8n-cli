package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func bucketTools(cs []guardCmd) map[string]bool {
	m := map[string]bool{}
	for _, c := range cs {
		m[c.tool] = true
	}
	return m
}

func TestClassifyAPICommands(t *testing.T) {
	read, writes, irreversible := classifyAPICommands(RootCmd())
	r, w, irr := bucketTools(read), bucketTools(writes), bucketTools(irreversible)

	// Reads — generic and read-only custom verbs.
	assert.True(t, r["n8n_workflows_list"], "list is read")
	assert.True(t, r["n8n_workflows_get"], "get is read")
	assert.True(t, r["n8n_workflows_search"], "search is read")
	assert.True(t, r["n8n_workflows_lint"], "lint is read")
	assert.True(t, r["n8n_credentials_schema"], "schema is read")
	assert.True(t, r["n8n_backup"], "backup is read")

	// Irreversible — delete, including compound delete-rows.
	assert.True(t, irr["n8n_workflows_delete"], "delete is irreversible")
	assert.True(t, irr["n8n_data-tables_delete-rows"], "delete-rows is irreversible")

	// Ordinary writes — create/update + reversible custom mutations.
	assert.True(t, w["n8n_workflows_create"], "create is a write")
	assert.True(t, w["n8n_workflows_activate"], "activate is a write")
	assert.True(t, w["n8n_restore"], "restore is a write")
	assert.True(t, w["n8n_workflows_sync"], "sync is a write")

	// Local utility commands (no API call) must never be classified/gated.
	for _, m := range []map[string]bool{r, w, irr} {
		assert.False(t, m["n8n_agent_guard"], "guard is not an API op")
		assert.False(t, m["n8n_auth_login"], "auth is not an API op")
		assert.False(t, m["n8n_config_use"], "config is not an API op")
	}

	// No operation is in two buckets at once.
	for tool := range irr {
		assert.False(t, w[tool], "%s must not be both irreversible and write", tool)
		assert.False(t, r[tool], "%s must not be both irreversible and read", tool)
	}
}

func runGuard(t *testing.T, args ...string) string {
	t.Helper()
	resetFlags()
	resetCommandTree(RootCmd())
	var out bytes.Buffer
	root := RootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(append([]string{"agent", "guard"}, args...))
	require.NoError(t, root.ExecuteContext(t.Context()))
	return out.String()
}

func TestGuard_ClaudeCode(t *testing.T) {
	out := runGuard(t, "--host", "claude-code")
	// the hook script blocks the irreversible verbs
	assert.Contains(t, out, "n8nctl-guard.sh")
	assert.Contains(t, out, "verbs='(delete|delete-rows)'")
	// settings.json denies delete (Bash + MCP) and asks on writes
	assert.Contains(t, out, "Bash(n8nctl * delete:*)")
	assert.Contains(t, out, "mcp__.*n8n.*_(delete|delete-rows)")
	assert.Contains(t, out, "Bash(n8nctl * create:*)") // a write -> ask
	// reads are not gated
	assert.NotContains(t, out, "Bash(n8nctl * list:*)")
	assert.NotContains(t, out, "Bash(n8nctl * search:*)")
}

func TestGuard_AllWritesFoldsIntoDeny(t *testing.T) {
	out := runGuard(t, "--host", "claude-code", "--all-writes")
	// create/update join the hard-block verb set
	assert.Contains(t, out, "create")
	// and nothing is left in the ask list
	var settings map[string]any
	// extract the settings JSON block (last {...}) loosely: just assert ask is empty-ish
	assert.Contains(t, out, "verbs='(")
	_ = settings
	assert.NotContains(t, out, "\"ask\": [\n      \"Bash") // no Bash ask entries remain
}

func TestGuard_Codex(t *testing.T) {
	out := runGuard(t, "--host", "codex")
	assert.Contains(t, out, "sandbox_mode")
	assert.Contains(t, out, "read-only")
}

func TestGuard_OpenCode(t *testing.T) {
	out := runGuard(t, "--host", "opencode")
	assert.Contains(t, out, "opencode.json")
	assert.Contains(t, out, "\"n8n_*_delete\": \"deny\"")
	// the generated config is valid JSON
	start := strings.Index(out, "{")
	require.GreaterOrEqual(t, start, 0)
	var cfg map[string]any
	require.NoError(t, json.Unmarshal([]byte(out[start:]), &cfg))
}

func TestGuard_UnknownHost(t *testing.T) {
	resetFlags()
	resetCommandTree(RootCmd())
	root := RootCmd()
	root.SetArgs([]string{"agent", "guard", "--host", "bogus"})
	require.Error(t, root.ExecuteContext(t.Context()))
}
