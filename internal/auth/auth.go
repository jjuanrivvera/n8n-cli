// Package auth stores n8n API keys out of plaintext. The OS keyring (macOS Keychain,
// Linux Secret Service, Windows Credential Manager) is primary, keyed by profile; an
// encrypted file (credentials.enc in the config dir) is the transparent fallback for
// headless hosts where no keyring is available. Keys are never written to the config
// file or the repo.
package auth

import (
	"path/filepath"
	"sync"

	"github.com/jjuanrivvera/n8n-cli/internal/config"
)

// defaultStore is the process-wide key store, built lazily so the config dir is resolved
// on first use (after any env overrides are in place) rather than at package init.
var (
	defaultStore Store
	storeOnce    sync.Once
)

// getStore returns the shared store, constructing it against the config dir on first call.
// The encrypted fallback file lives next to config.yaml (see config.DefaultPath).
func getStore() Store {
	storeOnce.Do(func() {
		defaultStore = New(filepath.Dir(config.DefaultPath()))
	})
	return defaultStore
}

// Set stores the API key for a profile.
func Set(profile, apiKey string) error {
	return getStore().Set(profile, apiKey)
}

// Get retrieves the API key for a profile, returning ErrNotFound when absent.
func Get(profile string) (string, error) {
	return getStore().Get(profile)
}

// Delete removes the stored API key for a profile. Deleting a missing key is a no-op.
func Delete(profile string) error {
	return getStore().Delete(profile)
}

// Lookup returns the stored key or "" (never an error) for best-effort resolution.
func Lookup(profile string) string {
	v, err := getStore().Get(profile)
	if err != nil {
		return ""
	}
	return v
}
