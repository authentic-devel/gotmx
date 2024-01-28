package gotmx

// ModelPathResolver is an interface that handles the resolution of model paths within templates.
// It provides methods to extract and resolve values from data models using path expressions.
// The default implementation is ModelPathResolverDefault, which handles paths enclosed in [[ ]] delimiters.
type ModelPathResolver interface {

	// TryResolve is a convenience function that attempts to extract a model value from an expression.
	// If the expression is not recognized as a path, it returns false as the second argument.
	// If it is a valid path, it returns the resolved model value as the first return value and true as the second return value.
	// For example, the implementation in ModelPathResolverDefault returns false as the second return value if
	// the expression does not start with "[[" and does not end with "]]".
	TryResolve(expression string, data any) (any, bool)

	// Resolve processes the given path and returns the corresponding value from the model.
	// The path parameter should be the actual plain path, without any markers like
	// "[[" prefix or "]]" suffix that are used in the ModelPathResolverDefault implementation.
	Resolve(path string, data any) any
}
