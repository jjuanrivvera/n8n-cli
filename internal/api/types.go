package api

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
)

// The n8n API is loosely typed in a few places: workflow IDs are strings while
// execution IDs and workflowId are numbers (and sometimes serialised as quoted
// strings in examples). These adapters absorb that inconsistency so resource
// structs can stay simple and round-trip cleanly regardless of the wire shape.

// ID decodes from a JSON string OR number and always marshals back as a string.
// This keeps table output consistent and avoids float64 precision loss for large
// integer IDs (values above 2^53).
type ID string

// UnmarshalJSON accepts "123", 123, or null.
func (i *ID) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*i = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*i = ID(s)
		return nil
	}
	// Bare number. Preserve integer ids exactly — n8n ids can exceed 2^53, so
	// routing through float64 would silently corrupt them. An integer literal is
	// already its own canonical string; only a float-looking value (e.g. 123.0)
	// is normalised through float64.
	s := string(data)
	if !strings.ContainsAny(s, ".eE") {
		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			*i = ID(s)
			return nil
		}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && f == float64(int64(f)) {
		*i = ID(strconv.FormatInt(int64(f), 10))
		return nil
	}
	*i = ID(strings.Trim(s, `"`))
	return nil
}

// MarshalJSON always emits a JSON string (or null when empty).
func (i ID) MarshalJSON() ([]byte, error) {
	if i == "" {
		return []byte("null"), nil
	}
	return json.Marshal(string(i))
}

func (i ID) String() string { return string(i) }

// Int decodes from a JSON number OR a numeric string and marshals as a number.
type Int int64

// UnmarshalJSON accepts 30, "30", or null (which becomes 0).
func (i *Int) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*i = 0
		return nil
	}
	s := strings.Trim(string(data), `"`)
	if s == "" {
		*i = 0
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		// Tolerate floats such as "30.0".
		f, ferr := strconv.ParseFloat(s, 64)
		if ferr != nil {
			return err
		}
		n = int64(f)
	}
	*i = Int(n)
	return nil
}

// MarshalJSON emits a bare JSON number.
func (i Int) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}

func (i Int) Int64() int64 { return int64(i) }

// Bool decodes from a JSON bool OR the strings "true"/"false"/"1"/"0".
type Bool bool

// UnmarshalJSON accepts true, false, "true", "false", 1, 0, or null.
func (b *Bool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*b = false
		return nil
	}
	switch s := strings.Trim(string(data), `"`); strings.ToLower(s) {
	case "true", "1":
		*b = true
	case "false", "0", "":
		*b = false
	default:
		var raw bool
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		*b = Bool(raw)
	}
	return nil
}

// MarshalJSON emits a bare JSON boolean.
func (b Bool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (b Bool) Bool() bool { return bool(b) }

// StringOrSlice accepts either a single JSON string or an array of strings.
type StringOrSlice []string

// UnmarshalJSON accepts "x", ["x","y"], or null.
func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*s = nil
		return nil
	}
	if data[0] == '[' {
		var arr []string
		if err := json.Unmarshal(data, &arr); err != nil {
			return err
		}
		*s = arr
		return nil
	}
	var single string
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*s = []string{single}
	return nil
}

// MarshalJSON emits a single string when there is exactly one element, an array
// otherwise. This mirrors how the API tends to send the field back.
func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	switch len(s) {
	case 0:
		return []byte("null"), nil
	case 1:
		return json.Marshal(s[0])
	default:
		return json.Marshal([]string(s))
	}
}
