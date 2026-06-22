package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression: `list` must apply the resource's default columns so the table does
// not dump nested fields (nodes/connections). Found during live testing where
// `workflows list` rendered the entire node graph inline.
func TestListAppliesDefaultColumns(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"Demo","active":true,"triggerCount":1,` +
			`"nodes":[{"type":"n8n-nodes-base.webhook"}],"connections":{"a":{"main":[]}}}],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	out, _, err := run(t, "workflows", "list", "--quiet", "--no-color")
	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "TRIGGERCOUNT")
	assert.NotContains(t, out, "NODES", "default columns should hide the nodes graph")
	assert.NotContains(t, out, "CONNECTIONS", "default columns should hide connections")

	// --columns still overrides the default set.
	out, _, err = run(t, "workflows", "list", "--columns", "id", "--quiet", "--no-color")
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.NotContains(t, out, "NAME")
}

// Regression: `backup` must tolerate a licensed-feature 403 (e.g. variables on a
// Community instance) and still back up the rest, recording what was skipped.
func TestBackupToleratesForbidden(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"WF"}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/workflows/w1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"w1","name":"WF","nodes":[],"connections":{},"settings":{}}`))
	})
	mux.HandleFunc("GET /api/v1/tags", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/variables", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"Your license does not allow for feat:variables"}`))
	})
	mux.HandleFunc("GET /api/v1/credentials", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	dir := t.TempDir()
	out, _, err := run(t, "backup", "--out", dir)
	require.NoError(t, err, "backup must not abort on a licensed-feature 403")
	assert.Contains(t, out, "skipped")

	manifest, rerr := os.ReadFile(filepath.Join(dir, "manifest.json"))
	require.NoError(t, rerr)
	assert.True(t, strings.Contains(string(manifest), "variables"), "manifest should record the skipped section")
	// The core workflow was still backed up.
	wfFiles, _ := os.ReadDir(filepath.Join(dir, "workflows"))
	assert.Len(t, wfFiles, 1)
}

// keySource should distinguish a config-file key from a real env var.
func TestDoctorKeySourceConfigVsEnv(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	assert.Equal(t, "from config file", keySource("default", "abc"))
	t.Setenv("N8NCTL_API_KEY", "envkey")
	assert.Equal(t, "from N8NCTL_API_KEY env", keySource("default", "abc"))
}
