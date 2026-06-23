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

type nodeEntry struct {
	DisplayName string   `json:"displayName"`
	Params      []string `json:"params"`
}

type catalog struct {
	GeneratedFrom map[string]string    `json:"generatedFrom"`
	Nodes         map[string]nodeEntry `json:"nodes"`
}

func main() {
	sources := []source{
		{basePkg, baseVer, "n8n-nodes-base."},
		{lcPkg, lcVer, "@n8n/n8n-nodes-langchain."},
	}
	out := catalog{
		GeneratedFrom: map[string]string{basePkg: baseVer, lcPkg: lcVer},
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
			names := map[string]bool{}
			collectNames(n["properties"], names)
			out.Nodes[fullType] = nodeEntry{
				DisplayName: stringOr(n["displayName"], name),
				Params:      sortedKeys(names),
			}
		}
		fmt.Fprintf(os.Stderr, "%s@%s: %d nodes\n", s.pkg, s.ver, len(nodes))
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var nodes []map[string]any
	if err := json.Unmarshal(body, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// collectNames walks a node's properties tree and records every string under a
// "name" key. This is a deliberate superset (it also picks up option-value names),
// which keeps the unknown-parameter check conservative (fewer false positives).
func collectNames(v any, into map[string]bool) {
	switch t := v.(type) {
	case map[string]any:
		if n, ok := t["name"].(string); ok && n != "" {
			into[n] = true
		}
		for _, vv := range t {
			collectNames(vv, into)
		}
	case []any:
		for _, vv := range t {
			collectNames(vv, into)
		}
	}
}

func stringOr(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "gen-node-schemas: "+format+"\n", a...)
	os.Exit(1)
}
