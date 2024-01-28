package gotmx

import "strings"

// attr_control_flow.go contains functions for control flow attributes:
// - g-ignore: controls whether/how elements are rendered
// - g-if: conditional rendering
// - g-with: data context switching

// ignoreResult holds the result of checking the g-ignore attribute.
type ignoreResult struct {
	Mode      string // "outer", "inner", "outer-only", "none", or ""
	HasIgnore bool   // whether the attribute was present
}

// checkIgnore evaluates the g-ignore attribute and returns the effective ignore mode.
// Uses ctx.ResolveText to resolve any model path expressions in the attribute value.
//
// The mode can be:
//   - "outer" or "": skip the entire element and its children
//   - "inner": render the element but skip its children
//   - "outer-only": skip the element but render its children
//   - "none": render everything normally (used to override inherited ignore)
func checkIgnore(ctx *RenderContext, attrs AttributeMap, data any) (ignoreResult, error) {
	value, found := attrs[attrDataGIgnore]
	if !found {
		return ignoreResult{Mode: "", HasIgnore: false}, nil
	}

	effectiveMode, err := ctx.ResolveText(value, data, false)
	if err != nil {
		return ignoreResult{Mode: "", HasIgnore: true}, err
	}

	return ignoreResult{Mode: effectiveMode, HasIgnore: true}, nil
}

// shouldSkipElement returns true if the ignore mode means we should skip the entire element.
func shouldSkipElement(result ignoreResult) bool {
	return result.HasIgnore && (result.Mode == attrGIgnoreOuter || result.Mode == "")
}

// shouldSkipChildren returns true if the ignore mode means we should skip the children.
func shouldSkipChildren(result ignoreResult) bool {
	return result.HasIgnore && result.Mode == attrGIgnoreInner
}

// shouldSkipOuterOnly returns true if the ignore mode means we should skip only the outer element.
func shouldSkipOuterOnly(result ignoreResult) bool {
	return result.HasIgnore && result.Mode == attrGIgnoreOuterOnly
}

// shouldRenderNormally returns true if ignore is "none" or not present.
func shouldRenderNormally(result ignoreResult) bool {
	return !result.HasIgnore || result.Mode == attrGIgnoreNone
}

// checkIf evaluates the g-if attribute and returns whether the element should be rendered.
// If no g-if attribute is present, returns true (render by default).
func checkIf(ctx *RenderContext, attrs AttributeMap, data any) (shouldRender bool, err error) {
	value, found := attrs[attrDataGIf]
	if !found {
		return true, nil
	}
	return evaluateCondition(ctx, value, data)
}

// evaluateCondition evaluates a condition expression and returns whether it's true.
// The resolved value is compared case-insensitively:
//   - "true" → true
//   - "false" → false
//   - "!true" → false (negation)
//   - "!false" → true (negation)
//   - anything else → false
func evaluateCondition(ctx *RenderContext, expression string, data any) (bool, error) {
	text, err := ctx.ResolveText(expression, data, false)
	if err != nil {
		return false, err
	}
	lower := strings.ToLower(text)

	// Support negation prefix
	if strings.HasPrefix(lower, "!") {
		inner := strings.TrimPrefix(lower, "!")
		return inner != "true", nil
	}
	return lower == "true", nil
}

// applyWith applies the g-with attribute to switch the data context.
// If g-with is present and points to a valid path, returns the new context.
// Otherwise returns the original data unchanged.
func applyWith(ctx *RenderContext, attrs AttributeMap, data any) any {
	value, found := attrs[attrDataGWith]
	if !found {
		return data
	}

	newContext, isModelPath := ctx.ResolveValue(value, data)
	if isModelPath && newContext != nil {
		return newContext
	}
	return data
}
