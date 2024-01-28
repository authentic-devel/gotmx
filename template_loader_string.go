package gotmx

import (
	"bufio"
	"strings"
)

type templateLoaderString struct {
	templateRegistry TemplateRegistry
	logger           Logger
}

func newTemplateLoaderString(templateRegistry TemplateRegistry) *templateLoaderString {
	templateLoader := templateLoaderString{
		templateRegistry: templateRegistry,
		logger:           &noopLogger{},
	}
	return &templateLoader
}

func (tl *templateLoaderString) Init() {
	// todo: currently we use this so that we can use it in tests. But we should not be forced to use it
	/* Intentionally blank. */
}

func (tl *templateLoaderString) Load(namespace Namespace, _ TemplateName) error {
	// todo: currently we use this so that we can use it in tests. But we should not be forced to use it
	/* Intentionally blank. */
	return nil
}

func (tl *templateLoaderString) SetLogger(logger Logger) {
	tl.logger = logger
}

// loadFromString parses the given html string and extracts all Gotmx HTML templates from it.
func (tl *templateLoaderString) LoadFromString(templateString string, sourceFile string) error {
	reader := strings.NewReader(templateString)
	bufReader := bufio.NewReader(reader)
	return loadTemplateFromReader(tl.templateRegistry, bufReader, sourceFile)
}
