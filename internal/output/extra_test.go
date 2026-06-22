package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_IDOnly(t *testing.T) {
	f, err := Parse("id")
	require.NoError(t, err)
	assert.Equal(t, IDOnly, f)
	f, err = Parse("id-only")
	require.NoError(t, err)
	assert.Equal(t, IDOnly, f)
}

func TestRenderIDOnly(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{{"id": "a", "name": "x"}, {"id": 2, "name": "y"}}
	require.NoError(t, Render(&buf, data, IDOnly, Options{}))
	assert.Equal(t, "a\n2\n", buf.String())

	// single object
	buf.Reset()
	require.NoError(t, Render(&buf, map[string]any{"id": "solo"}, IDOnly, Options{}))
	assert.Equal(t, "solo\n", buf.String())
}

func TestRenderTable_NoHeader(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{{"id": "1", "name": "Alpha"}}
	require.NoError(t, Render(&buf, data, Table, Options{NoHeader: true}))
	out := buf.String()
	assert.NotContains(t, out, "ID")
	assert.NotContains(t, out, "NAME")
	assert.Contains(t, out, "Alpha")
}

func TestApplyJQ(t *testing.T) {
	data := []map[string]any{{"id": "1", "name": "Alpha", "active": true}, {"id": "2", "name": "Beta", "active": false}}

	var buf bytes.Buffer
	require.NoError(t, ApplyJQ(&buf, data, ".[].name"))
	assert.Equal(t, "Alpha\nBeta\n", buf.String()) // bare strings unquoted (jq -r style)

	buf.Reset()
	require.NoError(t, ApplyJQ(&buf, data, ".[] | select(.active) | .id"))
	assert.Equal(t, "1\n", strings.TrimSpace(buf.String())+"\n")

	buf.Reset()
	require.NoError(t, ApplyJQ(&buf, data, "length"))
	assert.Equal(t, "2", strings.TrimSpace(buf.String()))

	// invalid program
	require.Error(t, ApplyJQ(&buf, data, "this is not jq |||"))
}
