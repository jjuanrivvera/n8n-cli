package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateAPI(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/templates/search", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "slack", r.URL.Query().Get("search"))
		_, _ = w.Write([]byte(`{"totalWorkflows":1,"workflows":[{"id":42,"name":"Slack thing","totalViews":99}]}`))
	})
	mux.HandleFunc("/templates/workflows/42", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"workflow":{"id":42,"name":"Slack thing","description":"d","workflow":{"nodes":[{"name":"A"}],"connections":{}}}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	ta := &TemplateAPI{BaseURL: srv.URL, HTTP: srv.Client()}

	hits, err := ta.Search(t.Context(), "slack", 5)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, "Slack thing", hits[0].Name)
	assert.Equal(t, 99, hits[0].TotalViews)

	d, err := ta.Get(t.Context(), "42")
	require.NoError(t, err)
	assert.Equal(t, "Slack thing", d.Name)
	assert.Contains(t, string(d.Definition), "\"nodes\"")
}

func TestNewTemplateAPIDefault(t *testing.T) {
	ta := NewTemplateAPI()
	if ta.BaseURL != TemplateAPIBase {
		t.Fatalf("default base = %q, want %q", ta.BaseURL, TemplateAPIBase)
	}
	if ta.HTTP == nil {
		t.Fatal("HTTP client must be set")
	}
}

func TestTemplateAPIErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	ta := &TemplateAPI{BaseURL: srv.URL, HTTP: srv.Client()}
	_, err := ta.Get(t.Context(), "999")
	require.Error(t, err) // non-200 -> error
	_, err = ta.Search(t.Context(), "x", 0)
	require.Error(t, err)
}

func TestTemplateGetMissingDefinition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"workflow":{"id":1,"name":"x"}}`)) // no inner workflow
	}))
	t.Cleanup(srv.Close)
	ta := &TemplateAPI{BaseURL: srv.URL, HTTP: srv.Client()}
	_, err := ta.Get(t.Context(), "1")
	require.Error(t, err)
}
