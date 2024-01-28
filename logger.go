package gotmx

// Logger is the interface used by gotmx for all logging operations.
//
// The Logger interface provides a consistent logging API within the gotmx library
// while allowing users to integrate their own logging implementations. This design
// enables gotmx to work with any logging library or framework that can be adapted
// to this interface.
//
// Users can provide their own Logger implementation via WithLogger().
// If no logger is provided, a default no-operation logger is used, which silently
// discards all log messages.
// The slog.Logger from the golang standard library can be used as-is.
//
// Example usage with a custom logger:
//
//	type MyCustomLogger struct {
//	    // Your logger fields
//	}
//
//	func (l *MyCustomLogger) Debug(msg string, keysAndValues ...any) {
//	    // Your implementation
//	}
//
//	func (l *MyCustomLogger) Info(msg string, keysAndValues ...any) {
//	    // Your implementation
//	}
//
//	func (l *MyCustomLogger) Error(msg string, keysAndValues ...any) {
//	    // Your implementation
//	}
//
//	// Then in your application:
//	engine, _ := gotmx.New(gotmx.WithLogger(&MyCustomLogger{}))
type Logger interface {
	// Debug logs a message at debug level with optional key-value pairs.
	// This method is used for detailed troubleshooting information.
	Debug(msg string, keysAndValues ...any)

	// Info logs a message at info level with optional key-value pairs.
	// This method is used for general operational information.
	Info(msg string, keysAndValues ...any)

	// Error logs a message at error level with optional key-value pairs.
	// This method is used for error conditions that should be investigated.
	Error(msg string, keysAndValues ...any)
}

// hasLogger is an internal interface for components that accept a logger after construction.
type hasLogger interface {
	SetLogger(logger Logger)
}

// noopLogger is a default no-operation implementation of the Logger interface.
// It silently discards all log messages and is used as the default logger when
// no other logger is provided.
type noopLogger struct{}

func (d *noopLogger) Debug(msg string, keysAndValues ...any) {}
func (d *noopLogger) Info(msg string, keysAndValues ...any)  {}
func (d *noopLogger) Error(msg string, keysAndValues ...any) {}
