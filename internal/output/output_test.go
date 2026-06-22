package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sample struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

var rows = []sample{
	{"1", "Alpha", true},
	{"2", "Beta", false},
}

func TestParse(t *testing.T) {
	f, err := Parse("")
	require.NoError(t, err)
	assert.Equal(t, Table, f)
	_, err = Parse("xml")
	require.Error(t, err)
	for _, ok := range []string{"table", "json", "yaml", "csv"} {
		f, err := Parse(ok)
		require.NoError(t, err)
		assert.True(t, f.Valid())
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, JSON, Options{}))
	var back []sample
	require.NoError(t, json.Unmarshal(buf.Bytes(), &back))
	assert.Equal(t, rows, back)
}

func TestRenderYAML(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, YAML, Options{}))
	assert.Contains(t, buf.String(), "name: Alpha")
}

func TestRenderCSV(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, CSV, Options{}))
	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 3) // header + 2 rows
	assert.Equal(t, "id,name,active", lines[0])
	assert.Contains(t, lines[1], "1,Alpha,true")
}

func TestRenderCSV_Columns(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, CSV, Options{Columns: []string{"name", "id"}}))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, "name,id", lines[0])
	assert.Equal(t, "Alpha,1", lines[1])
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, Table, Options{}))
	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "Alpha")
	// id should come before name (preferred ordering).
	assert.Less(t, strings.Index(out, "ID"), strings.Index(out, "NAME"))
}

func TestRenderTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, []sample{}, Table, Options{}))
	assert.Contains(t, buf.String(), "No results")
}

func TestRenderTable_SingleObjectAndNested(t *testing.T) {
	var buf bytes.Buffer
	obj := map[string]any{"id": "1", "meta": map[string]any{"k": "v"}}
	require.NoError(t, Render(&buf, obj, Table, Options{}))
	out := buf.String()
	assert.Contains(t, out, "META")
	// Nested object rendered as compact JSON on one line.
	assert.Contains(t, out, `{"k":"v"}`)
	assert.NotContains(t, out, "\n\n")
}

func TestRenderColor(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, rows, Table, Options{Color: true}))
	assert.Contains(t, buf.String(), "\x1b[1m") // bold header
	buf.Reset()
	require.NoError(t, Render(&buf, rows, Table, Options{Color: true, NoColor: true}))
	assert.NotContains(t, buf.String(), "\x1b[1m")
}

func TestRenderRawMessage(t *testing.T) {
	var buf bytes.Buffer
	raw := json.RawMessage(`{"id":"x","name":"Raw"}`)
	require.NoError(t, Render(&buf, raw, JSON, Options{}))
	assert.Contains(t, buf.String(), "Raw")
}
