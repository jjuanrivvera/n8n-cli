package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// APIError is a typed representation of a non-2xx n8n API response. Its Error()
// appends an actionable hint keyed by HTTP status so users get a next step, not
// just a status code.
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // human-readable message from the API (or a fallback)
	Code       string // optional machine code from the API error body
	Hint       string // actionable remediation, derived from StatusCode
	Body       string // raw response body, retained for --verbose debugging
}

// Error renders "n8n API error (<status>): <message> [code <code>]\n  hint: ...".
func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "n8n API error (%d): %s", e.StatusCode, e.Message)
	if e.Code != "" {
		fmt.Fprintf(&b, " [code %s]", e.Code)
	}
	if e.Hint != "" {
		fmt.Fprintf(&b, "\n  hint: %s", e.Hint)
	}
	return b.String()
}

func (e *APIError) IsNotFound() bool     { return e.StatusCode == 404 }
func (e *APIError) IsUnauthorized() bool { return e.StatusCode == 401 }
func (e *APIError) IsForbidden() bool    { return e.StatusCode == 403 }
func (e *APIError) IsConflict() bool     { return e.StatusCode == 409 }
func (e *APIError) IsRateLimited() bool  { return e.StatusCode == 429 }

// IsForbidden reports whether err is an API 403 (unlicensed/forbidden feature).
func IsForbidden(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.IsForbidden()
}

// n8n error bodies look like {"message": "...", "code": "...", "description": "..."}.
type wireError struct {
	Message     string          `json:"message"`
	Code        json.RawMessage `json:"code"` // string or number depending on layer
	Description string          `json:"description"`
	// Some internal errors surface under "error" / "hint" instead of "message".
	ErrorField string `json:"error"`
	HintField  string `json:"hint"`
}

// newAPIError builds an APIError from a raw response body and status code.
func newAPIError(statusCode int, body []byte) *APIError {
	e := &APIError{StatusCode: statusCode, Body: string(body)}

	var we wireError
	if err := json.Unmarshal(body, &we); err == nil {
		switch {
		case we.Message != "":
			e.Message = we.Message
		case we.ErrorField != "":
			e.Message = we.ErrorField
		}
		if we.Description != "" && we.Description != e.Message {
			if e.Message == "" {
				e.Message = we.Description
			} else {
				e.Message = e.Message + " — " + we.Description
			}
		}
		if len(we.Code) > 0 && string(we.Code) != "null" {
			e.Code = strings.Trim(string(we.Code), `"`)
		}
	}
	if e.Message == "" {
		e.Message = strings.TrimSpace(string(body))
	}
	if e.Message == "" {
		e.Message = statusText(statusCode)
	}
	e.Hint = hintForStatus(statusCode)
	return e
}

// hintForStatus maps an HTTP status to an actionable remedy for n8n.
func hintForStatus(status int) string {
	switch {
	case status == 401:
		return "unauthorized — run `n8nctl auth login` to set a valid API key (Settings → n8n API in your instance)"
	case status == 403:
		return "forbidden — your API key lacks the required scope, or the feature is not licensed on this instance"
	case status == 404:
		return "not found — verify the id with the matching `list` command, and check `--base-url`/profile points at the right instance"
	case status == 409:
		return "conflict — the resource already exists or has uncommitted changes; resolve and retry"
	case status == 415:
		return "unsupported media type — the request body must be JSON"
	case status == 429:
		return "rate limited — slow down; n8nctl backs off automatically, lower --rps if it persists"
	case status >= 500:
		return "server error — usually transient; retry shortly. Check the n8n instance logs if it persists"
	default:
		return ""
	}
}

func statusText(status int) string {
	switch status {
	case 400:
		return "bad request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not found"
	case 409:
		return "conflict"
	case 429:
		return "too many requests"
	case 500:
		return "internal server error"
	default:
		return fmt.Sprintf("unexpected status %d", status)
	}
}
