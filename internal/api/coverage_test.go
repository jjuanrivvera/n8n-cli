package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_RemainingActions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/workflows/1/deactivate":
			_, _ = w.Write([]byte(`{"id":"1","active":false}`))
		case "/api/v1/workflows/1/archive":
			_, _ = w.Write([]byte(`{"id":"1","isArchived":true}`))
		case "/api/v1/workflows/1/unarchive":
			_, _ = w.Write([]byte(`{"id":"1","isArchived":false}`))
		case "/api/v1/workflows/1/tags":
			assert.Equal(t, http.MethodGet, r.Method)
			_, _ = w.Write([]byte(`[{"id":"t1","name":"a"}]`))
		default:
			t.Fatalf("unexpected %s", r.URL.Path)
		}
	})
	_, err := c.DeactivateWorkflow(context.Background(), "1")
	require.NoError(t, err)
	_, err = c.ArchiveWorkflow(context.Background(), "1")
	require.NoError(t, err)
	_, err = c.UnarchiveWorkflow(context.Background(), "1")
	require.NoError(t, err)
	tags, err := c.GetWorkflowTags(context.Background(), "1")
	require.NoError(t, err)
	require.Len(t, tags, 1)
}

func TestCredential_SchemaAndTransfer(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/credentials/schema/githubApi":
			_, _ = w.Write([]byte(`{"type":"object","properties":{"accessToken":{"type":"string"}}}`))
		case "/api/v1/credentials/9/transfer":
			assert.Equal(t, http.MethodPut, r.Method)
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "p1", body["destinationProjectId"])
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected %s", r.URL.Path)
		}
	})
	schema, err := c.CredentialSchema(context.Background(), "githubApi")
	require.NoError(t, err)
	assert.Contains(t, string(schema), "accessToken")
	require.NoError(t, c.TransferCredential(context.Background(), "9", "p1"))
}

func TestProjects_Members(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/projects/p1/users" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"data":[{"id":"u1","email":"a@x.com","role":"project:admin"}]}`))
		case r.URL.Path == "/api/v1/projects/p1/users" && r.Method == http.MethodPost:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.NotNil(t, body["relations"])
			w.WriteHeader(http.StatusCreated)
		case r.URL.Path == "/api/v1/projects/p1/users/u1" && r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/api/v1/projects/p1/users/u1" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	members, _, err := c.ListProjectMembers(context.Background(), "p1", ListParams{})
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, "a@x.com", members[0].Email)

	require.NoError(t, c.AddProjectMembers(context.Background(), "p1",
		[]map[string]string{{"userId": "u2", "role": "project:viewer"}}))
	require.NoError(t, c.ChangeProjectMemberRole(context.Background(), "p1", "u1", "project:editor"))
	require.NoError(t, c.RemoveProjectMember(context.Background(), "p1", "u1"))
}

func TestUsers_InviteAndRole(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/users" && r.Method == http.MethodPost:
			var body []map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "new@x.com", body[0]["email"])
			_, _ = w.Write([]byte(`[{"user":{"id":"u9","email":"new@x.com"}}]`))
		case r.URL.Path == "/api/v1/users/u9/role" && r.Method == http.MethodPatch:
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "global:admin", body["newRoleName"])
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	_, err := c.InviteUsers(context.Background(), []map[string]string{{"email": "new@x.com", "role": "global:member"}})
	require.NoError(t, err)
	require.NoError(t, c.ChangeUserRole(context.Background(), "u9", "global:admin"))
}

func TestNewAPIError_Variants(t *testing.T) {
	// "error" field instead of "message", plus description and numeric code.
	e := newAPIError(400, []byte(`{"error":"boom","description":"more detail","code":1001}`))
	assert.Equal(t, "boom — more detail", e.Message)
	assert.Equal(t, "1001", e.Code)
	assert.Contains(t, e.Error(), "code 1001")

	// non-JSON body falls back to raw text.
	e2 := newAPIError(500, []byte("upstream exploded"))
	assert.Equal(t, "upstream exploded", e2.Message)
	assert.Contains(t, e2.Error(), "server error")

	// empty body falls back to status text.
	e3 := newAPIError(409, []byte(""))
	assert.Equal(t, "conflict", e3.Message)
	assert.True(t, e3.IsConflict())
}

func TestRetry_RetryAfterAndShouldRetry(t *testing.T) {
	p := DefaultRetryPolicy(slog.Default())

	// shouldRetry: 429 retries for POST; 500 does not.
	resp429 := &http.Response{StatusCode: 429}
	resp500 := &http.Response{StatusCode: 500}
	assert.True(t, p.shouldRetry(http.MethodPost, resp429, nil))
	assert.False(t, p.shouldRetry(http.MethodPost, resp500, nil))
	assert.True(t, p.shouldRetry(http.MethodGet, resp500, nil))
	assert.True(t, p.shouldRetry(http.MethodGet, nil, assert.AnError)) // transport error, idempotent

	// Retry-After: delta seconds.
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("Retry-After", "2")
	assert.Equal(t, 2*time.Second, p.backoff(0, resp))

	d, ok := retryAfter("")
	assert.False(t, ok)
	assert.Zero(t, d)
	d, ok = retryAfter("3")
	assert.True(t, ok)
	assert.Equal(t, 3*time.Second, d)
}

func TestRateLimiter_ThrottleRestore(t *testing.T) {
	rl := NewRateLimiter(10, slog.Default())
	rl.Throttle()
	rl.Throttle()
	rl.Restore()
	// Unlimited mode is a no-op and must not panic.
	un := NewRateLimiter(0, nil)
	un.Throttle()
	un.Restore()
	require.NoError(t, un.Wait(context.Background()))
}

func TestDecodeList_Empty(t *testing.T) {
	items, next, err := decodeList[Tag](nil)
	require.NoError(t, err)
	assert.Nil(t, items)
	assert.Empty(t, next)
}
