package gotmx

import "io"

// NewStringLiteralTemplate creates a template that is defined by a string literal.
// When rendered, it does not substitute any placeholders. The string literal is rendered as is.
func NewStringLiteralTemplate(name TemplateName, literal string, namespace Namespace) Template {
	return &stringTemplate{
		name:      name,
		template:  literal,
		namespace: namespace,
	}
}

func (s *stringTemplate) NewRenderable(_ any) (Renderable, error) {
	return s, nil
}

// Render outputs the string literal to the writer.
// This implements the Renderable interface for string literal templates.
//
// Parameters:
//   - ctx: The render context (unused for string literals since they have no dynamic content).
//   - writer: Destination for the string literal output.
//   - _: RenderType is ignored for string literals.
//
// Note: String literal templates output their content verbatim without any processing.
func (s *stringTemplate) Render(_ *RenderContext, writer io.Writer, _ RenderType) error {
	_, err := writer.Write([]byte(s.template))
	if err != nil {
		return err
	}
	return nil
}

func (s *stringTemplate) Namespace() Namespace {
	return s.namespace
}

func (s *stringTemplate) Name() TemplateName {
	return s.name
}

type stringTemplate struct {
	name      TemplateName
	template  string
	namespace Namespace
}
