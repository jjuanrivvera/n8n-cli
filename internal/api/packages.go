package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// n8n Packages (.n8np) bundle workflows for transport between instances. The
// feature is Beta and disabled by default (N8N_PUBLIC_API_PACKAGES_ENABLED=true);
// while off, the endpoints return 404.
//
// See https://docs.n8n.io/api/api-reference/#tag/N8nPackage

// ExportPackage exports the given workflows as a gzip .n8np archive (raw bytes).
func (c *Client) ExportPackage(ctx context.Context, workflowIDs []string) ([]byte, error) {
	body := map[string]any{"workflowIds": workflowIDs}
	return c.doRaw(ctx, http.MethodPost, "n8n-packages/export", nil, body)
}

// ImportOptions tune a package import. Only ConflictPolicy is required by the API.
type ImportOptions struct {
	ConflictPolicy         string // workflowConflictPolicy (required), e.g. fail, new-version
	ProjectID              string
	FolderID               string
	WorkflowIDPolicy       string
	CredentialMatchingMode string
	CredentialMissingMode  string
}

// ImportPackage uploads a .n8np archive (multipart) and returns the raw result.
func (c *Client) ImportPackage(ctx context.Context, archive []byte, opts ImportOptions) (json.RawMessage, error) {
	fields := map[string]string{
		"workflowConflictPolicy": opts.ConflictPolicy,
		"projectId":              opts.ProjectID,
		"folderId":               opts.FolderID,
		"workflowIdPolicy":       opts.WorkflowIDPolicy,
		"credentialMatchingMode": opts.CredentialMatchingMode,
		"credentialMissingMode":  opts.CredentialMissingMode,
	}
	raw, err := c.PostMultipart(ctx, "n8n-packages/import", fields, "package", "package.n8np", archive)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}
