package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResource_ListEnvelope(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/workflows", r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"data":[{"id":"1","name":"A","active":true},{"id":2,"name":"B"}],"nextCursor":"abc"}`))
	})
	items, next, err := c.Workflows().List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, ID("1"), items[0].ID)
	assert.Equal(t, ID("2"), items[1].ID) // numeric id normalized
	assert.True(t, items[0].Active.Bool())
	assert.Equal(t, "abc", next)
}

func TestResource_ListBareArray(t *testing.T) {
	// Some endpoints (e.g. workflow tags) return a bare array.
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"t1","name":"Prod"}]`))
	})
	items, next, err := c.Tags().List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Prod", items[0].Name)
	assert.Empty(t, next)
}

func TestResource_ListAll_WalksCursor(t *testing.T) {
	var page int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		switch cursor {
		case "":
			page++
			_, _ = w.Write([]byte(`{"data":[{"id":"1"}],"nextCursor":"c2"}`))
		case "c2":
			page++
			_, _ = w.Write([]byte(`{"data":[{"id":"2"}],"nextCursor":"c3"}`))
		case "c3":
			page++
			_, _ = w.Write([]byte(`{"data":[{"id":"3"}],"nextCursor":null}`))
		default:
			t.Fatalf("unexpected cursor %q", cursor)
		}
	})
	items, err := c.Workflows().ListAll(context.Background(), ListParams{}, 0)
	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, 3, page)
	assert.Equal(t, ID("3"), items[2].ID)
}

func TestResource_ListAllChecked_Truncates(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		// Always returns a next cursor -> never ends.
		_, _ = w.Write([]byte(`{"data":[{"id":"1"}],"nextCursor":"more"}`))
	})
	items, truncated, err := c.Workflows().ListAllChecked(context.Background(), ListParams{}, 2)
	require.NoError(t, err)
	assert.True(t, truncated)
	assert.Len(t, items, 2)
}

func TestResource_GetCreateDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/tags/5":
			_, _ = w.Write([]byte(`{"id":"5","name":"Five"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/tags":
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "New", body["name"])
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"9","name":"New"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/tags/9":
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	got, err := c.Tags().Get(context.Background(), "5", nil)
	require.NoError(t, err)
	assert.Equal(t, "Five", got.Name)

	created, err := c.Tags().Create(context.Background(), map[string]string{"name": "New"})
	require.NoError(t, err)
	assert.Equal(t, ID("9"), created.ID)

	require.NoError(t, c.Tags().Delete(context.Background(), "9"))
}

func TestResource_UpdateVerb(t *testing.T) {
	// Tags update with PUT; credentials update with PATCH.
	var sawTagMethod, sawCredMethod string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/tags/1":
			sawTagMethod = r.Method
			_, _ = w.Write([]byte(`{"id":"1","name":"x"}`))
		case "/api/v1/credentials/1":
			sawCredMethod = r.Method
			_, _ = w.Write([]byte(`{"id":"1","name":"x","type":"githubApi"}`))
		}
	})
	_, err := c.Tags().Update(context.Background(), "1", map[string]string{"name": "x"})
	require.NoError(t, err)
	_, err = c.Credentials().Update(context.Background(), "1", map[string]string{"name": "x"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, sawTagMethod)
	assert.Equal(t, http.MethodPatch, sawCredMethod)
}

func TestWorkflow_Actions(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/workflows/7/activate" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"id":"7","active":true}`))
		case r.URL.Path == "/api/v1/workflows/7/transfer" && r.Method == http.MethodPut:
			var body map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "proj99", body["destinationProjectId"])
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/api/v1/workflows/7/tags" && r.Method == http.MethodPut:
			var body []map[string]string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Len(t, body, 2)
			assert.Equal(t, "t1", body[0]["id"])
			_, _ = w.Write([]byte(`[{"id":"t1","name":"a"},{"id":"t2","name":"b"}]`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	wf, err := c.ActivateWorkflow(context.Background(), "7")
	require.NoError(t, err)
	assert.True(t, wf.Active.Bool())

	require.NoError(t, c.TransferWorkflow(context.Background(), "7", "proj99"))

	tags, err := c.SetWorkflowTags(context.Background(), "7", []string{"t1", "t2"})
	require.NoError(t, err)
	require.Len(t, tags, 2)
}

func TestExecution_RetryStopAndNumericID(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/executions" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"data":[{"id":1000,"workflowId":55,"status":"error","finished":false}]}`))
		case r.URL.Path == "/api/v1/executions/1000/retry" && r.Method == http.MethodPost:
			var body map[string]bool
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.True(t, body["loadWorkflow"])
			_, _ = w.Write([]byte(`{"id":1001,"status":"success"}`))
		case r.URL.Path == "/api/v1/executions/1000/stop" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"id":1000,"status":"canceled"}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	items, _, err := c.Executions().List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, ID("1000"), items[0].ID)
	assert.Equal(t, ID("55"), items[0].WorkflowID)

	ex, err := c.RetryExecution(context.Background(), "1000", true)
	require.NoError(t, err)
	assert.Equal(t, ID("1001"), ex.ID)

	ex, err = c.StopExecution(context.Background(), "1000")
	require.NoError(t, err)
	assert.Equal(t, "canceled", ex.Status)
}

func TestListParams_Values(t *testing.T) {
	p := ListParams{Limit: 500, Cursor: "abc"} // 500 -> capped to MaxLimit
	v := p.values(100)
	assert.Equal(t, fmt.Sprint(MaxLimit), v.Get("limit"))
	assert.Equal(t, "abc", v.Get("cursor"))

	p2 := ListParams{}
	assert.Equal(t, "100", p2.values(100).Get("limit"))
	assert.Empty(t, p2.values(100).Get("cursor"))
}

func TestOperations_AuditAndSourceControl(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/audit":
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			add := body["additionalOptions"].(map[string]any)
			assert.Equal(t, float64(30), add["daysAbandonedWorkflow"])
			_, _ = w.Write([]byte(`{"Credentials Risk Report":{}}`))
		case "/api/v1/source-control/pull":
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, true, body["force"])
			_, _ = w.Write([]byte(`[]`))
		}
	})
	rep, err := c.GenerateAudit(context.Background(), AuditOptions{DaysAbandonedWorkflow: 30})
	require.NoError(t, err)
	assert.Contains(t, string(rep), "Credentials Risk Report")

	res, err := c.SourceControlPull(context.Background(), true, nil)
	require.NoError(t, err)
	assert.Equal(t, "[]", string(res))
}
