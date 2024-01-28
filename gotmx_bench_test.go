package gotmx

import (
	"context"
	"io"
	"testing"
)

func setupBenchEngine(b *testing.B) *Engine {
	b.Helper()
	tr := NewTemplateRegistryDefault()
	tl := newTemplateLoaderString(tr)
	tr.SetLazyTemplateLoader(tl)
	e, err := New(WithCustomRegistry(tr))
	if err != nil {
		b.Fatal(err)
	}

	// Simple template
	if err := tl.LoadFromString(`<div data-g-define="simple"><span data-g-inner-text="[[ .Name ]]">placeholder</span></div>`, ""); err != nil {
		b.Fatal(err)
	}

	// Template with multiple attributes
	if err := tl.LoadFromString(`<div data-g-define="attrs" class="container" id="main" data-role="content"><span data-g-inner-text="[[ .Name ]]">placeholder</span></div>`, ""); err != nil {
		b.Fatal(err)
	}

	// Template with repeat
	if err := tl.LoadFromString(`<ul data-g-define="repeat"><li data-g-inner-repeat="[[ .Items ]]" data-g-inner-text="[[ . ]]">item</li></ul>`, ""); err != nil {
		b.Fatal(err)
	}

	// Template with slots
	if err := tl.LoadFromString(`<div data-g-define="with-slots"><header data-g-define-slot="header">Default</header><main data-g-define-slot="">Content</main></div>`, ""); err != nil {
		b.Fatal(err)
	}

	// Template with g-use composition
	if err := tl.LoadFromString(`<div data-g-define="composed"><div data-g-use="simple"></div></div>`, ""); err != nil {
		b.Fatal(err)
	}

	// Template with conditional
	if err := tl.LoadFromString(`<div data-g-define="conditional"><span data-g-if="[[ .Visible ]]" data-g-inner-text="[[ .Name ]]">placeholder</span></div>`, ""); err != nil {
		b.Fatal(err)
	}

	return e
}

func BenchmarkRenderSimple(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "simple", data)
	}
}

func BenchmarkRenderString(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.RenderString(ctx, "simple", data)
	}
}

func BenchmarkRenderWithAttributes(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "attrs", data)
	}
}

func BenchmarkRenderWithRepeat(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	items := make([]string, 100)
	for i := range items {
		items[i] = "item"
	}
	data := map[string]any{"Items": items}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "repeat", data)
	}
}

func BenchmarkRenderWithSlots(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "with-slots", nil,
			Slot("header", "Custom Header"))
	}
}

func BenchmarkRenderComposed(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "composed", data)
	}
}

func BenchmarkRenderConditional(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World", "Visible": true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "conditional", data)
	}
}

func BenchmarkRenderWithRepeat1000(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	items := make([]string, 1000)
	for i := range items {
		items[i] = "item"
	}
	data := map[string]any{"Items": items}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "repeat", data)
	}
}

func BenchmarkRenderDeepNesting(b *testing.B) {
	tr := NewTemplateRegistryDefault()
	tl := newTemplateLoaderString(tr)
	tr.SetLazyTemplateLoader(tl)
	e, err := New(WithCustomRegistry(tr))
	if err != nil {
		b.Fatal(err)
	}

	// Create a 5-level deep nesting chain
	if err := tl.LoadFromString(`
		<div data-g-define="level1"><div data-g-use="level2"></div></div>
		<div data-g-define="level2"><div data-g-use="level3"></div></div>
		<div data-g-define="level3"><div data-g-use="level4"></div></div>
		<div data-g-define="level4"><div data-g-use="level5"></div></div>
		<div data-g-define="level5"><span data-g-inner-text="[[ .Name ]]">placeholder</span></div>
	`, ""); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Render(ctx, io.Discard, "level1", data)
	}
}

func BenchmarkRenderStringPooling(b *testing.B) {
	e := setupBenchEngine(b)
	ctx := context.Background()
	data := map[string]any{"Name": "World"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			e.RenderString(ctx, "simple", data)
		}
	})
}

func BenchmarkIterateSlice(b *testing.B) {
	items := make([]any, 100)
	for i := range items {
		items[i] = i
	}
	ctx := &RenderContext{Context: context.Background()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iterateValue(ctx, items, func(item any) error {
			return nil
		})
	}
}

func BenchmarkIterateStringSlice(b *testing.B) {
	items := make([]string, 100)
	for i := range items {
		items[i] = "item"
	}
	ctx := &RenderContext{Context: context.Background()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iterateValue(ctx, items, func(item any) error {
			return nil
		})
	}
}

func BenchmarkIterateMap(b *testing.B) {
	m := map[string]any{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	ctx := &RenderContext{Context: context.Background()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iterateValue(ctx, m, func(item any) error {
			return nil
		})
	}
}
