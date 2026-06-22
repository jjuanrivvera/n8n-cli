// Package auth stores n8n API keys in the OS keyring (macOS Keychain, Linux
// Secret Service, Windows Credential Manager) with a per-profile key. The token
// is never written to the config file or the repo.
package auth

import (
	"errors"

	"github.com/zalando/go-keyring"
)

// service is the keyring service name under which tokens are stored.
const service = "n8nctl-cli"

// ErrNotFound is returned when no token exists for a profile.
var ErrNotFound = errors.New("no API key stored for this profile")

// Set stores the API key for a profile.
func Set(profile, apiKey string) error {
	return keyring.Set(service, profile, apiKey)
}

// Get retrieves the API key for a profile, returning ErrNotFound when absent.
func Get(profile string) (string, error) {
	v, err := keyring.Get(service, profile)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return v, nil
}

// Delete removes the stored API key for a profile. Deleting a missing key is a no-op.
func Delete(profile string) error {
	err := keyring.Delete(service, profile)
	if err != nil && errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// Lookup returns the stored key or "" (never an error) for best-effort resolution.
func Lookup(profile string) string {
	v, err := Get(profile)
	if err != nil {
		return ""
	}
	return v
}
