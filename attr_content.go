package gotmx

// attr_content.go contains functions for content replacement attributes:
// - g-inner-text: replaces innerHTML with escaped text
// - g-inner-html: replaces innerHTML with unescaped HTML
// - g-outer-text: replaces the entire element with text
// - g-define-slot: marks an element as a slot for child content

// contentResult holds the result of checking content replacement attributes.
type contentResult struct {
	Value string // The attribute value (may contain model paths or templates)
	Found bool   // Whether the attribute was present
}

// getInnerText checks for the g-inner-text attribute.
// Returns the value and whether it was found.
func getInnerText(attrs AttributeMap) contentResult {
	value, found := attrs[attrDataGInnerText]
	return contentResult{Value: value, Found: found}
}

// getInnerHtml checks for the g-inner-html attribute.
// Returns the value and whether it was found.
func getInnerHtml(attrs AttributeMap) contentResult {
	value, found := attrs[attrDataGInnerHtml]
	return contentResult{Value: value, Found: found}
}

// getOuterText checks for the g-outer-text attribute.
// Returns the value and whether it was found.
func getOuterText(attrs AttributeMap) contentResult {
	value, found := attrs[attrDataGOuterText]
	return contentResult{Value: value, Found: found}
}

// getDefineSlot checks for the g-define-slot attribute.
// Returns the slot name and whether the attribute was present.
// An empty slot name means the default slot.
func getDefineSlot(attrs AttributeMap) (slotName string, isSlot bool) {
	value, found := attrs[attrDataGDefineSlot]
	return value, found
}

// resolveInnerText resolves the g-inner-text value and returns the rendered text.
// Uses ctx.ResolveText to resolve any model path expressions in the attribute value.
// Always escapes HTML for XSS safety — this is the entire purpose of g-inner-text
// vs g-inner-html. The escaping is unconditional regardless of ctx.Escaped.
func resolveInnerText(ctx *RenderContext, attrs AttributeMap, data any) (text string, found bool, err error) {
	result := getInnerText(attrs)
	if !result.Found {
		return "", false, nil
	}

	// Always escape — g-inner-text is the safe text injection attribute
	rendered, err := ctx.ResolveText(result.Value, data, true)
	if err != nil {
		return "", true, err
	}
	return rendered, true, nil
}

// resolveInnerHtml resolves the g-inner-html value and returns the rendered HTML.
// Uses ctx.ResolveText to resolve any model path expressions in the attribute value.
// The HTML is NOT escaped (that's the purpose of inner-html vs inner-text).
func resolveInnerHtml(ctx *RenderContext, attrs AttributeMap, data any) (html string, found bool, err error) {
	result := getInnerHtml(attrs)
	if !result.Found {
		return "", false, nil
	}

	// Use ctx.ResolveText without escaping for raw HTML content
	rendered, err := ctx.ResolveText(result.Value, data, false)
	if err != nil {
		return "", true, err
	}
	return rendered, true, nil
}

// resolveOuterText resolves the g-outer-text value and returns the rendered text.
// Uses ctx.ResolveText to resolve any model path expressions in the attribute value.
// Escaping follows ctx.Escaped (the global setting), since g-outer-text replaces
// the entire element and is commonly used for raw text extraction (e.g., template
// references rendered to strings).
func resolveOuterText(ctx *RenderContext, attrs AttributeMap, data any) (text string, found bool, err error) {
	result := getOuterText(attrs)
	if !result.Found {
		return "", false, nil
	}

	rendered, err := ctx.ResolveText(result.Value, data, ctx.Escaped)
	if err != nil {
		return "", true, err
	}
	return rendered, true, nil
}
