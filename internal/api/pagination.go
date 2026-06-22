package api

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// DefaultLimit is the page size requested when the caller does not set one.
// MaxLimit is the n8n-enforced ceiling for the `limit` query parameter.
const (
	DefaultLimit = 100
	MaxLimit     = 250
)

// ListParams are the query parameters common to n8n list endpoints. n8n paginates
// with an opaque cursor: pass the previous response's nextCursor back as Cursor
// to fetch the following page.
type ListParams struct {
	Limit  int        // page size (capped at MaxLimit)
	Cursor string     // opaque cursor from a previous nextCursor
	Extra  url.Values // resource-specific filters (active, tags, name, status, ...)
}

// values renders the params to a URL query. defaultLimit supplies the page size
// when Limit is unset.
func (p ListParams) values(defaultLimit int) url.Values {
	v := cloneValues(p.Extra)
	v.Set("limit", strconv.Itoa(p.effectiveLimit(defaultLimit)))
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	}
	return v
}

// effectiveLimit reports the page size values() would use, capped at MaxLimit.
func (p ListParams) effectiveLimit(defaultLimit int) int {
	limit := p.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return limit
}

// cloneValues deep-copies url.Values so callers' filters aren't mutated when we
// add limit/cursor for a single page.
func cloneValues(v url.Values) url.Values {
	out := url.Values{}
	for k, vs := range v {
		out[k] = append([]string(nil), vs...)
	}
	return out
}

// listEnvelope is the standard n8n list response wrapper: {data:[...], nextCursor}.
type listEnvelope[T any] struct {
	Data       []T    `json:"data"`
	NextCursor string `json:"nextCursor"`
}

// decodeList normalizes an n8n list response into items plus the next cursor.
// It accepts the standard {data, nextCursor} envelope and also tolerates a bare
// JSON array (some endpoints, e.g. workflow tags, return one).
func decodeList[T any](raw []byte) ([]T, string, error) {
	trimmed := trimSpace(raw)
	if len(trimmed) == 0 {
		return nil, "", nil
	}
	if trimmed[0] == '[' {
		var arr []T
		if err := json.Unmarshal(trimmed, &arr); err != nil {
			return nil, "", err
		}
		return arr, "", nil
	}
	var env listEnvelope[T]
	if err := json.Unmarshal(trimmed, &env); err != nil {
		return nil, "", err
	}
	return env.Data, env.NextCursor, nil
}

// trimSpace strips leading ASCII whitespace without importing bytes everywhere.
func trimSpace(b []byte) []byte {
	i := 0
	for i < len(b) {
		switch b[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return b[i:]
		}
	}
	return b[i:]
}
