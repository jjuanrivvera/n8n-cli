package api

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendsAPIKeyHeader(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-key", r.Header.Get("X-N8N-API-KEY"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Contains(t, r.Header.Get("User-Agent"), "n8nctl/")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	_, _, err := c.Workflows().List(context.Background(), ListParams{})
	require.NoError(t, err)
}

func TestClient_NormalizeBaseURL(t *testing.T) {
	assert.Equal(t, "https://h/api/v1", normalizeBaseURL("https://h"))
	assert.Equal(t, "https://h/api/v1", normalizeBaseURL("https://h/"))
	assert.Equal(t, "https://h/api/v1", normalizeBaseURL("https://h/api/v1"))
	assert.Equal(t, "https://h/api/v2", normalizeBaseURL("https://h/api/v2"))
	assert.Equal(t, "", normalizeBaseURL(""))
}

func TestClient_APIErrorHint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"X-N8N-API-KEY header required"}`))
	})
	_, err := c.Tags().Get(context.Background(), "1", nil)
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 401, apiErr.StatusCode)
	assert.True(t, apiErr.IsUnauthorized())
	assert.Contains(t, apiErr.Error(), "auth login")
}

func TestClient_404Hint(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	})
	_, err := c.Workflows().Get(context.Background(), "nope", nil)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.True(t, apiErr.IsNotFound())
	assert.Contains(t, apiErr.Error(), "verify the id")
}

func TestClient_DryRunPrintsCurlNoRequest(t *testing.T) {
	var called atomic.Int32
	srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		called.Add(1)
		_, _ = w.Write([]byte(`{}`))
	})
	var buf bytes.Buffer
	// Re-wrap with dry-run enabled, same base URL.
	c := New(
		WithBaseURL(srv.BaseURL()),
		WithAPIKey("secret-key"),
		WithRequestsPerSecond(0),
		WithDryRun(true, false, &buf),
	)
	_, err := c.Tags().Create(context.Background(), map[string]string{"name": "Prod"})
	require.True(t, IsDryRun(err))
	assert.Equal(t, int32(0), called.Load(), "no request should be sent in dry-run")

	out := buf.String()
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "/tags")
	assert.Contains(t, out, "X-N8N-API-KEY: <redacted>")
	assert.NotContains(t, out, "secret-key", "token must be redacted by default")
}

func TestClient_DryRunShowToken(t *testing.T) {
	var buf bytes.Buffer
	c := New(WithBaseURL("https://h/api/v1"), WithAPIKey("secret-key"),
		WithDryRun(true, true, &buf))
	_, _ = c.Tags().Create(context.Background(), map[string]string{"name": "x"})
	assert.Contains(t, buf.String(), "X-N8N-API-KEY: secret-key")
}

func TestClient_RetriesOn5xxForIdempotent(t *testing.T) {
	var attempts atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	})
	// shrink backoff so the test is fast
	c.retryPolicy.InitialBackoff = 1
	c.retryPolicy.MaxBackoff = 2
	_, _, err := c.Workflows().List(context.Background(), ListParams{})
	require.NoError(t, err)
	assert.Equal(t, int32(3), attempts.Load())
}

func TestClient_DoesNotRetryPOST(t *testing.T) {
	var attempts atomic.Int32
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})
	c.retryPolicy.InitialBackoff = 1
	_, err := c.Tags().Create(context.Background(), map[string]string{"name": "x"})
	require.Error(t, err)
	assert.Equal(t, int32(1), attempts.Load(), "POST must not be auto-retried")
}

func TestClient_RawAPIPassthrough(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// The test base URL has /api/v1 appended by normalizeBaseURL.
		assert.Equal(t, "/api/v1/custom/path", r.URL.Path)
		assert.Equal(t, "5", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	var out map[string]any
	q := map[string][]string{"limit": {"5"}}
	err := c.Do(context.Background(), http.MethodGet, "/custom/path", q, nil, &out)
	require.NoError(t, err)
	assert.Equal(t, true, out["ok"])
}

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'a b'`, shellQuote("a b"))
	assert.Equal(t, `'it'\''s'`, shellQuote("it's"))
	assert.False(t, strings.Contains(shellQuote("x"), " "))
}
