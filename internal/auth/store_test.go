package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestStore_Keyring(t *testing.T) {
	keyring.MockInit() // in-memory keyring for tests
	s := New(t.TempDir())

	require.NoError(t, s.Set("default", "key-abc"))
	got, err := s.Get("default")
	require.NoError(t, err)
	assert.Equal(t, "key-abc", got)
	assert.Equal(t, "keyring", s.Backend())

	require.NoError(t, s.Delete("default"))
	_, err = s.Get("default")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_FallsBackToFileWhenKeyringUnavailable(t *testing.T) {
	keyring.MockInitWithError(errors.New("no secret service"))
	t.Cleanup(keyring.MockInit)

	dir := t.TempDir()
	s := New(dir)
	require.NoError(t, s.Set("prod", "key-secret"))
	assert.Equal(t, "file", s.Backend())

	// A fresh store (also without a keyring) must read it back from the encrypted file.
	s2 := New(dir)
	got, err := s2.Get("prod")
	require.NoError(t, err)
	assert.Equal(t, "key-secret", got)
	assert.Equal(t, "file", s2.Backend())

	require.NoError(t, s2.Delete("prod"))
	_, err = s2.Get("prod")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestStore_GetMissingReturnsNotFound(t *testing.T) {
	keyring.MockInit()
	s := New(t.TempDir())
	_, err := s.Get("ghost")
	assert.ErrorIs(t, err, ErrNotFound)
}
