package gotmx

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// TemplateNotFoundError is returned when a template cannot be found in the registry.
// It includes helpful context for debugging: the list of available templates
// and a "did you mean" suggestion based on Levenshtein distance.
type TemplateNotFoundError struct {
	Name       TemplateName
	Namespace  Namespace
	Available  []TemplateName
	DidYouMean TemplateName
}

func (e *TemplateNotFoundError) Error() string {
	var msg strings.Builder

	if e.Namespace != "" {
		fmt.Fprintf(&msg, "template %q not found in namespace %q", e.Name, e.Namespace)
	} else {
		fmt.Fprintf(&msg, "template %q not found", e.Name)
	}

	if e.DidYouMean != "" {
		fmt.Fprintf(&msg, "; did you mean %q?", e.DidYouMean)
	}

	if len(e.Available) > 0 && len(e.Available) <= 10 {
		names := make([]string, len(e.Available))
		for i, n := range e.Available {
			names[i] = string(n)
		}
		fmt.Fprintf(&msg, "; available: [%s]", strings.Join(names, ", "))
	} else if len(e.Available) > 10 {
		fmt.Fprintf(&msg, "; %d templates available", len(e.Available))
	}

	return msg.String()
}

// AmbiguousTemplateError is returned when a template name matches multiple templates
// across different namespaces and no namespace was specified to disambiguate.
type AmbiguousTemplateError struct {
	Name       TemplateName
	Namespaces []Namespace
}

func (e *AmbiguousTemplateError) Error() string {
	namespaces := make([]string, len(e.Namespaces))
	for i, ns := range e.Namespaces {
		namespaces[i] = string(ns)
	}
	return fmt.Sprintf("template %q is ambiguous: found in %d namespaces [%s]; use fully qualified name like \"namespace#%s\"",
		e.Name, len(e.Namespaces), strings.Join(namespaces, ", "), e.Name)
}

// ComponentNotFoundError is returned when a component cannot be created,
// typically because the referenced template does not exist.
type ComponentNotFoundError struct {
	ComponentRef string
	TemplateName TemplateName
	Cause        error
}

func (e *ComponentNotFoundError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("component %q not found: %v", e.ComponentRef, e.Cause)
	}
	return fmt.Sprintf("component %q not found", e.ComponentRef)
}

func (e *ComponentNotFoundError) Unwrap() error {
	return e.Cause
}

// MaxNestingDepthExceededError is returned when template nesting depth exceeds the configured limit.
// This typically indicates circular template references (e.g., template A uses template B, and B uses A)
// or excessively deep template hierarchies.
type MaxNestingDepthExceededError struct {
	TemplateName string // The template that was being rendered when the limit was exceeded
	CurrentDepth int    // The current nesting depth when the error occurred
	MaxDepth     int    // The configured maximum allowed depth
}

func (e *MaxNestingDepthExceededError) Error() string {
	return fmt.Sprintf("max template nesting depth exceeded: template %q at depth %d (max: %d); "+
		"this may indicate circular template references (e.g., template A uses B, and B uses A)",
		e.TemplateName, e.CurrentDepth, e.MaxDepth)
}

// RenderError provides context about where an error occurred during rendering.
// It includes the template name, element tag, and optionally the attribute
// that caused the error. This makes debugging template issues much easier
// by showing the exact location in the template hierarchy where the error occurred.
type RenderError struct {
	Template  string // The template being rendered (may be empty if unknown)
	Element   string // The HTML element tag name (e.g., "div", "span")
	Attribute string // The attribute being processed when the error occurred (may be empty)
	Cause     error  // The underlying error
}

func (e *RenderError) Error() string {
	var msg strings.Builder

	msg.WriteString("render error")

	if e.Template != "" {
		fmt.Fprintf(&msg, " in template %q", e.Template)
	}

	if e.Element != "" {
		fmt.Fprintf(&msg, " at <%s>", e.Element)
	}

	if e.Attribute != "" {
		fmt.Fprintf(&msg, " [%s]", e.Attribute)
	}

	if e.Cause != nil {
		fmt.Fprintf(&msg, ": %v", e.Cause)
	}

	return msg.String()
}

func (e *RenderError) Unwrap() error {
	return e.Cause
}

// wrapRenderError wraps an error with rendering context if it's not already a RenderError.
// If the error is nil, it returns nil. If the error is already a RenderError, it returns
// the error unchanged to preserve the original context (innermost error location).
// Context errors (context.Canceled, context.DeadlineExceeded) are returned unwrapped
// so callers can detect them with standard error checks.
func wrapRenderError(err error, element, attribute string) error {
	if err == nil {
		return nil
	}

	// Don't wrap context errors — they should propagate cleanly
	if err == context.Canceled || err == context.DeadlineExceeded {
		return err
	}

	// Don't double-wrap RenderErrors - preserve the innermost error location
	var renderErr *RenderError
	if errors.As(err, &renderErr) {
		return err
	}

	return &RenderError{
		Element:   element,
		Attribute: attribute,
		Cause:     err,
	}
}

// VoidElementChildError is returned when a void HTML element (like <br>, <img>, etc.)
// unexpectedly has child nodes. Void elements cannot have children according to the HTML spec.
type VoidElementChildError struct {
	Element string // The void element tag name (e.g., "br", "img")
}

func (e *VoidElementChildError) Error() string {
	return fmt.Sprintf("html: void element <%s> has child nodes", e.Element)
}

// DuplicateTemplateError is returned when attempting to register a template
// that already exists with the same name and namespace.
type DuplicateTemplateError struct {
	Name      TemplateName
	Namespace Namespace
}

func (e *DuplicateTemplateError) Error() string {
	return fmt.Sprintf("template with name %q and namespace %q already exists", e.Name, e.Namespace)
}

// InvalidPathError is returned when a file system path is invalid for the expected operation.
// For example, when a file path is provided where a directory is expected, or vice versa.
type InvalidPathError struct {
	Path         string // The problematic path
	ExpectedType string // What was expected: "file" or "directory"
	ActualType   string // What was found: "file" or "directory"
}

func (e *InvalidPathError) Error() string {
	return fmt.Sprintf("path %q must be a %s, not a %s", e.Path, e.ExpectedType, e.ActualType)
}

// FileError is returned when a file system operation fails.
// It wraps the underlying error and provides context about which file and operation failed.
type FileError struct {
	Path      string // The file path that caused the error
	Operation string // The operation that failed (e.g., "stat", "open", "read")
	Cause     error  // The underlying error
}

func (e *FileError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to %s file %q: %v", e.Operation, e.Path, e.Cause)
	}
	return fmt.Sprintf("failed to %s file %q", e.Operation, e.Path)
}

func (e *FileError) Unwrap() error {
	return e.Cause
}

// TemplateLoadError is returned when a template fails to load.
// It wraps the underlying error and provides context about which template failed.
type TemplateLoadError struct {
	TemplateName TemplateName
	Namespace    Namespace
	Cause        error
}

func (e *TemplateLoadError) Error() string {
	if e.Namespace != "" {
		return fmt.Sprintf("failed to load template %q from namespace %q: %v", e.TemplateName, e.Namespace, e.Cause)
	}
	return fmt.Sprintf("failed to load template %q: %v", e.TemplateName, e.Cause)
}

func (e *TemplateLoadError) Unwrap() error {
	return e.Cause
}

// ComponentCreationError is returned when a component fails to be created from a template.
// It wraps the underlying error and provides context about which template failed.
type ComponentCreationError struct {
	TemplateName TemplateRef
	Cause        error
}

func (e *ComponentCreationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to create component for template %q: %v", e.TemplateName, e.Cause)
	}
	return fmt.Sprintf("failed to create component for template %q", e.TemplateName)
}

func (e *ComponentCreationError) Unwrap() error {
	return e.Cause
}

// NilComponentError is returned when a template's NewRenderable method returns nil
// without an error. This indicates a bug in the template implementation.
type NilComponentError struct {
	TemplateName TemplateRef
}

func (e *NilComponentError) Error() string {
	return fmt.Sprintf("template %q returned nil component", e.TemplateName)
}

// TemplateSourceNotFoundError is returned when a template cannot be found in any
// of the configured template sources (file systems or directories).
type TemplateSourceNotFoundError struct {
	Name      TemplateName
	Namespace Namespace
	Sources   []string // Description of sources that were searched
}

func (e *TemplateSourceNotFoundError) Error() string {
	if len(e.Sources) > 0 {
		return fmt.Sprintf("template %s#%s not found in any source: %v", e.Namespace, e.Name, e.Sources)
	}
	return fmt.Sprintf("template %s#%s not found in any source", e.Namespace, e.Name)
}

// TemplateRetrievalError is returned when there's a failure retrieving a template from the registry.
// It wraps the underlying error (which could be TemplateNotFoundError, AmbiguousTemplateError, etc.).
type TemplateRetrievalError struct {
	TemplateName TemplateRef
	Cause        error
}

func (e *TemplateRetrievalError) Error() string {
	return fmt.Sprintf("failed to get template %q: %v", e.TemplateName, e.Cause)
}

func (e *TemplateRetrievalError) Unwrap() error {
	return e.Cause
}

// levenshteinDistance calculates the edit distance between two strings.
// This is used to find "did you mean" suggestions for misspelled template names.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize first column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}

	// Initialize first row
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// findClosestMatch finds the template name with the smallest Levenshtein distance
// to the target name. Returns empty string if no good match is found.
// A match is only returned if the distance is at most 3 edits and less than
// half the length of the target name.
func findClosestMatch(target TemplateName, candidates []TemplateName) TemplateName {
	if len(candidates) == 0 {
		return ""
	}

	targetStr := strings.ToLower(string(target))
	minDistance := len(targetStr) / 2 // threshold: at most half the string length
	if minDistance > 3 {
		minDistance = 3 // cap at 3 edits
	}
	if minDistance < 1 {
		minDistance = 1
	}

	var closest TemplateName
	for _, candidate := range candidates {
		candidateStr := strings.ToLower(string(candidate))
		distance := levenshteinDistance(targetStr, candidateStr)
		if distance <= minDistance {
			minDistance = distance
			closest = candidate
		}
	}

	return closest
}
