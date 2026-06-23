package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// Workflow models an n8n workflow. Free-form structural fields (nodes, connections,
// settings, staticData, pinData, meta) are kept as json.RawMessage so they round-trip
// byte-for-byte — n8n is the source of truth for their shape and we never need to
// interpret them to create/update/transfer a workflow.
//
// See https://docs.n8n.io/api/api-reference/#tag/Workflow
type Workflow struct {
	ID           ID              `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	Description  string          `json:"description,omitempty"`
	Active       Bool            `json:"active,omitempty"`
	IsArchived   Bool            `json:"isArchived,omitempty"`
	VersionID    string          `json:"versionId,omitempty"`
	TriggerCount Int             `json:"triggerCount,omitempty"`
	CreatedAt    string          `json:"createdAt,omitempty"`
	UpdatedAt    string          `json:"updatedAt,omitempty"`
	Nodes        json.RawMessage `json:"nodes,omitempty"`
	Connections  json.RawMessage `json:"connections,omitempty"`
	Settings     json.RawMessage `json:"settings,omitempty"`
	StaticData   json.RawMessage `json:"staticData,omitempty"`
	PinData      json.RawMessage `json:"pinData,omitempty"`
	Meta         json.RawMessage `json:"meta,omitempty"`
	Tags         []Tag           `json:"tags,omitempty"`
}

// Workflows returns a typed handle to the /workflows resource.
func (c *Client) Workflows() *Resource[Workflow] { return NewResource[Workflow](c, "workflows") }

// Activate enables a workflow (POST /workflows/{id}/activate).
func (c *Client) ActivateWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	err := c.Workflows().Action(ctx, id, "activate", nil, &out)
	return &out, err
}

// Deactivate disables a workflow (POST /workflows/{id}/deactivate).
func (c *Client) DeactivateWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	err := c.Workflows().Action(ctx, id, "deactivate", nil, &out)
	return &out, err
}

// Archive archives a workflow (POST /workflows/{id}/archive).
func (c *Client) ArchiveWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	err := c.Workflows().Action(ctx, id, "archive", nil, &out)
	return &out, err
}

// Unarchive restores an archived workflow (POST /workflows/{id}/unarchive).
func (c *Client) UnarchiveWorkflow(ctx context.Context, id string) (*Workflow, error) {
	var out Workflow
	err := c.Workflows().Action(ctx, id, "unarchive", nil, &out)
	return &out, err
}

// TransferWorkflow moves a workflow to another project (PUT /workflows/{id}/transfer).
func (c *Client) TransferWorkflow(ctx context.Context, id, destinationProjectID string) error {
	body := map[string]string{"destinationProjectId": destinationProjectID}
	return c.Workflows().ActionMethod(ctx, http.MethodPut, id, "transfer", body, nil)
}

// GetWorkflowTags returns the tags attached to a workflow (GET /workflows/{id}/tags).
func (c *Client) GetWorkflowTags(ctx context.Context, id string) ([]Tag, error) {
	var out []Tag
	err := c.Do(ctx, http.MethodGet, "workflows/"+url.PathEscape(id)+"/tags", nil, nil, &out)
	return out, err
}

// SetWorkflowTags replaces a workflow's tags (PUT /workflows/{id}/tags). The body
// is an array of {id} objects.
func (c *Client) SetWorkflowTags(ctx context.Context, id string, tagIDs []string) ([]Tag, error) {
	body := make([]map[string]string, 0, len(tagIDs))
	for _, t := range tagIDs {
		body = append(body, map[string]string{"id": t})
	}
	var out []Tag
	err := c.Do(ctx, http.MethodPut, "workflows/"+url.PathEscape(id)+"/tags", nil, body, &out)
	return out, err
}
