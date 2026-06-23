package wffile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

func sampleWF() *api.Workflow {
	return &api.Workflow{
		Name:        "sample",
		Nodes:       json.RawMessage(`[{"id":"a","name":"Code","type":"n8n-nodes-base.code","parameters":{"jsCode":"const x = 1;\nconst y = 2;\nreturn [{x, y}];"}}]`),
		Connections: json.RawMessage(`{}`),
		Settings:    json.RawMessage(`{"executionOrder":"v1"}`),
	}
}

func TestFormatFromPath(t *testing.T) {
	assert.Equal(t, YAML, FormatFromPath("wf.yaml"))
	assert.Equal(t, YAML, FormatFromPath("wf.YML"))
	assert.Equal(t, JSON, FormatFromPath("wf.json"))
	assert.Equal(t, JSON, FormatFromPath("wf"))
}

func TestEncodeDecode_JSONandYAML(t *testing.T) {
	for _, f := range []Format{JSON, YAML} {
		data, err := Encode(sampleWF(), f)
		require.NoError(t, err, f)
		assert.Contains(t, string(data), "sample")

		back, err := Decode(data, f)
		require.NoError(t, err, f)
		assert.Equal(t, "sample", back.Name)
		assert.Contains(t, string(back.Nodes), "jsCode")
		// connections round-trips as an object
		assert.Contains(t, string(back.Connections), "{")
	}
}

func TestYAML_IsReadable(t *testing.T) {
	data, err := Encode(sampleWF(), YAML)
	require.NoError(t, err)
	s := string(data)
	assert.Contains(t, s, "name: sample")
	assert.Contains(t, s, "nodes:")
	assert.NotContains(t, s, "!!binary") // RawMessage must not leak as binary
}

func TestExternalizeRoundTrip(t *testing.T) {
	// jsCode is 3 lines; threshold 2 -> externalized.
	main, subfiles, err := EncodeExternalized(sampleWF(), YAML, "sample", 2)
	require.NoError(t, err)
	require.Len(t, subfiles, 1)

	var ref string
	for rel, content := range subfiles {
		ref = rel
		assert.Contains(t, string(content), "const x = 1;")
		assert.True(t, strings.HasSuffix(rel, ".js"))
	}
	// main file references the subfile, not the inline code
	assert.Contains(t, string(main), "$n8nctl_file")
	assert.NotContains(t, string(main), "const x = 1;")

	// Re-inline via loader.
	wf, err := DecodeWithFiles(main, YAML, func(rel string) ([]byte, error) {
		assert.Equal(t, ref, rel)
		return subfiles[rel], nil
	})
	require.NoError(t, err)
	assert.Contains(t, string(wf.Nodes), "const x = 1;")
	assert.NotContains(t, string(wf.Nodes), "$n8nctl_file")
}

func TestExternalize_BelowThreshold(t *testing.T) {
	// threshold higher than the code's line count -> nothing externalized
	_, subfiles, err := EncodeExternalized(sampleWF(), JSON, "sample", 10)
	require.NoError(t, err)
	assert.Empty(t, subfiles)
}

func TestExternalize_FieldExtensions(t *testing.T) {
	long := "a\nb\nc\nd\n"
	wf := &api.Workflow{
		Name: "multi",
		Nodes: json.RawMessage(`[{"name":"N","type":"x","parameters":{` +
			`"query":"` + jsonEsc(long) + `","jsonBody":"` + jsonEsc(long) + `","content":"` + jsonEsc(long) + `","pythonCode":"` + jsonEsc(long) + `"}}]`),
		Connections: json.RawMessage(`{}`),
	}
	_, subfiles, err := EncodeExternalized(wf, JSON, "multi", 2)
	require.NoError(t, err)
	exts := map[string]bool{}
	for rel := range subfiles {
		i := strings.LastIndex(rel, ".")
		exts[rel[i+1:]] = true
	}
	assert.True(t, exts["sql"])  // query
	assert.True(t, exts["json"]) // jsonBody
	assert.True(t, exts["md"])   // content
	assert.True(t, exts["py"])   // pythonCode
}

func TestDecodeWithFiles_LoaderError(t *testing.T) {
	main, _, err := EncodeExternalized(sampleWF(), JSON, "sample", 2)
	require.NoError(t, err)
	_, err = DecodeWithFiles(main, JSON, func(string) ([]byte, error) {
		return nil, assert.AnError
	})
	require.Error(t, err)
}

func jsonEsc(s string) string { return strings.ReplaceAll(s, "\n", "\\n") }

func TestDirLoader_RejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ok.js"), []byte("inside"), 0o600))
	load := DirLoader(dir)

	// legitimate relative path within dir works
	b, err := load("ok.js")
	require.NoError(t, err)
	assert.Equal(t, "inside", string(b))

	// absolute / rooted paths are explicitly refused on every platform
	for _, bad := range []string{"/etc/passwd"} {
		_, err := load(bad)
		require.Error(t, err, bad)
		assert.Contains(t, err.Error(), "refusing")
	}
	// traversal attempts must never read outside the dir; depending on the OS they
	// are either refused or confined to a non-existent path, but never succeed.
	for _, bad := range []string{"../../../../etc/passwd", "a/../../escape"} {
		_, err := load(bad)
		require.Error(t, err, bad)
	}
}

func TestDecode_Errors(t *testing.T) {
	_, err := Decode([]byte(`{bad json`), JSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing")
	_, err = Decode([]byte("- a\n- b\n"), YAML) // a sequence is not a workflow object
	require.Error(t, err)
	_, err = Decode([]byte(`null`), JSON)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty workflow")
}

func TestDirLoader_RejectsSymlinkEscape(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret")
	require.NoError(t, os.WriteFile(outside, []byte("top-secret"), 0o600))
	link := filepath.Join(dir, "link.js")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symlinks unsupported")
	}
	_, err := DirLoader(dir)("link.js")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resolves outside")
}
