package gotmx

// TemplateRegistry is an interface for managing templates within the gotmx engine.
// It provides methods for registering, retrieving, and managing templates.
type TemplateRegistry interface {
	// GetTemplate returns the template with the given name or an error if no
	// matching template exists or if the template name is ambiguous.
	GetTemplate(ref TemplateRef) (Template, error)

	// RegisterTemplate adds a new template to the registry.
	// Returns an error if the template cannot be registered (e.g., duplicate template name in the same namespace).
	RegisterTemplate(template Template) error

	// ClearTemplates removes all templates from the registry.
	ClearTemplates()

	SetLazyTemplateLoader(ltl LazyTemplateLoader)
}
