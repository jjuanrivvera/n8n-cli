package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicCompletion(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	// Add a second profile so `config use` completion has options.
	_, _, err := run(t, "config", "set", "base_url", "https://b/api/v1", "--profile", "cloud")
	require.NoError(t, err)

	// profileNames via the `config use` ValidArgsFunction.
	out, _, err := run(t, "__complete", "config", "use", "")
	require.NoError(t, err)
	assert.Contains(t, out, "cloud")

	// fixedCompletions via a list filter with allowed values.
	out, _, err = run(t, "__complete", "executions", "list", "--status", "")
	require.NoError(t, err)
	assert.Contains(t, out, "error")
}

func TestLoggerLevelsAndUpdateDryRun(t *testing.T) {
	setupProfile(t, "https://n8n.example.com/api/v1")
	_, _, err := run(t, "config", "set", "log_level", "info")
	require.NoError(t, err)
	// info-level logger path + update dry-run (prints PUT curl).
	out, _, err := run(t, "workflows", "update", "w1", "--set", "name=x", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X PUT")
}
