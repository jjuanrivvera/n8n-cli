package api

import (
	"context"
	"encoding/json"
	"net/http"
)

// These endpoints are instance-level operations rather than CRUD collections, so
// they are exposed as direct client methods and surfaced as standalone commands.

// AuditOptions are the optional knobs for GenerateAudit.
type AuditOptions struct {
	// DaysAbandonedWorkflow flags workflows not executed within this many days.
	DaysAbandonedWorkflow int
	// Categories restricts the audit to a subset:
	// credentials, database, nodes, filesystem, instance.
	Categories []string
}

// GenerateAudit runs a security audit (POST /audit) and returns the raw report,
// which is a free-form, section-keyed object.
//
// See https://docs.n8n.io/api/api-reference/#tag/Audit
func (c *Client) GenerateAudit(ctx context.Context, opts AuditOptions) (json.RawMessage, error) {
	var body any
	add := map[string]any{}
	if opts.DaysAbandonedWorkflow > 0 {
		add["daysAbandonedWorkflow"] = opts.DaysAbandonedWorkflow
	}
	if len(opts.Categories) > 0 {
		add["categories"] = opts.Categories
	}
	if len(add) > 0 {
		body = map[string]any{"additionalOptions": add}
	}
	return c.doRaw(ctx, http.MethodPost, "audit", nil, body)
}

// SourceControlPull pulls changes from the connected remote repository
// (POST /source-control/pull). Requires the licensed Source Control feature.
//
// See https://docs.n8n.io/api/api-reference/#tag/SourceControl
func (c *Client) SourceControlPull(ctx context.Context, force bool, variables map[string]any) (json.RawMessage, error) {
	body := map[string]any{"force": force}
	if len(variables) > 0 {
		body["variables"] = variables
	}
	return c.doRaw(ctx, http.MethodPost, "source-control/pull", nil, body)
}
