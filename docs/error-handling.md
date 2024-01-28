# Error Handling

Gotmx uses typed errors that work with Go's `errors.As()` and `errors.Is()` for inspection and matching.

## Error Types

### TemplateNotFoundError

Returned when a template name doesn't match any registered template. Includes a "did you mean" suggestion based on Levenshtein distance.

```go
result, err := engine.RenderString(ctx, "usr-profile", data)
if err != nil {
    var notFound *gotmx.TemplateNotFoundError
    if errors.As(err, &notFound) {
        // notFound.Name: "usr-profile"
        // notFound.DidYouMean: "user-profile"
        // notFound.Available: list of registered templates
        log.Printf("Template %q not found; did you mean %q?", notFound.Name, notFound.DidYouMean)
    }
}
```

### AmbiguousTemplateError

Returned when a simple (unqualified) template name matches multiple templates across different namespaces. Use a fully qualified name to disambiguate.

```go
var ambiguous *gotmx.AmbiguousTemplateError
if errors.As(err, &ambiguous) {
    // ambiguous.Name: "header"
    // ambiguous.Namespaces: ["pages/home.html", "pages/about.html"]
    log.Printf("Use qualified name like %q#%s", ambiguous.Namespaces[0], ambiguous.Name)
}
```

### RenderError

Wraps errors that occur during template rendering with location context (template name, element, attribute). Preserves the innermost error location — nested `RenderError` values are not double-wrapped.

```go
var renderErr *gotmx.RenderError
if errors.As(err, &renderErr) {
    // renderErr.Template: "user-profile"
    // renderErr.Element: "div"
    // renderErr.Attribute: "g-if"
    // renderErr.Cause: the underlying error
    log.Printf("Error at <%s> [%s] in %q: %v",
        renderErr.Element, renderErr.Attribute, renderErr.Template, renderErr.Cause)
}
```

### MaxNestingDepthExceededError

Returned when template nesting exceeds the configured limit (default: 64). This typically indicates circular references.

```go
var depthErr *gotmx.MaxNestingDepthExceededError
if errors.As(err, &depthErr) {
    // depthErr.TemplateName: the template that triggered the limit
    // depthErr.CurrentDepth: current depth when error occurred
    // depthErr.MaxDepth: configured limit
    log.Printf("Circular reference detected: %s at depth %d", depthErr.TemplateName, depthErr.CurrentDepth)
}
```

### ComponentNotFoundError

Returned when a `data-g-use` or `data-g-inner-use` references a template that cannot be found or instantiated.

### TemplateRetrievalError

Wraps errors from template registry lookup. Use `errors.As()` to unwrap and inspect the underlying cause (which may be `TemplateNotFoundError`, `AmbiguousTemplateError`, etc.).

### Other Error Types

| Error Type | When |
|-----------|------|
| `DuplicateTemplateError` | Registering a template with a name+namespace that already exists |
| `FileError` | File system operations fail (open, read, stat) |
| `TemplateLoadError` | Lazy loading of a template fails |
| `VoidElementChildError` | A void HTML element (`<br>`, `<img>`, etc.) has children |
| `NilComponentError` | A template's `NewRenderable()` returned nil without an error |

## Context Cancellation

Gotmx respects Go's `context.Context` for cancellation and timeouts. When a context is cancelled, rendering stops promptly and returns the context error.

```go
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

err := engine.Render(ctx, w, "dashboard", data)
if errors.Is(err, context.DeadlineExceeded) {
    http.Error(w, "Render timed out", http.StatusGatewayTimeout)
    return
}
```

Context cancellation is checked:
- Before creating each component (`data-g-use`)
- Between iteration steps (`data-g-outer-repeat`, `data-g-inner-repeat`)
- Between sibling children during rendering

## Error Recovery in HTTP Handlers

A common pattern is to render to a buffer first, so errors can be caught before sending HTTP headers:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    result, err := engine.RenderString(r.Context(), "page", data,
        gotmx.WithLayout("layout", layoutData),
    )
    if err != nil {
        log.Printf("Render error: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    io.WriteString(w, result)
}
```

## Debug Logging

Enable a logger to get diagnostic output during development:

```go
engine, err := gotmx.New(
    gotmx.WithLogger(slog.Default()),
    // ...
)
```

The logger receives:
- **Debug**: Template registration, missing template references, initialization
- **Info**: Template listings
- **Error**: Template load failures, resolution errors
