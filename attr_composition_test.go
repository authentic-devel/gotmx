package gotmx

import (
	"testing"
)

// ============================================================================
// getUse Tests
// ============================================================================

func TestGetUse_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-use": "my-template",
	}

	ref, found := getUse(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if ref != "my-template" {
		t.Errorf("Expected 'my-template', got '%s'", ref)
	}
}

func TestGetUse_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getUse(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

func TestGetUse_ReturnsValue(t *testing.T) {
	attrs := AttributeMap{
		"data-g-use": "long-template",
	}

	ref, _ := getUse(attrs)

	if ref != "long-template" {
		t.Errorf("Expected 'long-template', got '%s'", ref)
	}
}

// ============================================================================
// getInnerUse Tests
// ============================================================================

func TestGetInnerUse_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-inner-use": "inner-template",
	}

	ref, found := getInnerUse(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if ref != "inner-template" {
		t.Errorf("Expected 'inner-template', got '%s'", ref)
	}
}

func TestGetInnerUse_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getInnerUse(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// getUseSlot Tests
// ============================================================================

func TestGetUseSlot_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-use-slot": "sidebar",
	}

	slot, found := getUseSlot(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if slot != "sidebar" {
		t.Errorf("Expected 'sidebar', got '%s'", slot)
	}
}

func TestGetUseSlot_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getUseSlot(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// getOverrideAtt Tests
// ============================================================================

func TestGetOverrideAtt_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-override-att": "class,id,disabled",
	}

	list, found := getOverrideAtt(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if list != "class,id,disabled" {
		t.Errorf("Expected 'class,id,disabled', got '%s'", list)
	}
}

func TestGetOverrideAtt_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getOverrideAtt(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// getAsTemplate Tests
// ============================================================================

func TestGetAsTemplate_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-as-template": "go-template-name",
	}

	name, found := getAsTemplate(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if name != "go-template-name" {
		t.Errorf("Expected 'go-template-name', got '%s'", name)
	}
}

func TestGetAsTemplate_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getAsTemplate(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// getAsUnsafeTemplate Tests
// ============================================================================

func TestGetAsUnsafeTemplate_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-as-unsafe-template": "unsafe-template-name",
	}

	name, found := getAsUnsafeTemplate(attrs)

	if !found {
		t.Error("Expected found to be true")
	}
	if name != "unsafe-template-name" {
		t.Errorf("Expected 'unsafe-template-name', got '%s'", name)
	}
}

func TestGetAsUnsafeTemplate_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getAsUnsafeTemplate(attrs)

	if found {
		t.Error("Expected found to be false")
	}
}

// ============================================================================
// parseOverrideAttributes Tests
// ============================================================================

func TestParseOverrideAttributes_ParsesCommaSeparatedList(t *testing.T) {
	result := parseOverrideAttributes("class,id,disabled")

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != "class" || result[1] != "id" || result[2] != "disabled" {
		t.Errorf("Unexpected values: %v", result)
	}
}

func TestParseOverrideAttributes_TrimsWhitespace(t *testing.T) {
	result := parseOverrideAttributes("class , id , disabled")

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != "class" || result[1] != "id" || result[2] != "disabled" {
		t.Errorf("Unexpected values (should be trimmed): %v", result)
	}
}

func TestParseOverrideAttributes_HandlesEmptyString(t *testing.T) {
	result := parseOverrideAttributes("")

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestParseOverrideAttributes_HandlesSingleAttribute(t *testing.T) {
	result := parseOverrideAttributes("class")

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0] != "class" {
		t.Errorf("Expected 'class', got '%s'", result[0])
	}
}

func TestParseOverrideAttributes_SkipsEmptyParts(t *testing.T) {
	result := parseOverrideAttributes("class,,id,")

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0] != "class" || result[1] != "id" {
		t.Errorf("Unexpected values: %v", result)
	}
}

// ============================================================================
// collectOverrideAttributes Tests
// ============================================================================

func TestCollectOverrideAttributes_CollectsSpecifiedAttributes(t *testing.T) {
	source := AttributeMap{
		"class":    "my-class",
		"id":       "my-id",
		"disabled": "",
		"other":    "ignored",
	}

	result := collectOverrideAttributes(source, []string{"class", "id"})

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result["class"] != "my-class" {
		t.Errorf("Expected 'my-class', got '%s'", result["class"])
	}
	if result["id"] != "my-id" {
		t.Errorf("Expected 'my-id', got '%s'", result["id"])
	}
	if _, exists := result["other"]; exists {
		t.Error("Did not expect 'other' to be collected")
	}
}

func TestCollectOverrideAttributes_IncludesGAttIfAttributes(t *testing.T) {
	source := AttributeMap{
		"disabled":              "",
		"data-g-attif-disabled": "[[ .IsDisabled ]]",
		"data-g-attif-hidden":   "[[ .IsHidden ]]",
	}

	result := collectOverrideAttributes(source, []string{"disabled", "hidden"})

	if result["disabled"] != "" {
		t.Errorf("Expected empty disabled, got '%s'", result["disabled"])
	}
	if result["data-g-attif-disabled"] != "[[ .IsDisabled ]]" {
		t.Errorf("Expected data-g-attif-disabled to be collected")
	}
	if result["data-g-attif-hidden"] != "[[ .IsHidden ]]" {
		t.Errorf("Expected data-g-attif-hidden to be collected")
	}
}

func TestCollectOverrideAttributes_HandlesEmptyAttrNames(t *testing.T) {
	source := AttributeMap{
		"class": "my-class",
	}

	result := collectOverrideAttributes(source, []string{})

	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %d", len(result))
	}
}

func TestCollectOverrideAttributes_SkipsMissingAttributes(t *testing.T) {
	source := AttributeMap{
		"class": "my-class",
	}

	result := collectOverrideAttributes(source, []string{"class", "nonexistent"})

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result["class"] != "my-class" {
		t.Errorf("Expected 'my-class', got '%s'", result["class"])
	}
}

// ============================================================================
// resolveTemplateRef Tests
// ============================================================================

func TestResolveTemplateRef_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()

	result, err := resolveTemplateRef(ctx, "my-template", nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "my-template" {
		t.Errorf("Expected 'my-template', got '%s'", result)
	}
}

func TestResolveTemplateRef_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	data := map[string]any{
		"TemplateName": "dynamic-template",
	}

	result, err := resolveTemplateRef(ctx, "[[ .TemplateName ]]", data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "dynamic-template" {
		t.Errorf("Expected 'dynamic-template', got '%s'", result)
	}
}

// ============================================================================
// resolveSlotName Tests
// ============================================================================

func TestResolveSlotName_ResolvesLiteral(t *testing.T) {
	ctx := newTestRenderContext()

	result, err := resolveSlotName(ctx, "sidebar", nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "sidebar" {
		t.Errorf("Expected 'sidebar', got '%s'", result)
	}
}

func TestResolveSlotName_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	data := map[string]any{
		"SlotName": "dynamic-slot",
	}

	result, err := resolveSlotName(ctx, "[[ .SlotName ]]", data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "dynamic-slot" {
		t.Errorf("Expected 'dynamic-slot', got '%s'", result)
	}
}
