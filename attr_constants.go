package gotmx

// attr_constants.go contains all attribute name constants used by gotmx.
// These are organized by category for easier navigation.
//
// All attribute names are in canonical "data-g-*" form. Short-form "g-*" attributes
// are normalized to "data-g-*" at parse time in getAttributes(), so rendering code
// only needs to check the canonical form.

// Canonical attribute names (data-g-*)
const (
	// Template definition
	attrDataGDefine     = "data-g-define"
	attrDataGDefineSlot = "data-g-define-slot"

	// Control flow
	attrDataGIgnore = "data-g-ignore"
	attrDataGIf     = "data-g-if"
	attrDataGWith   = "data-g-with"

	// Iteration
	attrDataGOuterRepeat = "data-g-outer-repeat"
	attrDataGInnerRepeat = "data-g-inner-repeat"

	// Content
	attrDataGInnerText = "data-g-inner-text"
	attrDataGInnerHtml = "data-g-inner-html"
	attrDataGOuterText = "data-g-outer-text"

	// Composition
	attrDataGUse              = "data-g-use"
	attrDataGInnerUse         = "data-g-inner-use"
	attrDataGUseSlot          = "data-g-use-slot"
	attrDataGAsTemplate       = "data-g-as-template"
	attrDataGAsUnsafeTemplate = "data-g-as-unsafe-template"
	attrDataGOverrideAtt      = "data-g-override-att"

	// Transformation
	attrDataGTrans = "data-g-trans"
	attrDataGAtt   = "data-g-att-"
	attrDataGAttIf = "data-g-attif-"

	// Shortcuts
	attrDataGClass = "data-g-class"
	attrDataGHref  = "data-g-href"
	attrDataGSrc   = "data-g-src"
)

// Short form attribute names (g-*) — only kept for raw HTML node processing
// in node_template.go processAttributes() and processGAtt(), which operate
// on raw html.Node attributes before normalization.
const (
	attrGDefine           = "g-define"
	attrGAsTemplate       = "g-as-template"
	attrGAsUnsafeTemplate = "g-as-unsafe-template"
	attrGAtt              = "g-att-"
)

// g-ignore attribute values
const (
	attrGIgnoreOuter     = "outer"
	attrGIgnoreInner     = "inner"
	attrGIgnoreOuterOnly = "outer-only"
	attrGIgnoreNone      = "none"
)

// ignoredRenderAttributes defines attributes that should never be rendered to output.
// These are gotmx control attributes that are processed during rendering.
// Only canonical (data-g-*) forms are needed since attributes are normalized at parse time.
var ignoredRenderAttributes = map[string]bool{
	attrDataGAsTemplate:       true,
	attrDataGAsUnsafeTemplate: true,
	attrDataGClass:            true,
	attrDataGDefine:           true,
	attrDataGDefineSlot:       true,
	attrDataGHref:             true,
	attrDataGIf:               true,
	attrDataGIgnore:           true,
	attrDataGInnerText:        true,
	attrDataGInnerHtml:        true,
	attrDataGInnerRepeat:      true,
	attrDataGInnerUse:         true,
	attrDataGOuterRepeat:      true,
	attrDataGOuterText:        true,
	attrDataGOverrideAtt:      true,
	attrDataGSrc:              true,
	attrDataGTrans:            true,
	attrDataGUse:              true,
	attrDataGUseSlot:          true,
	attrDataGWith:             true,
}
