package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitInteractive drives the wizard entirely through stdin, exercising the
// shared-reader fix (profile name, base URL, then secret key on one piped input).
func TestInitInteractive(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	feedStdin(t, "wizard\n"+srv.URL+"\ntest-key\n")
	out, _, err := run(t, "init")
	require.NoError(t, err)
	assert.Contains(t, out, "wizard")
}

// TestAuthLoginInteractive covers prompting for the base URL and key together.
func TestAuthLoginInteractive(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// Blank the profile's base URL so login prompts for it, then the key.
	feedStdin(t, srv.URL+"\ntest-key\n")
	out, _, err := run(t, "auth", "login", "--profile", "fresh")
	require.NoError(t, err)
	assert.Contains(t, out, "authenticated")
}

func TestVersionCheck(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}`))
	}))
	t.Cleanup(srv.Close)
	old := releaseURL
	releaseURL = srv.URL
	t.Cleanup(func() { releaseURL = old })

	out, _, err := run(t, "version", "--check")
	require.NoError(t, err)
	assert.Contains(t, out, "newer version is available")
}
