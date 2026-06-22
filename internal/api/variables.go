package api

// Variable models an n8n environment variable. type and id are read-only; create
// accepts key, value, and an optional projectId.
//
// See https://docs.n8n.io/api/api-reference/#tag/Variables
type Variable struct {
	ID        ID     `json:"id,omitempty"`
	Key       string `json:"key,omitempty"`
	Value     string `json:"value,omitempty"`
	Type      string `json:"type,omitempty"`
	ProjectID string `json:"projectId,omitempty"`
}

// Variables returns a typed handle to the /variables resource. The API exposes
// list/create/update/delete but no single-GET, so `get` is served from the list.
func (c *Client) Variables() *Resource[Variable] { return NewResource[Variable](c, "variables") }
