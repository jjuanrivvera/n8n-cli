package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletionAllShells(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	for _, sh := range []string{"bash", "zsh", "fish", "powershell"} {
		out, _, err := run(t, "completion", sh)
		require.NoError(t, err, sh)
		assert.NotEmpty(t, out, sh)
	}
	_, _, err := run(t, "completion", "nonsense")
	require.Error(t, err)
}

func TestAPIRawVariants(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// GET with query
	out, _, err := run(t, "api", "GET", "/workflows", "--query", "limit=1", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Demo")
	// POST with --data
	_, _, err = run(t, "api", "POST", "/tags", "--data", `{"name":"X"}`, "-o", "json")
	require.NoError(t, err)
	// POST with --file
	f := filepath.Join(t.TempDir(), "b.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"name":"Y"}`), 0o600))
	_, _, err = run(t, "api", "POST", "/tags", "--file", f, "-o", "json")
	require.NoError(t, err)
	// bad -q
	_, _, err = run(t, "api", "GET", "/workflows", "--query", "noequals")
	require.Error(t, err)
}

func TestListAllAndParam(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// --all auto-paginates (single page here)
	out, _, err := run(t, "workflows", "list", "--all", "-o", "json", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "Demo")
	// arbitrary --param passthrough
	_, _, err = run(t, "workflows", "list", "--param", "active=true", "-o", "json", "--quiet")
	require.NoError(t, err)
	// bad --param
	_, _, err = run(t, "workflows", "list", "--param", "noequals")
	require.Error(t, err)
}

func TestGetViaListMiss(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// variables get resolves via list; a non-matching id errors.
	_, _, err := run(t, "variables", "get", "NOPE")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "matched")
}

func TestTransferRequiresProject(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "workflows", "transfer", "w1")
	require.Error(t, err)
	_, _, err = run(t, "credentials", "transfer", "c1")
	require.Error(t, err)
}

func TestProjectMemberAndUserErrors(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "projects", "add-member", "p1") // missing --user/--role
	require.Error(t, err)
	_, _, err = run(t, "users", "invite") // missing --email
	require.Error(t, err)
	_, _, err = run(t, "users", "change-role", "u1") // missing --role
	require.Error(t, err)
}

func TestSourceControlWithVariables(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "source-control", "pull", "--variables", `{"k":"v"}`, "-o", "json")
	require.NoError(t, err)
	// invalid variables JSON
	_, _, err = run(t, "source-control", "pull", "--variables", `{bad`)
	require.Error(t, err)
}

func TestAuthLoginVerificationFails(t *testing.T) {
	// Point at a server that rejects the key -> login must fail before persisting.
	setupProfile(t, "https://127.0.0.1:1/api/v1") // unroutable
	_, _, err := run(t, "auth", "login", "--api-key", "bad", "--base-url", "https://127.0.0.1:1")
	require.Error(t, err)
}
