package gotmx

import (
	"io/fs"
	"os"
	"sync"
	"time"
)

// engineConfig holds the configuration for creating an Engine.
type engineConfig struct {
	logger              Logger
	devMode             bool
	templateDirs        []string
	fsystems            []fs.FS
	ignorePatterns      []string
	eagerExtensions     []string
	lazyExtensions      []string
	customRegistry      TemplateRegistry
	customResolver      ModelPathResolver
	maxNestingDepth     int
	deterministicOutput bool
	devDebounce         time.Duration
	reloadCallback      func(err error)
}

// Engine is the main entry point for the gotmx template engine.
// It provides template loading, rendering, and lifecycle management.
// Create one with New() and pass it around your application.
type Engine struct {
	registry            TemplateRegistry
	resolver            ModelPathResolver
	logger              Logger
	maxNestingDepth     int
	deterministicOutput bool
	config              *engineConfig
	loader              LazyTemplateLoader
}

// Option configures the Engine during construction.
type Option func(*engineConfig) error

// DefaultMaxNestingDepth is the default maximum template nesting depth.
// This prevents stack overflow from circular template references.
const DefaultMaxNestingDepth = 64

// New creates a new template engine with the given options.
// Returns an error if the configuration is invalid.
func New(opts ...Option) (*Engine, error) {
	cfg := &engineConfig{
		logger:          &noopLogger{},
		eagerExtensions: []string{".htm"},
		lazyExtensions:  []string{".html"},
		maxNestingDepth: DefaultMaxNestingDepth,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return buildEngine(cfg)
}

// buildEngine constructs the Engine from the configuration.
func buildEngine(cfg *engineConfig) (*Engine, error) {
	// Create registry (custom or default)
	var registry TemplateRegistry
	if cfg.customRegistry != nil {
		registry = cfg.customRegistry
	} else {
		registry = NewTemplateRegistryDefault()
	}

	// Propagate logger to registry if it supports it
	if hasLogger, ok := registry.(hasLogger); ok {
		hasLogger.SetLogger(cfg.logger)
	}

	// Create the appropriate loader based on dev vs prod mode
	var loader LazyTemplateLoader

	if cfg.devMode {
		// DEV MODE: Use template directories with file watching
		loader = newCompositeDevLoader(cfg.templateDirs, cfg.ignorePatterns, cfg.eagerExtensions, cfg.lazyExtensions, registry, cfg.logger, cfg.devDebounce, cfg.reloadCallback)
	} else {
		// PROD MODE: Use embedded filesystems or template directories
		if len(cfg.fsystems) > 0 {
			loader = newCompositeLoader(cfg.fsystems, registry, cfg.logger)
		} else if len(cfg.templateDirs) > 0 {
			// Fallback to directory-based loading without file watching
			loader = newCompositeDirLoader(cfg.templateDirs, registry, cfg.logger)
		}
	}

	if loader != nil {
		registry.SetLazyTemplateLoader(loader)
	}

	engine := &Engine{
		registry:            registry,
		logger:              cfg.logger,
		maxNestingDepth:     cfg.maxNestingDepth,
		deterministicOutput: cfg.deterministicOutput,
		config:              cfg,
		loader:              loader,
	}

	// Create resolver (custom or default with empaths)
	// The default resolver needs a reference back to the engine for template-based resolution.
	if cfg.customResolver != nil {
		engine.resolver = cfg.customResolver
	} else {
		referenceResolver := func(templateRef string, data any) any {
			result, err := engine.getValueFromTemplate(TemplateRef(templateRef), data, false)
			if err != nil {
				engine.logger.Error("Error while resolving external template", "templateRef", templateRef, "error", err)
				return nil
			}
			return result
		}
		engine.resolver = NewModelPathResolverDefault(referenceResolver)
	}

	// Propagate logger to resolver if it supports it
	if hasLogger, ok := engine.resolver.(hasLogger); ok {
		hasLogger.SetLogger(cfg.logger)
	}

	// Register default Go template functions (GTemplate, GTextTemplate)
	engine.registerDefaultGoFunctions()

	// Initialize loaders - this triggers eager loading for configured extensions
	if loader != nil {
		loader.Init()
	}

	return engine, nil
}

// Close releases any resources held by the engine (file watchers, etc.)
// Should be called via defer after New().
func (e *Engine) Close() error {
	if e.config.devMode {
		// Stop file watchers if in dev mode
		if stopper, ok := e.loader.(interface{ Stop() error }); ok {
			return stopper.Stop()
		}
	}
	return nil
}

// Logger returns the logger configured for this engine.
func (e *Engine) Logger() Logger {
	return e.logger
}

// compositeLoader implements LazyTemplateLoader for multiple FS sources.
type compositeLoader struct {
	sources  []fs.FS
	registry TemplateRegistry
	logger   Logger
	loaders  []*templateLoaderHTML
}

// newCompositeLoader creates a loader that searches multiple FS sources.
func newCompositeLoader(sources []fs.FS, registry TemplateRegistry, logger Logger) LazyTemplateLoader {
	if len(sources) == 0 {
		return nil
	}
	return &compositeLoader{
		sources:  sources,
		registry: registry,
		logger:   logger,
	}
}

func (c *compositeLoader) Init() {
	// Initialize loaders for all sources, loading eager templates
	for _, src := range c.sources {
		loader := newTemplateLoaderHTML(src, c.registry)
		loader.SetLogger(c.logger)
		loader.Init() // Loads all .htm files (eager)
		c.loaders = append(c.loaders, loader)
	}
}

func (c *compositeLoader) Load(namespace Namespace, name TemplateName) error {
	// Try loading from each source until one succeeds
	for _, loader := range c.loaders {
		err := loader.Load(namespace, name)
		if err == nil {
			return nil
		}
	}
	return &TemplateSourceNotFoundError{
		Name:      name,
		Namespace: namespace,
	}
}

// compositeDevLoader implements LazyTemplateLoader for dev mode with multiple directories.
type compositeDevLoader struct {
	dirs            []string
	ignores         []string
	eagerExtensions []string
	lazyExtensions  []string
	registry        TemplateRegistry
	logger          Logger
	watchers        []*templateLoaderHTMLDev
	initOnce        sync.Once
	debounce        time.Duration
	reloadCallback  func(err error)
}

// newCompositeDevLoader creates a loader for dev mode with multiple directories.
func newCompositeDevLoader(dirs []string, ignores []string, eagerExtensions []string, lazyExtensions []string, registry TemplateRegistry, logger Logger, debounce time.Duration, reloadCallback func(err error)) LazyTemplateLoader {
	if len(dirs) == 0 {
		return nil
	}
	return &compositeDevLoader{
		dirs:            dirs,
		ignores:         ignores,
		eagerExtensions: eagerExtensions,
		lazyExtensions:  lazyExtensions,
		registry:        registry,
		logger:          logger,
		watchers:        make([]*templateLoaderHTMLDev, 0, len(dirs)),
		debounce:        debounce,
		reloadCallback:  reloadCallback,
	}
}

func (c *compositeDevLoader) Init() {
	c.initOnce.Do(func() {
		// Start watching all directories
		for _, dir := range c.dirs {
			watcher := newTemplateLoaderHTMLDev(dir, c.ignores, c.eagerExtensions, c.lazyExtensions, c.registry)
			watcher.SetLogger(c.logger)
			if c.debounce > 0 {
				watcher.SetDebounce(c.debounce)
			}
			if c.reloadCallback != nil {
				watcher.SetReloadCallback(c.reloadCallback)
			}
			watcher.Init()
			c.watchers = append(c.watchers, watcher)
		}
	})
}

func (c *compositeDevLoader) Load(namespace Namespace, name TemplateName) error {
	// Try loading from each watcher
	for _, watcher := range c.watchers {
		err := watcher.Load(namespace, name)
		if err == nil {
			return nil
		}
	}
	return &TemplateSourceNotFoundError{
		Name:      name,
		Namespace: namespace,
		Sources:   c.dirs,
	}
}

func (c *compositeDevLoader) Stop() error {
	// Stop all watchers
	var lastErr error
	for _, watcher := range c.watchers {
		if err := watcher.Stop(); err != nil {
			c.logger.Error("Error stopping watcher", "error", err)
			lastErr = err
		}
	}
	return lastErr
}

// compositeDirLoader implements LazyTemplateLoader for multiple directories without file watching.
type compositeDirLoader struct {
	dirs     []string
	registry TemplateRegistry
	logger   Logger
	loaders  []*templateLoaderHTML
}

// newCompositeDirLoader creates a loader for multiple directories (prod mode, no watching).
func newCompositeDirLoader(dirs []string, registry TemplateRegistry, logger Logger) LazyTemplateLoader {
	if len(dirs) == 0 {
		return nil
	}
	return &compositeDirLoader{
		dirs:     dirs,
		registry: registry,
		logger:   logger,
	}
}

func (c *compositeDirLoader) Init() {
	// Initialize loaders for all directories
	for _, dir := range c.dirs {
		fsys := os.DirFS(dir)
		loader := newTemplateLoaderHTML(fsys, c.registry)
		loader.SetLogger(c.logger)
		loader.Init() // Loads all .htm files (eager)
		c.loaders = append(c.loaders, loader)
	}
}

func (c *compositeDirLoader) Load(namespace Namespace, name TemplateName) error {
	// Try loading from each loader until one succeeds
	for _, loader := range c.loaders {
		err := loader.Load(namespace, name)
		if err == nil {
			return nil
		}
	}
	return &TemplateSourceNotFoundError{
		Name:      name,
		Namespace: namespace,
		Sources:   c.dirs,
	}
}
