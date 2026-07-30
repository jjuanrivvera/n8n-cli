package auth

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// service is the keyring service name under which API keys are stored, keyed by profile.
const service = "n8nctl-cli"

// ErrNotFound is returned when no API key is stored for a profile.
var ErrNotFound = errors.New("no API key stored for this profile")

// Store persists and retrieves a per-profile secret API key.
type Store interface {
	Set(profile, apiKey string) error
	Get(profile string) (string, error)
	Delete(profile string) error
	// Backend names where the key currently lives ("keyring" or "file"), for doctor output.
	Backend() string
}

// store tries the keyring first and transparently falls back to an encrypted file when the
// keyring is unavailable (no Secret Service on a headless Linux box, for example).
type store struct {
	service string
	fb      *fileStore
	backend string
}

// New returns the default key store. dir is where the encrypted fallback file lives
// (the config dir); it is only touched if the keyring is unreachable.
func New(dir string) Store {
	return &store{service: service, fb: newFileStore(dir), backend: "keyring"}
}

func (s *store) Backend() string { return s.backend }

func (s *store) Set(profile, apiKey string) error {
	if err := keyring.Set(s.service, profile, apiKey); err != nil {
		s.backend = "file"
		return s.fb.Set(profile, apiKey)
	}
	return nil
}

func (s *store) Get(profile string) (string, error) {
	key, err := keyring.Get(s.service, profile)
	if err == nil {
		return key, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		// Keyring works but has nothing — check the fallback file before giving up, in case
		// the key was written on a host without a keyring.
		if key, ferr := s.fb.Get(profile); ferr == nil {
			s.backend = "file"
			return key, nil
		}
		return "", ErrNotFound
	}
	// Keyring is unavailable entirely → use the fallback file.
	s.backend = "file"
	key, ferr := s.fb.Get(profile)
	if ferr != nil {
		return "", ErrNotFound
	}
	return key, nil
}

func (s *store) Delete(profile string) error {
	// Remove from both backends. The keyring delete is best-effort: it may be entirely
	// unavailable on this host, in which case the encrypted file is the real store. We only
	// surface an error if the file backend itself fails for a reason other than "not found".
	_ = keyring.Delete(s.service, profile)
	ferr := s.fb.Delete(profile)
	if ferr == nil || errors.Is(ferr, ErrNotFound) {
		return nil
	}
	return fmt.Errorf("delete key: %w", ferr)
}
