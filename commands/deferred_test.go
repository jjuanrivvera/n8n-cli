package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
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
	// invalid-parameter-value is a Warning, so lint surfaces it but exits 0.
	out, _, err := run(t, "workflows", "lint", "-f", f)
	require.NoError(t, err)
	assert.Contains(t, out, "invalid-parameter-value")
}

func TestTemplatesDeployCommand(t *testing.T) {
	// mock gallery
	gallery := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search") {
			_, _ = w.Write([]byte(`{"workflows":[{"id":7,"name":"Tmpl"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"workflow":{"id":7,"name":"Tmpl","workflow":{"nodes":[{"name":"A","type":"n8n-nodes-base.noOp"}],"connections":{}}}}`))
	}))
	t.Cleanup(gallery.Close)
	t.Setenv("N8NCTL_TEMPLATE_API_URL", gallery.URL)

	// mock instance that accepts the create
	created := 0
	instance := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			created++
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), "\"name\":\"My copy\"")
			assert.Contains(t, string(body), "\"nodes\"")
		}
		_, _ = w.Write([]byte(`{"id":"new1","name":"My copy"}`))
	}))
	t.Cleanup(instance.Close)
	setupProfile(t, instance.URL)

	// search works through the env override
	out, _, err := run(t, "templates", "search", "anything")
	require.NoError(t, err)
	assert.Contains(t, out, "Tmpl")

	// deploy creates on the instance with the chosen name
	_, _, err = run(t, "templates", "deploy", "7", "--name", "My copy")
	require.NoError(t, err)
	assert.Equal(t, 1, created)
}

func TestTemplateCreateBody(t *testing.T) {
	d := &api.TemplateDetail{Name: "Tmpl", Definition: []byte(`{"nodes":[{"name":"A"}],"connections":{"A":{}}}`)}
	body, err := templateCreateBody(d, "")
	require.NoError(t, err)
	assert.Equal(t, "Tmpl", body["name"]) // falls back to template name
	assert.NotNil(t, body["settings"])    // defaults to empty object

	body, err = templateCreateBody(d, "Custom")
	require.NoError(t, err)
	assert.Equal(t, "Custom", body["name"])

	// wrong-typed nodes -> clear error, not an opaque API 422
	bad := &api.TemplateDetail{Definition: []byte(`{"nodes":"oops"}`)}
	_, err = templateCreateBody(bad, "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nodes")
}

func TestBreakingChangesDirAndEmpty(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/a.json", []byte(
		`{"name":"a","nodes":[{"name":"H","type":"n8n-nodes-base.httpRequest","typeVersion":1,"parameters":{}}],"connections":{},"settings":{}}`), 0o600))
	out, _, err := run(t, "workflows", "breaking-changes", "--dir", dir, "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "httpRequest")
	// no input -> error
	_, _, err = run(t, "workflows", "breaking-changes")
	require.Error(t, err)
}

func TestNewTemplateAPIEnvOverride(t *testing.T) {
	t.Setenv("N8NCTL_TEMPLATE_API_URL", "http://example.invalid")
	assert.Equal(t, "http://example.invalid", api.NewTemplateAPI().BaseURL)
}

func galleryServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search") {
			if strings.Contains(r.URL.RawQuery, "none") {
				_, _ = w.Write([]byte(`{"workflows":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"workflows":[{"id":7,"name":"Tmpl","totalViews":5}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"workflow":{"id":7,"name":"Tmpl","workflow":{"nodes":[{"name":"A","type":"n8n-nodes-base.noOp"}],"connections":{}}}}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("N8NCTL_TEMPLATE_API_URL", srv.URL)
	return srv.URL
}

func TestTemplatesGetAndDryRun(t *testing.T) {
	galleryServer(t)
	setupProfile(t, "https://a/api/v1")
	// get prints the definition
	out, _, err := run(t, "templates", "get", "7", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "nodes")
	// deploy --dry-run does not hit the instance (prints the curl)
	_, _, err = run(t, "--dry-run", "templates", "deploy", "7")
	require.NoError(t, err)
	// search with no results errors
	_, _, err = run(t, "templates", "search", "none")
	require.Error(t, err)
}

func TestBreakingChangesRemoteAndId(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows/42", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"42","name":"w","nodes":[{"name":"H","type":"n8n-nodes-base.httpRequest","typeVersion":1,"parameters":{}}],"connections":{},"settings":{}}`))
	})
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"42","name":"w","nodes":[{"name":"H","type":"n8n-nodes-base.httpRequest","typeVersion":1,"parameters":{}}],"connections":{},"settings":{}}],"nextCursor":null}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "breaking-changes", "42")
	require.NoError(t, err)
	assert.Contains(t, out, "typeVersion 1")
	out, _, err = run(t, "workflows", "breaking-changes", "--remote")
	require.NoError(t, err)
	assert.Contains(t, out, "httpRequest")
}

func TestTemplatesDeployActivate(t *testing.T) {
	galleryServer(t)
	activated := 0
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"new1","name":"Tmpl"}`))
	})
	mux.HandleFunc("POST /api/v1/workflows/{id}/activate", func(w http.ResponseWriter, _ *http.Request) {
		activated++
		_, _ = w.Write([]byte(`{"id":"new1","active":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "templates", "deploy", "7", "--activate")
	require.NoError(t, err)
	assert.Equal(t, 1, activated)
}

func TestTemplatesAndBreakingErrorPaths(t *testing.T) {
	// gallery returns 404 -> get/deploy error
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }))
	t.Cleanup(bad.Close)
	t.Setenv("N8NCTL_TEMPLATE_API_URL", bad.URL)
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "templates", "get", "1")
	require.Error(t, err)
	_, _, err = run(t, "templates", "deploy", "1")
	require.Error(t, err)
	// breaking-changes on a missing file -> error
	_, _, err = run(t, "workflows", "breaking-changes", "-f", "/no/such/file.json")
	require.Error(t, err)
}
