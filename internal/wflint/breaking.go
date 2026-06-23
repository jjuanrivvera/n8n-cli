package wflint

import (
	"encoding/json"
	"sort"

	"github.com/jjuanrivvera/n8n-cli/internal/api"
)

// VersionIssue describes one workflow node pinned to an older typeVersion than the
// catalog's latest, plus any parameters it uses that the catalog does not recognize
// for that node type — a signal of a potential breaking change on upgrade.
type VersionIssue struct {
	Node           string   `json:"node"`
	Type           string   `json:"type"`
	CurrentVersion int      `json:"currentVersion"`
	LatestVersion  int      `json:"latestVersion"`
	UnknownParams  []string `json:"unknownParams,omitempty"`
}

// BreakingChanges reports nodes whose typeVersion trails the catalog's latest for
// that node type. For each, UnknownParams lists parameters present on the node that
// the catalog does not define for any version of it (renamed, removed, or typos).
// It only considers catalogued node types with a known version, so community/custom
// nodes are never reported.
func BreakingChanges(wf *api.Workflow) []VersionIssue {
	var nodes []struct {
		Name        string         `json:"name"`
		Type        string         `json:"type"`
		TypeVersion float64        `json:"typeVersion"`
		Parameters  map[string]any `json:"parameters"`
	}
	_ = json.Unmarshal(wf.Nodes, &nodes)

	var out []VersionIssue
	for _, n := range nodes {
		latest, ok := NodeLatestVersion(n.Type)
		if !ok {
			continue
		}
		// Skip nodes with no typeVersion (0 = field omitted) and nodes already at or
		// past the catalog's latest — both are "nothing to report".
		cur := int(n.TypeVersion)
		if cur == 0 || cur >= latest {
			continue
		}
		vi := VersionIssue{Node: n.Name, Type: n.Type, CurrentVersion: cur, LatestVersion: latest}
		for p := range n.Parameters {
			if !paramKnown(n.Type, p) {
				vi.UnknownParams = append(vi.UnknownParams, p)
			}
		}
		sort.Strings(vi.UnknownParams)
		out = append(out, vi)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Node < out[j].Node })
	return out
}
