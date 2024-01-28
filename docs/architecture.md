# Gotmx Architecture

This document explains the architecture of Gotmx, its core building blocks, design philosophy, and how components interact during rendering.

## Design Philosophy

Gotmx is a component-based HTML template engine for Go that prioritizes simplicity, composability, and progressive enhancement.

### Core Principles

1. **HTML-first approach**: Templates are valid HTML. All directives use standard HTML attributes (`g-*` or `data-g-*`), so your templates can be opened directly in a browser and edited with standard HTML tooling.

2. **Progressive enhancement**: Start with static HTML, then add template directives incrementally where needed. You can design in the browser first and add dynamic behavior later.

3. **Composition over inheritance**: Build complex UIs by composing small, reusable components. Use slots to inject content rather than template inheritance chains.

4. **Separation of concerns**: Templates define structure and presentation; data and logic live in Go code. This keeps templates declarative and testable.

5. **Interoperability**: Gotmx coexists with native Go `text/template` and `html/template`. You can mix both approaches in the same project or even the same file.

6. **Modularity**: Core components (registry, model resolver, logger) are interfaces that can be swapped for custom implementations.

## High-Level Flow

1. **Parse**: HTML files are parsed to discover template definitions (elements with `g-define`).
2. **Register**: Templates are stored in a `TemplateRegistry` by name and namespace.
3. **Create**: At render time, a template creates a `Renderable` component bound to your data.
4. **Render**: The component tree renders to an `io.Writer`, resolving data bindings and processing directives.

```
HTML Files                    Data
    │                          │
    ▼                          ▼
TemplateRegistry ──► Template.NewRenderable(data) ──► Renderable
                                                            │
                                                            ▼
                                                       io.Writer
```

## Entry Points

`Engine` is the single entry point for all use cases. It handles template loading, dev mode with file watching, and provides a clean interface:

```go
engine, err := gotmx.New(
    gotmx.WithTemplateDir("templates"),
    gotmx.WithDevMode(true),
)
defer engine.Close()

err = engine.Render(ctx, w, "my-template", data)
```

Key methods:
- `Render(ctx, w, templateName, data, ...opts)` - Render to writer with HTML escaping
- `RenderString(ctx, templateName, data, ...opts)` - Render to string
- `Component(templateName, data)` - Get a Renderable for manual rendering
- `MustComponent(templateName, data)` - Like Component, panics on error
- `HasTemplate(name)` - Check if a template exists
- `LoadHTML(html)` - Parse templates from an HTML string
- `LoadFile(path)` - Load templates from a file
- `RegisterFunc(name, fn)` - Register a function for use in Go templates
- `Close()` - Release resources (file watchers in dev mode)

Render options:
- `Unescaped()` - Disable HTML escaping for text nodes and g-outer-text
- `WithLayout(layoutTemplate, layoutData)` - Wrap output in a layout template
- `WithLayoutSlot(name)` - Target a named slot in the layout
- `WithSlots(slots)` - Provide slot content programmatically
- `Slot(name, content)` - Convenience for a single slot

For advanced customization, Engine supports custom registries and resolvers via options:

```go
engine, err := gotmx.New(
    gotmx.WithCustomRegistry(myRegistry),
    gotmx.WithCustomResolver(myResolver),
    gotmx.WithLogger(logger),
)
```

Pluggable components:
- `TemplateRegistry` - Where templates are stored
- `ModelPathResolver` - Resolves `[[ .Path ]]` expressions
- `Logger` - For diagnostics (slog-compatible interface)

## Core Building Blocks

### Template and Renderable

These are the two fundamental interfaces:

**Template** - A factory that can create renderables:
- `Name()` - Returns the template name (e.g., `"button"`)
- `Namespace()` - Returns the source file path (e.g., `"components/button.html"`)
- `NewRenderable(data)` - Creates a Renderable bound to data

**Renderable** - An instance that can render itself:
- `Render(ctx, writer, renderType)` - Renders to the writer using a RenderContext. Escaping is controlled by `ctx.Escaped`.

### Template Types

Gotmx supports multiple template implementations:

- **NodeTemplate/NodeComponent** - HTML elements with `g-*` directives (the primary type)
- **GolangTemplate** - Native Go templates wrapped for use in Gotmx
- **StringLiteralTemplate** - Plain text/HTML strings that render as-is

### TemplateRegistry

The registry stores and retrieves templates. The default implementation (`TemplateRegistryDefault`) provides:

- Storage by name within namespaces
- Lookup by simple name (`"button"`) or fully-qualified name (`"components/button.html#button"`)
- Support for Go template registration and function registration
- Integration with lazy template loaders

### ModelPathResolver

Resolves data binding expressions in the form `[[ .Path ]]`. The default implementation uses the [empaths](https://github.com/authentic-devel/empaths) library for type-safe path resolution.

Supported expressions:
- Property access: `[[ .User.Name ]]`
- Nested paths: `[[ .User.Profile.Email ]]`
- Array indexing: `[[ .Items[0] ]]`
- Map access: `[[ .Settings.Theme ]]`
- Current data: `[[ . ]]` or `[[ "" ]]`

## Template Loading

### Two-Phase Loading Strategy

Gotmx uses a two-phase loading approach:

1. **Eager loading** (`.htm` files by default): Loaded at startup during `Init()`. Use for frequently-accessed templates.

2. **Lazy loading** (`.html` files by default): Loaded on first access. Use for large template sets where startup time matters.

Configure with:
```go
gotmx.WithEagerExtensions(".htm", ".htmx")
gotmx.WithLazyExtensions(".html")
```

### Template Sources

Templates can be loaded from multiple sources:

- **Embedded filesystems**: Use `WithFS(embedFS)` for production deployments
- **Directories**: Use `WithTemplateDir("path")` for file-based loading
- **Multiple sources**: Call `WithFS()` or `WithTemplateDir()` multiple times

### Dev Mode

Enable dev mode for automatic template reloading when files change:

```go
gotmx.WithDevMode(true)
```

In dev mode, Gotmx watches template directories and automatically reloads modified files. Call `engine.Close()` to stop watchers.

## Names, Namespaces, and Lookup

Templates have a **name** and a **namespace**:

- **Name**: The template identifier (from `g-define` attribute or `id` attribute)
- **Namespace**: The source file path (e.g., `"components/buttons.html"`)

### Template References

A `TemplateRef` can be:
- **Simple**: `"button"` - Must be globally unique across all namespaces
- **Fully-qualified**: `"components/buttons.html#button"` - Uses `#` to specify namespace

If two files define templates with the same name, use fully-qualified references to disambiguate:
```html
<!-- In components/primary.html -->
<button g-define="button">Primary</button>

<!-- In components/secondary.html -->
<button g-define="button">Secondary</button>

<!-- Usage -->
<div g-use="components/primary.html#button"></div>
<div g-use="components/secondary.html#button"></div>
```

## Slots and Composition

Slots enable component composition by allowing content injection:

### Defining Slots

Use `g-define-slot` to mark where content should be inserted:

```html
<div g-define="card">
    <div class="card-header">
        <slot g-define-slot="header" g-ignore></slot>
    </div>
    <div class="card-body">
        <slot g-define-slot g-ignore></slot>  <!-- Default slot (empty name) -->
    </div>
</div>
```

### Filling Slots

Use `g-use-slot` to direct content to specific slots:

```html
<div g-use="card">
    <h2 g-use-slot="header">My Title</h2>
    <p>This goes to the default slot</p>
</div>
```

### Attribute Passing

Use `g-override-att` to pass attributes from parent to component:

```html
<button g-define="btn" class="btn">Click me</button>

<!-- Usage: passes class to override the component's class -->
<div g-use="btn" g-override-att="class" class="btn-primary"></div>
<!-- Result: <button class="btn-primary">Click me</button> -->
```

## Attribute Syntax

Gotmx supports two attribute forms:

- **Short form**: `g-if`, `g-use`, `g-class`, etc.
- **Long form**: `data-g-if`, `data-g-use`, `data-g-class`, etc.

Both are equivalent. Use long form for stricter HTML validation. If both are present on an element, the long form takes precedence.

All `g-*` and `data-g-*` attributes are removed from the output. They are control directives, not rendered attributes.

## Go Template Integration

Gotmx integrates with native Go templates in two ways:

### Inline Go Templates

Use `g-as-template` or `g-as-unsafe-template` to treat element content as a Go template:

```html
<script type="text/template" g-as-template="my-go-template">
    Hello, {{.Name}}!
</script>
```

### Calling Templates

From Go templates, use the built-in functions to call Gotmx templates:

- `GTemplate` - Renders with HTML escaping
- `GTextTemplate` - Renders without HTML escaping

```html
<div g-as-template="wrapper">
    {{ GTemplate "button" .ButtonData }}
</div>
```

## Rendering Pipeline

When a component renders, it processes the element tree with this priority:

1. **g-ignore** - Check if element should be skipped entirely
2. **g-with** - Switch data context if specified
3. **g-if** - Evaluate condition; skip if false
4. **g-attif-*** - Process conditional attributes
5. **g-outer-repeat** - Repeat entire element for each item, or continue to:
6. **Element rendering** - Handle g-use, content, children, and attributes

### Render Types

- `RenderOuter` - Render opening tag, content, and closing tag (default)
- `RenderInner` - Render content only (used by `g-inner-use`)

### HTML Escaping

Escaping is controlled by `ctx.Escaped` in the `RenderContext`, set via the `Unescaped()` render option.
Individual attributes enforce their own escaping rules regardless of the global setting:

- `g-inner-text`: **Always escaped** — safe for user input, even with `Unescaped()`
- `g-inner-html`: **Never escaped** — for trusted HTML content only
- `g-outer-text`: Follows `ctx.Escaped` (escaped by default)
- Attribute values: **Always escaped**
- Text nodes: Follow `ctx.Escaped` (escaped by default)

The `Unescaped()` render option only affects text nodes and `g-outer-text`. It does **not** disable escaping for `g-inner-text` or attribute values, which remain safe regardless.

## Extensibility

### Custom Registry

Implement `TemplateRegistry` interface for custom storage:

```go
gotmx.WithCustomRegistry(myRegistry)
```

### Custom Model Resolver

Implement `ModelPathResolver` for a different expression language:

```go
gotmx.WithCustomResolver(myResolver)
```

### Custom Logger

Any slog-compatible logger works:

```go
gotmx.WithLogger(slog.Default())
```

## See Also

- [Quick Start](./quick-start.md) - Get up and running in 5 minutes
- [Browser Preview Guide](./browser-preview.md) - Techniques for preview-friendly templates
- [Attribute Reference](./attribute-reference.md) - Complete list of all `g-*` attributes
- [Common Patterns](./common-patterns.md) - Recipes for typical use cases
- [Integration Patterns](./integration-patterns.md) - Framework and HTMX integration
- [Go Templates Integration](./golang-templates.md) - Working with native Go templates
- [Customization Guide](./customization.md) - Advanced customization options
