package gotmx

import (
	"bufio"
	"errors"

	"golang.org/x/net/html"
)

type LazyTemplateLoader interface {
	Init()

	// Load is called by the template registry to lazily load a template.
	// It is intentional that this method does not return those templates. Instead, the template loader must
	// register them with the registry. We do this to have a more flexible approach what the template loader needs to
	// do when a template is requested.
	Load(namespace Namespace, name TemplateName) error
}

// notFoundTemplate is an internal sentinel type registered when a template cannot be found.
// This prevents repeated lazy-load attempts for the same missing template.
type notFoundTemplate struct {
	namespace Namespace
	name      TemplateName
}

func (n notFoundTemplate) Name() TemplateName {
	return n.name
}
func (n notFoundTemplate) Namespace() Namespace {
	return n.namespace
}
func (n notFoundTemplate) NewRenderable(data any) (Renderable, error) {
	return nil, &TemplateNotFoundError{
		Name:      n.name,
		Namespace: n.namespace,
	}
}

func loadTemplateFromReader(templateRegistry TemplateRegistry, reader *bufio.Reader, sourceFile string) error {
	if templateRegistry == nil {
		return errors.New("templateRegistry is not set")
	}
	doc, err := html.Parse(reader)
	if err != nil {
		return err
	}
	return processNode(templateRegistry, sourceFile, doc)
}
