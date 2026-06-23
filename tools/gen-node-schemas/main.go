// Command gen-node-schemas builds the embedded n8n node catalog used by the
// wflint node-schema rules (unknown-node-type, unknown-parameter). It fetches
// n8n's published node descriptions (the same data the editor uses) and distills
// each node to its display name and the set of valid parameter names.
//
// Run from the repo root to refresh the catalog:
//
//	go run ./tools/gen-node-schemas
//
// It writes internal/wflint/node_catalog.json, which is embedded at build time.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

// Pinned package versions — bump these to refresh the catalog against a newer
// n8n. The path serves the full node descriptions the editor loads.
const (
	basePkg = "n8n-nodes-base"
	baseVer = "2.15.1"
	lcPkg   = "@n8n/n8n-nodes-langchain"
	lcVer   = "2.27.3"
	cdn     = "https://cdn.jsdelivr.net/npm/%s@%s/dist/types/nodes.json"
)

type source struct {
	pkg, ver, prefix string
}

// paramSchema is the slice of an n8n node property the lint rules need: enough to
// validate parameter values without re-implementing n8n's whole property model.
type paramSchema struct {
	Type           string                         `json:"type,omitempty"`
	Options        []string                       `json:"options,omitempty"`        // allowed values for options/multiOptions
	Required       bool                           `json:"required,omitempty"`       // n8n "required" flag
	DisplayOptions map[string]map[string][]string `json:"displayOptions,omitempty"` // "show"/"hide" -> param -> values
}

type nodeEntry struct {
	DisplayName string                   `json:"displayName"`
	Version     int                      `json:"version,omitempty"` // latest typeVersion
	Params      map[string][]paramSchema `json:"params"`            // name -> property variants (n8n repeats a name per resource/operation)
}

type catalog struct {
	GeneratedFrom map[string]string    `json:"generatedFrom"`
	Nodes         map[string]nodeEntry `json:"nodes"`
}

func main() {
	// Versions default to the pinned constants but can be overridden, so the
	// scheduled refresh job (.github/workflows/node-catalog.yml) can track latest.
	sources := []source{
		{basePkg, envOr("N8N_NODES_BASE_VER", baseVer), "n8n-nodes-base."},
		{lcPkg, envOr("N8N_LANGCHAIN_VER", lcVer), "@n8n/n8n-nodes-langchain."},
	}
	out := catalog{
		GeneratedFrom: map[string]string{},
		Nodes:         map[string]nodeEntry{},
	}

	client := &http.Client{Timeout: 60 * time.Second}
	for _, s := range sources {
		nodes, err := fetch(client, fmt.Sprintf(cdn, s.pkg, s.ver))
		if err != nil {
			fatalf("fetch %s@%s: %v", s.pkg, s.ver, err)
		}
		for _, n := range nodes {
			name, _ := n["name"].(string)
			if name == "" {
				continue
			}
			fullType := s.prefix + name
			// n8n publishes one entry per major-version line of a versioned node, all
			// sharing the same name. Merge them: take the highest version and the union
			// of parameter variants, so the catalog covers every supported version (and
			// the lint rules don't false-positive on a param that only exists in one).
			entry := out.Nodes[fullType]
			if entry.Params == nil {
				entry.Params = map[string][]paramSchema{}
			}
			if entry.DisplayName == "" {
				entry.DisplayName = stringOr(n["displayName"], name)
			}
			if v := latestVersion(n); v > entry.Version {
				entry.Version = v
			}
			for pname, variants := range topLevelParams(n["properties"]) {
				entry.Params[pname] = append(entry.Params[pname], variants...)
			}
			out.Nodes[fullType] = entry
		}
		out.GeneratedFrom[s.pkg] = s.ver
		fmt.Fprintf(os.Stderr, "%s@%s: %d nodes\n", s.pkg, s.ver, len(nodes))
	}

	// Drop duplicate variants introduced by merging version entries.
	for _, e := range out.Nodes {
		for pname, variants := range e.Params {
			e.Params[pname] = dedupVariants(variants)
		}
	}

	b, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		fatalf("marshal: %v", err)
	}
	const dst = "internal/wflint/node_catalog.json"
	if err := os.WriteFile(dst, append(b, '\n'), 0o644); err != nil { //nolint:gosec // generated data, not secret
		fatalf("write %s: %v", dst, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d node types)\n", dst, len(out.Nodes))
}

func fetch(c *http.Client, url string) ([]map[string]any, error) {
	resp, err := c.Get(url) //nolint:noctx // one-shot generator
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	var nodes []map[string]any
	if err := json.Unmarshal(body, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// topLevelParams returns a node's top-level properties keyed by name — exactly the
// keys a workflow node's `parameters` object uses. For each it captures the type,
// the static allowed values (for options/multiOptions), the required flag, and the
// displayOptions visibility rules. Nested collection entries are excluded (they are
// values under a top-level key, not parameter keys themselves).
func topLevelParams(properties any) map[string][]paramSchema {
	props, ok := properties.([]any)
	if !ok {
		return map[string][]paramSchema{}
	}
	out := map[string][]paramSchema{}
	for _, p := range props {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		name, _ := pm["name"].(string)
		if name == "" {
			continue
		}
		ps := paramSchema{Type: stringOr(pm["type"], "")}
		if req, ok := pm["required"].(bool); ok {
			ps.Required = req
		}
		if ps.Type == "options" || ps.Type == "multiOptions" {
			ps.Options = optionValues(pm["options"])
		}
		ps.DisplayOptions = displayOptions(pm["displayOptions"])
		out[name] = append(out[name], ps)
	}
	return out
}

// optionValues collects the static `value`s of an options/multiOptions property.
// Returns nil when the options are loaded dynamically (loadOptionsMethod), so the
// value rule stays silent rather than guessing.
func optionValues(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, o := range arr {
		om, ok := o.(map[string]any)
		if !ok {
			continue
		}
		if val, ok := om["value"]; ok {
			out = append(out, stringifyValue(val))
		}
	}
	sort.Strings(out)
	return out
}

// displayOptions normalizes n8n's show/hide visibility map to
// section -> param -> []allowed-values (stringified).
func displayOptions(v any) map[string]map[string][]string {
	dm, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	out := map[string]map[string][]string{}
	for _, section := range []string{"show", "hide"} {
		sm, ok := dm[section].(map[string]any)
		if !ok {
			continue
		}
		conds := map[string][]string{}
		for param, vals := range sm {
			if arr, ok := vals.([]any); ok {
				var ss []string
				for _, x := range arr {
					// Keep only scalar conditions; n8n also uses comparator objects
					// like {_cnd: {gte: 1.2}} that we cannot evaluate — skip them so
					// they never become a bogus literal the lint rule compares against.
					switch x.(type) {
					case string, float64, bool:
						ss = append(ss, stringifyValue(x))
					}
				}
				if len(ss) > 0 {
					conds[param] = ss
				}
			}
		}
		if len(conds) > 0 {
			out[section] = conds
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// latestVersion returns a node's current typeVersion, preferring defaultVersion
// (the version n8n instantiates) and falling back to the max of the version field.
func latestVersion(n map[string]any) int {
	if v := nodeVersion(n["defaultVersion"]); v > 0 {
		return v
	}
	return nodeVersion(n["version"])
}

// dedupVariants removes property variants that are structurally identical (same
// type, options, required flag, and displayOptions), which merging version entries
// can introduce.
func dedupVariants(variants []paramSchema) []paramSchema {
	seen := map[string]bool{}
	var out []paramSchema
	for _, ps := range variants {
		sig, _ := json.Marshal(ps)
		if seen[string(sig)] {
			continue
		}
		seen[string(sig)] = true
		out = append(out, ps)
	}
	return out
}

// nodeVersion returns the highest typeVersion a node supports (the field is a
// number or an array of numbers).
func nodeVersion(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case []any:
		max := 0
		for _, x := range t {
			if f, ok := x.(float64); ok && int(f) > max {
				max = int(f)
			}
		}
		return max
	default:
		return 0
	}
}

// stringifyValue renders a JSON scalar (string/number/bool) as a string so option
// values and displayOptions conditions compare uniformly.
func stringifyValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", t)
	}
}

func stringOr(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-schemas: "+format+"\n", a...)
	os.Exit(1)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
