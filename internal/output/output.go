// Package output renders any resource (or arbitrary JSON-able value) as a table,
// JSON, YAML, or CSV. Everything is normalized through encoding/json first, so a
// single renderer serves every resource without per-type formatting code.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format is an output format identifier.
type Format string

const (
	Table Format = "table"
	JSON  Format = "json"
	YAML  Format = "yaml"
	CSV   Format = "csv"
)

// Valid reports whether f is a supported format.
func (f Format) Valid() bool {
	switch f {
	case Table, JSON, YAML, CSV:
		return true
	default:
		return false
	}
}

// Parse converts a string to a Format, defaulting to Table for "".
func Parse(s string) (Format, error) {
	if s == "" {
		return Table, nil
	}
	f := Format(strings.ToLower(s))
	if !f.Valid() {
		return "", fmt.Errorf("invalid output format %q (want table|json|yaml|csv)", s)
	}
	return f, nil
}

// Options tune rendering.
type Options struct {
	Columns []string // explicit column selection/order (table & csv)
	NoColor bool     // disable ANSI color even on a TTY
	Color   bool     // enable color (caller decides based on TTY + NO_COLOR)
}

// Render writes data to w in the requested format. data may be a struct, a slice,
// a map, or a json.RawMessage — it is normalized to generic maps/slices first.
func Render(w io.Writer, data any, format Format, opts Options) error {
	switch format {
	case JSON:
		return renderJSON(w, data)
	case YAML:
		return renderYAML(w, data)
	case CSV:
		return renderCSV(w, data, opts.Columns)
	case Table, "":
		return renderTable(w, data, opts)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

func renderJSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(data)
}

func renderYAML(w io.Writer, data any) error {
	norm, err := normalize(data)
	if err != nil {
		return err
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(norm)
}

// normalize round-trips data through JSON so YAML/CSV/table all see the same
// generic shape (maps, slices, scalars) and honor json tags + custom marshalers.
func normalize(data any) (any, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var out any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// asRows coerces normalized data into a slice of row maps for table/csv output.
// A single object becomes a one-row slice.
func asRows(data any) ([]map[string]any, bool) {
	norm, err := normalize(data)
	if err != nil {
		return nil, false
	}
	switch v := norm.(type) {
	case []any:
		rows := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				rows = append(rows, m)
			} else {
				rows = append(rows, map[string]any{"value": item})
			}
		}
		return rows, true
	case map[string]any:
		return []map[string]any{v}, true
	default:
		return []map[string]any{{"value": v}}, true
	}
}

// columnsFor determines column order: explicit selection wins, otherwise a
// preferred-key ordering followed by the remaining keys alphabetically.
func columnsFor(rows []map[string]any, explicit []string) []string {
	if len(explicit) > 0 {
		return explicit
	}
	seen := map[string]bool{}
	for _, r := range rows {
		for k := range r {
			seen[k] = true
		}
	}
	preferred := []string{
		"id", "key", "name", "email", "type", "active", "isArchived", "status",
		"mode", "workflowId", "value", "role", "triggerCount", "finished",
		"description", "startedAt", "stoppedAt", "createdAt", "updatedAt",
	}
	var cols []string
	for _, p := range preferred {
		if seen[p] {
			cols = append(cols, p)
			delete(seen, p)
		}
	}
	rest := make([]string, 0, len(seen))
	for k := range seen {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	return append(cols, rest...)
}

func renderCSV(w io.Writer, data any, explicit []string) error {
	rows, ok := asRows(data)
	if !ok {
		return fmt.Errorf("cannot render value as csv")
	}
	cols := columnsFor(rows, explicit)
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write(cols); err != nil {
		return err
	}
	for _, r := range rows {
		rec := make([]string, len(cols))
		for i, c := range cols {
			rec[i] = scalar(r[c])
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	return cw.Error()
}

func renderTable(w io.Writer, data any, opts Options) error {
	rows, ok := asRows(data)
	if !ok {
		return fmt.Errorf("cannot render value as table")
	}
	if len(rows) == 0 {
		fmt.Fprintln(w, "No results.")
		return nil
	}
	cols := columnsFor(rows, opts.Columns)

	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	cells := make([][]string, len(rows))
	for ri, r := range rows {
		cells[ri] = make([]string, len(cols))
		for ci, c := range cols {
			s := scalar(r[c])
			cells[ri][ci] = s
			if len(s) > widths[ci] {
				widths[ci] = len(s)
			}
		}
	}

	header := make([]string, len(cols))
	for i, c := range cols {
		header[i] = pad(strings.ToUpper(c), widths[i])
	}
	headerLine := strings.TrimRight(strings.Join(header, "  "), " ")
	if opts.Color && !opts.NoColor {
		headerLine = "\x1b[1m" + headerLine + "\x1b[0m"
	}
	fmt.Fprintln(w, headerLine)

	for _, row := range cells {
		out := make([]string, len(cols))
		for i, s := range row {
			out[i] = pad(s, widths[i])
		}
		fmt.Fprintln(w, strings.TrimRight(strings.Join(out, "  "), " "))
	}
	return nil
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// scalar renders a cell value as a compact single-line string. Nested
// objects/arrays are emitted as compact JSON so a table stays one row per record.
func scalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return collapse(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case json.Number:
		return t.String()
	case float64:
		return fmt.Sprintf("%v", t)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return collapse(string(raw))
	}
}

// collapse flattens newlines/tabs so a value never breaks the table grid.
func collapse(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}
