# Gotmx

[![Go Reference](https://pkg.go.dev/badge/github.com/authentic-devel/gotmx.svg)](https://pkg.go.dev/github.com/authentic-devel/gotmx)
[![CI](https://github.com/authentic-devel/gotmx/actions/workflows/ci.yml/badge.svg)](https://github.com/authentic-devel/gotmx/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/authentic-devel/gotmx)](https://goreportcard.com/report/github.com/authentic-devel/gotmx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


Gotmx is a component-based HTML template engine for Go that keeps your templates as valid HTML.

- **Plain HTML templates**: Use standard `data-g-*` attributes instead of custom syntax.
- **Browser-previewable**: Open templates directly in your browser to see structure and styling.
- **Composable components**: Build complex UIs from small, reusable pieces using slots.
- **Interoperable**: Mix with Go's `text/template` and `html/template` where needed.
- **No build step**: Works with any HTML editor, no special tooling required.

Gotmx is ideal for server-side rendering, HTMX-enhanced applications, HTML emails, or any scenario where you want to author templates as real HTML.

## Quick Start

Install Gotmx:

```shell
go get github.com/authentic-devel/gotmx
```

### Minimal Example

Create an HTML template file `templates/hello.htm` (`.htm` for eager loading at startup):

```html
<div data-g-define="hello">
  <span data-g-inner-text="[[ .Name ]]">placeholder</span>
</div>
```

Render it from Go:

```go
package main

import (
    "context"
    "fmt"
    "github.com/authentic-devel/gotmx"
)

func main() {
    // Create engine and load templates from directory
    engine, err := gotmx.New(
        gotmx.WithTemplateDir("./templates"),
    )
    if err != nil {
        panic(err)
    }
    defer engine.Close()

    // Render template with data
    data := map[string]string{"Name": "World"}
    result, err := engine.RenderString(context.Background(), "hello", data)
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
    // Output: <div><span>World</span></div>
}
```

### Inline Templates

You can also load templates directly from strings:

```go
engine, _ := gotmx.New()
engine.LoadHTML(`<div data-g-define="greeting">Hello, <span data-g-inner-text="[[ .Name ]]"></span>!</div>`)

result, _ := engine.RenderString(context.Background(), "greeting", map[string]string{"Name": "Gotmx"})
```

### Dev Mode with Auto-Reload

For development, enable file watching to automatically reload templates on changes:

```go
engine, _ := gotmx.New(
    gotmx.WithTemplateDir("./templates"),
    gotmx.WithDevMode(true),
)
defer engine.Close() // Important: stops file watchers
```

## Philosophy and Goals

Gotmx was created to make server-side rendering easy and ergonomic, especially for HTMX-enhanced applications. While Go's built-in templates are powerful, they use non-HTML syntax that breaks editor tooling and prevents browser preview.

### Design Principles

**Zero additional tooling**: Templates are plain HTML with data attributes. Any HTML editor works. No plugins, no build step.

**Browser-previewable**: Open your templates directly in a browser to see layout and styling. The `data-g-*` attributes are ignored by browsers, so the HTML renders normally with placeholder content.

**Progressive enhancement**: Start with static HTML for design and styling, then add template directives incrementally where you need dynamic behavior.

**Composition over inheritance**: Build complex UIs by composing small components. Use slots to inject content rather than complex inheritance chains.

**Separation of concerns**: Templates define structure and presentation. Business logic stays in Go code. Keep templates declarative and simple.

**No lock-in**: Gotmx coexists with Go's native templates. Use Gotmx where it fits and Go templates where they make more sense. You can even mix them in the same project.

**Minimal dependencies**: Gotmx avoids framework assumptions. Use it with Gin, Echo, Chi, net/http, or no framework at all.

## Browser-Previewable Templates

Because gotmx uses HTML attributes, templates remain valid HTML that you can open directly in a browser. Sample content stays visible in preview but gets replaced at runtime:

```html
<div data-g-define="user-list">
    <h2 data-g-inner-text="[[ .Title ]]">Team Members</h2>
    <ul>
        <!-- First item is the template — repeated for each user at runtime -->
        <li data-g-outer-repeat="[[ .Users ]]">
            <strong data-g-inner-text="[[ .Name ]]">Alice Johnson</strong> —
            <span data-g-inner-text="[[ .Role ]]">Engineer</span>
        </li>
        <!-- Extra items are preview-only, stripped at runtime -->
        <li data-g-ignore="outer">
            <strong>Bob Smith</strong> — <span>Designer</span>
        </li>
        <li data-g-ignore="outer">
            <strong>Carol Lee</strong> — <span>Product Manager</span>
        </li>
    </ul>
</div>
```

Open this file in a browser and you see a team list with three members. At runtime, the ignored items disappear and the first `<li>` repeats for each actual user.

Other preview techniques include:
- **Placeholder text** in `data-g-inner-text` elements (visible in preview, replaced at runtime)
- **Static `src`/`href`** with `data-g-src`/`data-g-href` overrides (relative paths for preview, absolute for server)
- **Slot default content** (shows layout structure in preview, replaced by caller at runtime)
- **Full HTML scaffolding** around `data-g-define` elements (the `<html>`, `<head>`, `<body>` wrapper is ignored at runtime)

See [Browser Preview Guide](./docs/browser-preview.md) for all techniques with examples.

## Accessing Data from Your Model

Gotmx provides a fast path syntax for accessing model properties: `[[ .Path ]]`. This uses square brackets instead of Go template's curly braces and is optimized for simple property access.

```html
<!-- Simple property access -->
<span data-g-inner-text="[[ .User.Name ]]"></span>

<!-- Nested properties -->
<span data-g-inner-text="[[ .Order.Customer.Email ]]"></span>

<!-- Array indexing -->
<span data-g-inner-text="[[ .Items[0].Title ]]"></span>

<!-- String concatenation -->
<span data-g-inner-text='[[ "Hello, " .Name "!" ]]'></span>

<!-- Current data context -->
<span data-g-inner-text="[[ . ]]"></span>
```

The `[[ ]]` syntax supports:
- Property access on structs and maps
- Array/slice indexing
- String literals (in single or double quotes)
- Concatenation of multiple values
- Simple comparisons and negation

For complex expressions, use Go templates with `data-g-as-template`. See [Working with Go Templates](./docs/golang-templates.md).

The path resolution is powered by the [empaths](https://github.com/authentic-devel/empaths) library.

## Slots: Component Composition

Slots let you define named injection points in your templates. This enables true component composition where the parent decides what content to inject.

### Defining Slots

Use `data-g-define-slot` to mark where content can be injected:

```html
<div data-g-define="card">
    <div class="card-header" data-g-define-slot="header">
        Default Header (preview only)
    </div>
    <div class="card-body" data-g-define-slot="">
        Default content (preview only)
    </div>
    <div class="card-footer" data-g-define-slot="footer">
        Default Footer (preview only)
    </div>
</div>
```

An empty slot name (`data-g-define-slot=""`) creates the default slot for content without an explicit slot assignment.

**Note:** The text inside slot elements is for browser preview only. At render time, slots display the injected content or nothing — the preview text is never rendered.

### Filling Slots

Use `data-g-use-slot` when calling a component to direct content to specific slots:

```html
<div data-g-use="card">
    <h2 data-g-use-slot="header">My Custom Title</h2>
    <p>This paragraph goes to the default slot.</p>
    <button data-g-use-slot="footer">Action Button</button>
</div>
```

Children without `data-g-use-slot` are placed in the default slot.

### Why Slots Matter

Slots invert the dependency direction. Instead of a layout template deciding which content template to include, the calling code decides what to inject. This makes components truly independent and reusable.

For HTMX applications, this is particularly useful: you can render a full page (layout + content) for initial requests, or just the content component for HTMX partial updates.

## Building Blocks Overview

Gotmx has a layered architecture:

**Engine**: The single entry point for using gotmx. Handles template loading, dev mode, and provides convenient render methods.

```go
engine, _ := gotmx.New(gotmx.WithTemplateDir("templates"))
engine.Render(ctx, w, "my-template", data)
```

**TemplateRegistry**: Stores templates by name and namespace. Supports both eager loading (at startup) and lazy loading (on first use). Pluggable via `WithCustomRegistry()`.

**Template/Renderable**: Core interfaces. A `Template` is a factory that creates `Renderable` components bound to data.

**ModelPathResolver**: Resolves `[[ .Path ]]` expressions. Uses empaths by default but can be replaced via `WithCustomResolver()`.

For detailed architecture information, see [Architecture](./docs/architecture.md).

## Go Template Integration

Gotmx is not a replacement for Go templates. It complements them. You can use Go templates anywhere Gotmx templates alone are insufficient.

### Inline Go Templates

Add `data-g-as-template` to treat an element's content as a Go HTML template:

```html
<ul data-g-as-template>
    {{ range .Items }}
        <li>{{ .Name }}</li>
    {{ end }}
</ul>
```

Use `data-g-as-unsafe-template` for text templates (no automatic HTML escaping).

### Calling Gotmx from Go Templates

Within Go templates, use `GTemplate` to render Gotmx templates:

```html
<div data-g-as-template>
    {{ GTemplate "user-card" .User }}
</div>
```

See [Working with Go Templates](./docs/golang-templates.md) for details.

## Template Directives

Gotmx uses HTML attributes to control template behavior. All attributes work in both short (`g-*`) and long (`data-g-*`) forms.

**Template Definition:**
- `data-g-define="name"` - Define a reusable template

**Control Flow:**
- `data-g-if="[[ .Condition ]]"` - Conditional rendering
- `data-g-with="[[ .Object ]]"` - Switch data context
- `data-g-ignore` - Skip element or children

**Iteration:**
- `data-g-outer-repeat="[[ .Items ]]"` - Repeat entire element
- `data-g-inner-repeat="[[ .Items ]]"` - Repeat children only

**Content:**
- `data-g-inner-text="[[ .Text ]]"` - Set text content (escaped)
- `data-g-inner-html="[[ .Html ]]"` - Set HTML content (unescaped)
- `data-g-outer-text="[[ .Text ]]"` - Replace element with text

**Composition:**
- `data-g-use="template"` - Render a different template
- `data-g-inner-use="template"` - Render a template's inner content only
- `data-g-define-slot="name"` - Define a slot location
- `data-g-use-slot="name"` - Fill a named slot
- `data-g-override-att="class,id"` - Pass attributes to component

**Attributes:**
- `data-g-class="[[ .Class ]]"` - Set class attribute
- `data-g-href="[[ .Url ]]"` - Set href attribute
- `data-g-src="[[ .Url ]]"` - Set src attribute
- `data-g-att-*="[[ .Value ]]"` - Set any attribute dynamically
- `data-g-attif-*="[[ .Condition ]]"` - Conditionally add attribute
- `data-g-trans="tagname"` - Transform element tag name

See [Attribute Reference](./docs/attribute-reference.md) for complete documentation.

## Layout Composition

Use `WithLayout` to wrap any template inside a layout in a single render call:

```go
engine.Render(ctx, w, "dashboard-page", pageData,
    gotmx.WithLayout("main-layout", layoutData),
)
```

The rendered page is placed into the layout's default slot. Use `WithLayoutSlot` to target a named slot:

```go
engine.Render(ctx, w, "dashboard-page", pageData,
    gotmx.WithLayout("main-layout", layoutData),
    gotmx.WithLayoutSlot("content"),
)
```

This is especially useful for HTMX applications where you render just the component for HTMX requests and the full page with layout for initial loads:

```go
if isHxRequest(r) {
    engine.Render(r.Context(), w, "page", data)
} else {
    engine.Render(r.Context(), w, "page", data,
        gotmx.WithLayout("layout", layoutData),
    )
}
```

## Rendering Behavior

**Attribute order**: By default, attributes render in map iteration order for performance. Enable `WithDeterministicOutput(true)` for sorted, reproducible output (useful for testing).

**Boolean attributes**: HTML boolean attributes like `disabled`, `checked`, `hidden`, `required`, `readonly`, and `selected` render without a value when present (e.g., `<button disabled>` instead of `<button disabled="true">`).

**HTML escaping**: All text content and attribute values are HTML-escaped by default to prevent XSS. Use `data-g-inner-html` only for trusted content.

**Whitespace**: Some whitespace may not be preserved exactly due to HTML parsing normalization.

## Customization

Gotmx components can be replaced with custom implementations:

```go
engine, _ := gotmx.New(
    gotmx.WithCustomRegistry(myRegistry),    // Custom template storage
    gotmx.WithCustomResolver(myResolver),    // Custom path resolution
    gotmx.WithLogger(slog.Default()),        // Custom logging
    gotmx.WithMaxNestingDepth(128),          // Max template nesting depth
)
```

### Max Nesting Depth (Circular Reference Protection)

Gotmx protects against circular template references that could cause stack overflow. When template A uses template B, and B uses A, this creates an infinite loop. Gotmx detects this by limiting the nesting depth of `g-use` calls.

The default limit is 64 levels, which is sufficient for complex component hierarchies. You can configure this:

```go
// Allow deeper nesting for very complex hierarchies
engine, _ := gotmx.New(gotmx.WithMaxNestingDepth(128))

// Use a stricter limit
engine, _ := gotmx.New(gotmx.WithMaxNestingDepth(32))

// Disable the limit (not recommended)
engine, _ := gotmx.New(gotmx.WithMaxNestingDepth(0))
```

If the limit is exceeded, rendering fails with a `MaxNestingDepthExceededError` that includes the template name and current depth, making it easy to diagnose circular references.

See [Customization Guide](./docs/customization.md).

## Performance

Gotmx prioritizes developer experience over raw performance. That said, it uses:
- Streaming output to `io.Writer` (no intermediate string allocation for large outputs)
- Buffered writing via `bufio.Writer` to batch many small write calls efficiently
- Buffer pooling (`sync.Pool`) for `RenderString` to reduce GC pressure
- Fast path resolution via empaths
- Type-switch fast paths for iteration (avoids reflection for common types)
- Lazy template loading to reduce startup time
- Context cancellation support to stop rendering when clients disconnect

For most server-side rendering use cases, performance is more than adequate.

## Security

Gotmx HTML-escapes all text content and attribute values by default to prevent XSS attacks. Characters like `<`, `>`, `&`, and `"` are converted to HTML entities.

**Unconditionally safe (always escaped, even with `Unescaped()`):**
- `data-g-inner-text` - Always escaped for XSS safety
- Attribute values - Always escaped

**Safe by default (follows global escaping setting):**
- `data-g-outer-text` - Escaped by default, respects `Unescaped()` option
- Text nodes - Escaped by default, respects `Unescaped()` option

**Unsafe (never escaped):**
- `data-g-inner-html` - Use only with trusted content
- `data-g-as-unsafe-template` - Use only with trusted templates

Always validate and sanitize user input in your Go code before passing it to templates.

## Further Reading

- [Quick Start](./docs/quick-start.md) - Get up and running in 5 minutes
- [Browser Preview Guide](./docs/browser-preview.md) - Techniques for preview-friendly templates
- [Attribute Reference](./docs/attribute-reference.md) - Complete attribute documentation
- [Common Patterns](./docs/common-patterns.md) - Recipes for typical use cases
- [Integration Patterns](./docs/integration-patterns.md) - Framework and HTMX integration
- [Architecture](./docs/architecture.md) - Detailed design and internals
- [Go Templates](./docs/golang-templates.md) - Integration with Go's template packages
- [Customization](./docs/customization.md) - Extending and customizing Gotmx

## License

MIT License - see LICENSE file for details.
