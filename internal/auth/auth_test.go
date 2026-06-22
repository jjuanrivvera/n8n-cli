package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestKeyringRoundTrip(t *testing.T) {
	keyring.MockInit() // in-memory provider; no real OS keyring touched

	// Missing key -> ErrNotFound and empty Lookup.
	_, err := Get("missing")
	require.ErrorIs(t, err, ErrNotFound)
	assert.Empty(t, Lookup("missing"))

	require.NoError(t, Set("homelab", "secret-123"))
	got, err := Get("homelab")
	require.NoError(t, err)
	assert.Equal(t, "secret-123", got)
	assert.Equal(t, "secret-123", Lookup("homelab"))

	// Per-profile isolation.
	require.NoError(t, Set("cloud", "other-456"))
	assert.Equal(t, "other-456", Lookup("cloud"))
	assert.Equal(t, "secret-123", Lookup("homelab"))

	// Delete is idempotent.
	require.NoError(t, Delete("homelab"))
	assert.Empty(t, Lookup("homelab"))
	require.NoError(t, Delete("homelab"))
}
