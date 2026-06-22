package api

import (
	"encoding/json"
	"testing"
)

// A realistic workflow-list page for decode benchmarks.
var benchPage = func() []byte {
	items := make([]map[string]any, 50)
	for i := range items {
		items[i] = map[string]any{
			"id": float64(1000 + i), "name": "Workflow", "active": true,
			"triggerCount": 1, "nodes": []any{}, "connections": map[string]any{},
		}
	}
	b, _ := json.Marshal(map[string]any{"data": items, "nextCursor": nil})
	return b
}()

func BenchmarkDecodeList(b *testing.B) {
	for b.Loop() {
		_, _, _ = decodeList[Workflow](benchPage)
	}
}

func BenchmarkID_Unmarshal(b *testing.B) {
	inputs := [][]byte{[]byte(`"abc123"`), []byte(`1000`), []byte(`null`)}
	n := 0
	for b.Loop() {
		var id ID
		_ = id.UnmarshalJSON(inputs[n%len(inputs)])
		n++
	}
}
