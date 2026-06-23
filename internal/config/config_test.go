package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withTempConfig points N8NCTL_CONFIG at a temp file and clears env overrides.
func withTempConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("N8NCTL_CONFIG", path)
	for _, k := range []string{"N8NCTL_PROFILE", "N8NCTL_BASE_URL", "N8NCTL_API_KEY", "N8NCTL_OUTPUT", "N8NCTL_RPS", "N8NCTL_LOG_LEVEL"} {
		t.Setenv(k, "")
	}
	return path
}

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	withTempConfig(t)
	c, err := Load()
	require.NoError(t, err)
	assert.Empty(t, c.Profiles)
	assert.Equal(t, "default", c.ActiveProfileName(""))
}

func TestSaveLoadRoundTrip(t *testing.T) {
	path := withTempConfig(t)
	c := New()
	c.DefaultProfile = "homelab"
	c.SetProfile(&Profile{Name: "homelab", BaseURL: "https://n8n.home/api/v1", Description: "lab"})
	c.SetProfile(&Profile{Name: "cloud", BaseURL: "https://x.app.n8n.cloud/api/v1"})
	c.Settings.OutputFormat = "json"
	c.Settings.RequestsPerSecond = 8
	require.NoError(t, c.Save())

	_, statErr := os.Stat(path)
	require.NoError(t, statErr)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "homelab", loaded.DefaultProfile)
	require.Contains(t, loaded.Profiles, "homelab")
	assert.Equal(t, "https://n8n.home/api/v1", loaded.Profiles["homelab"].BaseURL)
	assert.Equal(t, "homelab", loaded.Profiles["homelab"].Name) // name backfilled
	assert.Len(t, loaded.Profiles, 2)
}

func TestActiveProfilePrecedence(t *testing.T) {
	withTempConfig(t)
	c := New()
	c.DefaultProfile = "cloud"
	assert.Equal(t, "flagwins", c.ActiveProfileName("flagwins")) // flag beats default
	assert.Equal(t, "cloud", c.ActiveProfileName(""))            // default
	t.Setenv("N8NCTL_PROFILE", "envwins")
	assert.Equal(t, "envwins", c.ActiveProfileName("")) // env beats default
	assert.Equal(t, "flagwins", c.ActiveProfileName("flagwins"))
}

func TestResolveEnvOverrides(t *testing.T) {
	withTempConfig(t)
	c := New()
	c.SetProfile(&Profile{Name: "default", BaseURL: "https://file.example/api/v1", APIKey: "filekey"})
	c.Settings.OutputFormat = "table"

	r := c.Resolve("default")
	assert.Equal(t, "https://file.example/api/v1", r.BaseURL)
	assert.Equal(t, "filekey", r.APIKey)
	assert.Equal(t, "table", r.OutputFormat)
	assert.Equal(t, float64(5), r.RequestsPerSecond) // default when unset

	t.Setenv("N8NCTL_BASE_URL", "https://env.example/api/v1")
	t.Setenv("N8NCTL_API_KEY", "envkey")
	t.Setenv("N8NCTL_OUTPUT", "json")
	t.Setenv("N8NCTL_RPS", "12")
	r = c.Resolve("default")
	assert.Equal(t, "https://env.example/api/v1", r.BaseURL)
	assert.Equal(t, "envkey", r.APIKey)
	assert.Equal(t, "json", r.OutputFormat)
	assert.Equal(t, float64(12), r.RequestsPerSecond)
}

func TestProfileCreatesOnDemand(t *testing.T) {
	withTempConfig(t)
	c := New()
	p := c.Profile("brand-new")
	assert.Equal(t, "brand-new", p.Name)
	assert.Contains(t, c.Profiles, "brand-new")
}

func TestDefaultPath_XDG(t *testing.T) {
	t.Setenv("N8NCTL_CONFIG", "")
	xdg := filepath.Join(string(filepath.Separator)+"tmp", "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	// Build the expected path with filepath.Join so it is correct on Windows too.
	assert.Equal(t, filepath.Join(xdg, "n8nctl-cli", "config.yaml"), DefaultPath())
	t.Setenv("N8NCTL_CONFIG", "/explicit/path.yaml")
	assert.Equal(t, "/explicit/path.yaml", DefaultPath())
}

func TestDefaultPath_Home(t *testing.T) {
	t.Setenv("N8NCTL_CONFIG", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	p := DefaultPath()
	assert.Contains(t, p, filepath.Join(".n8nctl-cli", "config.yaml"))
}

func TestPathAndSetProfile(t *testing.T) {
	path := withTempConfig(t)
	c := New()
	assert.Equal(t, path, c.Path())

	c.SetProfile(&Profile{Name: "a", BaseURL: "https://a/api/v1"})
	c.SetProfile(&Profile{Name: "a", BaseURL: "https://a2/api/v1"}) // replace
	assert.Equal(t, "https://a2/api/v1", c.Profiles["a"].BaseURL)

	// SetProfile on a nil map still works.
	c2 := &Config{path: path}
	c2.SetProfile(&Profile{Name: "x"})
	assert.Contains(t, c2.Profiles, "x")
}

func TestSaveCreatesParentDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "deeper")
	t.Setenv("N8NCTL_CONFIG", filepath.Join(dir, "config.yaml"))
	c := New()
	c.SetProfile(&Profile{Name: "default", BaseURL: "https://x/api/v1"})
	require.NoError(t, c.Save())
	_, err := os.Stat(filepath.Join(dir, "config.yaml"))
	require.NoError(t, err)
}

func TestResolveUnknownProfile(t *testing.T) {
	withTempConfig(t)
	c := New()
	r := c.Resolve("does-not-exist")
	assert.Empty(t, r.BaseURL)
	assert.Equal(t, "table", r.OutputFormat) // default applied
	assert.Equal(t, "warn", r.LogLevel)
}

func TestLoad_MalformedYAML(t *testing.T) {
	path := withTempConfig(t)
	require.NoError(t, os.WriteFile(path, []byte(":\n bad: [yaml"), 0o600))
	_, err := Load()
	require.Error(t, err)
}

func TestResolve_BadRPSEnvIgnored(t *testing.T) {
	withTempConfig(t)
	c := New()
	c.Profiles["default"] = &Profile{BaseURL: "https://x/api/v1"}
	base := c.Resolve("default").RequestsPerSecond
	t.Setenv("N8NCTL_RPS", "notanumber")
	assert.Equal(t, base, c.Resolve("default").RequestsPerSecond) // invalid env left the value unchanged
	t.Setenv("N8NCTL_LOG_LEVEL", "debug")
	assert.Equal(t, "debug", c.Resolve("default").LogLevel)
}
