package gotmx

import (
	"context"
	"io"
)

// TemplateName is the simple name of a template within its namespace.
// Examples: "my-button", "dashboard-page"
type TemplateName string

// TemplateRef is a reference to a template, either simple or fully qualified with namespace.
// Simple: "my-button"
// Qualified: "components/button.html#my-button"
type TemplateRef string

// Namespace identifies the source of a template, typically the file path it was loaded from.
// Example: "components/button.html"
type Namespace string

func (n TemplateName) String() string {
	return string(n)
}

func (n TemplateRef) String() string {
	return string(n)
}

func (n Namespace) String() string {
	return string(n)
}

// RenderContext provides the rendering capabilities needed during template rendering.
// A single context is created when rendering begins (via Engine.Render or Engine.RenderString)
// and passed by pointer through the entire rendering tree. This design:
//   - Avoids per-node allocations (only one context object is created per render call)
//   - Decouples Renderable implementations from Engine, breaking bi-directional coupling
//   - Makes testing easier by allowing mock implementations of individual functions
//
// Implementations of Renderable should NOT create new RenderContext instances; they should
// pass the same context pointer to any child renderables they create or invoke.
type RenderContext struct {
	// Context is Go's standard context.Context for request cancellation, timeouts,
	// and passing request-scoped values through the rendering pipeline.
	// HTTP handlers should pass r.Context() to enable proper request lifecycle handling.
	Context context.Context

	// ResolveText resolves a value (which may contain model path expressions like "[[ .Name ]]")
	// to a string. The value is evaluated against the provided data context.
	// If escaped is true, HTML special characters in the result are escaped for XSS protection.
	//
	// Example: ResolveText("[[ .User.Name ]]", userData, true) might return "John &amp; Jane"
	ResolveText func(value string, data any, escaped bool) (string, error)

	// ResolveValue resolves a value as a model path expression and returns the raw result.
	// Returns (resolved value, true) if the value was a model path expression that was resolved,
	// or (original value, false) if the value was not a model path expression.
	//
	// This is useful when you need the actual typed value rather than a string representation,
	// for example when resolving iteration targets or conditional expressions.
	//
	// Example: ResolveValue("[[ .Items ]]", data) might return ([]Item{...}, true)
	ResolveValue func(value string, data any) (any, bool)

	// CreateRenderable creates a new Renderable for the given template reference.
	// The template reference can be either a simple name (e.g., "myTemplate") or a
	// fully qualified name with namespace (e.g., "templates/page.html#myTemplate").
	//
	// Returns an error if the template is not found or if component creation fails.
	CreateRenderable func(name TemplateRef, data any) (Renderable, error)

	// Logger provides logging capabilities during rendering for debug, info, and error messages.
	// This is the same logger configured on the Engine instance.
	Logger Logger

	// CurrentNestingDepth tracks the current depth of nested template rendering.
	// This is incremented when entering a nested template (via g-use) and decremented when exiting.
	// Used to prevent stack overflow from circular template references.
	CurrentNestingDepth int

	// MaxNestingDepth is the maximum allowed nesting depth for template rendering.
	// If CurrentNestingDepth reaches this limit, rendering will fail with MaxNestingDepthExceededError.
	// A value of 0 means no limit (not recommended). Default is 64.
	MaxNestingDepth int

	// DeterministicOutput controls whether HTML attributes are sorted alphabetically.
	// When true, attributes are sorted for predictable output (useful for testing).
	// When false (default), attributes render in map iteration order (faster).
	DeterministicOutput bool

	// Escaped controls whether text content is HTML-escaped by default.
	// This is set once at the start of rendering based on the render options.
	// Individual attributes may override this — for example, g-inner-text always
	// escapes regardless of this flag, while g-inner-html never escapes.
	Escaped bool

	// CurrentTemplate tracks the template being rendered, for error diagnostics.
	// Updated when entering a new template via g-use or g-inner-use.
	CurrentTemplate string
}

// Renderable represents something that can be rendered to HTML output.
// Built-in implementations handle HTML nodes, text content, Go templates, and literal strings.
//
// The RenderContext parameter is created once at the start of rendering and passed through
// the entire rendering tree. Implementations should pass this same context to any child
// renderables they invoke, without creating new context instances.
//
// Escaping behavior is controlled by ctx.Escaped rather than a method parameter.
// Specific attributes enforce their own escaping rules: g-inner-text always escapes,
// g-inner-html never escapes, and attribute values always escape.
type Renderable interface {
	// Render outputs the HTML representation to the provided writer.
	//
	// Parameters:
	//   - ctx: The render context providing resolution and component creation functions.
	//          This is created once per render call and should be passed unchanged to children.
	//          Escaping is controlled by ctx.Escaped.
	//   - writer: Destination for the rendered HTML output.
	//   - renderType: Controls whether to render outer tags (RenderOuter) or content only (RenderInner).
	//
	// Returns an error if rendering fails at any point.
	Render(ctx *RenderContext, writer io.Writer, renderType RenderType) error
}

// Template is a blueprint for creating Renderable instances.
// Each Template has a unique name within its namespace and can create
// Renderable instances with the provided data.
type Template interface {

	// Name returns the unique name of the template within its namespace.
	// This must always return a constant value. Unpredictable behavior will occur if it returns
	// a different value after the template has been registered.
	Name() TemplateName

	// Namespace returns the namespace of the template. Within a namespace, the name must be unique.
	// Typically, the namespace is the file path the template came from, e.g., an HTML file.
	// The fully qualified name of a template is the namespace and the name separated by a hash,
	// for example "frontend/index.html#myTemplate".
	// This must always return a constant value. Unpredictable behavior will occur if it returns
	// a different value after the template has been registered.
	Namespace() Namespace

	// NewRenderable creates a new Renderable instance from this template with the given data.
	// The data parameter provides the model context for rendering.
	NewRenderable(data any) (Renderable, error)
}

// HasChildren is an interface for a Renderable that can have children (in slots).
type HasChildren interface {
	AddChild(slot string, child Renderable)
	AddChildren(slottedChildren SlottedRenderables)
}

// HasAttributes is an interface for a Renderable that can have attributes.
type HasAttributes interface {
	SetAttributes(attributes AttributeMap)
}

// SlottedRenderables maps slot names to their renderable content.
type SlottedRenderables map[string][]Renderable

// RenderType controls whether a Renderable outputs its outer HTML tags or only its content.
type RenderType int

const (
	// RenderInner renders only the inner content of the element, without the opening/closing tags.
	// Used by g-inner-use to embed a template's content without its root element.
	RenderInner RenderType = iota

	// RenderOuter renders the complete element including its opening and closing tags.
	// This is the default render type used for most rendering operations.
	RenderOuter
)
