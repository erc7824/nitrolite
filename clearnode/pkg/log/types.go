package log

// Logger is a logger interface.
type Logger interface {
	// Debug logs a message at debug level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Debug(msg string, keysAndValues ...any)
	// Info logs a message at info level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Info(msg string, keysAndValues ...any)
	// Warn logs a message at warn level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Warn(msg string, keysAndValues ...any)
	// Error logs a message at error level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Error(msg string, keysAndValues ...any)
	// Fatal logs a message at fatal level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Fatal(msg string, keysAndValues ...any)
	// WithKV returns a new logger with the given key-value pair.
	WithKV(key string, value any) Logger
	// GetAllKV returns all key-value pairs associated with the logger.
	// This is useful for retrieving context information.
	GetAllKV() []any
	// WithName returns a new logger with the given name.
	// This is useful for creating a logger with a specific purpose, like a component or module name.
	WithName(name string) Logger
	// Name returns the name of the logger.
	Name() string
	// AddCallerSkip returns a new logger with increased caller skip.
	// This is useful for skipping the logger's own call stack frame.
	// If the logger does not support caller skipping, it should return itself.
	AddCallerSkip(skip int) Logger
}

// Level represents the severity level of a log message.
// It can be used to filter log output based on importance.
type Level string

const (
	// LevelDebug is the most verbose level, used for debugging purposes.
	LevelDebug Level = "debug"
	// LevelInfo is used for informational messages.
	LevelInfo  Level = "info"
	// LevelWarn is used for warning messages that indicate potential issues.
	LevelWarn  Level = "warn"
	// LevelError is used for error messages that indicate something went wrong.
	LevelError Level = "error"
	// LevelFatal is used for fatal errors that typically cause the program to exit.
	LevelFatal Level = "fatal"
)

// SpanEventRecorder is an interface for recording events and errors to a span.
type SpanEventRecorder interface {
	// TraceID returns the trace ID of the span.
	TraceID() string
	// SpanID returns the span ID of the span.
	SpanID() string

	// RecordEvent records an event to the span.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	RecordEvent(name string, keysAndValues ...any)
	// RecordError records an error to the span.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	RecordError(name string, keysAndValues ...any)
}
