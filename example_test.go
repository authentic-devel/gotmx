package gotmx_test

import (
	"context"
	"fmt"
	"os"

	"github.com/authentic-devel/gotmx"
)

func ExampleEngine_Render() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<div data-g-define="greeting" data-g-inner-text="[[ .Name ]]">placeholder</div>`)

	engine.Render(context.Background(), os.Stdout, "greeting", map[string]any{"Name": "World"})
	// Output: <div>World</div>
}

func ExampleEngine_RenderString() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<span data-g-define="badge" data-g-inner-text="[[ .Label ]]">x</span>`)

	result, _ := engine.RenderString(context.Background(), "badge", map[string]any{"Label": "New"})
	fmt.Println(result)
	// Output: <span>New</span>
}

func ExampleEngine_Render_iteration() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<ul data-g-define="list"><li data-g-outer-repeat="[[ .Items ]]" data-g-inner-text="[[ . ]]">placeholder</li></ul>`)

	data := map[string]any{
		"Items": []string{"Alice", "Bob", "Carol"},
	}
	engine.Render(context.Background(), os.Stdout, "list", data)
	// Output: <ul><li>Alice</li><li>Bob</li><li>Carol</li></ul>
}

func ExampleEngine_Render_conditional() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<div data-g-define="status"><span data-g-if="[[ .Active ]]">Active</span><span data-g-if="[[ .Inactive ]]">Inactive</span></div>`)

	engine.Render(context.Background(), os.Stdout, "status", map[string]any{"Active": true, "Inactive": false})
	// Output: <div><span>Active</span></div>
}

func ExampleEngine_Render_composition() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`
		<button data-g-define="btn" class="btn"><span data-g-define-slot="">Click</span></button>
		<div data-g-define="page"><div data-g-use="btn">Submit</div></div>
	`)

	engine.Render(context.Background(), os.Stdout, "page", nil)
	// Output: <div><button class="btn"><span>Submit</span></button></div>
}

func ExampleEngine_RenderString_withSlots() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<div data-g-define="card"><header data-g-define-slot="header">Default</header><main data-g-define-slot="">Body</main></div>`)

	result, _ := engine.RenderString(context.Background(), "card", nil,
		gotmx.Slot("header", "<h1>Title</h1>"),
		gotmx.Slot("", "<p>Content</p>"),
	)
	fmt.Println(result)
	// Output: <div><header><h1>Title</h1></header><main><p>Content</p></main></div>
}

func ExampleEngine_Render_withLayout() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`
		<div data-g-define="layout"><nav>Menu</nav><main data-g-define-slot="">default</main></div>
		<p data-g-define="page">Hello from page</p>
	`)

	engine.Render(context.Background(), os.Stdout, "page", nil,
		gotmx.WithLayout("layout", nil),
	)
	// Output: <div><nav>Menu</nav><main><p>Hello from page</p></main></div>
}

func ExampleEngine_Render_escaping() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<div data-g-define="safe" data-g-inner-text="[[ .Input ]]">placeholder</div>`)

	// g-inner-text always escapes HTML, protecting against XSS
	engine.Render(context.Background(), os.Stdout, "safe", map[string]any{
		"Input": `<script>alert("xss")</script>`,
	})
	// Output: <div>&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;</div>
}

func ExampleUnescaped() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`<div data-g-define="raw" data-g-inner-html="[[ .Html ]]">placeholder</div>`)

	// Use g-inner-html for trusted HTML content
	engine.Render(context.Background(), os.Stdout, "raw", map[string]any{
		"Html": "<strong>Bold</strong>",
	})
	// Output: <div><strong>Bold</strong></div>
}

func ExampleWithLayout() {
	engine, _ := gotmx.New()
	engine.LoadHTML(`
		<html data-g-define="base"><head><title data-g-inner-text="[[ .Title ]]">T</title></head><body data-g-define-slot="">default</body></html>
		<article data-g-define="page">Welcome!</article>
	`)

	engine.Render(context.Background(), os.Stdout, "page", nil,
		gotmx.WithLayout("base", map[string]any{"Title": "Home"}),
	)
	// Output: <!DOCTYPE html>
	// <html><head><title>Home</title></head><body><article>Welcome!</article></body></html>
}
