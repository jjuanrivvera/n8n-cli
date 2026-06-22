package api

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataTables_CRUDAndRows(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/data-tables" && r.Method == http.MethodGet:
			_, _ = w.Write([]byte(`{"data":[{"id":"d1","name":"orders"}],"nextCursor":null}`))
		case r.URL.Path == "/api/v1/data-tables/d1" && r.Method == http.MethodPatch:
			_, _ = w.Write([]byte(`{"id":"d1","name":"renamed"}`))
		case r.URL.Path == "/api/v1/data-tables/d1/rows" && r.Method == http.MethodGet:
			assert.Equal(t, "sku", r.URL.Query().Get("search"))
			_, _ = w.Write([]byte(`{"data":[{"id":1,"sku":"A"}],"nextCursor":null}`))
		case r.URL.Path == "/api/v1/data-tables/d1/rows" && r.Method == http.MethodPost:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "all", body["returnType"])
			_, _ = w.Write([]byte(`[{"id":1}]`))
		case r.URL.Path == "/api/v1/data-tables/d1/rows/update" && r.Method == http.MethodPatch:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, true, body["returnData"])
			assert.NotNil(t, body["filter"])
			_, _ = w.Write([]byte(`[{"id":1,"updated":true}]`))
		case r.URL.Path == "/api/v1/data-tables/d1/rows/upsert" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`[{"id":2}]`))
		case r.URL.Path == "/api/v1/data-tables/d1/rows/delete" && r.Method == http.MethodDelete:
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, true, body["returnData"])
			_, _ = w.Write([]byte(`{"deleted":1}`))
		default:
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})

	tables, _, err := c.DataTables().List(context.Background(), ListParams{})
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, "orders", tables[0].Name)

	updated, err := c.DataTables().Update(context.Background(), "d1", map[string]string{"name": "renamed"})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)

	rows, err := c.ListDataTableRows(context.Background(), "d1", ListParams{Extra: map[string][]string{"search": {"sku"}}})
	require.NoError(t, err)
	assert.Contains(t, string(rows), `"sku":"A"`)

	_, err = c.AddDataTableRows(context.Background(), "d1", json.RawMessage(`[{"sku":"B"}]`))
	require.NoError(t, err)
	_, err = c.UpdateDataTableRows(context.Background(), "d1", json.RawMessage(`{"filter":{"type":"and"},"data":{"sku":"C"}}`))
	require.NoError(t, err)
	_, err = c.UpsertDataTableRows(context.Background(), "d1", json.RawMessage(`{"filter":{},"data":{}}`))
	require.NoError(t, err)
	_, err = c.DeleteDataTableRows(context.Background(), "d1", `{"type":"and"}`)
	require.NoError(t, err)
}

func TestPackages_ExportImport(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/n8n-packages/export":
			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			ids := body["workflowIds"].([]any)
			assert.Len(t, ids, 2)
			w.Header().Set("Content-Type", "application/gzip")
			_, _ = w.Write([]byte("PKG-BYTES"))
		case "/api/v1/n8n-packages/import":
			// must be multipart with a "package" file part and the policy field
			ct := r.Header.Get("Content-Type")
			assert.True(t, strings.HasPrefix(ct, "multipart/form-data"))
			_, params, err := mime.ParseMediaType(ct)
			require.NoError(t, err)
			mr := multipart.NewReader(r.Body, params["boundary"])
			var sawPackage bool
			var policy string
			for {
				part, perr := mr.NextPart()
				if perr == io.EOF {
					break
				}
				require.NoError(t, perr)
				switch part.FormName() {
				case "package":
					sawPackage = true
				case "workflowConflictPolicy":
					b, _ := io.ReadAll(part)
					policy = string(b)
				}
			}
			assert.True(t, sawPackage, "package file part required")
			assert.Equal(t, "fail", policy)
			_, _ = w.Write([]byte(`{"imported":1}`))
		default:
			t.Fatalf("unexpected %s", r.URL.Path)
		}
	})

	archive, err := c.ExportPackage(context.Background(), []string{"w1", "w2"})
	require.NoError(t, err)
	assert.Equal(t, "PKG-BYTES", string(archive))

	res, err := c.ImportPackage(context.Background(), []byte("ARCHIVE"), ImportOptions{ConflictPolicy: "fail", ProjectID: "p1"})
	require.NoError(t, err)
	assert.Contains(t, string(res), `"imported":1`)
}

func TestPostMultipart_DryRun(t *testing.T) {
	var buf strings.Builder
	c := New(WithBaseURL("https://h/api/v1"), WithAPIKey("secret"),
		WithDryRun(true, false, &buf))
	_, err := c.PostMultipart(context.Background(), "n8n-packages/import",
		map[string]string{"workflowConflictPolicy": "fail"}, "package", "p.n8np", []byte("x"))
	require.True(t, IsDryRun(err))
	out := buf.String()
	assert.Contains(t, out, "curl -X POST")
	assert.Contains(t, out, "-F 'package=@p.n8np'")
	assert.Contains(t, out, "workflowConflictPolicy=fail")
	assert.NotContains(t, out, "secret")
}
