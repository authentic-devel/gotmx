package gotmx

import (
	"testing"
)

// ============================================================================
// getInnerText Tests
// ============================================================================

func TestGetInnerText_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-inner-text": "Hello World",
	}

	result := getInnerText(attrs)

	if !result.Found {
		t.Error("Expected Found to be true")
	}
	if result.Value != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result.Value)
	}
}

func TestGetInnerText_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	result := getInnerText(attrs)

	if result.Found {
		t.Error("Expected Found to be false")
	}
}

func TestGetInnerText_CanonicalForm(t *testing.T) {
	attrs := AttributeMap{
		"data-g-inner-text": "Long Form",
	}

	result := getInnerText(attrs)

	if result.Value != "Long Form" {
		t.Errorf("Expected 'Long Form', got '%s'", result.Value)
	}
}

// ============================================================================
// getInnerHtml Tests
// ============================================================================

func TestGetInnerHtml_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-inner-html": "<b>Bold</b>",
	}

	result := getInnerHtml(attrs)

	if !result.Found {
		t.Error("Expected Found to be true")
	}
	if result.Value != "<b>Bold</b>" {
		t.Errorf("Expected '<b>Bold</b>', got '%s'", result.Value)
	}
}

func TestGetInnerHtml_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	result := getInnerHtml(attrs)

	if result.Found {
		t.Error("Expected Found to be false")
	}
}

// ============================================================================
// getOuterText Tests
// ============================================================================

func TestGetOuterText_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-outer-text": "Replacement Text",
	}

	result := getOuterText(attrs)

	if !result.Found {
		t.Error("Expected Found to be true")
	}
	if result.Value != "Replacement Text" {
		t.Errorf("Expected 'Replacement Text', got '%s'", result.Value)
	}
}

func TestGetOuterText_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	result := getOuterText(attrs)

	if result.Found {
		t.Error("Expected Found to be false")
	}
}

// ============================================================================
// getDefineSlot Tests
// ============================================================================

func TestGetDefineSlot_ReturnsSlotNameWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-define-slot": "main",
	}

	name, isSlot := getDefineSlot(attrs)

	if !isSlot {
		t.Error("Expected isSlot to be true")
	}
	if name != "main" {
		t.Errorf("Expected 'main', got '%s'", name)
	}
}

func TestGetDefineSlot_ReturnsEmptyStringForDefaultSlot(t *testing.T) {
	attrs := AttributeMap{
		"data-g-define-slot": "",
	}

	name, isSlot := getDefineSlot(attrs)

	if !isSlot {
		t.Error("Expected isSlot to be true")
	}
	if name != "" {
		t.Errorf("Expected empty string for default slot, got '%s'", name)
	}
}

func TestGetDefineSlot_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, isSlot := getDefineSlot(attrs)

	if isSlot {
		t.Error("Expected isSlot to be false")
	}
}

func TestGetDefineSlot_CanonicalForm(t *testing.T) {
	attrs := AttributeMap{
		"data-g-define-slot": "data-slot",
	}

	name, _ := getDefineSlot(attrs)

	if name != "data-slot" {
		t.Errorf("Expected 'data-slot', got '%s'", name)
	}
}

// ============================================================================
// resolveInnerText Tests
// ============================================================================

func TestResolveInnerText_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-text": "Hello",
	}

	text, found, err := resolveInnerText(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if text != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", text)
	}
}

func TestResolveInnerText_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-text": "[[ .Message ]]",
	}
	data := map[string]any{
		"Message": "Dynamic Text",
	}

	text, found, err := resolveInnerText(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if text != "Dynamic Text" {
		t.Errorf("Expected 'Dynamic Text', got '%s'", text)
	}
}

func TestResolveInnerText_EscapesWhenRequested(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-text": "[[ .Content ]]",
	}
	data := map[string]any{
		"Content": "<script>alert('xss')</script>",
	}

	text, _, err := resolveInnerText(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if text == "<script>alert('xss')</script>" {
		t.Error("Expected HTML to be escaped")
	}
}

func TestResolveInnerText_ReturnsNotFoundWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}

	_, found, err := resolveInnerText(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// resolveInnerHtml Tests
// ============================================================================

func TestResolveInnerHtml_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-html": "<b>Bold</b>",
	}

	html, found, err := resolveInnerHtml(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if html != "<b>Bold</b>" {
		t.Errorf("Expected '<b>Bold</b>', got '%s'", html)
	}
}

func TestResolveInnerHtml_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-html": "[[ .HtmlContent ]]",
	}
	data := map[string]any{
		"HtmlContent": "<em>Emphasis</em>",
	}

	html, found, err := resolveInnerHtml(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if html != "<em>Emphasis</em>" {
		t.Errorf("Expected '<em>Emphasis</em>', got '%s'", html)
	}
}

func TestResolveInnerHtml_DoesNotEscape(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-inner-html": "<script>alert('ok')</script>",
	}

	html, _, err := resolveInnerHtml(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if html != "<script>alert('ok')</script>" {
		t.Errorf("Expected unescaped HTML, got '%s'", html)
	}
}

// ============================================================================
// resolveOuterText Tests
// ============================================================================

func TestResolveOuterText_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-outer-text": "Replacement",
	}

	text, found, err := resolveOuterText(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if text != "Replacement" {
		t.Errorf("Expected 'Replacement', got '%s'", text)
	}
}

func TestResolveOuterText_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-outer-text": "[[ .Text ]]",
	}
	data := map[string]any{
		"Text": "Dynamic Replacement",
	}

	text, found, err := resolveOuterText(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected found to be true")
	}
	if text != "Dynamic Replacement" {
		t.Errorf("Expected 'Dynamic Replacement', got '%s'", text)
	}
}

func TestResolveOuterText_ReturnsNotFoundWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}

	_, found, err := resolveOuterText(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if found {
		t.Error("Expected found to be false")
	}
}
