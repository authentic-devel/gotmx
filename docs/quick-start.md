# Gotmx Quick Start

Get up and running with Gotmx in 5 minutes.

## Install

```shell
go get github.com/authentic-devel/gotmx
```

## 1. Create a Template

Create `templates/hello.htm`:

```html
<div data-g-define="hello-page">
    <h1 data-g-inner-text="[[ .Title ]]">Example Title</h1>
    <p data-g-inner-text="[[ .Message ]]">Example message shown in browser preview</p>
</div>
```

Key points:
- `data-g-define="hello-page"` registers this element as a named template
- `data-g-inner-text="[[ .Title ]]"` binds the text content to your data model
- The placeholder text ("Example Title") is visible when previewing the HTML in a browser, but replaced at render time

## 2. Render from Go

```go
package main

import (
    "log"
    "net/http"

    "github.com/authentic-devel/gotmx"
)

func main() {
    engine, err := gotmx.New(
        gotmx.WithTemplateDir("./templates"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer engine.Close()

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        data := map[string]any{
            "Title":   "Hello",
            "Message": "Welcome to gotmx!",
        }
        if err := engine.Render(r.Context(), w, "hello-page", data); err != nil {
            http.Error(w, err.Error(), 500)
        }
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Output:
```html
<div>
    <h1>Hello</h1>
    <p>Welcome to gotmx!</p>
</div>
```

## 3. Add Iteration

```html
<ul data-g-define="user-list">
    <li data-g-outer-repeat="[[ .Users ]]" data-g-inner-text="[[ .Name ]]">
        Sample User
    </li>
</ul>
```

```go
data := map[string]any{
    "Users": []map[string]any{
        {"Name": "Alice"},
        {"Name": "Bob"},
    },
}
engine.Render(r.Context(), w, "user-list", data)
```

## 4. Add Conditionals

```html
<div data-g-define="status">
    <span data-g-if="[[ .IsOnline ]]" class="badge-green">Online</span>
    <span data-g-if="[[ .IsOffline ]]" class="badge-red">Offline</span>
</div>
```

## 5. Compose Components

Define a reusable card component:

```html
<div data-g-define="card" class="card">
    <div class="card-header" data-g-define-slot="header">
        Default Header (visible in browser preview)
    </div>
    <div class="card-body" data-g-define-slot="">
        Default Body
    </div>
</div>
```

Use it with custom content:

```html
<div data-g-define="dashboard">
    <div data-g-use="card">
        <h2 data-g-use-slot="header">Dashboard</h2>
        <p>Welcome back!</p>
    </div>
</div>
```

## 6. Use a Layout

Wrap page content in a layout with a single render call:

```go
engine.Render(r.Context(), w, "dashboard-page", pageData,
    gotmx.WithLayout("main-layout", layoutData),
)
```

The rendered `dashboard-page` is placed into the layout's default slot automatically.

## 7. Production Setup

Embed templates for production deployment:

```go
//go:embed templates/**/*.htm templates/**/*.html
var templateFS embed.FS

engine, err := gotmx.New(
    gotmx.WithFS(templateFS),
)
```

## 8. Development Mode

Enable file watching for automatic reload during development:

```go
engine, err := gotmx.New(
    gotmx.WithTemplateDir("./templates"),
    gotmx.WithDevMode(true),
)
defer engine.Close() // stops file watchers
```

## 9. Design for Browser Preview

Gotmx templates are valid HTML — open them directly in a browser to see the layout with sample content. Use these techniques to make previews realistic:

```html
<!DOCTYPE html>
<html>
<head>
    <!-- Relative path works when opening the file directly -->
    <link rel="stylesheet" href="../static/css/app.css" data-g-href="/static/css/app.css" />
</head>
<body>
    <!-- Only the data-g-define element is used at runtime.
         The surrounding HTML/head/body is scaffolding for preview. -->
    <div data-g-define="task-list">
        <h2 data-g-inner-text="[[ .Heading ]]">My Tasks</h2>
        <ul>
            <!-- This item repeats at runtime -->
            <li data-g-outer-repeat="[[ .Tasks ]]"
                data-g-inner-text="[[ .Title ]]">Configure database</li>
            <!-- These items are preview-only padding -->
            <li data-g-ignore="outer">Set up CI pipeline</li>
            <li data-g-ignore="outer">Write documentation</li>
        </ul>
    </div>
</body>
</html>
```

Open this file in a browser: you see a styled page with three tasks. At runtime, the ignored items disappear, the heading and list come from data, and `data-g-href` replaces the static CSS path.

See [Browser Preview Guide](./browser-preview.md) for all techniques.

## Key Concepts

| Concept | Description |
|---------|-------------|
| `data-g-define` | Register a named template |
| `data-g-inner-text` | Safe text binding (always escaped) |
| `data-g-inner-html` | Raw HTML binding (trusted content only) |
| `data-g-if` | Conditional rendering |
| `data-g-outer-repeat` | Repeat element per item |
| `data-g-use` | Include another template |
| `data-g-define-slot` | Define a content injection point |
| `data-g-use-slot` | Direct content to a named slot |
| `[[ .Path ]]` | Model path expression |

## File Extensions

- `.htm` — Eagerly loaded at startup. Templates can be referenced by simple name.
- `.html` — Lazily loaded on first use. Must use fully qualified name (`path/file.html#name`).

## Next Steps

- [Browser Preview Guide](./browser-preview.md) — All techniques for preview-friendly templates
- [Attribute Reference](./attribute-reference.md) — All template directives
- [Common Patterns](./common-patterns.md) — Recipes for typical use cases
- [Integration Patterns](./integration-patterns.md) — Framework integration guides
- [Architecture](./architecture.md) — How gotmx works internally
- [Customization](./customization.md) — Custom registries, resolvers, and loggers
