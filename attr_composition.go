package gotmx

import "strings"

// attr_composition.go contains functions for template composition attributes:
// - g-use: renders another template instead of this element
// - g-inner-use: renders another template's inner content
// - g-use-slot: specifies which slot children go into
// - g-override-att: specifies which attributes to pass to the referenced template
// - g-as-template: treats children as a Go HTML template
// - g-as-unsafe-template: treats children as a Go text template (unescaped)

// getUse checks for the g-use attribute.
// Returns the template reference and whether the attribute was found.
func getUse(attrs AttributeMap) (templateRef string, found bool) {
	v, ok := attrs[attrDataGUse]
	return v, ok
}

// getInnerUse checks for the g-inner-use attribute.
// Returns the template reference and whether the attribute was found.
func getInnerUse(attrs AttributeMap) (templateRef string, found bool) {
	v, ok := attrs[attrDataGInnerUse]
	return v, ok
}

// getUseSlot checks for the g-use-slot attribute.
// Returns the slot name and whether the attribute was found.
func getUseSlot(attrs AttributeMap) (slotName string, found bool) {
	v, ok := attrs[attrDataGUseSlot]
	return v, ok
}

// getOverrideAtt checks for the g-override-att attribute.
// Returns the comma-separated list of attribute names and whether the attribute was found.
func getOverrideAtt(attrs AttributeMap) (attrList string, found bool) {
	v, ok := attrs[attrDataGOverrideAtt]
	return v, ok
}

// getAsTemplate checks for the g-as-template attribute.
// Returns the template name and whether the attribute was found.
func getAsTemplate(attrs AttributeMap) (templateName string, found bool) {
	v, ok := attrs[attrDataGAsTemplate]
	return v, ok
}

// getAsUnsafeTemplate checks for the g-as-unsafe-template attribute.
// Returns the template name and whether the attribute was found.
func getAsUnsafeTemplate(attrs AttributeMap) (templateName string, found bool) {
	v, ok := attrs[attrDataGAsUnsafeTemplate]
	return v, ok
}

// resolveTemplateRef resolves a template reference (may contain model paths or templates).
// Uses ctx.ResolveText to resolve any model path expressions in the template reference.
// Returns the resolved template name.
func resolveTemplateRef(ctx *RenderContext, templateRef string, data any) (string, error) {
	// Use ctx.ResolveText to resolve model paths in the template reference
	return ctx.ResolveText(templateRef, data, false)
}

// parseOverrideAttributes parses the g-override-att value into a list of attribute names.
// Handles comma-separated values and trims whitespace.
// Attribute names are normalized to canonical data-g-* form where applicable,
// so they match the normalized keys in the AttributeMap.
func parseOverrideAttributes(attrList string) []string {
	if attrList == "" {
		return nil
	}

	parts := strings.Split(attrList, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, normalizeAttrName(trimmed))
		}
	}

	return result
}

// collectOverrideAttributes collects the specified attributes from the source map.
// Also includes any matching data-g-attif-* attributes for the overridden attributes.
func collectOverrideAttributes(sourceAttrs AttributeMap, attrNames []string) AttributeMap {
	result := make(AttributeMap, len(attrNames))

	for _, attrName := range attrNames {
		// Copy the attribute value if it exists
		if value, exists := sourceAttrs[attrName]; exists {
			result[attrName] = value
		}

		// Also copy any matching data-g-attif- attributes (canonical form after normalization)
		attIfKey := attrDataGAttIf + attrName
		if value, exists := sourceAttrs[attIfKey]; exists {
			result[attIfKey] = value
		}
	}

	return result
}

// resolveSlotName resolves a slot name (may contain model paths or templates).
// Uses ctx.ResolveText to resolve any model path expressions in the slot name.
// Returns the resolved slot name and any error.
func resolveSlotName(ctx *RenderContext, slotName string, data any) (string, error) {
	// Use ctx.ResolveText to resolve model paths in the slot name
	return ctx.ResolveText(slotName, data, false)
}
