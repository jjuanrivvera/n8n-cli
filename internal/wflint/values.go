package wflint

import (
	"slices"
	"strconv"
	"strings"
)

// invalidOptionValues validates one parameter's value against the node's schema.
// It returns a finding message per invalid value, and is deliberately conservative
// — it flags an `options`/`multiOptions` value only when it is absent from the
// union of static allowed-values across every variant that could plausibly be
// shown. Anything dynamic (loaded at runtime), expression-driven, or whose
// visibility it cannot resolve is left alone, so the rule never reports a false
// positive on a legitimate workflow.
func invalidOptionValues(nodeType, param string, value any, params map[string]any) []string {
	variants, ok := paramVariants(nodeType, param)
	if !ok {
		return nil
	}

	// Gather the allowed values across every variant that could plausibly be shown
	// for this node's current parameters. If any such variant is a dynamic options
	// list (no static values), we cannot validate safely — bail out.
	allowed := map[string]bool{}
	haveStatic := false
	for _, ps := range variants {
		if ps.Type != "options" && ps.Type != "multiOptions" {
			continue
		}
		if variantHidden(ps, params) {
			continue
		}
		if len(ps.Options) == 0 {
			return nil // dynamic options — don't guess
		}
		haveStatic = true
		for _, o := range ps.Options {
			allowed[o] = true
		}
	}
	if !haveStatic {
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

// variantHidden reports whether a property variant is *definitely* not shown for
// the node's current parameters. It errs toward visible: it returns true only when
// a controlling parameter is actually present and contradicts the variant's
// displayOptions. A missing controlling parameter (using its default), a meta-key
// reference (`@version` = the node's typeVersion, `/foo` = a root-context ref), or
// a non-scalar comparator condition (`{_cnd: …}`, stripped to an empty list) is
// treated as unknown — so the value rule never narrows the allowed set on a guess,
// and never false-positives on a parameter whose real options it cannot resolve.
func variantHidden(ps ParamSchema, params map[string]any) bool {
	for ctrl, allowed := range ps.DisplayOptions["show"] {
		if metaKey(ctrl) || len(allowed) == 0 {
			continue
		}
		if v, present := params[ctrl]; present && !valueMatchesAny(v, allowed) {
			return true
		}
	}
	for ctrl, vals := range ps.DisplayOptions["hide"] {
		if metaKey(ctrl) || len(vals) == 0 {
			continue
		}
		if v, present := params[ctrl]; present && valueMatchesAny(v, vals) {
			return true
		}
	}
	return false
}

// metaKey reports whether a displayOptions control key is a meta/context reference
// (`@version`, `@tool`, `/rootRef`) rather than a sibling node parameter.
func metaKey(k string) bool {
	return strings.HasPrefix(k, "@") || strings.HasPrefix(k, "/")
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

// maxSuggestionDist is the largest edit distance for a "did you mean" suggestion.
const maxSuggestionDist = 2

// closest returns the nearest allowed value to v within maxSuggestionDist.
func closest(v string, allowed map[string]bool) (string, bool) {
	best, bestDist := "", maxSuggestionDist+1
	for a := range allowed {
		if d := levenshtein(strings.ToLower(a), strings.ToLower(v)); d < bestDist {
			best, bestDist = a, d
		}
	}
	return best, best != ""
}
