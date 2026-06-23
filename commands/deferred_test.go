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
	_, _, err := run(t, "workflows", "lint", "-f", f)
	require.Error(t, err) // lint exits non-zero on the value error
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
