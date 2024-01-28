package gotmx

import (
	"testing"
)

// ============================================================================
// getTrans Tests
// ============================================================================

func TestGetTrans_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-trans": "span",
	}

	tag, found := getTrans(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if tag != "span" {
		t.Errorf("Expected 'span', got '%s'", tag)
	}
}

func TestGetTrans_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getTrans(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

func TestGetTrans_CanonicalForm(t *testing.T) {
	attrs := AttributeMap{
		"data-g-trans": "article",
	}

	tag, _ := getTrans(attrs)

	if tag != "article" {
		t.Errorf("Expected 'article', got '%s'", tag)
	}
}

// ============================================================================
// resolveTrans Tests
// ============================================================================

func TestResolveTrans_ReturnsOriginalWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}

	tag, err := resolveTrans(ctx, attrs, "div", nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tag != "div" {
		t.Errorf("Expected 'div', got '%s'", tag)
	}
}

func TestResolveTrans_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-trans": "span",
	}

	tag, err := resolveTrans(ctx, attrs, "div", nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tag != "span" {
		t.Errorf("Expected 'span', got '%s'", tag)
	}
}

func TestResolveTrans_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-trans": "[[ .TagName ]]",
	}
	data := map[string]any{
		"TagName": "article",
	}

	tag, err := resolveTrans(ctx, attrs, "div", data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if tag != "article" {
		t.Errorf("Expected 'article', got '%s'", tag)
	}
}

// ============================================================================
// processGAttIf Tests
// ============================================================================

func TestProcessGAttIf_AddsBooleanAttributeWhenTrue(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "true",
	}

	processGAttIf(ctx, attrs, nil)

	if _, exists := attrs["disabled"]; !exists {
		t.Error("Expected 'disabled' attribute to be added")
	}
	if attrs["disabled"] != "" {
		t.Errorf("Expected empty value for boolean attribute, got '%s'", attrs["disabled"])
	}
}

func TestProcessGAttIf_RemovesAttributeWhenFalse(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"disabled":              "",
		"data-g-attif-disabled": "false",
	}

	processGAttIf(ctx, attrs, nil)

	if _, exists := attrs["disabled"]; exists {
		t.Error("Expected 'disabled' attribute to be removed")
	}
}

func TestProcessGAttIf_RemovesGAttIfAttributes(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "true",
		"data-g-attif-readonly": "false",
	}

	processGAttIf(ctx, attrs, nil)

	if _, exists := attrs["data-g-attif-disabled"]; exists {
		t.Error("Expected 'data-g-attif-disabled' to be removed")
	}
	if _, exists := attrs["data-g-attif-readonly"]; exists {
		t.Error("Expected 'data-g-attif-readonly' to be removed")
	}
}

func TestProcessGAttIf_DoesNotAddIfAlreadyExists(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"disabled":              "disabled",
		"data-g-attif-disabled": "true",
	}

	processGAttIf(ctx, attrs, nil)

	// Should keep the original value, not overwrite with empty
	if attrs["disabled"] != "disabled" {
		t.Errorf("Expected 'disabled' to remain 'disabled', got '%s'", attrs["disabled"])
	}
}

func TestProcessGAttIf_DoesNotRemoveIfNotExists(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "false",
	}

	processGAttIf(ctx, attrs, nil)

	// Should not add 'disabled' when condition is false and it didn't exist
	if _, exists := attrs["disabled"]; exists {
		t.Error("Did not expect 'disabled' to be added when condition is false")
	}
}

func TestProcessGAttIf_HandlesMultipleAttributes(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "true",
		"data-g-attif-readonly": "true",
		"hidden":                "",
		"data-g-attif-hidden":   "false",
	}

	processGAttIf(ctx, attrs, nil)

	if _, exists := attrs["disabled"]; !exists {
		t.Error("Expected 'disabled' to be added")
	}
	if _, exists := attrs["readonly"]; !exists {
		t.Error("Expected 'readonly' to be added")
	}
	if _, exists := attrs["hidden"]; exists {
		t.Error("Expected 'hidden' to be removed")
	}
}

func TestProcessGAttIf_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "[[ .IsDisabled ]]",
	}
	data := map[string]any{
		"IsDisabled": true,
	}

	processGAttIf(ctx, attrs, data)

	if _, exists := attrs["disabled"]; !exists {
		t.Error("Expected 'disabled' to be added")
	}
}

func TestProcessGAttIf_HandlesDataGPrefix(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-attif-disabled": "true",
	}

	processGAttIf(ctx, attrs, nil)

	if _, exists := attrs["disabled"]; !exists {
		t.Error("Expected 'disabled' to be added")
	}
	if _, exists := attrs["data-g-attif-disabled"]; exists {
		t.Error("Expected 'data-g-attif-disabled' to be removed")
	}
}

// ============================================================================
// processGAtt Tests
// ============================================================================

func TestProcessGAtt_StripsGAttPrefix(t *testing.T) {
	result := processGAtt("g-att-disabled")

	if result != "disabled" {
		t.Errorf("Expected 'disabled', got '%s'", result)
	}
}

func TestProcessGAtt_StripsDataGAttPrefix(t *testing.T) {
	result := processGAtt("data-g-att-href")

	if result != "href" {
		t.Errorf("Expected 'href', got '%s'", result)
	}
}

func TestProcessGAtt_ReturnsUnchangedForNonPrefix(t *testing.T) {
	result := processGAtt("class")

	if result != "class" {
		t.Errorf("Expected 'class', got '%s'", result)
	}
}

func TestProcessGAtt_ReturnsUnchangedForPartialMatch(t *testing.T) {
	result := processGAtt("g-att") // No suffix

	if result != "g-att" {
		t.Errorf("Expected 'g-att' (unchanged), got '%s'", result)
	}
}

// ============================================================================
// applyShortcutAttributes Tests
// ============================================================================

func TestApplyShortcutAttributes_CopiesGClassToClass(t *testing.T) {
	attrs := AttributeMap{
		"data-g-class": "my-class",
	}

	applyShortcutAttributes(attrs)

	if attrs["class"] != "my-class" {
		t.Errorf("Expected 'my-class', got '%s'", attrs["class"])
	}
}

func TestApplyShortcutAttributes_CopiesDataGClassToClass(t *testing.T) {
	attrs := AttributeMap{
		"data-g-class": "my-class",
	}

	applyShortcutAttributes(attrs)

	if attrs["class"] != "my-class" {
		t.Errorf("Expected 'my-class', got '%s'", attrs["class"])
	}
}

func TestApplyShortcutAttributes_CopiesGHrefToHref(t *testing.T) {
	attrs := AttributeMap{
		"data-g-href": "/path/to/page",
	}

	applyShortcutAttributes(attrs)

	if attrs["href"] != "/path/to/page" {
		t.Errorf("Expected '/path/to/page', got '%s'", attrs["href"])
	}
}

func TestApplyShortcutAttributes_CopiesGSrcToSrc(t *testing.T) {
	attrs := AttributeMap{
		"data-g-src": "/images/logo.png",
	}

	applyShortcutAttributes(attrs)

	if attrs["src"] != "/images/logo.png" {
		t.Errorf("Expected '/images/logo.png', got '%s'", attrs["src"])
	}
}

func TestApplyShortcutAttributes_CanonicalForm(t *testing.T) {
	attrs := AttributeMap{
		"data-g-class": "data-class",
	}

	applyShortcutAttributes(attrs)

	if attrs["class"] != "data-class" {
		t.Errorf("Expected 'data-class', got '%s'", attrs["class"])
	}
}

func TestApplyShortcutAttributes_HandlesAllShortcuts(t *testing.T) {
	attrs := AttributeMap{
		"data-g-class": "test-class",
		"data-g-href":  "/test",
		"data-g-src":   "/test.png",
	}

	applyShortcutAttributes(attrs)

	if attrs["class"] != "test-class" {
		t.Errorf("Expected 'test-class', got '%s'", attrs["class"])
	}
	if attrs["href"] != "/test" {
		t.Errorf("Expected '/test', got '%s'", attrs["href"])
	}
	if attrs["src"] != "/test.png" {
		t.Errorf("Expected '/test.png', got '%s'", attrs["src"])
	}
}

// ============================================================================
// isIgnoredRenderAttribute Tests
// ============================================================================

func TestIsIgnoredRenderAttribute_TrueForControlAttributes(t *testing.T) {
	controlAttrs := []string{
		"data-g-if",
		"data-g-ignore",
		"data-g-with",
		"data-g-use",
		"data-g-inner-text",
		"data-g-define",
	}

	for _, attr := range controlAttrs {
		if !isIgnoredRenderAttribute(attr) {
			t.Errorf("Expected '%s' to be ignored", attr)
		}
	}
}

func TestIsIgnoredRenderAttribute_TrueForPrefixAttributes(t *testing.T) {
	// After normalization, only data-g-attif-* prefix attributes remain in the map.
	// g-att-* and data-g-att-* are stripped by processGAtt before normalization.
	prefixAttrs := []string{
		"data-g-attif-hidden",
		"data-g-attif-readonly",
		"data-g-attif-disabled",
	}

	for _, attr := range prefixAttrs {
		if !isIgnoredRenderAttribute(attr) {
			t.Errorf("Expected '%s' to be ignored", attr)
		}
	}
}

func TestIsIgnoredRenderAttribute_FalseForRegularAttributes(t *testing.T) {
	regularAttrs := []string{
		"class",
		"id",
		"href",
		"disabled",
		"data-custom",
	}

	for _, attr := range regularAttrs {
		if isIgnoredRenderAttribute(attr) {
			t.Errorf("Expected '%s' to NOT be ignored", attr)
		}
	}
}
