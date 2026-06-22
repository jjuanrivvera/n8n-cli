package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// DataTable models an n8n data table. Columns are kept raw since they are a
// free-form array of {id, name, type}. Data tables are an n8n feature that may be
// unlicensed on some editions (the API returns 403).
//
// See https://docs.n8n.io/api/api-reference/#tag/DataTable
type DataTable struct {
	ID        ID              `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Columns   json.RawMessage `json:"columns,omitempty"`
	ProjectID string          `json:"projectId,omitempty"`
	CreatedAt string          `json:"createdAt,omitempty"`
	UpdatedAt string          `json:"updatedAt,omitempty"`
}

// DataTables returns a typed handle to the /data-tables resource. Like credentials,
// data-table update uses PATCH.
func (c *Client) DataTables() *Resource[DataTable] {
	return NewResource[DataTable](c, "data-tables", WithUpdateMethod(http.MethodPatch))
}

// ListDataTableRows lists rows in a data table (GET /data-tables/{id}/rows).
func (c *Client) ListDataTableRows(ctx context.Context, tableID string, params ListParams) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodGet, "data-tables/"+url.PathEscape(tableID)+"/rows", params.values(c.defaultLimit), nil)
}

// AddDataTableRows inserts rows (POST /data-tables/{id}/rows). data is an array of
// row objects.
func (c *Client) AddDataTableRows(ctx context.Context, tableID string, data json.RawMessage) (json.RawMessage, error) {
	body := map[string]any{"data": data, "returnType": "all"}
	return c.doRaw(ctx, http.MethodPost, "data-tables/"+url.PathEscape(tableID)+"/rows", nil, body)
}

// UpdateDataTableRows updates rows matching a filter
// (PATCH /data-tables/{id}/rows/update) with {filter, data, returnData:true}.
func (c *Client) UpdateDataTableRows(ctx context.Context, tableID string, payload json.RawMessage) (json.RawMessage, error) {
	body, err := mergeReturnData(payload)
	if err != nil {
		return nil, err
	}
	return c.doRaw(ctx, http.MethodPatch, "data-tables/"+url.PathEscape(tableID)+"/rows/update", nil, body)
}

// UpsertDataTableRows inserts or updates rows
// (POST /data-tables/{id}/rows/upsert) with {filter, data, returnData:true}.
func (c *Client) UpsertDataTableRows(ctx context.Context, tableID string, payload json.RawMessage) (json.RawMessage, error) {
	body, err := mergeReturnData(payload)
	if err != nil {
		return nil, err
	}
	return c.doRaw(ctx, http.MethodPost, "data-tables/"+url.PathEscape(tableID)+"/rows/upsert", nil, body)
}

// DeleteDataTableRows deletes rows matching a filter
// (DELETE /data-tables/{id}/rows/delete) with {filter, returnData:true}.
func (c *Client) DeleteDataTableRows(ctx context.Context, tableID, filter string) (json.RawMessage, error) {
	body := map[string]any{"returnData": true}
	if filter != "" {
		body["filter"] = json.RawMessage(filter)
	}
	return c.doRaw(ctx, http.MethodDelete, "data-tables/"+url.PathEscape(tableID)+"/rows/delete", nil, body)
}

// mergeReturnData decodes a {filter, data} payload and sets returnData:true.
func mergeReturnData(payload json.RawMessage) (map[string]any, error) {
	body := map[string]any{}
	if len(payload) > 0 {
		if err := json.Unmarshal(payload, &body); err != nil {
			return nil, err
		}
	}
	body["returnData"] = true
	return body, nil
}
