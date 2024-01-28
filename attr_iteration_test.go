package gotmx

import (
	"context"
	"errors"
	"testing"
)

// ============================================================================
// getOuterRepeat Tests
// ============================================================================

func TestGetOuterRepeat_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-outer-repeat": "[[ .Items ]]",
	}

	path, found := getOuterRepeat(attrs)

	if !found {
		t.Error("Expected to find g-outer-repeat")
	}
	if path != "[[ .Items ]]" {
		t.Errorf("Expected '[[ .Items ]]', got '%s'", path)
	}
}

func TestGetOuterRepeat_ReturnsNotFoundWhenMissing(t *testing.T) {
	attrs := AttributeMap{}

	_, found := getOuterRepeat(attrs)

	if found {
		t.Error("Expected not to find g-outer-repeat")
	}
}

func TestGetOuterRepeat_CanonicalForm(t *testing.T) {
	attrs := AttributeMap{
		"data-g-outer-repeat": "[[ .DataItems ]]",
	}

	path, found := getOuterRepeat(attrs)

	if !found {
		t.Error("Expected to find attribute")
	}
	if path != "[[ .DataItems ]]" {
		t.Errorf("Expected '[[ .DataItems ]]', got '%s'", path)
	}
}

// ============================================================================
// getInnerRepeat Tests
// ============================================================================

func TestGetInnerRepeat_ReturnsValueWhenPresent(t *testing.T) {
	attrs := AttributeMap{
		"data-g-inner-repeat": "[[ .Items ]]",
	}

	path, found := getInnerRepeat(attrs)

	if !found {
		t.Error("Expected to find g-inner-repeat")
	}
	if path != "[[ .Items ]]" {
		t.Errorf("Expected '[[ .Items ]]', got '%s'", path)
	}
}

// ============================================================================
// resolveAsSlice Tests
// ============================================================================

func TestResolveAsSlice_HandlesNil(t *testing.T) {
	result := resolveAsSlice(nil)

	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestResolveAsSlice_HandlesSlice(t *testing.T) {
	input := []string{"a", "b", "c"}

	result := resolveAsSlice(input)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("Unexpected values: %v", result)
	}
}

func TestResolveAsSlice_HandlesArray(t *testing.T) {
	input := [3]int{1, 2, 3}

	result := resolveAsSlice(input)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("Unexpected values: %v", result)
	}
}

func TestResolveAsSlice_HandlesMap(t *testing.T) {
	input := map[string]int{
		"one": 1,
		"two": 2,
	}

	result := resolveAsSlice(input)

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}

	// Check that each item is a mapEntry
	for _, item := range result {
		entry, ok := item.(mapEntry)
		if !ok {
			t.Fatalf("Expected mapEntry, got %T", item)
		}
		key := entry.Key.(string)
		val := entry.Value.(int)
		if (key == "one" && val != 1) || (key == "two" && val != 2) {
			t.Errorf("Unexpected entry: key=%v, value=%v", key, val)
		}
	}
}

func TestResolveAsSlice_WrapsNonIterableInSlice(t *testing.T) {
	input := "single value"

	result := resolveAsSlice(input)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0] != "single value" {
		t.Errorf("Expected 'single value', got '%v'", result[0])
	}
}

func TestResolveAsSlice_HandlesEmptySlice(t *testing.T) {
	input := []string{}

	result := resolveAsSlice(input)

	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %d", len(result))
	}
}

func TestResolveAsSlice_HandlesIntSlice(t *testing.T) {
	input := []int{10, 20, 30}

	result := resolveAsSlice(input)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != 10 || result[1] != 20 || result[2] != 30 {
		t.Errorf("Unexpected values: %v", result)
	}
}

func TestResolveAsSlice_HandlesInterfaceSlice(t *testing.T) {
	input := []any{"string", 42, true}

	result := resolveAsSlice(input)

	if len(result) != 3 {
		t.Errorf("Expected 3 items, got %d", len(result))
	}
	if result[0] != "string" || result[1] != 42 || result[2] != true {
		t.Errorf("Unexpected values: %v", result)
	}
}

// ============================================================================
// iterateValue Tests
// ============================================================================

func TestIterateValue_HandlesNil(t *testing.T) {
	var count int
	err := iterateValue(&RenderContext{Context: context.Background()}, nil, func(item any) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 iterations, got %d", count)
	}
}

func TestIterateValue_IteratesSlice(t *testing.T) {
	input := []string{"a", "b", "c"}
	var items []string

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		items = append(items, item.(string))
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

func TestIterateValue_IteratesArray(t *testing.T) {
	input := [3]int{1, 2, 3}
	var sum int

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		sum += item.(int)
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if sum != 6 {
		t.Errorf("Expected sum=6, got %d", sum)
	}
}

func TestIterateValue_IteratesMap(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2}
	var entries []mapEntry

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		entries = append(entries, item.(mapEntry))
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestIterateValue_CallsOnceForNonIterable(t *testing.T) {
	input := "single"
	var count int
	var received any

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		count++
		received = item
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 call, got %d", count)
	}
	if received != "single" {
		t.Errorf("Expected 'single', got '%v'", received)
	}
}

func TestIterateValue_StopsOnError(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	var count int
	testErr := errors.New("test error")

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		count++
		if count == 3 {
			return testErr
		}
		return nil
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if count != 3 {
		t.Errorf("Expected to stop at 3, got %d", count)
	}
}

func TestIterateValue_StopsOnErrorForMap(t *testing.T) {
	input := map[string]int{"a": 1, "b": 2, "c": 3}
	var count int
	testErr := errors.New("test error")

	err := iterateValue(&RenderContext{Context: context.Background()}, input, func(item any) error {
		count++
		return testErr // Stop on first
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if count != 1 {
		t.Errorf("Expected to stop at 1, got %d", count)
	}
}

// ============================================================================
// resolveIterable Tests
// ============================================================================

func TestResolveIterable_ResolvesModelPath(t *testing.T) {
	ctx := newTestRenderContext()
	data := map[string]any{
		"Items": []string{"a", "b"},
	}

	result, ok := resolveIterable(ctx, "[[ .Items ]]", data)

	if !ok {
		t.Error("Expected ok to be true")
	}
	slice, isSlice := result.([]string)
	if !isSlice {
		t.Fatalf("Expected []string, got %T", result)
	}
	if len(slice) != 2 {
		t.Errorf("Expected 2 items, got %d", len(slice))
	}
}

func TestResolveIterable_ReturnsFalseForInvalidPath(t *testing.T) {
	ctx := newTestRenderContext()
	data := map[string]any{}

	_, ok := resolveIterable(ctx, "not a model path", data)

	if ok {
		t.Error("Expected ok to be false for non-model-path")
	}
}

// ============================================================================
// mapEntry Tests
// ============================================================================

func TestMapEntry_HoldsKeyAndValue(t *testing.T) {
	entry := mapEntry{Key: "myKey", Value: 42}

	if entry.Key != "myKey" {
		t.Errorf("Expected key 'myKey', got '%v'", entry.Key)
	}
	if entry.Value != 42 {
		t.Errorf("Expected value 42, got '%v'", entry.Value)
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestIterateValue_StopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := []string{"a", "b", "c", "d", "e"}
	var count int

	err := iterateValue(&RenderContext{Context: ctx}, input, func(item any) error {
		count++
		return nil
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
	// Should have stopped before processing all items
	if count > 1 {
		t.Errorf("Expected at most 1 iteration before cancellation check, got %d", count)
	}
}

func TestIterateValue_StopsOnCancelledContextForMap(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	input := map[string]any{"a": 1, "b": 2, "c": 3}
	var count int

	err := iterateValue(&RenderContext{Context: ctx}, input, func(item any) error {
		count++
		return nil
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
	if count > 1 {
		t.Errorf("Expected at most 1 iteration before cancellation check, got %d", count)
	}
}

func TestIterateValue_StopsOnCancelledContextForReflection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Use a custom struct slice to force the reflection path
	type item struct{ Name string }
	input := []item{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	var count int

	err := iterateValue(&RenderContext{Context: ctx}, input, func(item any) error {
		count++
		return nil
	})

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
	if count > 1 {
		t.Errorf("Expected at most 1 iteration before cancellation check, got %d", count)
	}
}
