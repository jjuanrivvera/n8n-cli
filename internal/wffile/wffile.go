// Package wffile reads and writes n8n workflows as local files in JSON or YAML,
// with optional "code externalization": long inline code fields (jsCode, query,
// jsonBody, sticky-note content, ...) can be split into sibling files so a
// workflow diffs cleanly in Git. It is the shared file layer behind the
// convert/apply/backup/restore/diff commands.
package wffile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// Format is a workflow file serialization.
type Format string

const (
	JSON Format = "json"
	YAML Format = "yaml"
)

// FormatFromPath infers the format from a file extension (default JSON).
func FormatFromPath(path string) Format {
	l := strings.ToLower(path)
	if strings.HasSuffix(l, ".yaml") || strings.HasSuffix(l, ".yml") {
		return YAML
	}
	return JSON
}

// Encode serializes a workflow to JSON or YAML bytes. The workflow is normalized
// through JSON first so the raw structural fields (nodes/connections/settings)
// render as readable data rather than opaque blobs.
func Encode(wf *api.Workflow, format Format) ([]byte, error) {
	main, _, err := EncodeExternalized(wf, format, "", 0)
	return main, err
}

// extMarker is the single-key object an externalized field becomes on disk:
// {"$n8nctl_file": "relpath"}. The key is namespaced so it cannot collide with a
// legitimate workflow parameter that happens to be shaped like {"$ref": "..."}.
const extMarker = "$n8nctl_file"

// DirLoader returns a loader for externalized $ref files that is confined to dir.
// It rejects absolute paths and any path that escapes dir, so a crafted workflow
// file cannot make the CLI read arbitrary files (and upload them on apply/restore).
func DirLoader(dir string) func(string) ([]byte, error) {
	// Canonicalize the base once so symlinks in its ancestry (e.g. macOS
	// /var -> /private/var) don't make legitimate reads look like escapes.
	base := dir
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		base = r
	}
	return func(rel string) ([]byte, error) {
		clean := filepath.Clean(filepath.FromSlash(rel))
		if filepath.IsAbs(clean) {
			return nil, fmt.Errorf("refusing absolute externalized-file path %q", rel)
		}
		full := filepath.Join(dir, clean)
		if escapesDir(dir, full) {
			return nil, fmt.Errorf("refusing externalized-file path %q that escapes its directory", rel)
		}
		// Resolve symlinks and re-check against the canonical base, so a symlink
		// inside dir cannot redirect the read outside it. EvalSymlinks fails for a
		// not-yet-existing path; only re-check when it resolves.
		if resolved, err := filepath.EvalSymlinks(full); err == nil && escapesDir(base, resolved) {
			return nil, fmt.Errorf("refusing externalized-file path %q that resolves outside its directory", rel)
		}
		return os.ReadFile(full) //nolint:gosec // confined to dir by the checks above
	}
}

// escapesDir reports whether path lies outside dir.
func escapesDir(dir, path string) bool {
	rp, err := filepath.Rel(dir, path)
	return err != nil || rp == ".." || strings.HasPrefix(rp, ".."+string(filepath.Separator))
}

// externalizableExt maps a parameter field name to a file extension.
func externalizableExt(field string) (ext string, ok bool) {
	switch field {
	case "jsCode":
		return "js", true
	case "pythonCode":
		return "py", true
	case "query", "sqlQuery":
		return "sql", true
	case "jsonBody":
		return "json", true
	case "content", "text", "html":
		return "md", true
	default:
		return "", false
	}
}

// EncodeExternalized serializes a workflow, optionally extracting top-level string
// node parameters longer than threshold lines into sibling files. It returns the
// main file bytes plus a map of relative-path -> file content for the subfiles.
// threshold <= 0 disables externalization. stem names the subfile directory.
func EncodeExternalized(wf *api.Workflow, format Format, stem string, threshold int) ([]byte, map[string][]byte, error) {
	// Work on a generic copy so we can rewrite node parameters.
	raw, err := json.Marshal(wf)
	if err != nil {
		return nil, nil, err
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, nil, err
	}

	subfiles := map[string][]byte{}
	if threshold > 0 {
		externalizeNodes(generic, stem, threshold, subfiles)
	}

	var out []byte
	switch format {
	case YAML:
		out, err = yaml.Marshal(generic)
	default:
		out, err = json.MarshalIndent(generic, "", "  ")
		if err == nil {
			out = append(out, '\n')
		}
	}
	if err != nil {
		return nil, nil, err
	}
	return out, subfiles, nil
}

// externalizeNodes rewrites long top-level string node parameters into $ref
// markers and records their content in subfiles.
func externalizeNodes(wf map[string]any, stem string, threshold int, subfiles map[string][]byte) {
	nodes, ok := wf["nodes"].([]any)
	if !ok {
		return
	}
	for _, n := range nodes {
		node, ok := n.(map[string]any)
		if !ok {
			continue
		}
		params, ok := node["parameters"].(map[string]any)
		if !ok {
			continue
		}
		nodeName := fmt.Sprintf("%v", node["name"])
		// Deterministic field order so output is stable.
		fields := make([]string, 0, len(params))
		for f := range params {
			fields = append(fields, f)
		}
		sort.Strings(fields)
		for _, field := range fields {
			s, ok := params[field].(string)
			if !ok {
				continue
			}
			ext, ok := externalizableExt(field)
			if !ok || lineCount(s) <= threshold { // externalize only when strictly longer than N lines
				continue
			}
			rel := "_subfiles/" + slug(stem) + "/" + slug(nodeName) + "-" + field + "." + ext
			subfiles[rel] = []byte(s)
			params[field] = map[string]any{extMarker: rel}
		}
	}
}

// Decode parses a workflow from JSON or YAML. Externalized $ref fields are left
// as-is; use DecodeWithFiles to re-inline them.
func Decode(data []byte, format Format) (*api.Workflow, error) {
	return DecodeWithFiles(data, format, nil)
}

// DecodeWithFiles parses a workflow and re-inlines any externalized $ref fields by
// loading them via loader(relpath). A nil loader leaves $ref markers untouched.
func DecodeWithFiles(data []byte, format Format, loader func(relpath string) ([]byte, error)) (*api.Workflow, error) {
	var generic map[string]any
	switch format {
	case YAML:
		if err := yaml.Unmarshal(data, &generic); err != nil {
			return nil, fmt.Errorf("parsing YAML workflow: %w", err)
		}
	default:
		if err := json.Unmarshal(data, &generic); err != nil {
			return nil, fmt.Errorf("parsing JSON workflow: %w", err)
		}
	}
	if generic == nil {
		return nil, fmt.Errorf("empty workflow file")
	}
	if loader != nil {
		if err := inlineNodes(generic, loader); err != nil {
			return nil, err
		}
	}
	jsonBytes, err := json.Marshal(generic)
	if err != nil {
		return nil, err
	}
	var wf api.Workflow
	if err := json.Unmarshal(jsonBytes, &wf); err != nil {
		return nil, err
	}
	return &wf, nil
}

// inlineNodes replaces {"$ref": path} parameter values with the file's contents.
func inlineNodes(wf map[string]any, loader func(string) ([]byte, error)) error {
	nodes, ok := wf["nodes"].([]any)
	if !ok {
		return nil
	}
	for _, n := range nodes {
		node, ok := n.(map[string]any)
		if !ok {
			continue
		}
		params, ok := node["parameters"].(map[string]any)
		if !ok {
			continue
		}
		for field, v := range params {
			ref, ok := refPath(v)
			if !ok {
				continue
			}
			content, err := loader(ref)
			if err != nil {
				return fmt.Errorf("loading externalized field %q: %w", ref, err)
			}
			params[field] = string(content)
		}
	}
	return nil
}

// refPath returns the path of a {"$ref": path} marker, if v is one.
func refPath(v any) (string, bool) {
	m, ok := v.(map[string]any)
	if !ok || len(m) != 1 {
		return "", false
	}
	p, ok := m[extMarker].(string)
	return p, ok
}

func lineCount(s string) int { return strings.Count(s, "\n") + 1 }

// slug makes a filesystem-friendly token.
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "workflow"
	}
	return out
}
