package api

import (
	"context"
	"net/http"
	"net/url"
)

// Resource is a generic, typed handle to one n8n REST collection (e.g. /workflows).
// The CRUD/pagination logic lives here once; a new resource is just a struct plus
// a Client accessor that returns NewResource[T].
type Resource[T any] struct {
	client *Client
	path   string // collection path, e.g. "workflows" or "executions"

	// updateMethod is PUT for most resources; credentials use PATCH. Set via
	// WithUpdateMethod so the generic core needs no per-resource branching.
	updateMethod string
}

// ResourceOption customizes a Resource at construction time.
type ResourceOption func(*resourceConfig)

type resourceConfig struct {
	updateMethod string
}

// WithUpdateMethod overrides the HTTP verb used by Update (default PUT).
func WithUpdateMethod(method string) ResourceOption {
	return func(rc *resourceConfig) { rc.updateMethod = method }
}

// NewResource constructs a typed resource handle bound to a collection path.
func NewResource[T any](c *Client, path string, opts ...ResourceOption) *Resource[T] {
	cfg := resourceConfig{updateMethod: http.MethodPut}
	for _, o := range opts {
		o(&cfg)
	}
	return &Resource[T]{client: c, path: path, updateMethod: cfg.updateMethod}
}

// Path returns the collection path.
func (r *Resource[T]) Path() string { return r.path }

// Raw returns the underlying client (for custom calls).
func (r *Resource[T]) Raw() *Client { return r.client }

// List fetches a single page and returns the items plus the next cursor ("" when
// the collection is exhausted).
func (r *Resource[T]) List(ctx context.Context, params ListParams) ([]T, string, error) {
	raw, err := r.client.doRaw(ctx, http.MethodGet, r.path, params.values(r.client.defaultLimit), nil)
	if err != nil {
		return nil, "", err
	}
	return decodeList[T](raw)
}

// ListAll walks the cursor until the collection is exhausted or maxPages is hit.
// maxPages <= 0 applies a safety cap of 10000 pages.
func (r *Resource[T]) ListAll(ctx context.Context, params ListParams, maxPages int) ([]T, error) {
	items, _, err := r.ListAllChecked(ctx, params, maxPages)
	return items, err
}

// ListAllChecked is ListAll plus a truncation flag: truncated is true when the
// page cap was reached while more pages remained.
func (r *Resource[T]) ListAllChecked(ctx context.Context, params ListParams, maxPages int) ([]T, bool, error) {
	if maxPages <= 0 {
		maxPages = DefaultMaxPages
	}
	var all []T
	cursor := params.Cursor
	for page := 0; page < maxPages; page++ {
		p := params
		p.Cursor = cursor
		items, next, err := r.List(ctx, p)
		if err != nil {
			return all, false, err
		}
		all = append(all, items...)
		if next == "" {
			return all, false, nil
		}
		cursor = next
	}
	return all, true, nil
}

// Get retrieves a single resource by id.
func (r *Resource[T]) Get(ctx context.Context, id string, query url.Values) (*T, error) {
	var out T
	if err := r.client.Do(ctx, http.MethodGet, r.path+"/"+url.PathEscape(id), query, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create POSTs a body to the collection and returns the created resource.
func (r *Resource[T]) Create(ctx context.Context, body any) (*T, error) {
	var out T
	if err := r.client.Do(ctx, http.MethodPost, r.path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update sends the configured update verb (PUT/PATCH) to /<path>/<id>.
func (r *Resource[T]) Update(ctx context.Context, id string, body any) (*T, error) {
	var out T
	if err := r.client.Do(ctx, r.updateMethod, r.path+"/"+url.PathEscape(id), nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a resource by id.
func (r *Resource[T]) Delete(ctx context.Context, id string) error {
	return r.client.Do(ctx, http.MethodDelete, r.path+"/"+url.PathEscape(id), nil, nil, nil)
}

// Action POSTs to /<path>/<id>/<action> (custom actions like activate, retry, stop).
// body and out are both optional.
func (r *Resource[T]) Action(ctx context.Context, id, action string, body, out any) error {
	return r.client.Do(ctx, http.MethodPost, r.path+"/"+url.PathEscape(id)+"/"+action, nil, body, out)
}

// ActionMethod is like Action but with an explicit verb (e.g. PUT for transfer/tags).
func (r *Resource[T]) ActionMethod(ctx context.Context, method, id, action string, body, out any) error {
	return r.client.Do(ctx, method, r.path+"/"+url.PathEscape(id)+"/"+action, nil, body, out)
}

// CollectionAction POSTs to /<path>/<action> (e.g. /executions/stop).
func (r *Resource[T]) CollectionAction(ctx context.Context, action string, body, out any) error {
	return r.client.Do(ctx, http.MethodPost, r.path+"/"+action, nil, body, out)
}
