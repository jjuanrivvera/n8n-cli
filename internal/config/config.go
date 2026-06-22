// Package config loads and persists n8nctl configuration with profiles and
// flag>env>file>default precedence. Secrets are not stored here — API keys live
// in the OS keyring (see internal/auth); only non-secret profile metadata is saved.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// envPrefix namespaces all environment overrides (e.g. N8NCTL_BASE_URL).
const envPrefix = "N8NCTL_"

// Profile is a single named n8n instance configuration.
type Profile struct {
	Name        string `yaml:"-"`
	BaseURL     string `yaml:"base_url"`
	Description string `yaml:"description,omitempty"`
	// APIKey is normally empty (the key lives in the keyring). It is honored if a
	// user insists on storing it in the file, but `config view` redacts it.
	APIKey string `yaml:"api_key,omitempty"`
}

// Settings holds global, non-profile preferences.
type Settings struct {
	OutputFormat      string  `yaml:"output_format,omitempty"`
	RequestsPerSecond float64 `yaml:"requests_per_second,omitempty"`
	LogLevel          string  `yaml:"log_level,omitempty"`
}

// Config is the on-disk configuration document.
type Config struct {
	DefaultProfile string              `yaml:"default_profile,omitempty"`
	Profiles       map[string]*Profile `yaml:"profiles,omitempty"`
	Settings       Settings            `yaml:"settings,omitempty"`
	Aliases        map[string]string   `yaml:"aliases,omitempty"`

	path string `yaml:"-"`
}

// Resolved is the effective configuration for one command invocation after the
// active profile and environment overrides are applied.
type Resolved struct {
	Profile           string
	BaseURL           string
	APIKey            string // from env only; keyring lookup happens in the command layer
	OutputFormat      string
	RequestsPerSecond float64
	LogLevel          string
}

// DefaultPath returns the config file path, honoring N8NCTL_CONFIG and XDG.
func DefaultPath() string {
	if p := os.Getenv(envPrefix + "CONFIG"); p != "" {
		return p
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "n8nctl-cli", "config.yaml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".n8nctl-cli", "config.yaml")
	}
	return filepath.Join(home, ".n8nctl-cli", "config.yaml")
}

// New returns an empty config bound to the default path.
func New() *Config {
	return &Config{Profiles: map[string]*Profile{}, Aliases: map[string]string{}, path: DefaultPath()}
}

// Load reads the config file (returning an empty config if none exists yet).
func Load() (*Config, error) {
	path := DefaultPath()
	data, err := os.ReadFile(path) //nolint:gosec // path is user/XDG controlled by design
	if err != nil {
		if os.IsNotExist(err) {
			c := New()
			c.path = path
			return c, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	if c.Aliases == nil {
		c.Aliases = map[string]string{}
	}
	for name, p := range c.Profiles {
		p.Name = name
	}
	c.path = path
	return &c, nil
}

// Path returns the file path this config loads from / saves to.
func (c *Config) Path() string { return c.path }

// Save writes the config atomically (temp file + rename), creating parent dirs.
func (c *Config) Save() error {
	if c.path == "" {
		c.path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	tmp := c.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := os.Rename(tmp, c.path); err != nil {
		return fmt.Errorf("replacing config: %w", err)
	}
	return nil
}

// ActiveProfileName resolves which profile to use given an optional override and
// the N8NCTL_PROFILE env var, falling back to DefaultProfile then "default".
func (c *Config) ActiveProfileName(override string) string {
	switch {
	case override != "":
		return override
	case os.Getenv(envPrefix+"PROFILE") != "":
		return os.Getenv(envPrefix + "PROFILE")
	case c.DefaultProfile != "":
		return c.DefaultProfile
	default:
		return "default"
	}
}

// Profile returns the named profile, creating an empty one if absent.
func (c *Config) Profile(name string) *Profile {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	p, ok := c.Profiles[name]
	if !ok {
		p = &Profile{Name: name}
		c.Profiles[name] = p
	}
	return p
}

// SetProfile stores or replaces a profile.
func (c *Config) SetProfile(p *Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]*Profile{}
	}
	c.Profiles[p.Name] = p
}

// Resolve computes the effective settings for a profile, applying env overrides.
// Flag-level overrides are layered on top by the command layer.
func (c *Config) Resolve(profileName string) *Resolved {
	p := c.Profiles[profileName]
	if p == nil {
		p = &Profile{Name: profileName}
	}
	r := &Resolved{
		Profile:           profileName,
		BaseURL:           p.BaseURL,
		APIKey:            p.APIKey,
		OutputFormat:      firstNonEmpty(c.Settings.OutputFormat, "table"),
		RequestsPerSecond: c.Settings.RequestsPerSecond,
		LogLevel:          firstNonEmpty(c.Settings.LogLevel, "warn"),
	}
	if r.RequestsPerSecond <= 0 {
		r.RequestsPerSecond = 5
	}
	// Environment overrides (env beats file).
	if v := os.Getenv(envPrefix + "BASE_URL"); v != "" {
		r.BaseURL = v
	}
	if v := os.Getenv(envPrefix + "API_KEY"); v != "" {
		r.APIKey = v
	}
	if v := os.Getenv(envPrefix + "OUTPUT"); v != "" {
		r.OutputFormat = v
	}
	if v := os.Getenv(envPrefix + "LOG_LEVEL"); v != "" {
		r.LogLevel = v
	}
	if v := os.Getenv(envPrefix + "RPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			r.RequestsPerSecond = f
		}
	}
	return r
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
