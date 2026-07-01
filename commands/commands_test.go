package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

// resetCommandTree restores every flag in the (singleton) command tree to its
// default so each test invocation starts clean — including local subcommand flags
// like version's --json and repeatable --set slices.
func resetCommandTree(cmd *cobra.Command) {
	reset := func(fs *pflag.FlagSet) {
		fs.VisitAll(func(f *pflag.Flag) {
			f.Changed = false
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				_ = sv.Replace([]string{})
				return
			}
			_ = f.Value.Set(f.DefValue)
		})
	}
	reset(cmd.Flags())
	reset(cmd.PersistentFlags())
	for _, c := range cmd.Commands() {
		resetCommandTree(c)
	}
}

// resetConfigForTest clears the process-cached config so each invocation reloads
// from the (per-test) N8NCTL_CONFIG file.
func resetConfigForTest() {
	cfgOnce = sync.Once{}
	cfgVal = nil
	cfgErr = nil
}

// resetFlags returns all global persistent flags to their zero values, since the
// root command is a shared singleton across test invocations.
func resetFlags() {
	flagInstance, flagProfile, flagOutput, flagBaseURL, flagAPIKey = "", "", "", "", ""
	flagRPS = 0
	flagDryRun, flagShowToken, flagVerbose, flagNoColor, flagQuiet = false, false, false, false, false
	flagColumns = nil
}

// run executes the root command with args, capturing stdout and stderr.
func run(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	resetFlags()
	resetCommandTree(RootCmd())
	resetConfigForTest()
	var out, errb bytes.Buffer
	root := RootCmd()
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), errb.String(), err
}

// newServer starts a mock n8n API covering the routes the tests exercise.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "test-key", r.Header.Get("X-N8N-API-KEY"))
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"new1","name":"Created"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"Demo","active":true,"triggerCount":1}],"nextCursor":null}`))
	})
	mux.HandleFunc("/api/v1/workflows/w1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"w1","name":"Demo","active":true}`))
	})
	mux.HandleFunc("/api/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"t1","name":"Prod"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"t1","name":"Prod"}]}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// setupProfile writes a temp config with a "default" profile pointing at baseURL
// and stores a key in the mocked keyring.
func setupProfile(t *testing.T, baseURL string) {
	t.Helper()
	keyring.MockInit()
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("N8NCTL_CONFIG", path)
	for _, k := range []string{"N8NCTL_INSTANCE", "N8NCTL_PROFILE", "N8NCTL_BASE_URL", "N8NCTL_API_KEY", "N8NCTL_OUTPUT", "N8NCTL_RPS", "N8NCTL_LOG_LEVEL"} {
		t.Setenv(k, "")
	}
	c := config.New()
	c.DefaultProfile = "default"
	c.SetProfile(&config.Profile{Name: "default", BaseURL: baseURL})
	require.NoError(t, c.Save())
	require.NoError(t, keyring.Set("n8nctl-cli", "default", "test-key"))
}

func TestCmd_Version(t *testing.T) {
	out, _, err := run(t, "version")
	require.NoError(t, err)
	assert.Contains(t, out, "n8nctl")
}

func TestCmd_ConfigPath(t *testing.T) {
	setupProfile(t, "http://example/api/v1")
	out, _, err := run(t, "config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "config.yaml")
}

func TestCmd_WorkflowsList_JSON(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "list", "-o", "json", "--quiet")
	require.NoError(t, err)
	var items []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &items))
	require.Len(t, items, 1)
	assert.Equal(t, "Demo", items[0]["name"])
}

func TestCmd_WorkflowsList_Table(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "list", "--quiet", "--no-color")
	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "Demo")
}

func TestCmd_WorkflowsList_Columns(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "workflows", "list", "--columns", "name", "-o", "csv", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "name")
	assert.Contains(t, out, "Demo")
}

func TestCmd_TagsCreate(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "tags", "create", "--set", "name=Prod", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "Prod")
}

func TestCmd_DryRunCreate(t *testing.T) {
	setupProfile(t, "https://n8n.example.com/api/v1")
	out, _, err := run(t, "tags", "create", "--set", "name=Prod", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "/tags")
	assert.Contains(t, out, "<redacted>")
}

func TestCmd_RawAPIDryRun(t *testing.T) {
	setupProfile(t, "https://n8n.example.com/api/v1")
	out, _, err := run(t, "api", "GET", "/workflows", "--query", "limit=5", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "curl -X GET")
	assert.Contains(t, out, "limit=5")
}

func TestCmd_MissingBaseURL(t *testing.T) {
	keyring.MockInit()
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("N8NCTL_CONFIG", path)
	t.Setenv("N8NCTL_BASE_URL", "")
	_, _, err := run(t, "workflows", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "base URL")
}

func TestCmd_AuthStatus(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "auth", "status", "-o", "json", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "\"valid\": true")
}

func TestCmd_Doctor(t *testing.T) {
	srv := newServer(t)
	setupProfile(t, srv.URL)
	out, _, err := run(t, "doctor", "--quiet")
	require.NoError(t, err)
	assert.Contains(t, out, "api auth")
}

func TestCmd_ConfigUseAndListProfiles(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "config", "set", "base_url", "https://b/api/v1", "--profile", "cloud")
	require.NoError(t, err)
	_, _, err = run(t, "config", "use", "cloud")
	require.NoError(t, err)
	out, _, err := run(t, "config", "list-profiles", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "cloud")
}

// The multi-instance selector is --instance; --profile is a hidden back-compat alias.
// Both must select the same named profile, and --profile must stay out of help.
func TestCmd_InstanceFlagAndProfileAlias(t *testing.T) {
	setupProfile(t, "https://a/api/v1")

	// --instance writes to the named profile.
	_, _, err := run(t, "config", "set", "base_url", "https://viaInstance/api/v1", "--instance", "prod")
	require.NoError(t, err)
	// --profile (deprecated alias) writes to a different named profile.
	_, _, err = run(t, "config", "set", "base_url", "https://viaProfile/api/v1", "--profile", "legacy")
	require.NoError(t, err)

	out, _, err := run(t, "config", "list-profiles", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "prod")
	assert.Contains(t, out, "legacy")

	help, _, err := run(t, "help")
	require.NoError(t, err)
	assert.Contains(t, help, "--instance")
	assert.NotContains(t, help, "--profile")
}

func TestCmd_AliasLifecycle(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "alias", "set", "ls", "workflows", "list")
	require.NoError(t, err)
	out, _, err := run(t, "alias", "list", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "ls")
	_, _, err = run(t, "alias", "remove", "ls")
	require.NoError(t, err)
}

func TestCmd_AliasCannotShadowBuiltin(t *testing.T) {
	setupProfile(t, "https://a/api/v1")
	_, _, err := run(t, "alias", "set", "workflows", "tags", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "built-in")
}
