package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestConfigSetURLAndKey(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "config", "set-url", "https://new.example.com/api/v1")
	require.NoError(t, err)
	_, _, err = run(t, "config", "set-api-key", "k-123")
	require.NoError(t, err)
	storedKey, _ := keyring.Get("n8nctl-cli", "default")
	assert.Equal(t, "k-123", storedKey)

	// config show is an alias for view
	out, _, err := run(t, "config", "show", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "new.example.com")
}

func TestTopLevelLoginLogout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "login", "--api-key", "test-key", "--base-url", srv.URL)
	require.NoError(t, err)
	assert.Contains(t, out, "authenticated")
	_, _, err = run(t, "logout")
	require.NoError(t, err)
}

func TestDataTableRowInput(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/data-tables/d1/rows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1}]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	// via --file
	f := filepath.Join(t.TempDir(), "rows.json")
	require.NoError(t, os.WriteFile(f, []byte(`[{"sku":"A"}]`), 0o600))
	_, _, err := run(t, "data-tables", "add-rows", "d1", "--file", f, "-o", "json")
	require.NoError(t, err)

	// via --stdin
	feedStdin(t, `[{"sku":"B"}]`)
	_, _, err = run(t, "data-tables", "add-rows", "d1", "--stdin", "-o", "json")
	require.NoError(t, err)

	// no body -> error
	_, _, err = run(t, "data-tables", "add-rows", "d1")
	require.Error(t, err)
}

func TestSkillsResolveAllAgents(t *testing.T) {
	for _, agent := range agentNames() {
		target, err := resolveSkillTarget(agent, "", true)
		require.NoError(t, err, agent)
		assert.Contains(t, target, "n8nctl-cli")
		// project-level too
		_, err = resolveSkillTarget(agent, "", false)
		require.NoError(t, err, agent)
	}
	// explicit --dir wins
	target, err := resolveSkillTarget("claude", "/tmp/x", false)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/tmp/x", "n8nctl-cli"), target)
}
