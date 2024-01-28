package gotmx

import "strings"

// attr_transform.go contains functions for attribute transformation:
// - g-trans: transforms the element tag name
// - g-attif-*: conditionally adds/removes attributes
// - g-att-*: dynamically sets attributes
// - g-class, g-href, g-src: shortcuts for common dynamic attributes

// getTrans checks for the g-trans attribute.
// Returns the new tag name expression and whether the attribute was found.
func getTrans(attrs AttributeMap) (tagExpr string, found bool) {
	v, ok := attrs[attrDataGTrans]
	return v, ok
}

// resolveTrans resolves the g-trans attribute to get the effective tag name.
// Uses ctx.ResolveText to resolve any model path expressions in the attribute value.
// Returns the resolved tag name. If no g-trans is present, returns the original tag name.
func resolveTrans(ctx *RenderContext, attrs AttributeMap, originalTag string, data any) (string, error) {
	expr, found := getTrans(attrs)
	if !found {
		return originalTag, nil
	}

	// Use ctx.ResolveText to resolve model paths in the tag name expression
	resolved, err := ctx.ResolveText(expr, data, false)
	if err != nil {
		return originalTag, err
	}
	return resolved, nil
}

// processGAttIf processes all data-g-attif-* attributes in the attribute map.
// Uses ctx to evaluate condition expressions via evaluateCondition.
//
// For each data-g-attif-{name} attribute:
//   - If the condition is true and {name} doesn't exist, adds {name}="" (boolean attribute)
//   - If the condition is false and {name} exists, removes {name}
//
// The data-g-attif-* attributes themselves are always removed from the map.
// This function modifies the attrs map in place and returns it.
func processGAttIf(ctx *RenderContext, attrs AttributeMap, data any) AttributeMap {
	var keysToAdd []string
	var keysToRemove []string

	for attrName, condition := range attrs {
		if !strings.HasPrefix(attrName, attrDataGAttIf) {
			continue
		}

		keysToRemove = append(keysToRemove, attrName)
		targetAttr := strings.TrimPrefix(attrName, attrDataGAttIf)

		// Use evaluateCondition which now takes ctx to evaluate the condition
		isTrue, err := evaluateCondition(ctx, condition, data)
		if err != nil {
			// On error, skip this attribute (don't modify)
			continue
		}

		if isTrue {
			// If true but attribute doesn't exist, add it as boolean attribute
			if _, exists := attrs[targetAttr]; !exists {
				keysToAdd = append(keysToAdd, targetAttr)
			}
		} else {
			// If false and attribute exists, remove it
			if _, exists := attrs[targetAttr]; exists {
				keysToRemove = append(keysToRemove, targetAttr)
			}
		}
	}

	// Apply changes
	for _, key := range keysToAdd {
		attrs[key] = ""
	}
	for _, key := range keysToRemove {
		delete(attrs, key)
	}

	return attrs
}

// processGAtt processes the g-att-* prefix attributes during attribute extraction.
// This is called by getAttributes and strips the g-att- prefix so that
// g-att-disabled="..." becomes disabled="...".
// Returns the processed attribute name (with prefix stripped if applicable).
func processGAtt(attrName string) string {
	if strings.HasPrefix(attrName, attrDataGAtt) {
		return strings.TrimPrefix(attrName, attrDataGAtt)
	}
	if strings.HasPrefix(attrName, attrGAtt) {
		return strings.TrimPrefix(attrName, attrGAtt)
	}
	return attrName
}

// applyShortcutAttributes applies the shortcut attributes (g-class, g-href, g-src)
// by copying their values to the target attributes.
// This function modifies the attrs map in place.
// Attributes are already in canonical data-g-* form after normalization.
func applyShortcutAttributes(attrs AttributeMap) {
	// data-g-class -> class
	if value, found := attrs[attrDataGClass]; found {
		attrs["class"] = value
	}

	// data-g-href -> href
	if value, found := attrs[attrDataGHref]; found {
		attrs["href"] = value
	}

	// data-g-src -> src
	if value, found := attrs[attrDataGSrc]; found {
		attrs["src"] = value
	}
}

// isIgnoredRenderAttribute checks if an attribute should not be rendered to output.
// This includes all gotmx control attributes and data-g-attif-* prefix attributes.
// After parse-time normalization, only canonical data-g-* forms exist in the map.
func isIgnoredRenderAttribute(attrName string) bool {
	// Check exact matches
	if _, ignored := ignoredRenderAttributes[attrName]; ignored {
		return true
	}

	// Check data-g-attif-* prefix (these remain in the map after processGAttIf
	// removes them, but only for attributes not yet processed)
	return strings.HasPrefix(attrName, attrDataGAttIf)
}
