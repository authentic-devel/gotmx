package gotmx

import "io"

// textComponent is a Renderable that wraps a text. The text can be a literal or a template.
// It will be rendered using ctx.ResolveText to resolve any model path expressions.
type textComponent struct {
	text string
}

// Render outputs the text content to the writer.
// Uses ctx.ResolveText to resolve any model path expressions in the text.
// Escaping is controlled by ctx.Escaped.
func (t *textComponent) Render(ctx *RenderContext, writer io.Writer, _ RenderType) error {
	renderedText, err := ctx.ResolveText(t.text, nil, ctx.Escaped)
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, renderedText)
	return err
}
