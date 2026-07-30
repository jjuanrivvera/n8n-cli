package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeSecret(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"bracketed paste wrappers stripped", "\x1b[200~KEY\x1b[201~\n", "KEY"},
		{"clean key unchanged", "KEY", "KEY"},
		{"surrounding whitespace trimmed", "  KEY  ", "KEY"},
		{"only opening marker", "\x1b[200~KEY", "KEY"},
		{"only closing marker", "KEY\x1b[201~", "KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeSecret(tt.in))
		})
	}
}
