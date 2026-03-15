package gotmx

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// ================================================================================================
// Circular template reference detection
// ================================================================================================

func TestMaxNestingDepthExceeded(t *testing.T) {
	engine, err := New(WithMaxNestingDepth(5))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Template A uses Template B, Template B uses Template A → circular
	err = engine.LoadHTML(`
		<div data-g-define="tmpl-a"><div data-g-use="tmpl-b"></div></div>
		<div data-g-define="tmpl-b"><div data-g-use="tmpl-a"></div></div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	_, err = engine.RenderString(context.Background(), "tmpl-a", nil)
	if err == nil {
		t.Fatal("Expected error from circular template references, got nil")
	}

	var maxDepthErr *MaxNestingDepthExceededError
	if !errors.As(err, &maxDepthErr) {
		t.Errorf("Expected MaxNestingDepthExceededError, got %T: %v", err, err)
	}
}

func TestMaxNestingDepthSelfReference(t *testing.T) {
	engine, err := New(WithMaxNestingDepth(3))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Template that references itself
	err = engine.LoadHTML(`<div data-g-define="self-ref"><div data-g-use="self-ref"></div></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	_, err = engine.RenderString(context.Background(), "self-ref", nil)
	if err == nil {
		t.Fatal("Expected error from self-referencing template, got nil")
	}

	var maxDepthErr *MaxNestingDepthExceededError
	if !errors.As(err, &maxDepthErr) {
		t.Errorf("Expected MaxNestingDepthExceededError, got %T: %v", err, err)
	}
}

func TestDeepNestingWithinLimit(t *testing.T) {
	engine, err := New(WithMaxNestingDepth(10))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// 3-level nesting chain, well within limit of 10
	err = engine.LoadHTML(`
		<div data-g-define="level1"><div data-g-use="level2"></div></div>
		<div data-g-define="level2"><div data-g-use="level3"></div></div>
		<div data-g-define="level3">Leaf</div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "level1", nil)
	if err != nil {
		t.Fatalf("Expected no error for nesting within limit, got: %v", err)
	}
	if result != "<div><div><div>Leaf</div></div></div>" {
		t.Errorf("Unexpected output: %q", result)
	}
}

// ================================================================================================
// Concurrent access
// ================================================================================================

func TestConcurrentRender(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<span data-g-define="concurrent" data-g-inner-text="[[ .Name ]]">placeholder</span>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			data := map[string]any{"Name": fmt.Sprintf("user-%d", n)}
			result, err := engine.RenderString(context.Background(), "concurrent", data)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", n, err)
				return
			}
			expected := fmt.Sprintf("<span>user-%d</span>", n)
			if result != expected {
				errs <- fmt.Errorf("goroutine %d: got %q, want %q", n, result, expected)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestConcurrentRenderWithSlots(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`
		<div data-g-define="slotted">
			<main data-g-define-slot="">default</main>
		</div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			content := fmt.Sprintf("content-%d", n)
			_, err := engine.RenderString(context.Background(), "slotted", nil,
				Slot("", content),
			)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", n, err)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// ================================================================================================
// Error scenarios
// ================================================================================================

func TestRenderNonExistentTemplate(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	_, err = engine.RenderString(context.Background(), "does-not-exist", nil)
	if err == nil {
		t.Fatal("Expected error for non-existent template, got nil")
	}

	var retrievalErr *TemplateRetrievalError
	if !errors.As(err, &retrievalErr) {
		t.Errorf("Expected TemplateRetrievalError, got %T: %v", err, err)
	}
}

func TestRenderWithNilData(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<div data-g-define="nil-data">Static content</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "nil-data", nil)
	if err != nil {
		t.Fatalf("Rendering with nil data should work for static templates: %v", err)
	}
	if result != "<div>Static content</div>" {
		t.Errorf("Unexpected output: %q", result)
	}
}

func TestWithLayoutNonExistentLayout(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<p data-g-define="page">Content</p>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	_, err = engine.RenderString(context.Background(), "page", nil,
		WithLayout("non-existent-layout", nil),
	)
	if err == nil {
		t.Fatal("Expected error for non-existent layout, got nil")
	}
}

func TestCloseInNonDevMode(t *testing.T) {
	engine, err := New() // non-dev mode by default
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Close should be a no-op in non-dev mode
	err = engine.Close()
	if err != nil {
		t.Errorf("Close() in non-dev mode should return nil, got: %v", err)
	}
}

func TestEmptyTemplate(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<div data-g-define="empty"></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "empty", nil)
	if err != nil {
		t.Fatalf("Rendering empty template should work: %v", err)
	}
	if result != "<div></div>" {
		t.Errorf("Unexpected output: %q", result)
	}
}

// ================================================================================================
// Custom implementations
// ================================================================================================

// mockResolver implements ModelPathResolver for testing.
// Only resolves expressions that look like model paths ([[ ... ]]).
type mockResolver struct {
	called bool
	value  any
}

func (m *mockResolver) TryResolve(expression string, data any) (any, bool) {
	if len(expression) >= 4 && expression[:2] == "[[" && expression[len(expression)-2:] == "]]" {
		m.called = true
		return m.value, true
	}
	return nil, false
}

func (m *mockResolver) Resolve(path string, data any) any {
	m.called = true
	return m.value
}

func TestCustomResolver(t *testing.T) {
	resolver := &mockResolver{value: "custom-resolved"}
	engine, err := New(WithCustomResolver(resolver))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<div data-g-define="test" data-g-inner-text="[[ .Anything ]]">placeholder</div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	result, err := engine.RenderString(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	if !resolver.called {
		t.Error("Custom resolver was not called")
	}
	if result != "<div>custom-resolved</div>" {
		t.Errorf("Expected custom resolver output, got: %q", result)
	}
}

// mockLogger implements Logger for testing
type mockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	errorMsgs []string
}

func (l *mockLogger) Debug(msg string, keysAndValues ...any) {
	l.debugMsgs = append(l.debugMsgs, msg)
}

func (l *mockLogger) Info(msg string, keysAndValues ...any) {
	l.infoMsgs = append(l.infoMsgs, msg)
}

func (l *mockLogger) Error(msg string, keysAndValues ...any) {
	l.errorMsgs = append(l.errorMsgs, msg)
}

func TestCustomLogger(t *testing.T) {
	logger := &mockLogger{}
	engine, err := New(WithLogger(logger))
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Trigger a template-not-found error to generate a log message
	_, _ = engine.RenderString(context.Background(), "nonexistent", nil)

	if len(logger.errorMsgs) == 0 {
		t.Error("Expected error log messages from rendering non-existent template")
	}
}

// ================================================================================================
// Condition evaluation with negation
// ================================================================================================

func TestConditionNegationTrue(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	err = engine.LoadHTML(`<div data-g-define="test"><span data-g-if="[[ .Show ]]">visible</span><span data-g-if="[[ .Hide ]]">hidden</span></div>`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	data := map[string]any{
		"Show": true,
		"Hide": false,
	}

	result, err := engine.RenderString(context.Background(), "test", data)
	if err != nil {
		t.Fatalf("RenderString() returned error: %v", err)
	}

	if result != "<div><span>visible</span></div>" {
		t.Errorf("Unexpected output: %q", result)
	}
}

// ================================================================================================
// Context cancellation in children
// ================================================================================================

func TestContextCancellationInChildren(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() { _ = engine.Close() }()

	// Template with many sibling children
	err = engine.LoadHTML(`
		<div data-g-define="many-children">
			<p>child1</p><p>child2</p><p>child3</p><p>child4</p><p>child5</p>
			<p>child6</p><p>child7</p><p>child8</p><p>child9</p><p>child10</p>
		</div>
	`)
	if err != nil {
		t.Fatalf("LoadHTML() returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = engine.RenderString(ctx, "many-children", nil)
	if err == nil {
		t.Error("Expected error from cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %T - %v", err, err)
	}
}
