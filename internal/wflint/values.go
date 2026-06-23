package wflint

import (
	"slices"
	"strconv"
	"strings"
)

// invalidOptionValues validates one parameter's value against the node's schema.
// It returns a finding message per invalid value, and is deliberately conservative
// — it only flags `options`/`multiOptions` parameters whose currently-active
// property variant has a *static* allowed-value list. Anything dynamic (loaded at
// runtime), expression-driven, or ambiguous is left alone, so the rule never
// reports a false positive on a legitimate workflow.
func invalidOptionValues(nodeType, param string, value any, params map[string]any) []string {
	variants, ok := paramVariants(nodeType, param)
	if !ok {
		return nil
	}

	// Gather the allowed values across the variants that are active for this node's
	// current parameters. If any active variant is a dynamic options list (no static
	// values), we cannot validate safely — bail out.
	allowed := map[string]bool{}
	active := false
	for _, ps := range variants {
		if ps.Type != "options" && ps.Type != "multiOptions" {
			continue
		}
		if !variantActive(ps, params) {
			continue
		}
		if len(ps.Options) == 0 {
			return nil // dynamic options — don't guess
		}
		active = true
		for _, o := range ps.Options {
			allowed[o] = true
		}
	}
	if !active {
		return nil
	}

	var out []string
	for _, v := range scalarStrings(value) {
		if isExpressionValue(v) || allowed[v] {
			continue
		}
		msg := "parameter \"" + param + "\" = \"" + v + "\" is not an allowed value"
		if s, ok := closest(v, allowed); ok {
			msg += " — did you mean \"" + s + "\"?"
		}
		out = append(out, msg)
	}
	return out
}

// variantActive reports whether a property variant is visible given the node's
// current parameters, by evaluating its displayOptions. A show condition must be
// satisfied by a present parameter; a missing controlling parameter makes the
// condition fail (conservative: we won't validate against a variant that may not
// apply). A hide condition that matches makes the variant inactive.
func variantActive(ps ParamSchema, params map[string]any) bool {
	if ps.DisplayOptions == nil {
		return true
	}
	for ctrl, allowed := range ps.DisplayOptions["show"] {
		if !valueMatchesAny(params[ctrl], allowed) {
			return false
		}
	}
	for ctrl, vals := range ps.DisplayOptions["hide"] {
		if valueMatchesAny(params[ctrl], vals) {
			return false
		}
	}
	return true
}

// valueMatchesAny reports whether v (a scalar or a list) intersects allowed.
func valueMatchesAny(v any, allowed []string) bool {
	for _, got := range scalarStrings(v) {
		if slices.Contains(allowed, got) {
			return true
		}
	}
	return false
}

// scalarStrings flattens a parameter value into the string scalars it carries:
// a string -> itself; a []any -> its string elements (multiOptions); numbers and
// bools -> their string form. Other shapes yield nothing.
func scalarStrings(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case bool:
		if t {
			return []string{"true"}
		}
		return []string{"false"}
	case float64:
		return []string{strconv.FormatFloat(t, 'f', -1, 64)}
	case []any:
		var out []string
		for _, e := range t {
			out = append(out, scalarStrings(e)...)
		}
		return out
	default:
		return nil
	}
}

func isExpressionValue(s string) bool {
	return strings.HasPrefix(s, "=") || strings.Contains(s, "{{")
}

// closest returns the nearest allowed value to v within edit distance 2.
func closest(v string, allowed map[string]bool) (string, bool) {
	best, bestDist := "", 3
	for a := range allowed {
		if d := levenshtein(strings.ToLower(a), strings.ToLower(v)); d < bestDist {
			best, bestDist = a, d
		}
	}
	return best, best != ""
}
