package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a workflow whose webhook node lacks webhookId -> a lint ERROR.
const badWF = `{"name":"bad","nodes":[{"name":"Hook","type":"n8n-nodes-base.webhook","parameters":{}}],"connections":{},"settings":{}}`

// a clean workflow -> no lint errors.
const goodWF = `{"name":"ok","nodes":[{"name":"Hook","type":"n8n-nodes-base.webhook","webhookId":"x","parameters":{}},{"name":"Set","type":"n8n-nodes-base.set","parameters":{}}],"connections":{"Hook":{"main":[[{"node":"Set","type":"main","index":0}]]}},"settings":{}}`

func proxyTo(t *testing.T, blockDestructive bool) (*httptest.Server, *httptest.Server, *int) {
	t.Helper()
	hits := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		assert.Equal(t, "secret", r.Header.Get("X-N8N-API-KEY"), "proxy must inject the key")
		_, _ = w.Write([]byte(`{"id":"new1"}`))
	}))
	t.Cleanup(backend.Close)
	tu, _ := url.Parse(backend.URL)
	proxy := httptest.NewServer(newLintProxy(&url.URL{Scheme: tu.Scheme, Host: tu.Host}, "secret", nil, blockDestructive, io.Discard))
	t.Cleanup(proxy.Close)
	return proxy, backend, &hits
}

func do(t *testing.T, method, url, body string) (*http.Response, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	b, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, string(b)
}

func TestProxy_RejectsBadWorkflowCreate(t *testing.T) {
	proxy, _, hits := proxyTo(t, false)
	resp, body := do(t, http.MethodPost, proxy.URL+"/api/v1/workflows", badWF)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Contains(t, body, "rejected by lint")
	assert.Contains(t, body, "webhook-id-required")
	assert.Equal(t, 0, *hits, "a rejected write must not reach the instance")
}

func TestProxy_ForwardsCleanWorkflowCreate(t *testing.T) {
	proxy, _, hits := proxyTo(t, false)
	resp, body := do(t, http.MethodPost, proxy.URL+"/api/v1/workflows", goodWF)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, "new1")
	assert.Equal(t, 1, *hits, "a clean write must be forwarded")
}

func TestProxy_LintsUpdateButNotSubResources(t *testing.T) {
	proxy, _, hits := proxyTo(t, false)
	// PUT /workflows/{id} with a bad body is linted -> 422
	resp, _ := do(t, http.MethodPut, proxy.URL+"/api/v1/workflows/42", badWF)
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	assert.Equal(t, 0, *hits)
	// PUT /workflows/{id}/tags is a sub-resource, not a workflow body -> pass through
	resp, _ = do(t, http.MethodPut, proxy.URL+"/api/v1/workflows/42/tags", `[{"id":"1"}]`)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, *hits)
}

func TestProxy_ReadsPassThrough(t *testing.T) {
	proxy, _, hits := proxyTo(t, false)
	resp, _ := do(t, http.MethodGet, proxy.URL+"/api/v1/workflows", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, *hits, "reads are never gated")
}

func TestProxy_BlockDestructive(t *testing.T) {
	proxy, _, hits := proxyTo(t, true)
	resp, body := do(t, http.MethodDelete, proxy.URL+"/api/v1/workflows/42", "")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Contains(t, body, "DELETE blocked")
	assert.Equal(t, 0, *hits)
	// without the flag, delete passes through
	proxy2, _, hits2 := proxyTo(t, false)
	resp, _ = do(t, http.MethodDelete, proxy2.URL+"/api/v1/workflows/42", "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, *hits2)
}

func TestProxy_DisableRule(t *testing.T) {
	tu, _ := url.Parse("http://example.invalid")
	_ = tu
	// build a proxy with webhook-id-required disabled; the bad WF now passes lint
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	t.Cleanup(backend.Close)
	b, _ := url.Parse(backend.URL)
	h := newLintProxy(&url.URL{Scheme: b.Scheme, Host: b.Host}, "secret", map[string]bool{"webhook-id-required": true}, false, io.Discard)
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	resp, _ := do(t, http.MethodPost, srv.URL+"/api/v1/workflows", badWF)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "disabled rule must not block")
}

func TestProxyCmd_RequiresTarget(t *testing.T) {
	// no profile/base-url configured -> proxy errors out before serving
	resetFlags()
	resetCommandTree(RootCmd())
	resetConfigForTest()
	var out bytes.Buffer
	root := RootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"proxy", "--base-url", ""})
	require.Error(t, root.ExecuteContext(t.Context()))
}
