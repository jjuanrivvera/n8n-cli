package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionStrings(t *testing.T) {
	assert.Equal(t, Version, Short())
	assert.Contains(t, Info(), "n8nctl")
	assert.Contains(t, Info(), Version)
	ua := UserAgent()
	assert.True(t, strings.HasPrefix(ua, "n8nctl/"))
	assert.Contains(t, ua, "/") // platform suffix
}
