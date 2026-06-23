package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func parse(t *testing.T, s string) any {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestTopLevelParams(t *testing.T) {
	props := parse(t, `[
		{"name":"resource","type":"options","options":[{"name":"Msg","value":"message"},{"name":"Ch","value":"channel"}]},
		{"name":"operation","type":"options","options":[{"name":"Post","value":"post"}],"displayOptions":{"show":{"resource":["message"]}}},
		{"name":"text","type":"string"},
		{"type":"notice"}
	]`)
	got := topLevelParams(props)
	if len(got["resource"]) != 1 || got["resource"][0].Type != "options" {
		t.Fatalf("resource variant wrong: %+v", got["resource"])
	}
	if !reflect.DeepEqual(got["resource"][0].Options, []string{"channel", "message"}) {
		t.Fatalf("option values = %v", got["resource"][0].Options)
	}
	if got["operation"][0].DisplayOptions["show"]["resource"][0] != "message" {
		t.Fatalf("displayOptions not captured: %+v", got["operation"][0].DisplayOptions)
	}
	if _, ok := got["text"]; !ok {
		t.Fatal("string param dropped")
	}
}

func TestOptionValuesDynamicAndStatic(t *testing.T) {
	if optionValues(parse(t, `"loadOptionsMethod"`)) != nil {
		t.Fatal("dynamic options should yield nil")
	}
	got := optionValues(parse(t, `[{"value":"b"},{"value":"a"},{"name":"no-value"}]`))
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDisplayOptionsSkipsComparators(t *testing.T) {
	// comparator objects ({_cnd:...}) must be dropped, not stringified
	do := displayOptions(parse(t, `{"show":{"@version":[{"_cnd":{"gte":1.2}}],"resource":["message"]}}`))
	if _, ok := do["show"]["@version"]; ok {
		t.Fatalf("comparator condition should be dropped: %+v", do)
	}
	if do["show"]["resource"][0] != "message" {
		t.Fatalf("scalar condition lost: %+v", do)
	}
}

func TestNodeAndLatestVersion(t *testing.T) {
	if nodeVersion(parse(t, `[3,4,4.2]`)) != 4 {
		t.Fatal("max of version array wrong")
	}
	if nodeVersion(parse(t, `2`)) != 2 {
		t.Fatal("scalar version wrong")
	}
	if nodeVersion(parse(t, `"x"`)) != 0 {
		t.Fatal("non-numeric version should be 0")
	}
	// latestVersion prefers defaultVersion
	n := parse(t, `{"version":[1,2],"defaultVersion":4.4}`).(map[string]any)
	if latestVersion(n) != 4 {
		t.Fatal("should prefer defaultVersion")
	}
}

func TestDedupVariants(t *testing.T) {
	v := paramSchema{Type: "options", Options: []string{"a"}}
	out := dedupVariants([]paramSchema{v, v, {Type: "string"}})
	if len(out) != 2 {
		t.Fatalf("expected 2 unique variants, got %d", len(out))
	}
}

func TestStringifyValue(t *testing.T) {
	if stringifyValue(true) != "true" || stringifyValue(float64(3)) != "3" || stringifyValue("x") != "x" {
		t.Fatal("scalar stringify wrong")
	}
}

func TestEnvOr(t *testing.T) {
	t.Setenv("GEN_TEST_KEY", "v")
	if envOr("GEN_TEST_KEY", "d") != "v" || envOr("GEN_MISSING_KEY", "d") != "d" {
		t.Fatal("envOr wrong")
	}
}
