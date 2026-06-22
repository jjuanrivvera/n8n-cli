package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// Credential models an n8n credential. Secret material lives under Data (writeOnly):
// it is sent on create/update but never returned by the API, so it is omitempty.
//
// See https://docs.n8n.io/api/api-reference/#tag/Credential
type Credential struct {
	ID                      ID              `json:"id,omitempty"`
	Name                    string          `json:"name,omitempty"`
	Type                    string          `json:"type,omitempty"`
	Data                    json.RawMessage `json:"data,omitempty"`
	ProjectID               string          `json:"projectId,omitempty"`
	IsManaged               Bool            `json:"isManaged,omitempty"`
	IsGlobal                Bool            `json:"isGlobal,omitempty"`
	IsResolvable            Bool            `json:"isResolvable,omitempty"`
	ResolvableAllowFallback Bool            `json:"resolvableAllowFallback,omitempty"`
	CreatedAt               string          `json:"createdAt,omitempty"`
	UpdatedAt               string          `json:"updatedAt,omitempty"`
}

// Credentials returns a typed handle to the /credentials resource. Unlike most
// n8n resources, credential update uses PATCH rather than PUT.
func (c *Client) Credentials() *Resource[Credential] {
	return NewResource[Credential](c, "credentials", WithUpdateMethod(http.MethodPatch))
}

// CredentialSchema returns the JSON schema describing a credential type's fields
// (GET /credentials/schema/{credentialTypeName}). The raw schema is returned so
// callers can render it as-is.
func (c *Client) CredentialSchema(ctx context.Context, credentialTypeName string) (json.RawMessage, error) {
	return c.doRaw(ctx, http.MethodGet, "credentials/schema/"+credentialTypeName, nil, nil)
}

// TransferCredential moves a credential to another project (PUT /credentials/{id}/transfer).
func (c *Client) TransferCredential(ctx context.Context, id, destinationProjectID string) error {
	body := map[string]string{"destinationProjectId": destinationProjectID}
	return c.Credentials().ActionMethod(ctx, http.MethodPut, id, "transfer", body, nil)
}
