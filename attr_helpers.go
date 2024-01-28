package gotmx

import "strings"

// attr_helpers.go contains common helper functions for attribute handling.
// These are pure functions with no allocations beyond what's necessary.

// normalizeAttrName converts short-form g-* attributes to their canonical data-g-* form.
// This is called once per attribute at parse/extraction time in getAttributes(),
// so all downstream rendering code only needs to check the canonical form.
//
// Examples:
//
//	"g-if"           → "data-g-if"
//	"g-inner-text"   → "data-g-inner-text"
//	"data-g-if"      → "data-g-if"   (already canonical, unchanged)
//	"class"          → "class"        (not a g-* attribute, unchanged)
//	"disabled"       → "disabled"     (result of processGAtt stripping g-att-, unchanged)
func normalizeAttrName(name string) string {
	if strings.HasPrefix(name, "g-") {
		return "data-" + name
	}
	return name
}

// applyOverrides copies all attributes from overrides into attrs, replacing existing values.
func applyOverrides(attrs AttributeMap, overrides AttributeMap) {
	for name, value := range overrides {
		attrs[name] = value
	}
}
