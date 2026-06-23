package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func TestNodesCommand(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	out, _, err := run(t, "nodes", "search", "slack")
	require.NoError(t, err)
	assert.Contains(t, out, "n8n-nodes-base.slack")

	out, _, err = run(t, "nodes", "show", "n8n-nodes-base.slack")
	require.NoError(t, err)
	assert.Contains(t, out, "Slack")

	_, _, err = run(t, "nodes", "show", "n8n-nodes-base.slak")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did you mean")
}

func TestAutofixWorkflow(t *testing.T) {
	wf := &api.Workflow{
		Name: "x",
		Nodes: json.RawMessage(`[` +
			`{"name":"S","type":"n8n-nodes-base.slak","parameters":{"text":"{{ $json.x }}"}},` +
			`{"name":"H","type":"n8n-nodes-base.webhook","parameters":{}}]`),
	}
	fixes, err := autofixWorkflow(wf)
	require.NoError(t, err)
	assert.Len(t, fixes, 3) // type typo, expression prefix, webhookId
	s := string(wf.Nodes)
	assert.Contains(t, s, "n8n-nodes-base.slack") // typo corrected
	assert.Contains(t, s, "={{ $json.x }}")       // expression prefixed
	assert.Contains(t, s, "webhookId")            // generated

	// a clean workflow yields no fixes
	clean := &api.Workflow{Nodes: json.RawMessage(`[{"name":"A","type":"n8n-nodes-base.set","parameters":{}}]`)}
	fixes, err = autofixWorkflow(clean)
	require.NoError(t, err)
	assert.Empty(t, fixes)
}

func TestAgeCutoff(t *testing.T) {
	c, err := ageCutoff("30d")
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(-30*24*time.Hour), c, time.Minute)
	c, err = ageCutoff("720h")
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(-720*time.Hour), c, time.Minute)
	_, err = ageCutoff("nonsense")
	require.Error(t, err)
}

// execServer mocks the executions endpoints for prune/stats.
func execServer(t *testing.T, executions string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"a","active":true},{"id":"w2","name":"b","active":false}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/executions", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[` + executions + `],"nextCursor":null}`))
	})
	mux.HandleFunc("DELETE /api/v1/executions/{id}", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestExecutionsPruneDryRun(t *testing.T) {
	old := `{"id":"1","status":"error","startedAt":"2020-01-01T00:00:00.000Z"}`
	recent := `{"id":"2","status":"success","startedAt":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	srv := execServer(t, old+","+recent)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "executions", "prune", "--older-than", "30d", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "would delete 1") // only the 2020 one is old
}

func TestStats(t *testing.T) {
	srv := execServer(t, `{"id":"1","status":"success","startedAt":"2020-01-01T00:00:00.000Z"}`)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "workflows.total")
	assert.Contains(t, out, "workflows.active")
	assert.Contains(t, out, "executions.success")
}

func TestWorkflowsBulkDryRun(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"prod-api"},{"id":"w2","name":"prod-web"}],"nextCursor":null}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "bulk", "activate", "--tag", "prod", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "would activate 2 workflow(s)")
	assert.Contains(t, out, "prod-api")
}

func TestProxyRejectDuplicateName(t *testing.T) {
	created := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"e1","name":"dup"}],"nextCursor":null}`))
			return
		}
		created++
		_, _ = w.Write([]byte(`{"id":"new"}`))
	}))
	t.Cleanup(backend.Close)
	tu, _ := url.Parse(backend.URL)
	h := newLintProxy(&url.URL{Scheme: tu.Scheme, Host: tu.Host}, "secret", nil, false, true, io.Discard)
	proxy := httptest.NewServer(h)
	t.Cleanup(proxy.Close)

	// creating "dup" (already exists) -> 409, not forwarded
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/api/v1/workflows",
		strings.NewReader(`{"name":"dup","nodes":[],"connections":{},"settings":{}}`))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	assert.Equal(t, 0, created, "a duplicate-named create must not be forwarded")
}

func TestNodesList(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	out, _, err := run(t, "nodes", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "n8n-nodes-base.")
}

func TestAutofixCommandWrite(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	dir := t.TempDir()
	f := dir + "/wf.json"
	require.NoError(t, os.WriteFile(f, []byte(
		`{"name":"x","nodes":[{"name":"S","type":"n8n-nodes-base.slak","parameters":{}}],"connections":{},"settings":{}}`), 0o600))
	// report-only first
	out, _, err := run(t, "workflows", "autofix", "-f", f)
	require.NoError(t, err)
	assert.Contains(t, out, "would fix")
	// then --write
	_, _, err = run(t, "workflows", "autofix", "-f", f, "--write")
	require.NoError(t, err)
	b, _ := os.ReadFile(f)
	assert.Contains(t, string(b), "n8n-nodes-base.slack")
	// nothing left to fix
	out, _, err = run(t, "workflows", "autofix", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "nothing to fix")
}

func TestExecutionsPruneDelete(t *testing.T) {
	deleted := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/executions", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"1","status":"error","startedAt":"2020-01-01T00:00:00.000Z"}],"nextCursor":null}`))
	})
	mux.HandleFunc("DELETE /api/v1/executions/{id}", func(w http.ResponseWriter, _ *http.Request) {
		deleted++
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "executions", "prune", "--older-than", "30d", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "pruned 1")
	assert.Equal(t, 1, deleted)
}

func TestWorkflowsBulkActivate(t *testing.T) {
	activated := 0
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"prod-api"}],"nextCursor":null}`))
	})
	mux.HandleFunc("POST /api/v1/workflows/{id}/activate", func(w http.ResponseWriter, _ *http.Request) {
		activated++
		_, _ = w.Write([]byte(`{"id":"w1","active":true}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "bulk", "activate", "--tag", "prod", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "activated 1")
	assert.Equal(t, 1, activated)
}

func TestExecutionsWatchCancels(t *testing.T) {
	srv := execServer(t, `{"id":"1","status":"error","startedAt":"2020-01-01T00:00:00.000Z"}`)
	setupProfile(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	resetFlags()
	resetCommandTree(RootCmd())
	resetConfigForTest()
	root := RootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"executions", "watch", "--interval", "20ms"})
	require.NoError(t, root.ExecuteContext(ctx)) // exits cleanly on cancellation
}

func TestProxyDupNameAllowsNew(t *testing.T) {
	created := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"e1","name":"existing"}],"nextCursor":null}`))
			return
		}
		created++
		_, _ = w.Write([]byte(`{"id":"new"}`))
	}))
	t.Cleanup(backend.Close)
	tu, _ := url.Parse(backend.URL)
	proxy := httptest.NewServer(newLintProxy(&url.URL{Scheme: tu.Scheme, Host: tu.Host}, "secret", nil, false, true, io.Discard))
	t.Cleanup(proxy.Close)
	req, _ := http.NewRequest(http.MethodPost, proxy.URL+"/api/v1/workflows",
		strings.NewReader(`{"name":"brand-new","nodes":[],"connections":{},"settings":{}}`))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, created, "a new name must be forwarded")
}

func TestNewFeatureEdgeCases(t *testing.T) {
	// nodes search with no match -> error
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "nodes", "search", "zzz-no-such-node-xyz")
	require.Error(t, err)

	// autofix with no target -> error
	_, _, err = run(t, "workflows", "autofix")
	require.Error(t, err)

	// prune with neither --older-than nor --status -> error
	_, _, err = run(t, "executions", "prune")
	require.Error(t, err)

	// bulk requires --tag
	_, _, err = run(t, "workflows", "bulk", "activate")
	require.Error(t, err)
}

func TestPruneNothingAndBulkEmpty(t *testing.T) {
	// prune finds nothing old
	srv := execServer(t, `{"id":"2","status":"success","startedAt":"`+time.Now().UTC().Format(time.RFC3339)+`"}`)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "executions", "prune", "--older-than", "30d", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "nothing to prune")

	// bulk with a tag that matches no workflows
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[],"nextCursor":null}`))
	})
	s2 := httptest.NewServer(mux)
	t.Cleanup(s2.Close)
	setupProfile(t, s2.URL)
	out, _, err = run(t, "workflows", "bulk", "deactivate", "--tag", "none", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "no workflows carry")
}

func TestStatsJSON(t *testing.T) {
	srv := execServer(t, `{"id":"1","status":"error","startedAt":"2020-01-01T00:00:00.000Z"}`)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "stats", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "executions.error")
}

func TestFormatExecLine(t *testing.T) {
	e := api.Execution{ID: "1", WorkflowID: "w", Status: "error", StartedAt: "t"}
	assert.Contains(t, formatExecLine(e, false), "error")
	assert.NotContains(t, formatExecLine(e, false), "\x1b[") // no color
	assert.Contains(t, formatExecLine(e, true), "\x1b[31m")  // red for error
	assert.Contains(t, formatExecLine(api.Execution{Status: "success"}, true), "\x1b[32m")
	assert.Contains(t, formatExecLine(api.Execution{Status: "running"}, true), "\x1b[33m")
}

func TestFixExpressionsNested(t *testing.T) {
	v := map[string]any{
		"a":   "{{ $json.x }}",
		"arr": []any{"{{ $node.z }}", map[string]any{"b": "plain"}},
	}
	n := fixExpressions(v)
	assert.Equal(t, 2, n) // a and arr[0]
	assert.Equal(t, "={{ $json.x }}", v["a"])
	assert.Equal(t, "={{ $node.z }}", v["arr"].([]any)[0])
}

func TestStatsExecForbidden(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workflows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"a","active":true}],"nextCursor":null}`))
	})
	mux.HandleFunc("GET /api/v1/executions", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"forbidden"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "stats")
	require.NoError(t, err) // degrades gracefully on 403
	assert.Contains(t, out, "workflows.total")
	assert.NotContains(t, out, "executions.success")
}
