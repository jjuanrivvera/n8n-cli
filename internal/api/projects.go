package api

import (
	"context"
	"net/http"
	"net/url"
)

// Project models an n8n project (Enterprise feature). type is read-only.
//
// See https://docs.n8n.io/api/api-reference/#tag/Projects
type Project struct {
	ID   ID     `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// Projects returns a typed handle to the /projects resource.
func (c *Client) Projects() *Resource[Project] { return NewResource[Project](c, "projects") }

// ProjectMember is a user's membership of a project, including their project role.
type ProjectMember struct {
	ID    ID     `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Role  string `json:"role,omitempty"`
}

// ListProjectMembers lists the members of a project (GET /projects/{id}/users).
func (c *Client) ListProjectMembers(ctx context.Context, projectID string, params ListParams) ([]ProjectMember, string, error) {
	raw, err := c.doRaw(ctx, http.MethodGet, "projects/"+url.PathEscape(projectID)+"/users", params.values(c.defaultLimit), nil)
	if err != nil {
		return nil, "", err
	}
	return decodeList[ProjectMember](raw)
}

// AddProjectMembers adds users to a project (POST /projects/{id}/users). relations
// is a list of {userId, role} pairs.
func (c *Client) AddProjectMembers(ctx context.Context, projectID string, relations []map[string]string) error {
	body := map[string]any{"relations": relations}
	return c.Do(ctx, http.MethodPost, "projects/"+url.PathEscape(projectID)+"/users", nil, body, nil)
}

// ChangeProjectMemberRole updates a member's role (PATCH /projects/{id}/users/{userId}).
func (c *Client) ChangeProjectMemberRole(ctx context.Context, projectID, userID, role string) error {
	body := map[string]string{"role": role}
	return c.Do(ctx, http.MethodPatch, "projects/"+url.PathEscape(projectID)+"/users/"+url.PathEscape(userID), nil, body, nil)
}

// RemoveProjectMember removes a user from a project (DELETE /projects/{id}/users/{userId}).
func (c *Client) RemoveProjectMember(ctx context.Context, projectID, userID string) error {
	return c.Do(ctx, http.MethodDelete, "projects/"+url.PathEscape(projectID)+"/users/"+url.PathEscape(userID), nil, nil, nil)
}
