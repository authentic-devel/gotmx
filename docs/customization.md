# Customizing Gotmx

This guide shows how to customize Gotmx behavior by providing your own
- template registry
- model path resolver
- logger

It also explains how to use development mode with file watching.

## Custom Template Registry

Implement the TemplateRegistry interface if you want to change how templates are stored and resolved.

Key responsibilities:
- Store templates keyed by name and optional namespace
- GetTemplate(ref TemplateRef) returns the resolved Template, or an error if not found/ambiguous
- RegisterTemplate(t Template) stores a template, detecting duplicates in the same namespace
- ClearTemplates() resets internal state

Optional: implement GoTemplateRegistry if you want to support Go text/HTML templates and shared funcs:
- RegisterGoTemplate(name TemplateName, template string, sourceFile string)
- RegisterFunc(name string, f interface{})

Tip: Use TemplateRegistryDefault as a reference implementation.

Example: Plug a custom registry into the Engine

```go
reg := mypackage.NewMyRegistry() // implements gotmx.TemplateRegistry
// If it also implements gotmx.GoTemplateRegistry, the engine will register default funcs automatically
engine, err := gotmx.New(
    gotmx.WithCustomRegistry(reg),
)
```

## Dev Mode: Hot Reloading

In development mode, the engine watches template directories for changes and automatically
reloads modified templates. No manual registry setup is needed.

```go
engine, err := gotmx.New(
    gotmx.WithTemplateDir("./templates"),
    gotmx.WithDevMode(true),
    gotmx.WithIgnore("node_modules", ".git"),
    gotmx.WithLogger(slog.Default()),
)
if err != nil {
    log.Fatal(err)
}
defer engine.Close() // stops file watchers
```

In production, use embedded filesystems instead:

```go
//go:embed templates/**/*.htm templates/**/*.html
var templateFS embed.FS

engine, err := gotmx.New(
    gotmx.WithFS(templateFS),
)
```

The same Engine API is used in both cases. Only the template source differs.

## Custom Model Path Resolver

If you want to change or extend the `[[ .path ]]` syntax, implement ModelPathResolver.

Default behavior:
- ModelPathResolverDefault understands expressions enclosed in `[[` and `]]`
- Resolve(path, data) resolves a plain path like `.User.Name` against your model
- TryResolve(expression, data) checks for `[[ ... ]]` and returns the resolved value and true on success

Example: custom resolver that supports a different delimiter

```go
type CurlyResolver struct{}

func (CurlyResolver) TryResolve(expr string, data any) (any, bool) {
    if !strings.HasPrefix(expr, "{{") || !strings.HasSuffix(expr, "}}") {
        return nil, false
    }
    p := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expr, "{{"), "}}"))
    // Use the empaths library or your own resolution logic
    result := empaths.Resolve(p, data)
    return result, true
}

func (CurlyResolver) Resolve(path string, data any) any {
    result := empaths.Resolve(path, data)
    return result
}
```

Wire it in via the WithCustomResolver option:

```go
engine, err := gotmx.New(
    gotmx.WithCustomResolver(CurlyResolver{}),
    gotmx.WithTemplateDir("./templates"),
)
```

## Custom Logger

Gotmx uses a minimal Logger interface compatible with `*slog.Logger` from the standard library.
You can pass `slog.Default()` directly, or provide any implementation that satisfies the interface:

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
}
```

The default is NoopLogger which discards logs.

Example: using slog directly

```go
engine, err := gotmx.New(
    gotmx.WithLogger(slog.Default()),
    gotmx.WithTemplateDir("./templates"),
)
```

The logger is automatically propagated to all internal components (registry, loaders, resolver).

## Max Nesting Depth

Gotmx protects against circular template references that could cause stack overflow. When template A uses template B via `g-use`, and B uses A, this creates an infinite loop that would normally crash the application.

Gotmx prevents this by tracking the nesting depth of template composition and failing fast when a limit is exceeded.

### Configuration

```go
// Default is 64 levels
engine, _ := gotmx.New(
    gotmx.WithTemplateDir("./templates"),
    gotmx.WithMaxNestingDepth(128),  // Allow deeper nesting
)
```

### How It Works

Every time a template uses another template via `g-use` or `g-inner-use`, the nesting depth counter increments. When the nested template finishes rendering, the counter decrements. If the counter reaches the configured maximum, rendering fails immediately with a `MaxNestingDepthExceededError`.

The error includes:
- The template name that triggered the limit
- The current depth when the error occurred
- The configured maximum depth

### Choosing a Limit

- **64 (default)**: Suitable for most applications with complex component hierarchies
- **32**: Stricter limit for simpler applications; catches problems earlier
- **128+**: For very deep component compositions (rare)
- **0**: Disables the check (not recommended - risks stack overflow)

### Example Error

When circular references are detected:

```
max template nesting depth exceeded: template "component-a" at depth 64 (max: 64);
this may indicate circular template references (e.g., template A uses B, and B uses A)
```

## Putting it together

- Swap registries to change storage, discovery, or enable dev-time reloading.
- Swap the model resolver to change the expression language.
- Plug in your logger for observability.
- Configure max nesting depth to protect against circular template references.

These extension points are intentionally small and focused to keep Gotmx modular and easy to integrate.
