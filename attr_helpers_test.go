package gotmx

import (
	"testing"
)

// ============================================================================
// normalizeAttrName Tests
// ============================================================================

func TestNormalizeAttrName_ConvertsShortFormToLong(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"g-if", "data-g-if"},
		{"g-ignore", "data-g-ignore"},
		{"g-inner-text", "data-g-inner-text"},
		{"g-outer-repeat", "data-g-outer-repeat"},
		{"g-use", "data-g-use"},
		{"g-define-slot", "data-g-define-slot"},
		{"g-attif-disabled", "data-g-attif-disabled"},
	}

	for _, tt := range tests {
		result := normalizeAttrName(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeAttrName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeAttrName_LeavesLongFormUnchanged(t *testing.T) {
	tests := []string{
		"data-g-if",
		"data-g-ignore",
		"data-g-inner-text",
		"data-g-use",
	}

	for _, input := range tests {
		result := normalizeAttrName(input)
		if result != input {
			t.Errorf("normalizeAttrName(%q) = %q, want unchanged", input, result)
		}
	}
}

func TestNormalizeAttrName_LeavesRegularAttributesUnchanged(t *testing.T) {
	tests := []string{
		"class",
		"id",
		"disabled",
		"data-custom",
		"href",
	}

	for _, input := range tests {
		result := normalizeAttrName(input)
		if result != input {
			t.Errorf("normalizeAttrName(%q) = %q, want unchanged", input, result)
		}
	}
}

// ============================================================================
// applyOverrides Tests
// ============================================================================

func TestApplyOverrides_AddsNewAttributes(t *testing.T) {
	attrs := AttributeMap{
		"existing": "value",
	}
	overrides := AttributeMap{
		"new": "newValue",
	}

	applyOverrides(attrs, overrides)

	if attrs["new"] != "newValue" {
		t.Errorf("Expected 'newValue', got '%s'", attrs["new"])
	}
	if attrs["existing"] != "value" {
		t.Errorf("Expected 'value', got '%s'", attrs["existing"])
	}
}

func TestApplyOverrides_ReplacesExistingAttributes(t *testing.T) {
	attrs := AttributeMap{
		"key": "oldValue",
	}
	overrides := AttributeMap{
		"key": "newValue",
	}

	applyOverrides(attrs, overrides)

	if attrs["key"] != "newValue" {
		t.Errorf("Expected 'newValue', got '%s'", attrs["key"])
	}
}

func TestApplyOverrides_HandlesEmptyOverrides(t *testing.T) {
	attrs := AttributeMap{
		"key": "value",
	}
	overrides := AttributeMap{}

	applyOverrides(attrs, overrides)

	if attrs["key"] != "value" {
		t.Errorf("Expected 'value', got '%s'", attrs["key"])
	}
}
