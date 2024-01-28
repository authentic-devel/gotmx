package gotmx

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestNewEngineWithDefaults(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	if engine == nil {
		t.Fatal("New() returned nil engine")
	}
	if engine.registry == nil {
		t.Fatal("engine.registry is nil")
	}
}

func TestEngineLoadHTML(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test-template">Hello, World!</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "test-template", nil)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	expected := `<div>Hello, World!</div>`
	if result != expected {
		t.Errorf("RenderString() = %q, want %q", result, expected)
	}
}

func TestEngineRenderWithData(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="greeting"><span data-g-inner-text="[[ .Name ]]">placeholder</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Name": "Alice"}
	result, err := engine.RenderString(context.Background(), "greeting", data)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	expected := `<div><span>Alice</span></div>`
	if result != expected {
		t.Errorf("RenderString() = %q, want %q", result, expected)
	}
}

func TestEngineRenderEscapesHTML(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test"><span data-g-inner-text="[[ .Text ]]">placeholder</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Text": "<script>alert('xss')</script>"}
	result, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	// HTML should be escaped
	if bytes.Contains([]byte(result), []byte("<script>")) {
		t.Errorf("RenderString() did not escape HTML: %q", result)
	}
}

func TestEngineRenderInnerTextAlwaysEscapes(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test"><span data-g-inner-text="[[ .Text ]]">placeholder</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Text": "<b>bold</b>"}
	// g-inner-text always escapes for XSS safety, even with Unescaped()
	result, err := engine.RenderString(context.Background(), "test", data, Unescaped())
	if err != nil {
		t.Fatalf("RenderString(Unescaped()) returned error: %v", err)
	}

	// HTML should still be escaped because g-inner-text is unconditionally safe
	if bytes.Contains([]byte(result), []byte("<b>bold</b>")) {
		t.Errorf("g-inner-text should always escape, but got raw HTML: %q", result)
	}
}

func TestEngineRenderInnerHtmlDoesNotEscape(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test"><span data-g-inner-html="[[ .Html ]]">placeholder</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Html": "<b>bold</b>"}
	result, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	// g-inner-html should render raw HTML
	if !bytes.Contains([]byte(result), []byte("<b>bold</b>")) {
		t.Errorf("g-inner-html should not escape, but got: %q", result)
	}
}

func TestEngineRenderToWriter(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<p data-g-define="hello">Hello</p>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	var buf bytes.Buffer
	err = engine.Render(context.Background(), &buf, "hello", nil)
	if err != nil {
		t.Fatalf("Render() returned error: %v", err)
	}

	expected := `<p>Hello</p>`
	if buf.String() != expected {
		t.Errorf("Render() = %q, want %q", buf.String(), expected)
	}
}

func TestEngineHasTemplate(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="exists">Content</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	if !engine.HasTemplate("exists") {
		t.Error("HasTemplate('exists') = false, want true")
	}

	if engine.HasTemplate("does-not-exist") {
		t.Error("HasTemplate('does-not-exist') = true, want false")
	}
}

func TestEngineComponent(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="component">Component</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	component, err := engine.Component("component", nil)
	if err != nil {
		t.Fatalf("Component() returned error: %v", err)
	}

	if component == nil {
		t.Error("Component() returned nil")
	}
}

func TestEngineWithSlots(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`
		<div data-g-define="layout">
			<header data-g-define-slot="header">Default Header</header>
			<main data-g-define-slot="">Default Content</main>
		</div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "layout", nil,
		WithSlots(Slots{
			"header": "<h1>Custom Header</h1>",
			"":       "<p>Custom Content</p>",
		}),
	)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("Custom Header")) {
		t.Errorf("Slots content not rendered: %q", result)
	}
	if !bytes.Contains([]byte(result), []byte("Custom Content")) {
		t.Errorf("Default slot content not rendered: %q", result)
	}
}

func TestEngineSingleSlot(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`
		<div data-g-define="card">
			<div class="body" data-g-define-slot="body">Default</div>
		</div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "card", nil,
		Slot("body", "<p>Card Body</p>"),
	)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("Card Body")) {
		t.Errorf("Slot content not rendered: %q", result)
	}
}

func TestEngineWithLogger(t *testing.T) {
	logger := &noopLogger{}

	engine, err := New(WithLogger(logger))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	if engine.config.logger != logger {
		t.Error("Logger was not set correctly")
	}
}

func TestEngineWithDevMode(t *testing.T) {
	// Test dev mode enabled
	engine, err := New(WithDevMode(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	engine.Close()

	if !engine.config.devMode {
		t.Error("DevMode was not enabled")
	}

	// Test dev mode disabled
	engine2, err := New(WithDevMode(false))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine2.Close()

	if engine2.config.devMode {
		t.Error("DevMode should be disabled")
	}
}

func TestEngineWithDevDebounce(t *testing.T) {
	d := 500 * time.Millisecond
	engine, err := New(WithDevDebounce(d))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	if engine.config.devDebounce != d {
		t.Errorf("Expected devDebounce=%v, got %v", d, engine.config.devDebounce)
	}
}

func TestEngineWithReloadCallback(t *testing.T) {
	called := false
	cb := func(err error) {
		called = true
	}
	engine, err := New(WithReloadCallback(cb))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	if engine.config.reloadCallback == nil {
		t.Error("Expected reloadCallback to be set")
	}

	// Invoke to verify it works
	engine.config.reloadCallback(nil)
	if !called {
		t.Error("Expected callback to have been called")
	}
}

func TestEngineRegisterFunc(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	// Register a custom function
	engine.RegisterFunc("uppercase", func(s string) string {
		result := ""
		for _, r := range s {
			if r >= 'a' && r <= 'z' {
				result += string(r - 32)
			} else {
				result += string(r)
			}
		}
		return result
	})

	// Load a template that uses the function
	err = engine.LoadHTML(`<div data-g-define="test" data-g-as-template="">{{ uppercase .Name }}</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Name": "alice"}
	result, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("ALICE")) {
		t.Errorf("Custom function not applied: %q", result)
	}
}

// ================================================================================================
// Escaping behavior tests (via Engine API)
// ================================================================================================

func TestEngineUnescapedAffectsTextNodes(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	// Template with raw text nodes containing special characters.
	// The HTML parser will unescape &amp; etc during parsing, so we use model paths
	// to inject content with special chars at render time.
	err = engine.LoadHTML(`<div data-g-define="test" data-g-inner-text="[[ .Text ]]">placeholder</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Text": "<b>bold</b>"}

	// With default escaping, g-inner-text always escapes (unconditionally safe)
	escaped, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}
	if bytes.Contains([]byte(escaped), []byte("<b>")) {
		t.Errorf("g-inner-text should always escape, got: %q", escaped)
	}

	// With Unescaped(), g-inner-text STILL escapes
	unescaped, err := engine.RenderString(context.Background(), "test", data, Unescaped())
	if err != nil {
		t.Fatalf("RenderString(Unescaped()) error: %v", err)
	}
	if bytes.Contains([]byte(unescaped), []byte("<b>")) {
		t.Errorf("g-inner-text should escape even with Unescaped(), got: %q", unescaped)
	}
}

func TestEngineUnescapedAffectsOuterText(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test"><span data-g-outer-text="[[ .Text ]]">placeholder</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Text": "<b>bold</b>"}

	// Default: g-outer-text escapes
	escaped, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}
	if bytes.Contains([]byte(escaped), []byte("<b>bold</b>")) {
		t.Errorf("g-outer-text should escape by default, got: %q", escaped)
	}

	// With Unescaped(): g-outer-text does NOT escape (follows global setting)
	unescaped, err := engine.RenderString(context.Background(), "test", data, Unescaped())
	if err != nil {
		t.Fatalf("RenderString(Unescaped()) error: %v", err)
	}
	if !bytes.Contains([]byte(unescaped), []byte("<b>bold</b>")) {
		t.Errorf("g-outer-text should respect Unescaped(), got: %q", unescaped)
	}
}

func TestEngineContextCancellationDuringRender(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<ul data-g-define="list"><li data-g-outer-repeat="[[ .Items ]]" data-g-inner-text="[[ . ]]">item</li></ul>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	items := make([]string, 10000)
	for i := range items {
		items[i] = "item"
	}
	data := map[string]any{"Items": items}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var buf bytes.Buffer
	err = engine.Render(ctx, &buf, "list", data)
	if err == nil {
		t.Error("Expected error from cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %T - %v", err, err)
	}
}

func TestEngineAttributeValuesAlwaysEscaped(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test" data-g-att-title="[[ .Title ]]">content</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]string{"Title": `He said "hello" & goodbye`}

	// Even with Unescaped(), attribute values should be escaped
	result, err := engine.RenderString(context.Background(), "test", data, Unescaped())
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}
	if bytes.Contains([]byte(result), []byte(`"hello"`)) {
		t.Errorf("Attribute values should always be escaped, got: %q", result)
	}
}

// ================================================================================================
// WithLayout tests
// ================================================================================================

func TestEngineWithLayout(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`
		<div data-g-define="layout">
			<header>Layout Header</header>
			<main data-g-define-slot="">default content</main>
		</div>
		<p data-g-define="page" data-g-inner-text="[[ .Message ]]">placeholder</p>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "page",
		map[string]string{"Message": "Hello"},
		WithLayout("layout", nil),
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("Layout Header")) {
		t.Errorf("Expected layout header in output, got: %q", result)
	}
	if !bytes.Contains([]byte(result), []byte("Hello")) {
		t.Errorf("Expected page content in output, got: %q", result)
	}
}

func TestEngineWithLayoutNamedSlot(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`
		<div data-g-define="layout">
			<header>Static Header</header>
			<main data-g-define-slot="content">default content</main>
		</div>
		<p data-g-define="page">Page Content</p>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "page", nil,
		WithLayout("layout", nil),
		WithLayoutSlot("content"),
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	// Page content should be in the "content" slot
	if !bytes.Contains([]byte(result), []byte("Page Content")) {
		t.Errorf("Expected page content in named slot, got: %q", result)
	}
	// The static header (not a slot) should always render
	if !bytes.Contains([]byte(result), []byte("Static Header")) {
		t.Errorf("Expected static header, got: %q", result)
	}
}

func TestEngineWithLayoutPassesData(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`
		<div data-g-define="layout">
			<title data-g-inner-text="[[ .Title ]]">Default Title</title>
			<main data-g-define-slot="">default</main>
		</div>
		<p data-g-define="page">Page</p>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	layoutData := map[string]string{"Title": "My App"}
	result, err := engine.RenderString(context.Background(), "page", nil,
		WithLayout("layout", layoutData),
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("My App")) {
		t.Errorf("Expected layout data to be rendered, got: %q", result)
	}
}

// ================================================================================================
// Boolean HTML attribute tests
// ================================================================================================

func TestBooleanAttributeRenderedWithoutValue(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<button data-g-define="btn" data-g-attif-disabled="[[ .Disabled ]]">Click</button>`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	// When disabled=true, should render <button disabled>
	result, err := engine.RenderString(context.Background(), "btn",
		map[string]any{"Disabled": true},
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("disabled")) {
		t.Errorf("Expected disabled attribute, got: %q", result)
	}
	if bytes.Contains([]byte(result), []byte(`disabled="`)) {
		t.Errorf("Boolean attribute should not have value, got: %q", result)
	}
}

func TestBooleanAttributeNotRenderedWhenFalse(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<button data-g-define="btn" data-g-attif-disabled="[[ .Disabled ]]">Click</button>`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	// When disabled=false, should render <button>
	result, err := engine.RenderString(context.Background(), "btn",
		map[string]any{"Disabled": false},
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if bytes.Contains([]byte(result), []byte("disabled")) {
		t.Errorf("Disabled should not be present when false, got: %q", result)
	}
}

func TestNonBooleanAttributeStillHasValue(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<div data-g-define="test" data-g-att-data-custom="myvalue">content</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte(`data-custom="myvalue"`)) {
		t.Errorf("Non-boolean attribute should have value, got: %q", result)
	}
}

func TestMultipleBooleanAttributes(t *testing.T) {
	engine, err := New(WithDeterministicOutput(true))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer engine.Close()

	err = engine.LoadHTML(`<input data-g-define="inp" type="checkbox" data-g-attif-checked="[[ .Checked ]]" data-g-attif-required="[[ .Required ]]" data-g-attif-disabled="[[ .Disabled ]]" />`)
	if err != nil {
		t.Fatalf("LoadHTML() error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "inp",
		map[string]any{"Checked": true, "Required": true, "Disabled": false},
	)
	if err != nil {
		t.Fatalf("RenderString() error: %v", err)
	}

	if !bytes.Contains([]byte(result), []byte("checked")) {
		t.Errorf("Expected checked attribute, got: %q", result)
	}
	if !bytes.Contains([]byte(result), []byte("required")) {
		t.Errorf("Expected required attribute, got: %q", result)
	}
	if bytes.Contains([]byte(result), []byte("disabled")) {
		t.Errorf("Disabled should not be present, got: %q", result)
	}
}
