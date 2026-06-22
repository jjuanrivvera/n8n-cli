package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient returns a Client wired to an httptest server running handler.
// Rate limiting is effectively disabled so tests don't sleep.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(
		WithBaseURL(srv.URL),
		WithAPIKey("test-key"),
		WithRequestsPerSecond(0), // unlimited
	)
}
