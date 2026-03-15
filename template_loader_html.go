package gotmx

import (
	"bufio"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// templateLoaderHTML is an implementation of the LazyTemplateLoader interface.
// It loads all HTML files below a fs.FS (usually an os.DirFS) when its Init function is called and can lazily load
// templates as long as the namespace given is a relative path to a file within that FS.
// Implements
// - LazyTemplateLoader
// - HasLogger
type templateLoaderHTML struct {
	templateRegistry TemplateRegistry
	logger           Logger
	rootFs           fs.FS
}

// newTemplateLoaderHTML creates a new templateLoaderHTML
func newTemplateLoaderHTML(rootFs fs.FS, templateRegistry TemplateRegistry) *templateLoaderHTML {
	templateLoader := templateLoaderHTML{
		templateRegistry: templateRegistry,
		rootFs:           rootFs,
		logger:           &noopLogger{},
	}
	return &templateLoader
}

func (tl *templateLoaderHTML) SetLogger(logger Logger) {
	tl.logger = logger
}

func (tl *templateLoaderHTML) Init() {
	if tl.templateRegistry == nil {
		tl.logger.Error("templateRegistry in templateLoaderHTML is missing")
		return
	}
	tl.templateRegistry.ClearTemplates()
	if tl.rootFs != nil {
		if err := tl.loadRecursive(tl.rootFs); err != nil {
			tl.logger.Error("Error loading templates recursively", "error", err)
		}
	}
}

// Note: This implementation ignores the template name since it loads all templates found in the html file (namespace)
// at once.
func (tl *templateLoaderHTML) Load(namespace Namespace, _ TemplateName) error {
	if tl.templateRegistry == nil {
		tl.logger.Error("templateRegistry in templateLoaderHTML is missing")
		return errors.New("templateRegistry in templateLoaderHTML is missing")
	}
	if namespace == "" {
		return errors.New("cannot lazily load template that has no namespace")
	}
	return tl.loadFromFsPath(namespace.String())
}

// loadFromFile parses the given html file and extracts all Gotmx HTML templates from it.
func (tl *templateLoaderHTML) loadFromFile(filePath string) error {
	tl.logger.Debug("Parsing html document", "file", filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return &FileError{Path: filePath, Operation: "stat", Cause: err}
	}
	if fileInfo.IsDir() {
		return &InvalidPathError{Path: filePath, ExpectedType: "file", ActualType: "directory"}
	}
	file, err := os.Open(filePath)
	if err != nil {
		return &FileError{Path: filePath, Operation: "open", Cause: err}
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			tl.logger.Error("Error closing file", "filePath", filePath, "error", closeErr)
		}
	}()
	reader := bufio.NewReader(file)
	normalizedPath := filepath.ToSlash(filePath)
	return loadTemplateFromReader(tl.templateRegistry, reader, normalizedPath)
}

func (tl *templateLoaderHTML) loadFromFsPath(path string) error {

	if tl.rootFs == nil {
		return errors.New("rootFs is not set")
	}

	tl.logger.Debug("Parsing html document", "file", path)
	b, err := fs.ReadFile(tl.rootFs, path)
	if err != nil {
		return err // or panic or ignore
	}

	normalizedPath := filepath.ToSlash(path)

	reader := bufio.NewReader(bytes.NewReader(b))
	return loadTemplateFromReader(tl.templateRegistry, reader, normalizedPath)
}

func (tl *templateLoaderHTML) loadRecursive(fsys fs.FS) error {
	var walkFunc fs.WalkDirFunc = func(path string, info fs.DirEntry, entryErr error) error {
		if entryErr != nil {
			return entryErr
		}
		if !info.IsDir() && filepath.Ext(path) == ".htm" {
			return tl.loadFromFsPath(path)
		}
		return nil
	}
	return fs.WalkDir(fsys, ".", walkFunc)

}
