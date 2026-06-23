package commands

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

// feedStdin replaces os.Stdin with a pipe carrying s for one read.
func feedStdin(t *testing.T, s string) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	go func() { _, _ = w.WriteString(s); _ = w.Close() }()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old; _ = r.Close() })
}

func TestPromptHelpers(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetErr(os.Stderr)

	feedStdin(t, "homelab\n")
	assert.Equal(t, "homelab", prompt(cmd, "name: "))

	feedStdin(t, "secret-key\n")
	s, err := promptSecret(cmd, "key: ")
	require.NoError(t, err)
	assert.Equal(t, "secret-key", s)

	feedStdin(t, "y\n")
	assert.True(t, confirm(cmd, "go?"))

	feedStdin(t, "n\n")
	assert.False(t, confirm(cmd, "go?"))
}

func TestExpandAliases(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	c, err := config.Load()
	require.NoError(t, err)
	c.Aliases = map[string]string{"ls": "workflows list --all"}
	require.NoError(t, c.Save())
	resetConfigForTest()

	old := os.Args
	os.Args = []string{"n8nctl", "ls", "--quiet"}
	t.Cleanup(func() { os.Args = old })
	expandAliases()
	assert.Equal(t, []string{"n8nctl", "workflows", "list", "--all", "--quiet"}, os.Args)
}

func TestExecuteFunc(t *testing.T) {
	old := os.Args
	os.Args = []string{"n8nctl", "version"}
	t.Cleanup(func() { os.Args = old })
	resetCommandTree(RootCmd())
	RootCmd().SetArgs(nil)
	require.NoError(t, Execute(context.Background()))
}

func TestCreateBodyVariants(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// --data JSON
	out, _, err := run(t, "credentials", "create", "--data", `{"name":"D","type":"githubApi","data":{}}`, "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "New")
	// --data plus --set merges
	_, _, err = run(t, "credentials", "create", "--data", `{"type":"githubApi"}`, "--set", "name=Merged", "-o", "json")
	require.NoError(t, err)
	// no body -> error
	_, _, err = run(t, "credentials", "create")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no body")
	// invalid --set
	_, _, err = run(t, "tags", "create", "--set", "noequalsign")
	require.Error(t, err)
}

func TestDeleteAbort(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	feedStdin(t, "n\n")
	out, _, err := run(t, "tags", "delete", "t1") // no --yes -> prompts, we answer no
	require.NoError(t, err)
	assert.NotContains(t, out, "deleted")
}

func TestConfigSetVariants(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	for _, kv := range [][2]string{
		{"output_format", "yaml"},
		{"log_level", "debug"},
		{"requests_per_second", "9"},
		{"description", "my lab"},
	} {
		_, _, err := run(t, "config", "set", kv[0], kv[1])
		require.NoError(t, err, kv[0])
	}
	_, _, err := run(t, "config", "set", "bogus_key", "x")
	require.Error(t, err)
	_, _, err = run(t, "config", "set", "requests_per_second", "not-a-number")
	require.Error(t, err)
}

func TestDoctorKeySources(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	// key via flag
	_, _, err := run(t, "doctor", "--api-key", "test-key", "--quiet", "--json")
	require.NoError(t, err)
	// key via env
	t.Setenv("N8NCTL_API_KEY", "test-key")
	_, _, err = run(t, "doctor", "--quiet")
	require.NoError(t, err)
}

func TestVerboseLogging(t *testing.T) {
	srv := fullServer(t)
	setupProfile(t, srv.URL)
	_, _, err := run(t, "workflows", "list", "-v", "--quiet", "-o", "json")
	require.NoError(t, err)
}
