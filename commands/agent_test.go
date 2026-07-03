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

	// packages import creates workflows/credentials on the instance — it must
	// be gated as a write, never fall through as an unannotated local command.
	assert.True(t, w["n8n_packages_import"], "packages import is a remote write")
	assert.True(t, r["n8n_packages_export"], "packages export only reads the instance")

	// Local utility commands (no API call) must never be classified/gated.
	for _, m := range []map[string]bool{r, w, irr} {
		assert.False(t, m["n8n_agent_guard"], "guard is not an API op")
		assert.False(t, m["n8n_auth_login"], "auth is not an API op")
		assert.False(t, m["n8n_config_use"], "config is not an API op")
		// The raw `api` escape hatch cannot be classified by verb; it is
		// covered separately (method-position patterns in the hook + rules).
		assert.False(t, m["n8n_api"], "raw api escape is not verb-classifiable")
	}

	// Alias spellings are enumerated so `n8nctl wf delete` cannot bypass the
	// Bash-surface rules built from these commands.
	for _, gc := range irreversible {
		if gc.tool == "n8n_workflows_delete" {
			assert.Contains(t, gc.spellings, "workflows delete")
			assert.Contains(t, gc.spellings, "wf delete")
			assert.Contains(t, gc.spellings, "workflow delete")
		}
		if gc.tool == "n8n_executions_prune" {
			assert.Contains(t, gc.spellings, "exec prune")
		}
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
	// the hook script blocks the irreversible operations by exact path
	assert.Contains(t, out, "n8nctl-guard.sh")
	assert.Contains(t, out, "'workflows delete'") // hook blocked_cmds entry
	assert.Contains(t, out, "'wf delete'")        // alias spelling is covered too
	// settings.json denies delete (Bash + MCP) and asks on writes
	assert.Contains(t, out, "Bash(n8nctl workflows delete:*)")
	assert.Contains(t, out, "Bash(n8nctl wf delete:*)")
	assert.Contains(t, out, "Bash(n8nctl api DELETE:*)") // raw-api escape hatch
	assert.Contains(t, out, "Bash(n8nctl api POST:*)")
	assert.Contains(t, out, "mcp__n8nctl__n8n_workflows_delete") // exact MCP deny
	assert.Contains(t, out, "Bash(n8nctl workflows create:*)")   // a write -> ask
	assert.Contains(t, out, "Bash(n8nctl packages import:*)")    // remote write -> ask
	// reads are not gated
	assert.NotContains(t, out, "Bash(n8nctl workflows list:*)")
	assert.NotContains(t, out, "Bash(n8nctl workflows search:*)")
	// no verb-wildcard rules remain (they can't match top-level commands and
	// false-match unrelated paths)
	assert.NotContains(t, out, "Bash(n8nctl * ")
}

func TestGuard_AllWritesFoldsIntoDeny(t *testing.T) {
	out := runGuard(t, "--host", "claude-code", "--all-writes")
	// create/update join the hard-block set
	assert.Contains(t, out, "'workflows create'")
	assert.Contains(t, out, "Bash(n8nctl workflows create:*)")
	// and nothing is left in the ask list
	assert.Contains(t, out, `"ask": []`)
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
	assert.Contains(t, out, "\"n8n_workflows_delete\": \"deny\"")
	assert.Contains(t, out, "\"n8nctl wf delete\": \"deny\"")
	assert.Contains(t, out, "\"n8nctl api DELETE\": \"deny\"")
	assert.Contains(t, out, "\"n8nctl packages import\": \"ask\"")
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
