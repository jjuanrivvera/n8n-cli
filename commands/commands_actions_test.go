package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fullServer mounts every route the action tests touch, using Go 1.22 method+path
// patterns. Each handler returns a minimal valid body.
func fullServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	ok := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }
	}
	// workflows
	mux.HandleFunc("GET /api/v1/workflows", ok(`{"data":[{"id":"w1","name":"Demo","active":true}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/workflows", ok(`{"id":"w2","name":"Imported"}`))
	mux.HandleFunc("GET /api/v1/workflows/w1", ok(`{"id":"w1","name":"Demo"}`))
	mux.HandleFunc("PUT /api/v1/workflows/w1", ok(`{"id":"w1","name":"Renamed"}`))
	mux.HandleFunc("DELETE /api/v1/workflows/w1", ok(`{"id":"w1"}`))
	mux.HandleFunc("POST /api/v1/workflows/w1/activate", ok(`{"id":"w1","active":true}`))
	mux.HandleFunc("POST /api/v1/workflows/w1/deactivate", ok(`{"id":"w1","active":false}`))
	mux.HandleFunc("POST /api/v1/workflows/w1/archive", ok(`{"id":"w1","isArchived":true}`))
	mux.HandleFunc("POST /api/v1/workflows/w1/unarchive", ok(`{"id":"w1","isArchived":false}`))
	mux.HandleFunc("PUT /api/v1/workflows/w1/transfer", ok(``))
	mux.HandleFunc("GET /api/v1/workflows/w1/tags", ok(`[{"id":"t1","name":"Prod"}]`))
	mux.HandleFunc("PUT /api/v1/workflows/w1/tags", ok(`[{"id":"t1","name":"Prod"}]`))
	// executions
	mux.HandleFunc("GET /api/v1/executions", ok(`{"data":[{"id":1000,"workflowId":55,"status":"error"}],"nextCursor":null}`))
	mux.HandleFunc("GET /api/v1/executions/1000", ok(`{"id":1000,"status":"error"}`))
	mux.HandleFunc("DELETE /api/v1/executions/1000", ok(`{"id":1000}`))
	mux.HandleFunc("POST /api/v1/executions/1000/retry", ok(`{"id":1001,"status":"success"}`))
	mux.HandleFunc("POST /api/v1/executions/1000/stop", ok(`{"id":1000,"status":"canceled"}`))
	// credentials
	mux.HandleFunc("GET /api/v1/credentials", ok(`{"data":[{"id":"c1","name":"GH","type":"githubApi"}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/credentials", ok(`{"id":"c2","name":"New","type":"githubApi"}`))
	mux.HandleFunc("GET /api/v1/credentials/c1", ok(`{"id":"c1","name":"GH","type":"githubApi"}`))
	mux.HandleFunc("PATCH /api/v1/credentials/c1", ok(`{"id":"c1","name":"GH2","type":"githubApi"}`))
	mux.HandleFunc("DELETE /api/v1/credentials/c1", ok(`{"id":"c1"}`))
	mux.HandleFunc("GET /api/v1/credentials/schema/githubApi", ok(`{"type":"object","properties":{"accessToken":{}}}`))
	mux.HandleFunc("PUT /api/v1/credentials/c1/transfer", ok(``))
	// tags
	mux.HandleFunc("GET /api/v1/tags", ok(`{"data":[{"id":"t1","name":"Prod"}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/tags", ok(`{"id":"t2","name":"X"}`))
	mux.HandleFunc("PUT /api/v1/tags/t1", ok(`{"id":"t1","name":"Staging"}`))
	mux.HandleFunc("DELETE /api/v1/tags/t1", ok(`{"id":"t1"}`))
	// variables
	mux.HandleFunc("GET /api/v1/variables", ok(`{"data":[{"id":"v1","key":"API_BASE","value":"x"}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/variables", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201) })
	mux.HandleFunc("PUT /api/v1/variables/v1", ok(`{"id":"v1","key":"API_BASE","value":"y"}`))
	mux.HandleFunc("DELETE /api/v1/variables/v1", ok(``))
	// projects
	mux.HandleFunc("GET /api/v1/projects", ok(`{"data":[{"id":"p1","name":"Team","type":"team"}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/projects", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201) })
	mux.HandleFunc("PUT /api/v1/projects/p1", ok(`{"id":"p1","name":"Team2"}`))
	mux.HandleFunc("DELETE /api/v1/projects/p1", ok(``))
	mux.HandleFunc("GET /api/v1/projects/p1/users", ok(`{"data":[{"id":"u1","email":"a@x.com","role":"project:admin"}],"nextCursor":null}`))
	mux.HandleFunc("POST /api/v1/projects/p1/users", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(201) })
	mux.HandleFunc("PATCH /api/v1/projects/p1/users/u1", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
	mux.HandleFunc("DELETE /api/v1/projects/p1/users/u1", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
	// users
	mux.HandleFunc("GET /api/v1/users", ok(`{"data":[{"id":"u1","email":"a@x.com","role":"global:owner"}],"nextCursor":null}`))
	mux.HandleFunc("GET /api/v1/users/u1", ok(`{"id":"u1","email":"a@x.com"}`))
	mux.HandleFunc("POST /api/v1/users", ok(`[{"user":{"id":"u9","email":"new@x.com"}}]`))
	mux.HandleFunc("DELETE /api/v1/users/u1", ok(``))
	mux.HandleFunc("PATCH /api/v1/users/u1/role", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	// operations
	mux.HandleFunc("POST /api/v1/audit", ok(`{"Credentials Risk Report":{}}`))
	mux.HandleFunc("POST /api/v1/source-control/pull", ok(`[]`))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestActions_Workflows(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	cases := [][]string{
		{"workflows", "get", "w1", "-o", "json", "--quiet"},
		{"workflows", "update", "w1", "--set", "name=Renamed", "-o", "json"},
		{"workflows", "delete", "w1", "--yes"},
		{"workflows", "activate", "w1", "-o", "json"},
		{"workflows", "deactivate", "w1", "-o", "json"},
		{"workflows", "archive", "w1", "-o", "json"},
		{"workflows", "unarchive", "w1", "-o", "json"},
		{"workflows", "transfer", "w1", "--project", "p1"},
		{"workflows", "tags", "w1", "-o", "json"},
		{"workflows", "tags", "w1", "--set", "t1,t2", "-o", "json"},
	}
	for _, args := range cases {
		_, _, err := run(t, args...)
		require.NoError(t, err, "args: %v", args)
	}
}

func TestActions_WorkflowsCreateFromFile(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	f := filepath.Join(t.TempDir(), "wf.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"name":"Imported","nodes":[],"connections":{},"settings":{}}`), 0o600))
	out, _, err := run(t, "workflows", "create", "--file", f, "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Imported")
}

func TestActions_Executions(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	cases := [][]string{
		{"executions", "list", "--status", "error", "-o", "json", "--quiet"},
		{"executions", "get", "1000", "--include-data", "-o", "json"},
		{"executions", "retry", "1000", "--load-workflow", "-o", "json"},
		{"executions", "stop", "1000", "-o", "json"},
		{"executions", "delete", "1000", "--yes"},
	}
	for _, args := range cases {
		_, _, err := run(t, args...)
		require.NoError(t, err, "args: %v", args)
	}
}

func TestActions_Credentials(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	cases := [][]string{
		{"credentials", "list", "-o", "json", "--quiet"},
		{"credentials", "get", "c1", "-o", "json"},
		{"credentials", "create", "--set", "name=New", "--set", "type=githubApi", "--set", `data={"accessToken":"x"}`, "-o", "json"},
		{"credentials", "update", "c1", "--set", "name=GH2", "-o", "json"},
		{"credentials", "schema", "githubApi", "-o", "json"},
		{"credentials", "transfer", "c1", "--project", "p1"},
		{"credentials", "delete", "c1", "--yes"},
	}
	for _, args := range cases {
		_, _, err := run(t, args...)
		require.NoError(t, err, "args: %v", args)
	}
}

func TestActions_TagsVariablesProjectsUsers(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	cases := [][]string{
		{"tags", "update", "t1", "--set", "name=Staging", "-o", "json"},
		{"tags", "delete", "t1", "--yes"},
		{"variables", "list", "-o", "json", "--quiet"},
		{"variables", "get", "API_BASE", "-o", "json"}, // resolved via list by key
		{"variables", "create", "--set", "key=API_BASE", "--set", "value=x"},
		{"variables", "update", "v1", "--set", "value=y", "-o", "json"},
		{"variables", "delete", "v1", "--yes"},
		{"projects", "list", "-o", "json", "--quiet"},
		{"projects", "create", "--set", "name=Team"},
		{"projects", "update", "p1", "--set", "name=Team2", "-o", "json"},
		{"projects", "members", "p1", "-o", "json"},
		{"projects", "add-member", "p1", "--user", "u2", "--role", "project:viewer"},
		{"projects", "set-member-role", "p1", "u1", "--role", "project:editor"},
		{"projects", "remove-member", "p1", "u1"},
		{"projects", "delete", "p1", "--yes"},
		{"users", "list", "-o", "json", "--quiet"},
		{"users", "get", "u1", "-o", "json"},
		{"users", "invite", "--email", "new@x.com", "--role", "global:member", "-o", "json"},
		{"users", "change-role", "u1", "--role", "global:admin"},
		{"users", "delete", "u1", "--yes"},
	}
	for _, args := range cases {
		_, _, err := run(t, args...)
		require.NoError(t, err, "args: %v", args)
	}
}

func TestActions_Operations(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "audit", "--days", "30", "--categories", "credentials,nodes", "-o", "json")
	require.NoError(t, err)
	_, _, err = run(t, "source-control", "pull", "--force", "-o", "json")
	require.NoError(t, err)
}

func TestMeta_InitConfigViewCompletion(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// init non-interactively (verifies against the live server, stores in keyring).
	_, _, err := run(t, "init", "--profile", "ci", "--base-url", srv.URL, "--api-key", "test-key")
	require.NoError(t, err)
	// config view redacts secrets.
	out, _, err := run(t, "config", "view", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "profiles")
	// completion generation must not error.
	_, _, err = run(t, "completion", "bash")
	require.NoError(t, err)
	// version --json.
	out, _, err = run(t, "version", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "version")
}

func TestMeta_AuthLoginLogout(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "auth", "login", "--api-key", "test-key", "--base-url", srv.URL)
	require.NoError(t, err)
	_, _, err = run(t, "auth", "logout")
	require.NoError(t, err)
}
