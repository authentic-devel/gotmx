package gotmx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// bufferPool reuses bytes.Buffer instances to reduce GC pressure from RenderString
// and renderToStringInternal calls. Buffers are reset before reuse.
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// RenderOption configures individual render calls.
type RenderOption func(*renderConfig)

// renderConfig holds configuration for a single render call.
type renderConfig struct {
	slots      SlottedRenderables
	escaped    bool
	layout     string
	layoutData any
	layoutSlot string // slot name for the rendered content in the layout (default: "")
}

// Render writes the rendered template to the given writer.
// Uses HTML escaping by default. Returns an error if template not found.
// The ctx parameter is Go's standard context.Context for request cancellation,
// timeouts, and passing request-scoped values. HTTP handlers should pass r.Context().
func (e *Engine) Render(ctx context.Context, w io.Writer, templateName string, data any, opts ...RenderOption) error {
	cfg := &renderConfig{escaped: true}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.layout != "" {
		return e.renderWithLayout(ctx, w, cfg, templateName, data)
	}

	return e.renderInternal(ctx, w, TemplateRef(templateName), data, cfg.slots, cfg.escaped)
}

// renderWithLayout renders a template wrapped inside a layout template.
// The rendered content is injected into the layout's slot (default slot by default).
func (e *Engine) renderWithLayout(ctx context.Context, w io.Writer, cfg *renderConfig, templateName string, data any) error {
	// Create the inner content component
	inner, err := e.createComponent(TemplateRef(templateName), data)
	if err != nil {
		return err
	}

	// Apply any slots to the inner component
	if hasChildren, ok := inner.(HasChildren); ok && len(cfg.slots) > 0 {
		hasChildren.AddChildren(cfg.slots)
	}

	// Create the layout component
	layoutSlots := SlottedRenderables{
		cfg.layoutSlot: {inner},
	}
	return e.renderInternal(ctx, w, TemplateRef(cfg.layout), cfg.layoutData, layoutSlots, cfg.escaped)
}

// RenderString renders a template to a string.
// Convenience wrapper around Render() using a pooled bytes.Buffer.
// The ctx parameter is Go's standard context.Context for request cancellation,
// timeouts, and passing request-scoped values.
func (e *Engine) RenderString(ctx context.Context, templateName string, data any, opts ...RenderOption) (string, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := e.Render(ctx, buf, templateName, data, opts...); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Unescaped disables HTML escaping for the render call.
// Use only when rendering trusted content. By default, all text is HTML-escaped.
//
// Example:
//
//	engine.Render(ctx, w, "template", trustedData, gotmx.Unescaped())
func Unescaped() RenderOption {
	return func(cfg *renderConfig) {
		cfg.escaped = false
	}
}

// WithLayout wraps the rendered template inside a layout template.
// The rendered content is placed into the layout's default slot (empty name).
// Use WithLayoutSlot to target a named slot instead.
//
// Example:
//
//	engine.Render(ctx, w, "dashboard-page", pageData,
//	    gotmx.WithLayout("main-layout", layoutData),
//	)
func WithLayout(layoutTemplate string, layoutData any) RenderOption {
	return func(cfg *renderConfig) {
		cfg.layout = layoutTemplate
		cfg.layoutData = layoutData
	}
}

// WithLayoutSlot sets which slot in the layout receives the rendered content.
// By default, content goes to the default slot (empty name).
// Must be used together with WithLayout.
//
// Example:
//
//	engine.Render(ctx, w, "dashboard-page", pageData,
//	    gotmx.WithLayout("main-layout", layoutData),
//	    gotmx.WithLayoutSlot("content"),
//	)
func WithLayoutSlot(slotName string) RenderOption {
	return func(cfg *renderConfig) {
		cfg.layoutSlot = slotName
	}
}

// HasTemplate checks if a template exists without triggering lazy loading.
// Useful for conditional rendering logic.
func (e *Engine) HasTemplate(templateName string) bool {
	_, err := e.registry.GetTemplate(TemplateRef(templateName))
	return err == nil
}

// Component returns a Renderable for the given template.
// Useful when you need to pass a component as slot content.
func (e *Engine) Component(templateName string, data any) (Renderable, error) {
	return e.createComponent(TemplateRef(templateName), data)
}

// MustComponent is like Component but panics on error.
// Use only for templates known to exist at development time.
func (e *Engine) MustComponent(templateName string, data any) Renderable {
	r, err := e.Component(templateName, data)
	if err != nil {
		panic(fmt.Sprintf("gotmx: %s: %v", templateName, err))
	}
	return r
}

// ComponentWithSlots creates a Renderable and populates its slots.
// This is a convenience method combining Component() and AddChildren().
func (e *Engine) ComponentWithSlots(templateName string, data any, slots map[string][]Renderable) (Renderable, error) {
	component, err := e.Component(templateName, data)
	if err != nil {
		return nil, err
	}
	if hasChildren, ok := component.(HasChildren); ok && len(slots) > 0 {
		hasChildren.AddChildren(slots)
	}
	return component, nil
}
