package gotmx

import (
	"context"
	"testing"
)

// Helper to create a test Engine instance
func newTestEngine() *Engine {
	e, _ := New()
	return e
}

// newTestRenderContext creates a RenderContext from an Engine instance for testing.
// Uses context.Background() since tests don't need request cancellation.
func newTestRenderContext() *RenderContext {
	return newTestEngine().NewRenderContext(context.Background())
}

// ============================================================================
// checkIgnore Tests
// ============================================================================

func TestCheckIgnore_ReturnsNotFoundWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}

	result, err := checkIgnore(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.HasIgnore {
		t.Error("Expected HasIgnore to be false")
	}
	if result.Mode != "" {
		t.Errorf("Expected empty mode, got '%s'", result.Mode)
	}
}

func TestCheckIgnore_ReturnsOuterMode(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-ignore": "outer",
	}

	result, err := checkIgnore(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result.HasIgnore {
		t.Error("Expected HasIgnore to be true")
	}
	if result.Mode != "outer" {
		t.Errorf("Expected 'outer', got '%s'", result.Mode)
	}
}

func TestCheckIgnore_ReturnsInnerMode(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-ignore": "inner",
	}

	result, err := checkIgnore(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Mode != "inner" {
		t.Errorf("Expected 'inner', got '%s'", result.Mode)
	}
}

func TestCheckIgnore_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-ignore": "[[ .IgnoreMode ]]",
	}
	data := map[string]any{
		"IgnoreMode": "outer-only",
	}

	result, err := checkIgnore(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Mode != "outer-only" {
		t.Errorf("Expected 'outer-only', got '%s'", result.Mode)
	}
}

func TestCheckIgnore_CanonicalFormWorks(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-ignore": "inner",
	}

	result, err := checkIgnore(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Mode != "inner" {
		t.Errorf("Expected 'inner' (canonical form), got '%s'", result.Mode)
	}
}

// ============================================================================
// shouldSkipElement Tests
// ============================================================================

func TestShouldSkipElement_TrueForOuterMode(t *testing.T) {
	result := ignoreResult{Mode: "outer", HasIgnore: true}
	if !shouldSkipElement(result) {
		t.Error("Expected true for outer mode")
	}
}

func TestShouldSkipElement_TrueForEmptyMode(t *testing.T) {
	result := ignoreResult{Mode: "", HasIgnore: true}
	if !shouldSkipElement(result) {
		t.Error("Expected true for empty mode")
	}
}

func TestShouldSkipElement_FalseForInnerMode(t *testing.T) {
	result := ignoreResult{Mode: "inner", HasIgnore: true}
	if shouldSkipElement(result) {
		t.Error("Expected false for inner mode")
	}
}

func TestShouldSkipElement_FalseWhenNoIgnore(t *testing.T) {
	result := ignoreResult{Mode: "outer", HasIgnore: false}
	if shouldSkipElement(result) {
		t.Error("Expected false when HasIgnore is false")
	}
}

// ============================================================================
// shouldSkipChildren Tests
// ============================================================================

func TestShouldSkipChildren_TrueForInnerMode(t *testing.T) {
	result := ignoreResult{Mode: "inner", HasIgnore: true}
	if !shouldSkipChildren(result) {
		t.Error("Expected true for inner mode")
	}
}

func TestShouldSkipChildren_FalseForOuterMode(t *testing.T) {
	result := ignoreResult{Mode: "outer", HasIgnore: true}
	if shouldSkipChildren(result) {
		t.Error("Expected false for outer mode")
	}
}

// ============================================================================
// shouldSkipOuterOnly Tests
// ============================================================================

func TestShouldSkipOuterOnly_TrueForOuterOnlyMode(t *testing.T) {
	result := ignoreResult{Mode: "outer-only", HasIgnore: true}
	if !shouldSkipOuterOnly(result) {
		t.Error("Expected true for outer-only mode")
	}
}

func TestShouldSkipOuterOnly_FalseForOtherModes(t *testing.T) {
	modes := []string{"outer", "inner", "none", ""}
	for _, mode := range modes {
		result := ignoreResult{Mode: mode, HasIgnore: true}
		if shouldSkipOuterOnly(result) {
			t.Errorf("Expected false for mode '%s'", mode)
		}
	}
}

// ============================================================================
// shouldRenderNormally Tests
// ============================================================================

func TestShouldRenderNormally_TrueWhenNoIgnore(t *testing.T) {
	result := ignoreResult{Mode: "", HasIgnore: false}
	if !shouldRenderNormally(result) {
		t.Error("Expected true when no ignore")
	}
}

func TestShouldRenderNormally_TrueForNoneMode(t *testing.T) {
	result := ignoreResult{Mode: "none", HasIgnore: true}
	if !shouldRenderNormally(result) {
		t.Error("Expected true for none mode")
	}
}

func TestShouldRenderNormally_FalseForOtherModes(t *testing.T) {
	modes := []string{"outer", "inner", "outer-only"}
	for _, mode := range modes {
		result := ignoreResult{Mode: mode, HasIgnore: true}
		if shouldRenderNormally(result) {
			t.Errorf("Expected false for mode '%s'", mode)
		}
	}
}

// ============================================================================
// checkIf Tests
// ============================================================================

func TestCheckIf_ReturnsTrueWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected true when g-if is missing")
	}
}

func TestCheckIf_ReturnsTrueForLiteralTrue(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "true",
	}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected true")
	}
}

func TestCheckIf_ReturnsTrueForUppercaseTrue(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "TRUE",
	}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected true for uppercase TRUE")
	}
}

func TestCheckIf_ReturnsFalseForLiteralFalse(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "false",
	}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected false")
	}
}

func TestCheckIf_ReturnsFalseForArbitraryString(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "anything",
	}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result {
		t.Error("Expected false for arbitrary string")
	}
}

func TestCheckIf_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "[[ .IsVisible ]]",
	}
	data := map[string]any{
		"IsVisible": true,
	}

	result, err := checkIf(ctx, attrs, data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected true from model path")
	}
}

func TestCheckIf_ReturnsTrueForNotFalse(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-if": "!false",
	}

	result, err := checkIf(ctx, attrs, nil)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !result {
		t.Error("Expected true for !false")
	}
}

// ============================================================================
// evaluateCondition Tests
// ============================================================================

func TestEvaluateCondition_TrueVariants(t *testing.T) {
	ctx := newTestRenderContext()
	trueValues := []string{"true", "TRUE", "True", "!false", "!FALSE", "!anything", "!0", "!"}

	for _, val := range trueValues {
		result, err := evaluateCondition(ctx, val, nil)
		if err != nil {
			t.Errorf("Unexpected error for '%s': %v", val, err)
		}
		if !result {
			t.Errorf("Expected true for '%s'", val)
		}
	}
}

func TestEvaluateCondition_FalseVariants(t *testing.T) {
	ctx := newTestRenderContext()
	falseValues := []string{"false", "FALSE", "anything", "", "0", "1", "yes", "no", "!true", "!TRUE"}

	for _, val := range falseValues {
		result, err := evaluateCondition(ctx, val, nil)
		if err != nil {
			t.Errorf("Unexpected error for '%s': %v", val, err)
		}
		if result {
			t.Errorf("Expected false for '%s'", val)
		}
	}
}

// ============================================================================
// applyWith Tests
// ============================================================================

func TestApplyWith_ReturnsOriginalDataWhenMissing(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{}
	originalData := map[string]any{"key": "value"}

	result := applyWith(ctx, attrs, originalData)

	// Compare by checking the actual values since maps can't be compared directly
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected map result")
	}
	if resultMap["key"] != "value" {
		t.Error("Expected original data when g-with is missing")
	}
}

func TestApplyWith_SwitchesContext(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-with": "[[ .Nested ]]",
	}
	originalData := map[string]any{
		"Nested": map[string]any{
			"inner": "value",
		},
	}

	result := applyWith(ctx, attrs, originalData)

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected map result")
	}
	if resultMap["inner"] != "value" {
		t.Errorf("Expected 'value', got '%v'", resultMap["inner"])
	}
}

func TestApplyWith_ReturnsOriginalForInvalidPath(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-with": "[[ .NonExistent ]]",
	}
	originalData := map[string]any{"key": "value"}

	result := applyWith(ctx, attrs, originalData)

	// When path doesn't exist, should return original data
	// Compare by checking the actual values since maps can't be compared directly
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected map result")
	}
	if resultMap["key"] != "value" {
		t.Error("Expected original data for invalid path")
	}
}

func TestApplyWith_CanonicalFormWorks(t *testing.T) {
	ctx := newTestRenderContext()
	attrs := AttributeMap{
		"data-g-with": "[[ .DataNested ]]",
	}
	originalData := map[string]any{
		"DataNested": map[string]any{"source": "data-g-with"},
	}

	result := applyWith(ctx, attrs, originalData)

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected map result")
	}
	if resultMap["source"] != "data-g-with" {
		t.Errorf("Expected 'data-g-with' (canonical form), got '%v'", resultMap["source"])
	}
}
