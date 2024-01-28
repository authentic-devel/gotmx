# Integration Patterns

How to use Gotmx with common Go web frameworks and tools.

## net/http (Standard Library)

Gotmx works with any `io.Writer`, so it integrates naturally with `http.ResponseWriter`:

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
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        data := map[string]any{"Title": "Home"}
        if err := engine.Render(r.Context(), w, "home-page", data); err != nil {
            log.Printf("render error: %v", err)
            http.Error(w, "Internal Server Error", 500)
        }
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Always pass `r.Context()` so that rendering stops if the client disconnects.

## Chi / Gorilla Mux / Any Router

Gotmx doesn't assume a specific router. The same pattern works with any router that gives you an `http.ResponseWriter` and `*http.Request`:

```go
r := chi.NewRouter()

r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
    userID := chi.URLParam(r, "id")
    user, err := userService.Get(userID)
    if err != nil {
        http.Error(w, "Not found", 404)
        return
    }
    engine.Render(r.Context(), w, "user-profile", user)
})
```

## HTMX Partial Rendering

HTMX sends an `HX-Request` header for AJAX requests. Use this to decide whether to render a full page (with layout) or just the component:

```go
func handleDashboard(w http.ResponseWriter, r *http.Request) {
    data := getDashboardData()

    if r.Header.Get("HX-Request") != "" {
        // HTMX request: render just the component
        engine.Render(r.Context(), w, "dashboard-content", data)
    } else {
        // Full page request: wrap in layout
        engine.Render(r.Context(), w, "dashboard-content", data,
            gotmx.WithLayout("main-layout", layoutData),
        )
    }
}
```

This pattern lets you use the same template for both full-page loads and HTMX partial updates.

## Embedded Filesystems (Production)

For production, embed templates into the binary:

```go
//go:embed templates/**/*.htm templates/**/*.html
var templateFS embed.FS

func main() {
    engine, err := gotmx.New(
        gotmx.WithFS(templateFS),
    )
    if err != nil {
        log.Fatal(err)
    }
    // No need for Close() — no file watchers
}
```

## Development vs Production

Use a single flag to switch between dev mode (file watching) and production mode (embedded):

```go
func initEngine(isDev bool) (*gotmx.Engine, error) {
    if isDev {
        return gotmx.New(
            gotmx.WithTemplateDir("./templates"),
            gotmx.WithDevMode(true),
            gotmx.WithLogger(slog.Default()),
        )
    }
    return gotmx.New(
        gotmx.WithFS(templateFS),
    )
}
```

## Error Pages

Render error pages using the same engine:

```go
func renderError(w http.ResponseWriter, r *http.Request, code int, message string) {
    w.WriteHeader(code)
    data := map[string]any{
        "Code":    code,
        "Message": message,
    }
    if err := engine.Render(r.Context(), w, "error-page", data,
        gotmx.WithLayout("main-layout", nil),
    ); err != nil {
        // Fallback if template rendering fails
        http.Error(w, message, code)
    }
}
```

## Dependency Injection (Uber Fx, Wire, etc.)

Gotmx's `*Engine` is a plain struct with no global state. Pass it as a dependency:

```go
// Uber Fx example
func provideEngine(config *Config) (*gotmx.Engine, error) {
    return gotmx.New(
        gotmx.WithTemplateDir(config.TemplateDir),
        gotmx.WithDevMode(config.DevMode),
        gotmx.WithLogger(slog.Default()),
    )
}

// Register lifecycle hook to stop file watchers
fx.Invoke(func(lc fx.Lifecycle, engine *gotmx.Engine) {
    lc.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            return engine.Close()
        },
    })
})
```

## Multiple Template Sources

Templates can come from multiple directories. This is useful when modules define their own templates:

```go
engine, err := gotmx.New(
    gotmx.WithTemplateDir("./layout"),         // Shared layouts
    gotmx.WithTemplateDir("./modules/users"),  // User module templates
    gotmx.WithTemplateDir("./modules/admin"),  // Admin module templates
    gotmx.WithDevMode(config.IsDev),
)
```

## Testing Templates

Use `RenderString` and `WithDeterministicOutput` for predictable test assertions:

```go
func TestDashboardRendering(t *testing.T) {
    engine, err := gotmx.New(
        gotmx.WithDeterministicOutput(true), // Sorted attributes for stable output
    )
    if err != nil {
        t.Fatal(err)
    }

    engine.LoadHTML(`<div data-g-define="greeting" data-g-inner-text="[[ .Name ]]">x</div>`)

    result, err := engine.RenderString(context.Background(), "greeting",
        map[string]string{"Name": "Test"},
    )
    if err != nil {
        t.Fatal(err)
    }

    expected := `<div>Test</div>`
    if result != expected {
        t.Errorf("got %q, want %q", result, expected)
    }
}
```
