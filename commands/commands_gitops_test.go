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

const wfJSON = `{"name":"order-sync","nodes":[` +
	`{"id":"a","name":"Hook","type":"n8n-nodes-base.webhook","parameters":{"path":"o"}},` +
	`{"id":"b","name":"Code","type":"n8n-nodes-base.code","parameters":{"jsCode":"const a=1;\nconst b=2;\nreturn [];"}}],` +
	`"connections":{"Hook":{"main":[[{"node":"Code","type":"main","index":0}]]}},"settings":{}}`

func TestConvertAndExternalize(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	dir := t.TempDir()
	in := filepath.Join(dir, "wf.json")
	require.NoError(t, os.WriteFile(in, []byte(wfJSON), 0o600))

	_, _, err := run(t, "workflows", "convert", in, "--to", "yaml", "--externalize", "2")
	require.NoError(t, err)
	yaml, rerr := os.ReadFile(filepath.Join(dir, "wf.yaml"))
	require.NoError(t, rerr)
	assert.Contains(t, string(yaml), "$ref")         // jsCode externalized
	assert.NotContains(t, string(yaml), "const a=1") // not inline
	_, serr := os.Stat(filepath.Join(dir, "_subfiles", "wf", "Code-jsCode.js"))
	require.NoError(t, serr)

	// round-trip back to JSON re-inlines the code
	_, _, err = run(t, "workflows", "convert", filepath.Join(dir, "wf.yaml"), "--to", "json", "--out", t.TempDir())
	require.NoError(t, err)
}

func TestLintFile(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	f := filepath.Join(t.TempDir(), "wf.json")
	require.NoError(t, os.WriteFile(f, []byte(wfJSON), 0o600)) // Hook lacks webhookId -> error
	_, _, err := run(t, "workflows", "lint", "-f", f)
	require.Error(t, err) // non-zero on lint errors
	assert.Contains(t, err.Error(), "error")

	// disabling the rule passes
	_, _, err = run(t, "workflows", "lint", "-f", f, "--disable-rule", "webhook-id-required")
	require.NoError(t, err)

	// list-rules works and shows provenance
	out, _, err := run(t, "workflows", "lint", "--list-rules")
	require.NoError(t, err)
	assert.Contains(t, out, "basis:")
}

func gitopsServer(t *testing.T, existing string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[` + existing + `],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/workflows/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"` + r.PathValue("id") + `","name":"keeper","nodes":[],"connections":{},"settings":{}}`))
	})
	mux.HandleFunc("POST /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"new1","name":"created"}`))
	})
	mux.HandleFunc("DELETE /api/v1/workflows/{id}", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestApplyReconcile(t *testing.T) {
	// instance has "keeper" (id k1); desired dir has only "fresh" -> create fresh, prune keeper
	srv := gitopsServer(t, `{"id":"k1","name":"keeper"}`)
	setupProfile(t, srv.URL)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "fresh.yaml"),
		[]byte("name: fresh\nnodes: []\nconnections: {}\nsettings: {}\n"), 0o600))

	// dry-run plan (no writes): create fresh, and with --prune, plan to remove keeper
	out, _, err := run(t, "workflows", "apply", "--dir", dir, "--dry-run", "--prune")
	require.NoError(t, err)
	assert.Contains(t, out, "+ create fresh")
	assert.Contains(t, out, "- prune keeper")
	assert.Contains(t, out, "1 created, 0 updated, 0 unchanged, 1 pruned")

	// real apply without prune: creates fresh, leaves keeper
	out, _, err = run(t, "workflows", "apply", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "1 created")
}

func TestApplyUnchanged(t *testing.T) {
	// instance "keeper" matches the desired file -> unchanged
	srv := gitopsServer(t, `{"id":"k1","name":"keeper"}`)
	setupProfile(t, srv.URL)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "keeper.json"),
		[]byte(`{"name":"keeper","nodes":[],"connections":{},"settings":{}}`), 0o600))
	out, _, err := run(t, "workflows", "apply", "--dir", dir, "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "1 unchanged")
}

func TestDiffAgainstFile(t *testing.T) {
	srv := gitopsServer(t, `{"id":"k1","name":"keeper"}`)
	setupProfile(t, srv.URL)
	f := filepath.Join(t.TempDir(), "keeper.json")
	require.NoError(t, os.WriteFile(f,
		[]byte(`{"name":"keeper","nodes":[{"id":"n","name":"New","type":"x","parameters":{}}],"connections":{},"settings":{}}`), 0o600))
	out, _, err := run(t, "workflows", "diff", "k1", "--file", f)
	require.NoError(t, err)
	assert.Contains(t, out, "+++")
	assert.Contains(t, out, "New")
}

func TestLintDirAndRemote(t *testing.T) {
	// --dir
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ok.json"),
		[]byte(`{"name":"ok","nodes":[],"connections":{}}`), 0o600))
	srv := gitopsServer(t, `{"id":"k1","name":"keeper"}`)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "workflows", "lint", "--dir", dir, "-o", "json")
	require.NoError(t, err)

	// --remote exercises the live-lint path (keeper has no nodes; disable that rule)
	_, _, err = run(t, "workflows", "lint", "--remote", "--disable-rule", "required-fields")
	require.NoError(t, err)

	// no source -> error
	_, _, err = run(t, "workflows", "lint")
	require.Error(t, err)
}

func TestDiffAcrossProfiles(t *testing.T) {
	left := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"x","name":"shared","nodes":[],"connections":{},"settings":{"executionOrder":"v1"}}`))
	}))
	t.Cleanup(left.Close)
	right := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// list (find by name) then it is read from the list directly
		_, _ = w.Write([]byte(`{"data":[{"id":"y","name":"shared","nodes":[],"connections":{},"settings":{"executionOrder":"v0"}}],"nextCursor":null}`))
	}))
	t.Cleanup(right.Close)
	setupProfile(t, left.URL)
	addProfile(t, "other", right.URL)
	out, _, err := run(t, "workflows", "diff", "x", "--to", "other")
	require.NoError(t, err)
	assert.Contains(t, out, "executionOrder")
}

func TestConvertBadFormat(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	f := filepath.Join(t.TempDir(), "wf.json")
	require.NoError(t, os.WriteFile(f, []byte(wfJSON), 0o600))
	_, _, err := run(t, "workflows", "convert", f, "--to", "xml")
	require.Error(t, err)
}

func TestApplyRequiresDir(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "workflows", "apply")
	require.Error(t, err)
}

func TestBackupYAMLAndRestore(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"order-sync"}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/workflows/w1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(wfJSON))
	})
	mux.HandleFunc("POST /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"r1","name":"order-sync"}`))
	})
	mux.HandleFunc("GET /api/v1/tags", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"data":[]}`)) })
	mux.HandleFunc("GET /api/v1/variables", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"data":[]}`)) })
	mux.HandleFunc("GET /api/v1/credentials", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"data":[]}`)) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	dir := t.TempDir()
	_, _, err := run(t, "backup", "--out", dir, "--format", "yaml", "--externalize", "2")
	require.NoError(t, err)
	files, _ := os.ReadDir(filepath.Join(dir, "workflows"))
	var hasYAML bool
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			hasYAML = true
		}
	}
	assert.True(t, hasYAML, "backup should write .yaml workflow files")

	// restore reads the YAML (re-inlining externalized code) and recreates
	out, _, err := run(t, "restore", "--in", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "restored")
}
