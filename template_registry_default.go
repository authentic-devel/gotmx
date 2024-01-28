package gotmx

import (
	"fmt"
	htmlTemplate "html/template"
	"strings"
	"sync"
	"sync/atomic"
	textTemplate "text/template"
)

// TemplateRegistryDefault is a default implementation of TemplateRegistry and also
// implements GoTemplateRegistry
// It supports
//   - registering any type of template that implements the Template interface
//   - parsing HTML files from the filesystem or other sources using the golang builtin HTML parser and automatically
//     registering any found template definitions as NodeTemplate
//   - registering golang templates
//
// Thread Safety: This registry is safe for concurrent use. All template access
// and registration operations are protected by an internal mutex.
type TemplateRegistryDefault struct {
	mu                 sync.RWMutex
	lazyTemplateLoader LazyTemplateLoader
	templates          templateMap
	// This is where all the Golang text and HTML templates we discover are collected
	// Every template will be added to both instances so that they can be neither
	// rendered as text or HTML and also can be called from either a HTML template or text template.
	goTextTemplate *textTemplate.Template
	goHtmlTemplate *htmlTemplate.Template
	goFunctions    map[string]any
	initialized    atomic.Bool
	logger         Logger
}

func (tr *TemplateRegistryDefault) SetLazyTemplateLoader(ltl LazyTemplateLoader) {
	tr.lazyTemplateLoader = ltl
}

// templateMap is used to store all registered templates.
// The key is the unqualified name of the template, the value is a slice of templates with the same name from
// different namespaces.
// We chose this structure to be able to quickly find a template by its simple (unqualified) name because that will be
// the most common use case. So if globally unique names are used for the template names, even across all namespaces,
// then each of the map entries will only contain one single slice element and finding a template is very fast.
// A downside of using a map here is, that if a template does not behave well and returns a different name or namespace
// with the functions Name() or Namespace() after it has been registered, then we will have unpredictable behavior.
type templateMap map[TemplateName][]Template

type noopLazyTemplateLoader struct{}

func (n noopLazyTemplateLoader) Init() {
	/* Intentionally blank because as the name "noop" implies, we do not do anything. */
}

func (n noopLazyTemplateLoader) Load(namespace Namespace, name TemplateName) error {
	/* Intentionally blank because as the name "noop" implies, we do not do anything. */
	return nil
}

// NewTemplateRegistryDefault creates a new ModelPathResolverDefault
func NewTemplateRegistryDefault() *TemplateRegistryDefault {

	return &TemplateRegistryDefault{
		lazyTemplateLoader: noopLazyTemplateLoader{},
		templates:          make(templateMap),
		goHtmlTemplate:     htmlTemplate.New("GoTmx"),
		goTextTemplate:     textTemplate.New("GoTmx"),
		goFunctions:        make(map[string]any),
		logger:             &noopLogger{},
	}
}

func (tr *TemplateRegistryDefault) SetLogger(logger Logger) {
	tr.logger = logger
}

// GetTemplate returns the template with the given name or an error if no
// matching template exists or if the template name is ambiguous.
// The name can either be an unqualified template name without namespace like "myTemplate" or a fully qualified name
// with namespace like "frontend/templates/index.html#myTemplate".
// If an unqualified name is given, then that name must be globally unique across all namespaces. Otherwise, an error
// will be returned.
func (tr *TemplateRegistryDefault) GetTemplate(templateRef TemplateRef) (Template, error) {

	// If the templateRef contains a hash symbol, then we consider it a qualified templateRef, otherwise we consider it an
	// unqualified templateRef.
	simpleName := templateRef
	namespaceString := ""
	hashIndex := strings.Index(templateRef.String(), "#")
	if hashIndex != -1 {
		namespaceString = templateRef.String()[:hashIndex]
		simpleName = templateRef[hashIndex+1:]
	}
	return tr.GetTemplateExt(Namespace(namespaceString), TemplateName(simpleName))
}

// GetTemplateExt returns a template by namespace and name, with extended lookup logic.
// Returns an error if the template is not found, is ambiguous, or if loading fails.
func (tr *TemplateRegistryDefault) GetTemplateExt(namespace Namespace, templateName TemplateName) (Template, error) {
	// Fast path: check initialization without acquiring any lock.
	// atomic.Bool provides race-free reads across goroutines.
	if !tr.initialized.Load() {
		tr.mu.Lock()
		// Double-check after acquiring write lock
		if !tr.initialized.Load() {
			tr.lazyTemplateLoader.Init()
			tr.initialized.Store(true)
		}
		tr.mu.Unlock()
	}

	// Try to find the template with read lock
	tr.mu.RLock()
	templateSlice, ok := tr.templates[templateName]
	tr.mu.RUnlock()

	if !ok || len(templateSlice) == 0 {
		// If we were not able to find the template in the registry, then we need to ask the template
		// loader to load it, then we try again
		err := tr.loadTemplate(namespace, templateName)
		if err != nil {
			return nil, &TemplateLoadError{TemplateName: templateName, Namespace: namespace, Cause: err}
		}
		return tr.GetTemplateExt(namespace, templateName)
	}

	if namespace == "" {
		if len(templateSlice) == 1 {
			// if the template in the templateSlice is a "notFoundTemplate" it is an indicator that we already tried
			// to lazy load it, but were not able to find it.
			if _, notFound := templateSlice[0].(notFoundTemplate); notFound {
				available := tr.getAvailableTemplateNamesLocked()
				return nil, &TemplateNotFoundError{
					Name:       templateName,
					Available:  available,
					DidYouMean: findClosestMatch(templateName, available),
				}
			}
			return templateSlice[0], nil
		} else {
			tr.logger.Debug("Found multiple templates with the same name",
				"templateName", templateName, "templates", templateSlice)
			namespaces := make([]Namespace, len(templateSlice))
			for i, t := range templateSlice {
				namespaces[i] = t.Namespace()
			}
			return nil, &AmbiguousTemplateError{
				Name:       templateName,
				Namespaces: namespaces,
			}
		}
	}

	// a namespaceString is given, so we need to find the first template, that has the given namespaceString
	for _, template := range templateSlice {
		if template.Namespace().String() == namespace.String() {
			// if the template in the templateSlice is a "notFoundTemplate" it is an indicator that we already tried
			// to lazy load it, but were not able to find it.
			if _, notFound := template.(notFoundTemplate); notFound {
				available := tr.getAvailableTemplateNamesLocked()
				return nil, &TemplateNotFoundError{
					Name:       templateName,
					Namespace:  namespace,
					Available:  available,
					DidYouMean: findClosestMatch(templateName, available),
				}
			}
			return template, nil
		}
	}
	tr.logger.Debug("Was not able to find template", "templateName",
		templateName, "namespace", namespace, "templates", templateSlice)
	available := tr.getAvailableTemplateNamesLocked()
	return nil, &TemplateNotFoundError{
		Name:       templateName,
		Namespace:  namespace,
		Available:  available,
		DidYouMean: findClosestMatch(templateName, available),
	}
}

func (tr *TemplateRegistryDefault) loadTemplate(namespace Namespace, templateName TemplateName) error {
	err := tr.lazyTemplateLoader.Load(namespace, templateName)
	if err != nil {
		tr.logger.Error("Error loading template", "templateName", templateName, "namespace", namespace, "error", err)
		return err
	}

	// Now let's check whether the template exists. If not, we register a notFoundTemplate for it so that we don't
	// try to load it over and over again the next time it is requested.
	tr.mu.RLock()
	templateSlice, ok := tr.templates[templateName]
	tr.mu.RUnlock()

	if !ok || len(templateSlice) == 0 {
		tr.registernotFoundTemplate(namespace, templateName)
		return nil
	}

	for _, template := range templateSlice {
		if template.Namespace().String() == namespace.String() {
			return nil
		}
	}
	tr.registernotFoundTemplate(namespace, templateName)
	return nil
}

func (tr *TemplateRegistryDefault) ClearTemplates() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.templates = make(templateMap)
	tr.goHtmlTemplate = htmlTemplate.New("GoTmx")
	tr.goTextTemplate = textTemplate.New("GoTmx")
	tr.initialized.Store(false)
}

// ReplaceFrom atomically replaces all templates in this registry with the templates from source.
// This is used by dev mode to perform zero-downtime reloads: a new registry is built in isolation,
// then swapped in with a single write-lock acquisition. Concurrent readers never see an empty registry.
func (tr *TemplateRegistryDefault) ReplaceFrom(source *TemplateRegistryDefault) {
	source.mu.RLock()
	newTemplates := source.templates
	newGoHtml := source.goHtmlTemplate
	newGoText := source.goTextTemplate
	newGoFunctions := source.goFunctions
	source.mu.RUnlock()

	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.templates = newTemplates
	tr.goHtmlTemplate = newGoHtml
	tr.goTextTemplate = newGoText
	tr.goFunctions = newGoFunctions
	tr.initialized.Store(true)
}

func (tr *TemplateRegistryDefault) RegisterTemplate(template Template) error {
	templateName := template.Name()
	namespace := template.Namespace()
	tr.logger.Debug("Registering template", "templateName", templateName, "nameSpace", namespace)

	tr.mu.Lock()
	defer tr.mu.Unlock()

	templatesWithSameName, exists := tr.templates[templateName]
	if !exists {
		tr.templates[templateName] = []Template{template}
		return nil
	} else {
		// if a template exists in templatesWithSameName that has the same namespace as the new template
		// then we will return an error
		for _, t := range templatesWithSameName {
			if t.Namespace() == namespace {
				return &DuplicateTemplateError{Name: templateName, Namespace: namespace}
			}
		}
		tr.templates[templateName] = append(tr.templates[templateName], template)
	}
	return nil
}

func (tr *TemplateRegistryDefault) registernotFoundTemplate(namespace Namespace, templateName TemplateName) {
	template := notFoundTemplate{
		namespace: namespace,
		name:      templateName,
	}
	if err := tr.RegisterTemplate(template); err != nil {
		tr.logger.Error("Error registering not found template", "templateName",
			templateName, "namespace", namespace, "error", err)
	}
}

func (tr *TemplateRegistryDefault) RegisterFunc(name string, f interface{}) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.goFunctions[name] = f
	tr.goTextTemplate.Funcs(tr.goFunctions)
	tr.goHtmlTemplate.Funcs(tr.goFunctions)
}

func (tr *TemplateRegistryDefault) RegisterGoTemplate(name TemplateName, template string, sourceFile string) error {
	tr.mu.Lock()
	goTemplate := golangTemplate{
		goTextTemplate: tr.goTextTemplate,
		goHtmlTemplate: tr.goHtmlTemplate,
		name:           name,
		namespace:      Namespace(sourceFile),
	}

	if err := tr.registerGoHtmlTemplateLocked(name.String(), template); err != nil {
		tr.mu.Unlock()
		return err
	}
	if err := tr.registerGoTextTemplateLocked(name.String(), template); err != nil {
		tr.mu.Unlock()
		return err
	}
	tr.mu.Unlock()

	return tr.RegisterTemplate(&goTemplate)
}

// templateIDCounter is used to generate unique internal IDs for anonymous templates.
var templateIDCounter atomic.Uint64

// nextTemplateID generates a unique identifier for internal template use.
func nextTemplateID() string {
	return fmt.Sprintf("_tmpl_%d", templateIDCounter.Add(1))
}

func (tr *TemplateRegistryDefault) LogTemplates() {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	tr.logger.Info("Currently known templates", "templates", tr.templates)
}

// getAvailableTemplateNamesLocked returns a list of all registered template names,
// excluding notFoundTemplate entries. Used for error messages and suggestions.
// This method acquires a read lock internally.
func (tr *TemplateRegistryDefault) getAvailableTemplateNamesLocked() []TemplateName {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	result := make([]TemplateName, 0, len(tr.templates))
	for name, templates := range tr.templates {
		hasRealTemplate := false
		for _, t := range templates {
			if _, notFound := t.(notFoundTemplate); !notFound {
				hasRealTemplate = true
				break
			}
		}
		if hasRealTemplate {
			result = append(result, name)
		}
	}
	return result
}

// registerGoHtmlTemplateLocked registers a Go HTML template.
// Caller must hold tr.mu lock.
func (tr *TemplateRegistryDefault) registerGoHtmlTemplateLocked(name string, template string) error {
	_, err := tr.goHtmlTemplate.New(name).Funcs(tr.goFunctions).Parse(template)
	return err
}

// registerGoTextTemplateLocked registers a Go text template.
// Caller must hold tr.mu lock.
func (tr *TemplateRegistryDefault) registerGoTextTemplateLocked(name string, template string) error {
	_, err := tr.goTextTemplate.New(name).Funcs(tr.goFunctions).Parse(template)
	return err
}
