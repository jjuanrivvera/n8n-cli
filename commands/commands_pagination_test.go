package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pagingServer always returns a next cursor, so list emits a "more results" hint
// and --all hits the page cap (truncation warning).
func pagingServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"Demo"}],"nextCursor":"more"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestListMoreResultsHint(t *testing.T) {
	srv := pagingServer(t)
	setupProfile(t, srv.URL)
	_, errOut, err := run(t, "workflows", "list", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, errOut, "more results available")
}

func TestListAllTruncationWarning(t *testing.T) {
	srv := pagingServer(t)
	setupProfile(t, srv.URL)
	_, errOut, err := run(t, "workflows", "list", "--all", "--max-pages", "2", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, errOut, "truncated")
}

func TestDeleteDryRun(t *testing.T) {
	setupProfile(t, "https://n8n.example.com/api/v1")
	out, _, err := run(t, "workflows", "delete", "w1", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X DELETE")
}

func TestWorkflowTagsClear(t *testing.T) {
	setupProfile(t, "https://n8n.example.com/api/v1")
	// --set with empty value clears tags; dry-run shows the PUT.
	out, _, err := run(t, "workflows", "tags", "w1", "--set", "", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X PUT")
}
