package gotmx

import (
	"errors"
	"strings"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"button", "buttn", 1},
		{"button", "buton", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			result := levenshteinDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFindClosestMatch(t *testing.T) {
	candidates := []TemplateName{"button", "card", "header", "footer", "navigation"}

	tests := []struct {
		target   TemplateName
		expected TemplateName
	}{
		{"buttn", "button"},  // 1 edit away
		{"buton", "button"},  // 1 edit away
		{"botton", "button"}, // 1 edit away
		{"card", "card"},     // exact match
		{"cardd", "card"},    // 1 edit away
		{"xyz", ""},          // too far from any candidate
		{"completely", ""},   // too far from any candidate
		{"heater", "header"}, // 2 edits away
		{"footr", "footer"},  // 1 edit away
		{"nav", ""},          // too short to match navigation reliably
	}

	for _, tt := range tests {
		t.Run(string(tt.target), func(t *testing.T) {
			result := findClosestMatch(tt.target, candidates)
			if result != tt.expected {
				t.Errorf("findClosestMatch(%q) = %q, want %q", tt.target, result, tt.expected)
			}
		})
	}
}

func TestTemplateNotFoundError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &TemplateNotFoundError{
			Name: "mytemplate",
		}
		if !strings.Contains(err.Error(), "mytemplate") {
			t.Errorf("error should contain template name, got: %s", err.Error())
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error should contain 'not found', got: %s", err.Error())
		}
	})

	t.Run("with namespace", func(t *testing.T) {
		err := &TemplateNotFoundError{
			Name:      "mytemplate",
			Namespace: "frontend/components",
		}
		if !strings.Contains(err.Error(), "frontend/components") {
			t.Errorf("error should contain namespace, got: %s", err.Error())
		}
	})

	t.Run("with did you mean", func(t *testing.T) {
		err := &TemplateNotFoundError{
			Name:       "buttn",
			DidYouMean: "button",
		}
		if !strings.Contains(err.Error(), "did you mean") {
			t.Errorf("error should contain 'did you mean', got: %s", err.Error())
		}
		if !strings.Contains(err.Error(), "button") {
			t.Errorf("error should contain suggestion, got: %s", err.Error())
		}
	})

	t.Run("with available templates", func(t *testing.T) {
		err := &TemplateNotFoundError{
			Name:      "unknown",
			Available: []TemplateName{"button", "card", "header"},
		}
		if !strings.Contains(err.Error(), "available") {
			t.Errorf("error should contain 'available', got: %s", err.Error())
		}
		if !strings.Contains(err.Error(), "button") {
			t.Errorf("error should list available templates, got: %s", err.Error())
		}
	})

	t.Run("many available templates", func(t *testing.T) {
		available := make([]TemplateName, 15)
		for i := range available {
			available[i] = TemplateName("template" + string(rune('a'+i)))
		}
		err := &TemplateNotFoundError{
			Name:      "unknown",
			Available: available,
		}
		// Should show count instead of listing all
		if !strings.Contains(err.Error(), "15 templates available") {
			t.Errorf("error should show count for many templates, got: %s", err.Error())
		}
	})
}

func TestAmbiguousTemplateError(t *testing.T) {
	err := &AmbiguousTemplateError{
		Name:       "button",
		Namespaces: []Namespace{"frontend/components", "admin/components"},
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "button") {
		t.Errorf("error should contain template name, got: %s", errStr)
	}
	if !strings.Contains(errStr, "ambiguous") {
		t.Errorf("error should contain 'ambiguous', got: %s", errStr)
	}
	if !strings.Contains(errStr, "frontend/components") {
		t.Errorf("error should list namespaces, got: %s", errStr)
	}
	if !strings.Contains(errStr, "namespace#button") {
		t.Errorf("error should suggest qualified name format, got: %s", errStr)
	}
}

func TestComponentNotFoundError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &ComponentNotFoundError{
			ComponentRef: "my-button",
		}
		if !strings.Contains(err.Error(), "my-button") {
			t.Errorf("error should contain component ref, got: %s", err.Error())
		}
	})

	t.Run("with cause", func(t *testing.T) {
		cause := &TemplateNotFoundError{Name: "button"}
		err := &ComponentNotFoundError{
			ComponentRef: "my-button",
			Cause:        cause,
		}
		if !strings.Contains(err.Error(), "my-button") {
			t.Errorf("error should contain component ref, got: %s", err.Error())
		}

		// Test Unwrap
		unwrapped := errors.Unwrap(err)
		if unwrapped != cause {
			t.Errorf("Unwrap should return cause, got: %v", unwrapped)
		}
	})
}

func TestRenderError(t *testing.T) {
	t.Run("full context", func(t *testing.T) {
		cause := errors.New("value is nil")
		err := &RenderError{
			Template:  "page.html",
			Element:   "div",
			Attribute: "g-if",
			Cause:     cause,
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "page.html") {
			t.Errorf("error should contain template, got: %s", errStr)
		}
		if !strings.Contains(errStr, "<div>") {
			t.Errorf("error should contain element in brackets, got: %s", errStr)
		}
		if !strings.Contains(errStr, "[g-if]") {
			t.Errorf("error should contain attribute in brackets, got: %s", errStr)
		}
		if !strings.Contains(errStr, "value is nil") {
			t.Errorf("error should contain cause, got: %s", errStr)
		}
	})

	t.Run("element only", func(t *testing.T) {
		err := &RenderError{
			Element: "span",
			Cause:   errors.New("failed"),
		}
		errStr := err.Error()
		if !strings.Contains(errStr, "<span>") {
			t.Errorf("error should contain element, got: %s", errStr)
		}
		if strings.Contains(errStr, "template") {
			t.Errorf("error should not mention template when empty, got: %s", errStr)
		}
	})

	t.Run("unwrap", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &RenderError{
			Element: "div",
			Cause:   cause,
		}
		unwrapped := errors.Unwrap(err)
		if unwrapped != cause {
			t.Errorf("Unwrap should return cause, got: %v", unwrapped)
		}
	})
}

func TestWrapRenderError(t *testing.T) {
	t.Run("wraps regular error", func(t *testing.T) {
		cause := errors.New("something went wrong")
		wrapped := wrapRenderError(cause, "div", "g-if")

		var renderErr *RenderError
		if !errors.As(wrapped, &renderErr) {
			t.Fatal("wrapped error should be RenderError")
		}
		if renderErr.Element != "div" {
			t.Errorf("Element = %q, want %q", renderErr.Element, "div")
		}
		if renderErr.Attribute != "g-if" {
			t.Errorf("Attribute = %q, want %q", renderErr.Attribute, "g-if")
		}
		if renderErr.Cause != cause {
			t.Errorf("Cause should be original error")
		}
	})

	t.Run("does not double wrap", func(t *testing.T) {
		original := &RenderError{
			Element:   "span",
			Attribute: "g-with",
			Cause:     errors.New("inner error"),
		}
		wrapped := wrapRenderError(original, "div", "g-if")

		// Should return the original error unchanged
		if wrapped != original {
			t.Error("should not double-wrap RenderError")
		}
	})

	t.Run("returns nil for nil error", func(t *testing.T) {
		result := wrapRenderError(nil, "div", "g-if")
		if result != nil {
			t.Error("should return nil for nil error")
		}
	})
}
