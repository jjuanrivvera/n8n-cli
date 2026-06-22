package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

// addProfile registers an extra profile (with a keyring key) on top of the one
// created by setupProfile.
func addProfile(t *testing.T, name, baseURL string) {
	t.Helper()
	c, err := config.Load()
	require.NoError(t, err)
	c.SetProfile(&config.Profile{Name: name, BaseURL: baseURL})
	require.NoError(t, c.Save())
	require.NoError(t, keyring.Set("n8nctl-cli", name, "test-key"))
}

const sampleWorkflow = `{
  "id":"w1","name":"Order Sync","active":true,
  "nodes":[
    {"name":"Webhook","type":"n8n-nodes-base.webhook","parameters":{"path":"orders"}},
    {"name":"Slack","type":"n8n-nodes-base.slack","parameters":{},"credentials":{"slackApi":{"id":"c1","name":"Team Slack"}}}
  ],
  "connections":{},"settings":{"executionOrder":"v1"}
}`

func TestWorkflowSync(t *testing.T) {
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/workflows/w1" {
			_, _ = w.Write([]byte(sampleWorkflow))
			return
		}
		t.Fatalf("unexpected src %s", r.URL.Path)
	}))
	t.Cleanup(src.Close)

	var posted map[string]any
	dst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodPost:
			_ = json.NewDecoder(r.Body).Decode(&posted)
			_, _ = w.Write([]byte(`{"id":"new1","name":"Order Sync"}`))
		case r.URL.Path == "/api/v1/workflows/new1/activate":
			_, _ = w.Write([]byte(`{"id":"new1","active":true}`))
		default:
			t.Fatalf("unexpected dst %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(dst.Close)

	setupProfile(t, src.URL)
	addProfile(t, "prod", dst.URL)

	out, _, err := run(t, "workflows", "sync", "w1", "--to", "prod", "--activate", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "new1")
	// Read-only fields stripped; structural fields carried over.
	assert.Equal(t, "Order Sync", posted["name"])
	assert.NotContains(t, posted, "id")
	assert.NotContains(t, posted, "active")
	assert.Contains(t, posted, "nodes")
}

func TestWorkflowSyncDryRun(t *testing.T) {
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sampleWorkflow))
	}))
	t.Cleanup(src.Close)
	setupProfile(t, src.URL)
	addProfile(t, "prod", "https://prod.example.com/api/v1")

	out, _, err := run(t, "workflows", "sync", "w1", "--to", "prod", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "prod.example.com")
}

func TestWorkflowSyncSameProfileErrors(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "workflows", "sync", "w1", "--from", "default", "--to", "default")
	require.Error(t, err)
}

func TestWorkflowSearch(t *testing.T) {
	list := `{"data":[` + sampleWorkflow + `],"nextCursor":null}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(list))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	// by node type
	out, _, err := run(t, "workflows", "search", "--node", "slack", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Order Sync")
	// by credential name
	out, _, err = run(t, "workflows", "search", "--credential", "Team Slack", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Order Sync")
	// by webhook path
	out, _, err = run(t, "workflows", "search", "--webhook", "/orders", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Order Sync")
	// by name regex
	out, _, err = run(t, "workflows", "search", "--name", "^Order", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Order Sync")
	// no filter -> error
	_, _, err = run(t, "workflows", "search")
	require.Error(t, err)
	// bad regex -> error
	_, _, err = run(t, "workflows", "search", "--name", "(")
	require.Error(t, err)
}

func TestBackupAndRestore(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[` + sampleWorkflow + `],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/workflows/w1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sampleWorkflow))
	})
	mux.HandleFunc("POST /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"r1","name":"Order Sync"}`))
	})
	mux.HandleFunc("GET /api/v1/tags", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"t1","name":"prod"}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/variables", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"v1","key":"K","value":"V"}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/credentials", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"c1","name":"Team Slack","type":"slackApi"}],"nextCursor":null}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	dir := t.TempDir()
	_, _, err := run(t, "backup", "--out", dir)
	require.NoError(t, err)
	for _, f := range []string{"manifest.json", "tags.json", "variables.json"} {
		_, statErr := os.Stat(filepath.Join(dir, f))
		require.NoError(t, statErr, f)
	}
	wfFiles, _ := os.ReadDir(filepath.Join(dir, "workflows"))
	require.Len(t, wfFiles, 1)

	// restore from the same directory
	out, _, err := run(t, "restore", "--in", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "restored")
}

func TestBackupRequiresOut(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "backup")
	require.Error(t, err)
	_, _, err = run(t, "restore")
	require.Error(t, err)
}

func TestWorkflowSyncUpdateByName(t *testing.T) {
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sampleWorkflow))
	}))
	t.Cleanup(src.Close)

	var putHit bool
	dst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/workflows" && r.Method == http.MethodGet:
			// destination already has a workflow with the same name
			_, _ = w.Write([]byte(`{"data":[{"id":"existing9","name":"Order Sync"}],"nextCursor":null}`))
		case r.URL.Path == "/api/v1/workflows/existing9" && r.Method == http.MethodPut:
			putHit = true
			_, _ = w.Write([]byte(`{"id":"existing9","name":"Order Sync"}`))
		default:
			t.Fatalf("unexpected dst %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(dst.Close)

	setupProfile(t, src.URL)
	addProfile(t, "prod", dst.URL)
	out, _, err := run(t, "workflows", "sync", "w1", "--to", "prod", "--update-by-name", "-o", "json")
	require.NoError(t, err)
	assert.True(t, putHit, "should PUT-update the same-named workflow")
	assert.Contains(t, out, "existing9")
}

func TestRestoreUpdateByName(t *testing.T) {
	var putHit bool
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"ex1","name":"Order Sync"}],"nextCursor":null}`))
	})
	mux.HandleFunc("PUT /api/v1/workflows/ex1", func(w http.ResponseWriter, _ *http.Request) {
		putHit = true
		_, _ = w.Write([]byte(`{"id":"ex1","name":"Order Sync"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "workflows"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "workflows", "order.w1.json"), []byte(sampleWorkflow), 0o600))

	out, _, err := run(t, "restore", "--in", dir, "--update-by-name")
	require.NoError(t, err)
	assert.True(t, putHit)
	assert.Contains(t, out, "1 updated")
}

func TestSearchByCredentialID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[` + sampleWorkflow + `],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "search", "--credential", "c1", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Order Sync")
}

func TestSyncExplicitFrom(t *testing.T) {
	src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sampleWorkflow))
	}))
	t.Cleanup(src.Close)
	dst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"n2","name":"Order Sync"}`))
	}))
	t.Cleanup(dst.Close)
	setupProfile(t, "https://unused/api/v1")
	addProfile(t, "dev", src.URL)
	addProfile(t, "prod", dst.URL)
	out, _, err := run(t, "workflows", "sync", "w1", "--from", "dev", "--to", "prod", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "n2")
}

func TestSearchNoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[` + sampleWorkflow + `],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "search", "--node", "nonexistent", "-o", "json")
	require.NoError(t, err)
	assert.NotContains(t, out, "Order Sync")
}
