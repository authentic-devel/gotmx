// Package gotmx is a component-based HTML template engine for Go.
//
// It enables server-side rendering with templates that remain valid HTML.
// Templates use standard HTML attributes (g-* or data-g-*) instead of custom syntax,
// making them viewable and editable in any HTML editor and previewable in browsers.
//
// # Getting Started
//
// Use the Engine API to create and render templates:
//
//	engine, err := gotmx.New(gotmx.WithTemplateDir("templates"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer engine.Close()
//
//	// Render a template
//	engine.Render(ctx, w, "my-template", data)
//
// # Advanced Customization
//
// For advanced use cases, Engine supports custom registries and resolvers:
//
//	engine, err := gotmx.New(
//	    gotmx.WithCustomRegistry(myRegistry),
//	    gotmx.WithCustomResolver(myResolver),
//	)
//
// # Template Attributes
//
// Templates use special attributes to control rendering:
//
//   - g-if: Conditional rendering
//   - g-outer-repeat, g-inner-repeat: Iteration over collections
//   - g-inner-text, g-inner-html: Content replacement
//   - g-use: Component composition
//   - g-define-slot, g-use-slot: Slot-based content injection
//
// For more information on available attributes, see the attribute reference
// in the docs/ directory.
package gotmx

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"golang.org/x/net/html"
)

// GoTemplateRegistry is an additional interface that template registries can implement to handle
// native Go templates. Go templates may reference each other, and in such cases they need to be in the same
// textTemplate.Template or htmlTemplate.Template instance. Therefore, they need to be treated separately
// from regular gotmx templates.
//
// Note: Golang templates require you to register any functions BEFORE the templates are actually parsed.
type GoTemplateRegistry interface {
	// RegisterGoTemplate adds a native Go template to the registry.
	// Returns an error if the template cannot be registered.
	RegisterGoTemplate(templateName TemplateName, template string, sourceFile string) error

	// RegisterFunc registers a function that can be used within Go templates.
	// The function must be registered before any templates that use it are parsed.
	RegisterFunc(name string, fun interface{})
}

// clearTemplates removes all templates from the Engine's registry.
func (e *Engine) clearTemplates() {
	e.logger.Debug("Clearing all templates")
	e.registry.ClearTemplates()
}

// NewRenderContext creates a new RenderContext that wraps this Engine's capabilities.
// This context is created once per render call and passed through the entire rendering tree,
// avoiding per-node allocations while decoupling renderables from the Engine type.
//
// The ctx parameter is Go's standard context.Context for request cancellation, timeouts,
// and passing request-scoped values. HTTP handlers should pass r.Context() to enable
// proper request lifecycle handling.
//
// Use this method when you need to call Renderable.Render() directly on a component
// obtained via Component(). For normal template rendering, use Render() or
// RenderString() which create the context automatically.
//
// Example:
//
//	component, _ := engine.Component("my-template", data)
//	renderCtx := engine.NewRenderContext(r.Context())
//	component.Render(renderCtx, writer, gotmx.RenderOuter)
func (e *Engine) NewRenderContext(ctx context.Context) *RenderContext {
	renderCtx := &RenderContext{
		// Context is Go's standard context for cancellation and request-scoped values.
		Context: ctx,

		// ResolveText delegates to getEffectiveText which handles model path resolution
		// and optional HTML escaping.
		ResolveText: e.getEffectiveText,

		// ResolveValue delegates to getEffectiveValue for raw value resolution.
		ResolveValue: e.getEffectiveValue,

		// Logger is the same logger configured on this Engine instance.
		Logger: e.logger,

		// Initialize nesting depth tracking
		CurrentNestingDepth: 0,
		MaxNestingDepth:     e.maxNestingDepth,

		// DeterministicOutput controls attribute sorting
		DeterministicOutput: e.deterministicOutput,
	}

	// Wrap createComponent to return depth-tracking renderables.
	// This ensures that nested template rendering (via g-use) is tracked and limited.
	renderCtx.CreateRenderable = func(name TemplateRef, data any) (Renderable, error) {
		component, err := e.createComponent(name, data)
		if err != nil {
			return nil, err
		}
		// Wrap the component in a depth-tracking renderable
		return &depthTrackingRenderable{
			inner:        component,
			templateName: string(name),
		}, nil
	}

	return renderCtx
}

// depthTrackingRenderable wraps a Renderable to track and limit nesting depth.
// It increments the depth before rendering and decrements after, checking against
// the max depth limit to prevent stack overflow from circular template references.
type depthTrackingRenderable struct {
	inner        Renderable
	templateName string
}

func (d *depthTrackingRenderable) Render(ctx *RenderContext, writer io.Writer, renderType RenderType) error {
	// Check depth limit before rendering (0 means no limit)
	if ctx.MaxNestingDepth > 0 && ctx.CurrentNestingDepth >= ctx.MaxNestingDepth {
		return &MaxNestingDepthExceededError{
			TemplateName: d.templateName,
			CurrentDepth: ctx.CurrentNestingDepth,
			MaxDepth:     ctx.MaxNestingDepth,
		}
	}

	// Track current template for error diagnostics
	prev := ctx.CurrentTemplate
	ctx.CurrentTemplate = d.templateName
	defer func() { ctx.CurrentTemplate = prev }()

	// Increment depth before rendering
	ctx.CurrentNestingDepth++
	defer func() { ctx.CurrentNestingDepth-- }()

	return d.inner.Render(ctx, writer, renderType)
}

// Implement HasChildren interface by delegating to inner
func (d *depthTrackingRenderable) AddChild(slot string, child Renderable) {
	if hasChildren, ok := d.inner.(HasChildren); ok {
		hasChildren.AddChild(slot, child)
	}
}

func (d *depthTrackingRenderable) AddChildren(slottedChildren SlottedRenderables) {
	if hasChildren, ok := d.inner.(HasChildren); ok {
		hasChildren.AddChildren(slottedChildren)
	}
}

// Implement HasAttributes interface by delegating to inner
func (d *depthTrackingRenderable) SetAttributes(attributes AttributeMap) {
	if hasAttrs, ok := d.inner.(HasAttributes); ok {
		hasAttrs.SetAttributes(attributes)
	}
}

// renderInternal writes the rendered template to the provided io.Writer.
// This is the core render method used by Engine.Render and other render methods.
//
// Parameters:
//   - ctx: Go's standard context.Context for request cancellation, timeouts, and request-scoped values
//   - writer: The writer where the rendered output will be written
//   - templateName: The reference to the template to render
//   - data: The data model to use for rendering
//   - slottedChildren: Optional child components to be slotted into the template
//   - escaped: Whether HTML special characters should be escaped in the output
func (e *Engine) renderInternal(ctx context.Context, writer io.Writer, templateName TemplateRef, data any, slottedChildren SlottedRenderables, escaped bool) error {
	template, err := e.registry.GetTemplate(templateName)
	if err != nil {
		e.logger.Error("Template does not exist", "templateName", templateName, "error", err)
		return &TemplateRetrievalError{TemplateName: templateName, Cause: err}
	}

	renderable, err := template.NewRenderable(data)
	if err != nil {
		return err
	}

	if hasChildren, ok := renderable.(HasChildren); ok {
		hasChildren.AddChildren(slottedChildren)
	}

	// Wrap writer in bufio.Writer to batch the many small io.WriteString calls
	// (tag opens, attribute names/values, closing tags) into efficient 4KB chunks.
	bw := bufio.NewWriterSize(writer, 4096)

	// Create a single render context for the entire rendering tree.
	// This context is passed by pointer to all child renderables, avoiding per-node allocations.
	renderCtx := e.NewRenderContext(ctx)
	renderCtx.Escaped = escaped
	renderCtx.CurrentTemplate = string(templateName)

	if err := renderable.Render(renderCtx, bw, RenderOuter); err != nil {
		return err
	}
	return bw.Flush()
}

// renderToStringInternal renders a template to a string.
// This is the core render-to-string method used by Engine.RenderString and other methods.
// Uses a pooled buffer to reduce GC pressure.
func (e *Engine) renderToStringInternal(ctx context.Context, templateRef TemplateRef, data any, slottedChildren SlottedRenderables, escaped bool) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	err := e.renderInternal(ctx, buf, templateRef, data, slottedChildren, escaped)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// createComponent returns a Renderable instance for the given template name.
//
// Parameters:
//   - name: The template reference, which can be either:
//   - An unqualified template name without namespace (e.g., "myTemplate")
//   - A fully qualified name with namespace (e.g., "frontend/templates/index.html#myTemplate")
//   - data: The data model to use for the component
//
// Returns:
//   - A Renderable instance that can be used to render the template
//   - An error if the template was not found or component creation failed
func (e *Engine) createComponent(name TemplateRef, data any) (Renderable, error) {
	template, err := e.registry.GetTemplate(name)
	if err != nil {
		return nil, &TemplateRetrievalError{TemplateName: name, Cause: err}
	}

	component, err := template.NewRenderable(data)
	if err != nil {
		return nil, &ComponentCreationError{TemplateName: name, Cause: err}
	}
	if component == nil {
		return nil, &NilComponentError{TemplateName: name}
	}

	return component, nil
}

// getEffectiveValue attempts to resolve a value as a model path expression.
// It delegates to the ModelPathResolver to handle the actual resolution.
func (e *Engine) getEffectiveValue(value string, data any) (any, bool) {
	return e.resolver.TryResolve(value, data)
}

// getEffectiveText resolves a value that might be a model path expression and returns it as a string.
// If the value is a model path expression, it resolves it using the model data.
// If escaped is true, HTML special characters in the result will be escaped.
func (e *Engine) getEffectiveText(value string, data any, escaped bool) (string, error) {
	modelValue, isModelPath := e.getEffectiveValue(value, data)
	text := value

	if isModelPath {
		text = fmt.Sprintf("%v", modelValue)
	}

	if escaped {
		return html.EscapeString(text), nil
	}
	return text, nil
}

// getValueFromTemplate renders the given template to a string using the provided data.
// The template can be either a native Go template or a gotmx template.
//
// Note: Uses context.Background() since this is called during model path resolution
// where no request context is available.
func (e *Engine) getValueFromTemplate(templateRef TemplateRef, data any, escaped bool) (string, error) {
	_, err := e.registry.GetTemplate(templateRef)
	if err != nil {
		// Template doesn't exist - return empty string (this is used for optional template references)
		e.logger.Debug("Template reference not found, returning empty", "ref", templateRef)
		return "", nil
	}

	result, err := e.renderToStringInternal(context.Background(), templateRef, data, nil, escaped)
	if err != nil {
		return "", &RenderError{Template: string(templateRef), Cause: err}
	}
	return result, nil
}

// registerDefaultGoFunctions registers standard functions for use in Go templates.
// This method only has an effect if the provided registry implements the GoTemplateRegistry interface.
//
// The registered functions are:
//   - GTemplate: Renders a gotmx template with HTML escaping
//   - GTextTemplate: Renders a gotmx template without HTML escaping
//
// Note: These functions use context.Background() since Go template functions don't have
// access to the request context. Consider using gotmx templates directly if you need
// context propagation.
func (e *Engine) registerDefaultGoFunctions() {
	goTemplateRegistry, ok := e.registry.(GoTemplateRegistry)
	if !ok {
		return
	}

	// Register a function to render templates with HTML escaping.
	goTemplateRegistry.RegisterFunc("GTemplate", func(templateRef string, data any) (string, error) {
		result, err := e.renderToStringInternal(context.Background(), TemplateRef(templateRef), data, nil, true)
		if err != nil {
			return "", err
		}
		return result, nil
	})

	// Register a function to render templates without HTML escaping.
	goTemplateRegistry.RegisterFunc("GTextTemplate", func(templateRef string, data any) (string, error) {
		result, err := e.renderToStringInternal(context.Background(), TemplateRef(templateRef), data, nil, false)
		if err != nil {
			return "", err
		}
		return result, nil
	})
}
