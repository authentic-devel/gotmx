package gotmx

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
)

// LoadHTML parses HTML string(s) and registers any template definitions found.
// Templates are available immediately after this call.
// Returns an error if parsing fails.
func (e *Engine) LoadHTML(html string) error {
	loader := newTemplateLoaderString(e.registry)
	loader.SetLogger(e.config.logger)
	return loader.LoadFromString(html, "inline")
}

// LoadFile parses a single HTML file and registers any template definitions.
// The file path becomes the namespace for the templates.
func (e *Engine) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	normalizedPath := filepath.ToSlash(path)
	return loadTemplateFromReader(e.registry, reader, normalizedPath)
}

// LoadFS loads templates from a filesystem with an optional prefix.
// Useful for the registration pattern where packages register their own templates.
// The prefix is prepended to template namespaces for disambiguation.
//
//	engine.LoadFS(users.TemplateFS, "users")  // templates namespaced as "users/..."
//	engine.LoadFS(shared.TemplateFS, "")      // no prefix
func (e *Engine) LoadFS(fsys fs.FS, prefix string) error {
	loader := newTemplateLoaderHTML(fsys, e.registry)
	loader.SetLogger(e.config.logger)

	// Walk the filesystem and load all templates
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .htm files (eager loading)
		ext := filepath.Ext(path)
		if ext == ".htm" {
			// Prepend prefix if provided
			sourcePath := path
			if prefix != "" {
				sourcePath = prefix + "/" + path
			}

			// Load the file
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			loaderString := newTemplateLoaderString(e.registry)
			loaderString.SetLogger(e.config.logger)
			return loaderString.LoadFromString(string(data), sourcePath)
		}

		return nil
	})
}

// Preload forces immediate loading of templates matching the given patterns.
// Useful for preloading critical templates while keeping others lazy.
// Patterns support glob syntax: "components/*.htm", "layouts/**/*.html"
func (e *Engine) Preload(patterns ...string) error {
	// For each configured source, load templates matching patterns
	for _, pattern := range patterns {
		if err := e.preloadPattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

// preloadPattern loads templates matching a single pattern.
func (e *Engine) preloadPattern(pattern string) error {
	// Try to load from configured template directories
	for _, dir := range e.config.templateDirs {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return err
		}

		for _, match := range matches {
			if err := e.LoadFile(match); err != nil {
				return err
			}
		}
	}

	// Try to load from configured filesystems
	for _, fsys := range e.config.fsystems {
		if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}

			// Check if path matches the pattern
			matched, err := filepath.Match(pattern, path)
			if err != nil {
				return err
			}
			if matched {
				data, err := fs.ReadFile(fsys, path)
				if err != nil {
					return err
				}
				loaderString := newTemplateLoaderString(e.registry)
				loaderString.SetLogger(e.config.logger)
				return loaderString.LoadFromString(string(data), path)
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

// RegisterFunc registers a function for use in Go templates.
// Must be called before templates using the function are parsed.
func (e *Engine) RegisterFunc(name string, fn any) {
	if goReg, ok := e.registry.(GoTemplateRegistry); ok {
		goReg.RegisterFunc(name, fn)
	}
}
