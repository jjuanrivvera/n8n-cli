package output

import (
	"io"
	"testing"
)

// A representative list payload (50 workflow-like records) for render benchmarks.
func benchData(n int) []map[string]any {
	out := make([]map[string]any, n)
	for i := range out {
		out[i] = map[string]any{
			"id": "wkfl000000000000", "name": "Some Workflow", "active": true,
			"isArchived": false, "triggerCount": 1, "updatedAt": "2026-06-22T00:00:00.000Z",
		}
	}
	return out
}

func BenchmarkRenderTable(b *testing.B) {
	data := benchData(50)
	cols := []string{"id", "name", "active", "triggerCount", "updatedAt"}
	for b.Loop() {
		_ = Render(io.Discard, data, Table, Options{Columns: cols})
	}
}

func BenchmarkRenderJSON(b *testing.B) {
	data := benchData(50)
	for b.Loop() {
		_ = Render(io.Discard, data, JSON, Options{})
	}
}

func BenchmarkApplyJQ(b *testing.B) {
	data := benchData(50)
	for b.Loop() {
		_ = ApplyJQ(io.Discard, data, ".[] | select(.active) | .id")
	}
}
