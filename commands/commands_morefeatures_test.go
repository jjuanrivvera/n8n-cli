package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- output: --jq, -o id, --no-header ---

func TestOutputJQAndIDOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"w1","name":"A","active":true},{"id":"w2","name":"B","active":false}],"nextCursor":null}`))
	}))
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	out, _, err := run(t, "workflows", "list", "-o", "id", "--quiet")
	require.NoError(t, err)
	assert.Equal(t, "w1\nw2", strings.TrimSpace(out))

	out, _, err = run(t, "workflows", "list", "--jq", ".[] | select(.active) | .name", "--quiet")
	require.NoError(t, err)
	assert.Equal(t, "A", strings.TrimSpace(out))

	out, _, err = run(t, "workflows", "list", "--no-header", "--no-color", "--quiet")
	require.NoError(t, err)
	assert.NotContains(t, out, "NAME")
	assert.Contains(t, out, "A")
}

// --- skills install/path/print ---

func TestSkillsCommands(t *testing.T) {
	setupProfile(t, "https://a/api/v1")

	out, _, err := run(t, "skills", "path", "--global", "--agent", "cursor")
	require.NoError(t, err)
	assert.Contains(t, out, filepath.Join(".cursor", "skills", "n8nctl-cli"))

	out, _, err = run(t, "skills", "print")
	require.NoError(t, err)
	assert.Contains(t, out, "n8nctl")

	dir := t.TempDir()
	out, _, err = run(t, "skills", "install", "--dir", dir)
	require.NoError(t, err)
	assert.Contains(t, out, "Installed")
	_, statErr := os.Stat(filepath.Join(dir, "n8nctl-cli", "SKILL.md"))
	require.NoError(t, statErr)
	// references copied too
	refs, _ := os.ReadDir(filepath.Join(dir, "n8nctl-cli", "references"))
	assert.NotEmpty(t, refs)

	// dry-run lists files without writing
	out, _, err = run(t, "skills", "install", "--dir", t.TempDir(), "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, out, "Would install")

	// unknown agent
	_, _, err = run(t, "skills", "install", "--agent", "nope")
	require.Error(t, err)
}

// --- data-tables ---

func dataTableServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/data-tables", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"d1","name":"orders"}],"nextCursor":null}`))
	})
	mux.HandleFunc("POST /api/v1/data-tables", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"d2","name":"new"}`))
	})
	mux.HandleFunc("GET /api/v1/data-tables/d1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"d1","name":"orders"}`))
	})
	mux.HandleFunc("PATCH /api/v1/data-tables/d1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"d1","name":"renamed"}`))
	})
	mux.HandleFunc("DELETE /api/v1/data-tables/d1", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"id":"d1"}`))
	})
	mux.HandleFunc("GET /api/v1/data-tables/d1/rows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":1,"sku":"A"}],"nextCursor":null}`))
	})
	mux.HandleFunc("POST /api/v1/data-tables/d1/rows", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1}]`))
	})
	mux.HandleFunc("PATCH /api/v1/data-tables/d1/rows/update", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":1}]`))
	})
	mux.HandleFunc("POST /api/v1/data-tables/d1/rows/upsert", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"id":2}]`))
	})
	mux.HandleFunc("DELETE /api/v1/data-tables/d1/rows/delete", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"deleted":1}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestDataTableCommands(t *testing.T) {
	srv := dataTableServer(t)
	setupProfile(t, srv.URL)
	cases := [][]string{
		{"data-tables", "list", "-o", "json", "--quiet"},
		{"data-tables", "get", "d1", "-o", "json"},
		{"data-tables", "create", "--set", "name=new", "--set", `columns=[{"name":"sku","type":"string"}]`, "-o", "json"},
		{"data-tables", "update", "d1", "--set", "name=renamed", "-o", "json"},
		{"data-tables", "rows", "d1", "-o", "json", "--quiet"},
		{"data-tables", "add-rows", "d1", "--data", `[{"sku":"B"}]`, "-o", "json"},
		{"data-tables", "update-rows", "d1", "--data", `{"filter":{"type":"and"},"data":{"sku":"C"}}`, "-o", "json"},
		{"data-tables", "upsert-rows", "d1", "--data", `{"filter":{},"data":{}}`, "-o", "json"},
		{"data-tables", "delete-rows", "d1", "--filter", `{"type":"and"}`, "-o", "json"},
		{"data-tables", "delete", "d1", "--yes"},
	}
	for _, args := range cases {
		_, _, err := run(t, args...)
		require.NoError(t, err, "args: %v", args)
	}
	// missing filter on delete-rows errors
	_, _, err := run(t, "data-tables", "delete-rows", "d1")
	require.Error(t, err)
}

// --- packages ---

func TestPackageCommands(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/n8n-packages/export", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("GZIP-ARCHIVE-BYTES"))
	})
	mux.HandleFunc("POST /api/v1/n8n-packages/import", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"imported":2}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	setupProfile(t, srv.URL)

	out := filepath.Join(t.TempDir(), "pkg.n8np")
	_, _, err := run(t, "packages", "export", "--workflow", "w1", "--workflow", "w2", "--out", out)
	require.NoError(t, err)
	data, rerr := os.ReadFile(out)
	require.NoError(t, rerr)
	assert.Equal(t, "GZIP-ARCHIVE-BYTES", string(data))

	res, _, err := run(t, "packages", "import", "--file", out, "--conflict-policy", "fail", "--project", "p1", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, res, "imported")

	// validation
	_, _, err = run(t, "packages", "export", "--out", out)
	require.Error(t, err)
	_, _, err = run(t, "packages", "import", "--file", out)
	require.Error(t, err)
}
