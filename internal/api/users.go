package api

import (
	"context"
	"net/http"
	"net/url"
)

// User models an n8n user. Most fields are read-only; the API creates users by
// invite (email + global role) and changes roles via a dedicated endpoint.
//
// See https://docs.n8n.io/api/api-reference/#tag/User
type User struct {
	ID         ID     `json:"id,omitempty"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Role       string `json:"role,omitempty"`
	IsPending  Bool   `json:"isPending,omitempty"`
	MFAEnabled Bool   `json:"mfaEnabled,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
	UpdatedAt  string `json:"updatedAt,omitempty"`
}

// Users returns a typed handle to the /users resource. Generic create/update are
// disabled at the command layer because n8n invites users with a bespoke payload.
func (c *Client) Users() *Resource[User] { return NewResource[User](c, "users") }

// InviteUsers invites one or more users (POST /users) with an array of {email, role}.
// The response is a heterogeneous array, returned raw for the caller to render.
func (c *Client) InviteUsers(ctx context.Context, invites []map[string]string) (any, error) {
	var out any
	err := c.Do(ctx, http.MethodPost, "users", nil, invites, &out)
	return out, err
}

// ChangeUserRole changes a user's global role (PATCH /users/{id}/role) with
// {newRoleName: "global:admin"|"global:member"}.
func (c *Client) ChangeUserRole(ctx context.Context, id, newRoleName string) error {
	body := map[string]string{"newRoleName": newRoleName}
	return c.Do(ctx, http.MethodPatch, "users/"+url.PathEscape(id)+"/role", nil, body, nil)
}
