package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestID_Unmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want ID
	}{
		{`"abc123"`, "abc123"},
		{`123`, "123"},
		{`1000`, "1000"},
		{`123.0`, "123"},
		{`null`, ""},
		{`""`, ""},
	}
	for _, c := range cases {
		var got ID
		require.NoError(t, json.Unmarshal([]byte(c.in), &got), "input %s", c.in)
		assert.Equal(t, c.want, got, "input %s", c.in)
	}
}

func TestID_Marshal(t *testing.T) {
	b, err := json.Marshal(ID("42"))
	require.NoError(t, err)
	assert.Equal(t, `"42"`, string(b))

	b, err = json.Marshal(ID(""))
	require.NoError(t, err)
	assert.Equal(t, `null`, string(b))
}

func TestID_RoundTripInStruct(t *testing.T) {
	// Execution-like payload: numeric id and workflowId.
	var ex struct {
		ID         ID `json:"id"`
		WorkflowID ID `json:"workflowId"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"id":1000,"workflowId":"55"}`), &ex))
	assert.Equal(t, ID("1000"), ex.ID)
	assert.Equal(t, ID("55"), ex.WorkflowID)
}

func TestInt_Unmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want Int
	}{
		{`30`, 30},
		{`"30"`, 30},
		{`30.0`, 30},
		{`null`, 0},
		{`""`, 0},
	}
	for _, c := range cases {
		var got Int
		require.NoError(t, json.Unmarshal([]byte(c.in), &got), "input %s", c.in)
		assert.Equal(t, c.want, got, "input %s", c.in)
	}
	b, err := json.Marshal(Int(7))
	require.NoError(t, err)
	assert.Equal(t, "7", string(b))
}

func TestBool_Unmarshal(t *testing.T) {
	cases := []struct {
		in   string
		want Bool
	}{
		{`true`, true},
		{`false`, false},
		{`"true"`, true},
		{`"false"`, false},
		{`1`, true},
		{`0`, false},
		{`null`, false},
	}
	for _, c := range cases {
		var got Bool
		require.NoError(t, json.Unmarshal([]byte(c.in), &got), "input %s", c.in)
		assert.Equal(t, c.want, got, "input %s", c.in)
	}
	b, err := json.Marshal(Bool(true))
	require.NoError(t, err)
	assert.Equal(t, "true", string(b))
}

func TestStringOrSlice_Unmarshal(t *testing.T) {
	var single StringOrSlice
	require.NoError(t, json.Unmarshal([]byte(`"client"`), &single))
	assert.Equal(t, StringOrSlice{"client"}, single)

	var multi StringOrSlice
	require.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &multi))
	assert.Equal(t, StringOrSlice{"a", "b"}, multi)

	var none StringOrSlice
	require.NoError(t, json.Unmarshal([]byte(`null`), &none))
	assert.Nil(t, none)

	// Marshal: single element -> string, multiple -> array, empty -> null.
	b, _ := json.Marshal(StringOrSlice{"x"})
	assert.Equal(t, `"x"`, string(b))
	b, _ = json.Marshal(StringOrSlice{"x", "y"})
	assert.Equal(t, `["x","y"]`, string(b))
	b, _ = json.Marshal(StringOrSlice{})
	assert.Equal(t, `null`, string(b))
}

// --- fuzz the flexible decoders: they must never panic and must round-trip. ---

func FuzzID(f *testing.F) {
	for _, s := range []string{`"x"`, `1`, `null`, `123.0`, `""`, `-5`, `true`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		var id ID
		if err := json.Unmarshal([]byte(in), &id); err != nil {
			return // invalid JSON is fine; we only care that it never panics
		}
		// Marshalling the result must always succeed and re-parse.
		b, err := json.Marshal(id)
		require.NoError(t, err)
		var again ID
		require.NoError(t, json.Unmarshal(b, &again))
		assert.Equal(t, id, again)
	})
}

func FuzzInt(f *testing.F) {
	for _, s := range []string{`1`, `"2"`, `null`, `3.0`, `""`, `bad`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		var n Int
		_ = json.Unmarshal([]byte(in), &n) // must not panic
	})
}

func FuzzStringOrSlice(f *testing.F) {
	for _, s := range []string{`"x"`, `["a","b"]`, `null`, `[]`, `1`} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		var s StringOrSlice
		_ = json.Unmarshal([]byte(in), &s) // must not panic
	})
}
