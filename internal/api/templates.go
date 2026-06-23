package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// TemplateAPIBase is the public n8n template gallery host (separate from any
// instance — it needs no API key).
const TemplateAPIBase = "https://api.n8n.io"

// TemplateAPI is a thin client for the n8n.io template gallery.
type TemplateAPI struct {
	BaseURL string
	HTTP    *http.Client
}

// NewTemplateAPI returns a client pointed at the public gallery. The base URL can
// be overridden with N8NCTL_TEMPLATE_API_URL (used by tests).
func NewTemplateAPI() *TemplateAPI {
	base := TemplateAPIBase
	if v := os.Getenv("N8NCTL_TEMPLATE_API_URL"); v != "" {
		base = v
	}
	return &TemplateAPI{BaseURL: base, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// TemplateSummary is one search hit.
type TemplateSummary struct {
	ID         ID     `json:"id"`
	Name       string `json:"name"`
	TotalViews int    `json:"totalViews"`
	User       struct {
		Username string `json:"username"`
	} `json:"user"`
}

// TemplateDetail is a single template with its workflow definition.
type TemplateDetail struct {
	ID          ID
	Name        string
	Description string
	Definition  json.RawMessage // the {nodes, connections, settings} object
}

// Search returns up to rows template summaries matching query.
func (t *TemplateAPI) Search(ctx context.Context, query string, rows int) ([]TemplateSummary, error) {
	if rows <= 0 {
		rows = 20
	}
	q := url.Values{}
	q.Set("page", "1")
	q.Set("rows", fmt.Sprintf("%d", rows))
	q.Set("search", query)
	var out struct {
		TotalWorkflows int               `json:"totalWorkflows"`
		Workflows      []TemplateSummary `json:"workflows"`
	}
	if err := t.get(ctx, "/templates/search?"+q.Encode(), &out); err != nil {
		return nil, err
	}
	return out.Workflows, nil
}

// Get fetches one template and extracts its workflow definition. The gallery wraps
// the definition as {workflow: {…metadata…, workflow: {nodes, connections}}}.
func (t *TemplateAPI) Get(ctx context.Context, id string) (*TemplateDetail, error) {
	var out struct {
		Workflow struct {
			ID          ID              `json:"id"`
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Workflow    json.RawMessage `json:"workflow"`
		} `json:"workflow"`
	}
	if err := t.get(ctx, "/templates/workflows/"+url.PathEscape(id), &out); err != nil {
		return nil, err
	}
	if len(out.Workflow.Workflow) == 0 {
		return nil, fmt.Errorf("template %s has no workflow definition", id)
	}
	return &TemplateDetail{
		ID:          out.Workflow.ID,
		Name:        out.Workflow.Name,
		Description: out.Workflow.Description,
		Definition:  out.Workflow.Workflow,
	}, nil
}

func (t *TemplateAPI) get(ctx context.Context, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.BaseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := t.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("template API %s: HTTP %d", path, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return err
	}
	return json.Unmarshal(body, into)
}
