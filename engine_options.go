package gotmx

import (
	"io/fs"
	"time"
)

// WithLogger sets the logger for all engine components.
// The logger is automatically propagated to all internal components.
// Compatible with slog.Logger from the standard library.
func WithLogger(logger Logger) Option {
	return func(cfg *engineConfig) error {
		cfg.logger = logger
		return nil
	}
}

// WithDevMode enables or disables development mode with automatic template reloading.
// When enabled (true), the engine watches template directories for changes.
// Should be disabled (false) in production due to overhead.
// Accepts a boolean for easy integration with config flags:
//
//	gotmx.WithDevMode(config.IsDev)
//	gotmx.WithDevMode(os.Getenv("ENV") == "development")
func WithDevMode(enabled bool) Option {
	return func(cfg *engineConfig) error {
		cfg.devMode = enabled
		return nil
	}
}

// WithTemplateDir adds a directory as a template source.
// Can be called multiple times to add multiple directories - templates can be
// scattered across the codebase (e.g., next to their corresponding Go code).
// In dev mode: all directories are watched for changes.
// In prod mode: templates are loaded lazily on first access.
func WithTemplateDir(dir string) Option {
	return func(cfg *engineConfig) error {
		cfg.templateDirs = append(cfg.templateDirs, dir)
		return nil
	}
}

// WithFS adds an embedded or virtual filesystem as a template source.
// IMPORTANT: The engine automatically walks the ENTIRE filesystem and discovers
// all template files (*.htm, *.html) at any depth. You do NOT need to specify
// where templates are located within the FS - they are found automatically.
//
// This enables the "single embed.FS" pattern where one embed directive with glob
// patterns captures templates scattered across the codebase:
//
//	//go:embed */*.htm */*/*.htm */*/*.html
//	var templateFs embed.FS
//
// Can be called multiple times to add multiple filesystems if needed.
// Templates are loaded lazily on first access by default.
func WithFS(fsys fs.FS) Option {
	return func(cfg *engineConfig) error {
		cfg.fsystems = append(cfg.fsystems, fsys)
		return nil
	}
}

// WithIgnore specifies directory patterns to skip when loading templates.
// Commonly used to ignore "node_modules", "vendor", ".git", etc.
func WithIgnore(patterns ...string) Option {
	return func(cfg *engineConfig) error {
		cfg.ignorePatterns = append(cfg.ignorePatterns, patterns...)
		return nil
	}
}

// WithEagerExtensions specifies which file extensions are parsed at startup.
// Templates in these files are loaded when New() is called and can be referenced
// by simple name (e.g., "button" instead of "components/button.htm#button").
// Default: []string{".htm"}
func WithEagerExtensions(exts ...string) Option {
	return func(cfg *engineConfig) error {
		cfg.eagerExtensions = exts
		return nil
	}
}

// WithLazyExtensions specifies which file extensions are loaded on demand.
// Templates in these files are only parsed when first requested and must be
// referenced by fully qualified name (e.g., "pages/home.html#home-page").
// This enables faster startup for large template sets.
// Default: []string{".html"}
func WithLazyExtensions(exts ...string) Option {
	return func(cfg *engineConfig) error {
		cfg.lazyExtensions = exts
		return nil
	}
}

// WithCustomRegistry allows using a custom TemplateRegistry implementation.
// Most users should not need this.
func WithCustomRegistry(registry TemplateRegistry) Option {
	return func(cfg *engineConfig) error {
		cfg.customRegistry = registry
		return nil
	}
}

// WithCustomResolver allows using a custom ModelPathResolver implementation.
// Most users should not need this.
func WithCustomResolver(resolver ModelPathResolver) Option {
	return func(cfg *engineConfig) error {
		cfg.customResolver = resolver
		return nil
	}
}

// WithDeterministicOutput enables sorted attribute output for deterministic HTML.
// When enabled, HTML attributes are sorted alphabetically on each element.
// This is useful for testing where you need predictable output for assertions.
//
// By default, attributes are rendered in map iteration order (faster, non-deterministic).
// Production code should leave this disabled for better performance.
//
// Example:
//
//	// For tests:
//	engine, _ := gotmx.New(gotmx.WithDeterministicOutput(true))
func WithDeterministicOutput(enabled bool) Option {
	return func(cfg *engineConfig) error {
		cfg.deterministicOutput = enabled
		return nil
	}
}

// WithDevDebounce sets the debounce duration for dev mode file watching.
// When template files change rapidly (e.g., during a save-all operation), the engine waits
// this long after the last change before reloading. This prevents redundant reloads.
// Default: 1 second. Only effective when WithDevMode(true) is set.
//
// Example:
//
//	gotmx.WithDevDebounce(500 * time.Millisecond) // Faster feedback
//	gotmx.WithDevDebounce(2 * time.Second)        // Less churn
func WithDevDebounce(d time.Duration) Option {
	return func(cfg *engineConfig) error {
		cfg.devDebounce = d
		return nil
	}
}

// WithReloadCallback sets a callback that is invoked after every dev mode template reload.
// If the reload succeeded, err is nil. If it failed, err describes what went wrong.
// The previous templates remain intact on failure (see atomic reload).
// Only effective when WithDevMode(true) is set.
//
// Example:
//
//	gotmx.WithReloadCallback(func(err error) {
//	    if err != nil {
//	        log.Printf("Template reload failed: %v", err)
//	    } else {
//	        log.Println("Templates reloaded successfully")
//	    }
//	})
func WithReloadCallback(fn func(err error)) Option {
	return func(cfg *engineConfig) error {
		cfg.reloadCallback = fn
		return nil
	}
}

// WithMaxNestingDepth sets the maximum allowed template nesting depth.
// This prevents stack overflow from circular template references (e.g., template A uses B, and B uses A).
//
// When a template uses another template via g-use or g-inner-use, the nesting depth increases.
// If the depth exceeds this limit, rendering will fail with MaxNestingDepthExceededError.
//
// The default is 64, which is sufficient for most use cases while providing protection
// against accidental circular references. Set to 0 to disable the limit (not recommended).
//
// Example:
//
//	// Allow deeper nesting for complex component hierarchies
//	engine, _ := gotmx.New(gotmx.WithMaxNestingDepth(128))
//
//	// Use a lower limit for stricter protection
//	engine, _ := gotmx.New(gotmx.WithMaxNestingDepth(32))
func WithMaxNestingDepth(depth int) Option {
	return func(cfg *engineConfig) error {
		cfg.maxNestingDepth = depth
		return nil
	}
}
