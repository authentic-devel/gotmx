package gotmx

import (
	htmlTemplate "html/template"
	"io"
	textTemplate "text/template"
)

// golangTemplate implements the Template interface for Go's standard template packages.
// It wraps both text/template and html/template to support different rendering modes.
type golangTemplate struct {
	goTextTemplate *textTemplate.Template
	goHtmlTemplate *htmlTemplate.Template
	name           TemplateName
	namespace      Namespace
}

func (g *golangTemplate) Name() TemplateName {
	return g.name
}

func (g *golangTemplate) Namespace() Namespace {
	return g.namespace
}

func (g *golangTemplate) NewRenderable(data any) (Renderable, error) {
	component := golangTemplateComponent{
		goTextTemplate: g.goTextTemplate,
		goHtmlTemplate: g.goHtmlTemplate,
		templateRef:    TemplateRef(g.name),
		data:           data,
	}
	return &component, nil
}

// golangTemplateComponent implements the Renderable interface for Go templates.
type golangTemplateComponent struct {
	goTextTemplate *textTemplate.Template
	goHtmlTemplate *htmlTemplate.Template
	templateRef    TemplateRef
	data           any
}

// Render executes the template and writes the result to the provided writer.
// If ctx.Escaped is true, uses html/template for HTML-escaped output; otherwise uses text/template.
func (gt *golangTemplateComponent) Render(ctx *RenderContext, writer io.Writer, _ RenderType) error {
	templateRef := gt.templateRef.String()

	if ctx.Escaped {
		if gt.goHtmlTemplate.Lookup(templateRef) == nil {
			return &TemplateNotFoundError{Name: TemplateName(templateRef)}
		}
		err := gt.goHtmlTemplate.ExecuteTemplate(writer, templateRef, gt.data)
		if err != nil {
			return err
		}
	} else {
		if gt.goTextTemplate.Lookup(templateRef) == nil {
			return &TemplateNotFoundError{Name: TemplateName(templateRef)}
		}
		err := gt.goTextTemplate.ExecuteTemplate(writer, templateRef, gt.data)
		if err != nil {
			return err
		}
	}
	return nil
}
