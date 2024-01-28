package gotmx

import (
	"errors"
	"io"
	"maps"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// nodeComponent wraps an HTML node for rendering with gotmx directives.
// Implements Renderable, HasChildren, and HasAttributes interfaces.
// nodeComponent wraps an html.Node and serves as the primary unit of rendering. It enables
// component-based templating by supporting slot-based composition (parent components can
// inject children into named slots) and attribute overrides (parent can pass attributes
// to child components).
//
// The rendering pipeline processes special g-* attributes in a defined order to ensure
// predictable behavior when multiple directives are combined on a single element.
// See handleElementNode for the processing order.
type nodeComponent struct {
	node               *html.Node
	data               any                // Template data passed to this component and its children
	slottedChildren    SlottedRenderables // Children organized by slot name; empty string is the default slot
	attributeOverrides AttributeMap       // Attributes passed from parent component via g-override-att
}

// newNodeComponent creates a nodeComponent from an HTML node.
// The node parameter is the parsed HTML element to wrap.
// The data parameter is the template context available during rendering.
func newNodeComponent(node *html.Node, data any) *nodeComponent {
	return &nodeComponent{
		node:               node,
		data:               data,
		slottedChildren:    SlottedRenderables{},
		attributeOverrides: AttributeMap{},
	}
}

// SetAttributes stores attribute overrides passed from a parent component.
// These take precedence over the node's original attributes during rendering.
func (nc *nodeComponent) SetAttributes(attributes AttributeMap) {
	nc.attributeOverrides = attributes
}

// Render outputs the HTML representation of this component.
//
// Parameters:
//   - ctx: The render context providing resolution and component creation functions.
//     This context is created once per render call and should be passed unchanged to children.
//   - writer: Destination for rendered HTML output.
//   - renderType: Controls whether to render outer tags (RenderOuter) or content only (RenderInner).
//   - escaped: When true, text content is HTML-escaped for XSS protection.
//
// When a nodeComponent is rendered directly (not as a child), g-if and g-ignore attributes
// on its root element are overridden to ensure it always renders. This prevents confusing
// behavior where a component call would silently produce no output due to conditional
// attributes on the template's root element.
func (nc *nodeComponent) Render(ctx *RenderContext, writer io.Writer, renderType RenderType) error {

	// Optimize allocation: only copy when we have overrides to copy
	var attributeMap AttributeMap
	if len(nc.attributeOverrides) > 0 {
		// Pre-allocate with expected size: overrides + 2 forced attributes
		attributeMap = make(AttributeMap, len(nc.attributeOverrides)+2)
		for attributeName, attributeValue := range nc.attributeOverrides {
			attributeMap[attributeName] = attributeValue
		}
	} else {
		// Small allocation for the 2 forced attributes only
		attributeMap = make(AttributeMap, 2)
	}

	// Force rendering by disabling conditionals on the root element. Conditionals should
	// be evaluated at the call site (g-use), not inside the component template.
	// Only canonical data-g-* form needed since attributes are normalized at parse time.
	attributeMap[attrDataGIgnore] = attrGIgnoreNone
	attributeMap[attrDataGIf] = "true"

	return nc.render(ctx, writer, nc.node, renderType, nc.slottedChildren, attributeMap, false, nc.data)
}

// AddChild adds a renderable to the specified slot.
// Use empty string for the default slot. Multiple children can be added to the same slot.
func (nc *nodeComponent) AddChild(slot string, renderable Renderable) {
	if nc.slottedChildren == nil {
		nc.slottedChildren = SlottedRenderables{}
	}
	if nc.slottedChildren[slot] == nil {
		nc.slottedChildren[slot] = []Renderable{renderable}
	} else {
		nc.slottedChildren[slot] = append(nc.slottedChildren[slot], renderable)
	}
}

// AddChildren merges multiple slotted children into this component.
func (nc *nodeComponent) AddChildren(slottedChildren SlottedRenderables) {
	for slot, children := range slottedChildren {
		for _, child := range children {
			nc.AddChild(slot, child)
		}
	}
}

// render dispatches to the appropriate renderer based on node type.
// The ctx parameter is the render context created once at the start of rendering.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) render(ctx *RenderContext, writer io.Writer, node *html.Node, renderType RenderType,
	slottedNodes SlottedRenderables, overriddenAttributes AttributeMap, isTemplate bool, data any) error {

	switch node.Type {
	case html.TextNode:
		return nc.renderTextNode(ctx, writer, node, isTemplate, data)
	case html.DocumentNode:
		return nc.renderDocumentNode(ctx, writer, node, slottedNodes, isTemplate, data)
	case html.ElementNode:
		return nc.handleElementNode(ctx, writer, node, renderType, slottedNodes, overriddenAttributes,
			isTemplate, data)
	case html.CommentNode:
		// Comments are not rendered
		return nil
	default:
		return errors.New("html: unknown node type")
	}
}

// renderUse renders a referenced template instead of the current element.
// The ctx parameter is passed through to the created component for rendering.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderUse(ctx *RenderContext, writer io.Writer, templateRef string, renderType RenderType,
	childrenOfNode *html.Node, data any) error {

	// Check for context cancellation before creating a new component
	if err := ctx.Context.Err(); err != nil {
		return err
	}

	// Resolve the template reference (may contain model paths or templates)
	resolvedRef, err := resolveTemplateRef(ctx, templateRef, data)
	if err != nil {
		return err
	}

	// Use the context's CreateRenderable function to create the target component
	targetComp, err := ctx.CreateRenderable(TemplateRef(resolvedRef), data)
	if err != nil {
		return &ComponentNotFoundError{
			ComponentRef: resolvedRef,
			Cause:        err,
		}
	}

	nodeAttributeMap := nc.getAttributes(childrenOfNode)

	// If the target component supports children, process them
	if hasChildren, ok := targetComp.(HasChildren); ok {
		// Check if g-inner-text should be used for the default slot
		if innerText := getInnerText(nodeAttributeMap); innerText.Found {
			textRenderable := &textComponent{innerText.Value}
			hasChildren.AddChild("", textRenderable)
		} else {
			// Process each child node
			for child := childrenOfNode.FirstChild; child != nil; child = child.NextSibling {
				// Skip whitespace-only text nodes
				if child.Type == html.TextNode && strings.TrimSpace(child.Data) == "" {
					continue
				}

				childAttributeMap := nc.getAttributes(child)
				slotName, usesSlot := getUseSlot(childAttributeMap)
				if usesSlot {
					slotName, _ = resolveSlotName(ctx, slotName, data)
				}

				childRenderable := newNodeComponent(child, data)
				hasChildren.AddChild(slotName, childRenderable)
			}
		}
	}

	// If the target component supports attributes, process overrides
	if hasAttributes, ok := targetComp.(HasAttributes); ok {
		if attrList, found := getOverrideAtt(nodeAttributeMap); found {
			attrNames := parseOverrideAttributes(attrList)
			overrides := collectOverrideAttributes(nodeAttributeMap, attrNames)
			hasAttributes.SetAttributes(overrides)
		}
	}

	// Pass the same context through to the child component
	return targetComp.Render(ctx, writer, renderType)
}

// renderDocumentNode renders a document node and its children.
// The ctx parameter is passed through unchanged to child renders.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderDocumentNode(ctx *RenderContext, writer io.Writer, n *html.Node,
	slottedNodes SlottedRenderables, isTemplate bool, data any) error {

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		err := nc.render(ctx, writer, child, RenderOuter, slottedNodes, AttributeMap{}, isTemplate, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// renderTextNode renders a text node, optionally treating it as a template.
// Uses ctx.ResolveText to resolve model path expressions in the text content.
// Escaping is controlled by ctx.Escaped.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderTextNode(ctx *RenderContext, writer io.Writer, n *html.Node,
	asTemplate bool, data any) error {

	if asTemplate {
		// Use the context's ResolveText function to resolve model paths and apply escaping
		effectiveText, err := ctx.ResolveText(n.Data, data, ctx.Escaped)
		if err != nil {
			return err
		}
		_, err = io.WriteString(writer, effectiveText)
		return err
	}

	return renderText(writer, n.Data, ctx.Escaped)
}

func renderText(writer io.Writer, text string, escaped bool) error {
	if escaped {
		_, err := io.WriteString(writer, html.EscapeString(text))
		return err
	}
	_, err := io.WriteString(writer, text)
	return err
}

// handleElementNode is the entry point for element node processing.
// It processes control flow attributes in order:
//  1. g-ignore (early exit)
//  2. g-with (context switch)
//  3. g-if (conditional)
//  4. g-attif-* (attribute manipulation)
//  5. g-outer-repeat (iteration) or render
//
// The ctx parameter provides resolution functions and is passed through to child renders.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) handleElementNode(ctx *RenderContext, writer io.Writer, n *html.Node,
	renderType RenderType, slottedNodes SlottedRenderables,
	overriddenAttributes AttributeMap, isTemplate bool, data any) error {

	if n.Data == "plaintext" {
		return nil
	}

	// Collect and merge attributes
	attrs := nc.getAttributes(n)
	if len(overriddenAttributes) > 0 {
		if attrs == nil {
			attrs = make(AttributeMap, len(overriddenAttributes))
		}
		applyOverrides(attrs, overriddenAttributes)
	}

	// Phase 1: g-ignore (early exit)
	ignoreResult, err := checkIgnore(ctx, attrs, data)
	if err != nil {
		return wrapRenderError(err, n.Data, "g-ignore")
	}
	if shouldSkipElement(ignoreResult) {
		return nil
	}

	// Phase 2: g-with (context switch)
	data = applyWith(ctx, attrs, data)

	// Phase 3: g-if (conditional)
	shouldRender, err := checkIf(ctx, attrs, data)
	if err != nil {
		return wrapRenderError(err, n.Data, "g-if")
	}
	if !shouldRender {
		return nil
	}

	// Phase 4: g-attif-* (attribute manipulation)
	processGAttIf(ctx, attrs, data)

	// Phase 5: g-outer-repeat or render
	if repeatPath, found := getOuterRepeat(attrs); found {
		if err := nc.renderOuterRepeat(ctx, writer, n, attrs, slottedNodes, repeatPath, ignoreResult, isTemplate, data); err != nil {
			return wrapRenderError(err, n.Data, "g-outer-repeat")
		}
		return nil
	}

	return nc.renderElement(ctx, writer, n, renderType, slottedNodes, attrs, ignoreResult, isTemplate, data)
}

// renderElement renders an element node with all its attributes and content.
// The ctx parameter provides resolution functions and is passed through to child renders.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderElement(ctx *RenderContext, writer io.Writer, n *html.Node,
	renderType RenderType, slottedNodes SlottedRenderables, attrs AttributeMap,
	ignoreResult ignoreResult, isTemplate bool, data any) error {

	// Check for g-use (render another template instead)
	if templateRef, found := getUse(attrs); found {
		if err := nc.renderUse(ctx, writer, templateRef, RenderOuter, n, data); err != nil {
			return wrapRenderError(err, n.Data, "g-use")
		}
		return nil
	}

	// Check for g-inner-use
	if templateRef, found := getInnerUse(attrs); found {
		if err := nc.renderUse(ctx, writer, templateRef, RenderInner, n, data); err != nil {
			return wrapRenderError(err, n.Data, "g-inner-use")
		}
		return nil
	}

	// Render DOCTYPE before <html> element
	if n.Type == html.ElementNode && n.Data == "html" {
		if _, err := writer.Write([]byte("<!DOCTYPE html>\n")); err != nil {
			return err
		}
	}

	// Check for g-outer-text (replace entire element with text)
	if outerText, found, err := resolveOuterText(ctx, attrs, data); found {
		if err != nil {
			return wrapRenderError(err, n.Data, "g-outer-text")
		}
		// If ignore is set (but not "none"), skip output
		if ignoreResult.HasIgnore && ignoreResult.Mode != attrGIgnoreNone {
			return nil
		}
		_, err = writer.Write([]byte(outerText))
		return err
	}

	// Resolve tag name (may be transformed by g-trans)
	tagName, err := resolveTrans(ctx, attrs, n.Data, data)
	if err != nil {
		return wrapRenderError(err, n.Data, "g-trans")
	}

	// Render opening tag (unless outer-only ignore or inner-only render)
	skipOuter := shouldSkipOuterOnly(ignoreResult) || renderType == RenderInner
	if !skipOuter {
		isVoid, err := nc.renderOpeningTag(ctx, writer, tagName, attrs, data)
		if err != nil {
			return wrapRenderError(err, n.Data, "")
		}
		if isVoid {
			return nil
		}
	}

	// Render content (unless inner ignore)
	if !shouldSkipChildren(ignoreResult) {
		if repeatPath, found := getInnerRepeat(attrs); found {
			if err := nc.renderInnerRepeat(ctx, writer, n, slottedNodes, repeatPath, attrs, isTemplate, data); err != nil {
				return wrapRenderError(err, n.Data, "g-inner-repeat")
			}
		} else {
			if err := nc.renderElementContent(ctx, writer, n, slottedNodes, attrs, isTemplate, data); err != nil {
				return wrapRenderError(err, n.Data, "")
			}
		}
	}

	// Render closing tag
	if !skipOuter {
		return nc.renderClosingTag(writer, tagName)
	}
	return nil
}

// renderInnerRepeat renders children multiple times, once per item in the collection.
// The ctx parameter is passed through to child renders.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderInnerRepeat(ctx *RenderContext, writer io.Writer, n *html.Node,
	slottedNodes SlottedRenderables, repeatPath string, attrs AttributeMap,
	isTemplate bool, data any) error {

	modelValue, ok := resolveIterable(ctx, repeatPath, data)
	if !ok || modelValue == nil {
		return nil
	}

	return iterateValue(ctx, modelValue, func(item any) error {
		return nc.renderChildren(ctx, writer, n, attrs, slottedNodes, isTemplate, item)
	})
}

// renderOpeningTag renders the opening tag with attributes.
// Returns true if this is a void element (no closing tag needed).
// Uses ctx.ResolveText for attribute value resolution.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderOpeningTag(ctx *RenderContext, writer io.Writer, tagName string,
	attrs AttributeMap, data any) (bool, error) {

	if _, err := io.WriteString(writer, "<"); err != nil {
		return false, err
	}
	if _, err := io.WriteString(writer, tagName); err != nil {
		return false, err
	}
	if err := nc.renderAttributes(ctx, writer, attrs, data); err != nil {
		return false, err
	}

	if voidElements[tagName] {
		_, err := io.WriteString(writer, " />")
		return true, err
	}
	_, err := io.WriteString(writer, ">")
	return false, err
}

// renderClosingTag renders the closing tag.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderClosingTag(writer io.Writer, tagName string) error {
	if _, err := io.WriteString(writer, "</"); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, tagName); err != nil {
		return err
	}
	_, err := io.WriteString(writer, ">")
	return err
}

// renderElementContent renders the content inside an element.
// Handles g-inner-html, g-inner-text, slots, and regular children.
// The ctx parameter is passed through to child renders and slot content.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderElementContent(ctx *RenderContext, writer io.Writer, n *html.Node,
	slottedNodes SlottedRenderables, attrs AttributeMap, isTemplate bool, data any) error {

	// Check for g-inner-html (unescaped HTML content)
	if innerHtml, found, err := resolveInnerHtml(ctx, attrs, data); found {
		if err != nil {
			return err
		}
		_, err = io.WriteString(writer, innerHtml)
		return err
	}

	// Check for g-inner-text (always escapes for XSS safety, regardless of ctx.Escaped)
	if innerText, found, err := resolveInnerText(ctx, attrs, data); found {
		if err != nil {
			return err
		}
		_, err = io.WriteString(writer, innerText)
		return err
	}

	// Check for slot definition
	if slotName, isSlot := getDefineSlot(attrs); isSlot {
		if slotChildren, ok := slottedNodes[slotName]; ok && len(slotChildren) > 0 {
			for _, slottedNode := range slotChildren {
				// Pass the same context through to slotted children
				if err := slottedNode.Render(ctx, writer, RenderOuter); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Render normal children
	return nc.renderChildren(ctx, writer, n, attrs, slottedNodes, isTemplate, data)
}

// renderGoHtmlTemplate renders content as a Go HTML template (escaped).
// Uses ctx.CreateRenderable to create the component and passes ctx through for rendering.
// Temporarily sets ctx.Escaped=true for the duration of the render.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderGoHtmlTemplate(ctx *RenderContext, writer io.Writer, templateName string, data any) error {
	component, err := ctx.CreateRenderable(TemplateRef(templateName), data)
	if err != nil {
		return &ComponentNotFoundError{
			ComponentRef: templateName,
			Cause:        err,
		}
	}
	prev := ctx.Escaped
	ctx.Escaped = true
	defer func() { ctx.Escaped = prev }()
	return component.Render(ctx, writer, RenderOuter)
}

// renderGoTextTemplate renders content as a Go text template (unescaped).
// Uses ctx.CreateRenderable to create the component and passes ctx through for rendering.
// Temporarily sets ctx.Escaped=false for the duration of the render.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderGoTextTemplate(ctx *RenderContext, writer io.Writer, templateName string, data any) error {
	component, err := ctx.CreateRenderable(TemplateRef(templateName), data)
	if err != nil {
		return &ComponentNotFoundError{
			ComponentRef: templateName,
			Cause:        err,
		}
	}
	prev := ctx.Escaped
	ctx.Escaped = false
	defer func() { ctx.Escaped = prev }()
	return component.Render(ctx, writer, RenderOuter)
}

// renderChildren renders the children of an element.
// Handles g-as-template, g-as-unsafe-template, literal content, and normal children.
// The ctx parameter is passed through to all child renders.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderChildren(ctx *RenderContext, writer io.Writer, n *html.Node, attrs AttributeMap,
	slottedNodes SlottedRenderables, isTemplate bool, data any) error {

	// Check for g-as-template (render as Go HTML template)
	if templateName, found := getAsTemplate(attrs); found {
		return nc.renderGoHtmlTemplate(ctx, writer, templateName, data)
	}

	// Check for g-as-unsafe-template (render as Go text template)
	if templateName, found := getAsUnsafeTemplate(attrs); found {
		return nc.renderGoTextTemplate(ctx, writer, templateName, data)
	}

	// Handle elements with literal text children (script, style, iframe, etc.)
	// Write directly to the writer — since renderInternal wraps the writer in a
	// bufio.Writer, the many small writes are already batched efficiently.
	if childTextNodesAreLiteral(n) {
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode {
				if _, err := io.WriteString(writer, child.Data); err != nil {
					return err
				}
			} else {
				if err := nc.render(ctx, writer, child, RenderOuter, slottedNodes, AttributeMap{}, isTemplate, data); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Render normal children, checking for context cancellation between siblings
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if err := ctx.Context.Err(); err != nil {
			return err
		}
		if err := nc.render(ctx, writer, child, RenderOuter, slottedNodes, AttributeMap{}, isTemplate, data); err != nil {
			return err
		}
	}
	return nil
}

// renderAttributes renders all non-ignored attributes.
// Uses ctx.ResolveText to resolve model path expressions in attribute values.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderAttributes(ctx *RenderContext, writer io.Writer, attrs AttributeMap, data any) error {
	var keys []string
	if ctx.DeterministicOutput {
		// Sort for deterministic output (useful for testing)
		keys = slices.Sorted(maps.Keys(attrs))
	} else {
		// Unsorted: faster, no sort allocation
		keys = make([]string, 0, len(attrs))
		for k := range attrs {
			keys = append(keys, k)
		}
	}

	for _, name := range keys {
		if isIgnoredRenderAttribute(name) {
			continue
		}

		// Use ctx.ResolveText to resolve model paths and apply HTML escaping for attribute values
		value, err := ctx.ResolveText(attrs[name], data, true)
		if err != nil {
			return err
		}

		if _, err := io.WriteString(writer, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, name); err != nil {
			return err
		}

		// Render boolean HTML attributes without a value (e.g., <button disabled>)
		if booleanHTMLAttributes[name] && (value == "" || value == "true") {
			continue
		}

		if _, err := io.WriteString(writer, `="`); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, value); err != nil {
			return err
		}
		if _, err := io.WriteString(writer, `"`); err != nil {
			return err
		}
	}
	return nil
}

// getAttributes extracts attributes from an HTML node into a map.
// Normalizes short-form g-* attributes to canonical data-g-* form,
// processes g-att-* prefixes, and applies shortcut attributes.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) getAttributes(n *html.Node) AttributeMap {
	if len(n.Attr) == 0 {
		return nil
	}

	attrs := make(AttributeMap, len(n.Attr))

	for _, a := range n.Attr {
		// Process g-att-* prefixes first (g-att-disabled -> disabled),
		// then normalize remaining g-* to data-g-* canonical form.
		name := normalizeAttrName(processGAtt(a.Key))
		attrs[name] = a.Val
	}

	// Apply shortcut attributes
	applyShortcutAttributes(attrs)

	return attrs
}

// renderOuterRepeat renders the element multiple times, once per item in the collection.
// The ctx parameter is passed through to each iteration's render.
// ---------------------------------------------------------------------------------------------------------------------
func (nc *nodeComponent) renderOuterRepeat(ctx *RenderContext, writer io.Writer, n *html.Node,
	attrs AttributeMap, slottedNodes SlottedRenderables, repeatPath string,
	ignoreResult ignoreResult, isTemplate bool, data any) error {

	modelValue, ok := resolveIterable(ctx, repeatPath, data)
	if !ok || modelValue == nil {
		return nil
	}

	return iterateValue(ctx, modelValue, func(item any) error {
		return nc.renderElement(ctx, writer, n, RenderOuter, slottedNodes, attrs, ignoreResult, isTemplate, item)
	})
}
