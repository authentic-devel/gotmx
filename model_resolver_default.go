package gotmx

import (
	"strings"

	"github.com/authentic-devel/empaths"
)

// ModelPathResolverDefault is a default implementation of ModelPathResolver.
// it uses the empaths library for resolving model paths:
// https://github.com/authentic-devel/empaths
type ModelPathResolverDefault struct {
	referenceResolver empaths.ReferenceResolver
}

// NewModelPathResolverDefault creates a new ModelPathResolverDefault
func NewModelPathResolverDefault(
	referenceResolver empaths.ReferenceResolver,
) ModelPathResolver {
	return &ModelPathResolverDefault{
		referenceResolver: referenceResolver,
	}
}

// Resolve processes the given path string and returns the corresponding value from the model.
// The path should be the actual plain path, without "[[" prefix or "]]" suffix.
func (r *ModelPathResolverDefault) Resolve(path string, data any) any {
	return empaths.Resolve(path, data, r.referenceResolver)
}

// TryResolve attempts to extract a model value from an expression.
// It returns the resolved value and true if the expression is a valid path (starts with "[[" and ends with "]]").
// Otherwise, it returns nil and false.
func (r *ModelPathResolverDefault) TryResolve(path string, data any) (any, bool) {
	// Check if the path is a valid model path
	if !strings.HasPrefix(path, "[[") || !strings.HasSuffix(path, "]]") {
		return nil, false
	}

	// Extract the actual path by removing the [[ and ]] delimiters and trimming spaces
	cleanPath := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(path, "[["), "]]"))

	// If the path is empty, return the data itself
	if cleanPath == "" || cleanPath == "." {
		return data, true
	}

	result := empaths.Resolve(cleanPath, data, r.referenceResolver)
	return result, true
}
