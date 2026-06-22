package api

// Tag models an n8n workflow tag.
//
// See https://docs.n8n.io/api/api-reference/#tag/Tags
type Tag struct {
	ID        ID     `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// Tags returns a typed handle to the /tags resource (full CRUD, PUT update).
func (c *Client) Tags() *Resource[Tag] { return NewResource[Tag](c, "tags") }
