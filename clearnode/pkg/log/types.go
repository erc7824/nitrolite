package log

// Logger is a logger interface.
type Logger interface {
	// Debug logs a message for low-level debugging.
	// Use for detailed information useful during development.
	// keysAndValues lets you add structured context (e.g., "user", id).
	Debug(msg string, keysAndValues ...any)
	// Info logs general information about application progress.
	// Use for routine events or state changes.
	// keysAndValues lets you add structured context (e.g., "module", name).
	Info(msg string, keysAndValues ...any)
	// Warn logs a message for unexpected situations that aren't errors.
	// Use when something might be wrong but the app can continue.
	// keysAndValues lets you add structured context (e.g., "attempt", n).
	Warn(msg string, keysAndValues ...any)
	// Error logs an error that prevents normal operation.
	// Use for failures or problems that need attention.
	// keysAndValues lets you add structured context (e.g., "error", err).
	Error(msg string, keysAndValues ...any)
	// Fatal logs a critical error and may terminate the program.
	// Use for unrecoverable failures.
	// keysAndValues lets you add structured context (e.g., "reason", reason).
	Fatal(msg string, keysAndValues ...any)
	// WithKV returns a logger with an extra key-value pair for all future logs.
	// Use to add persistent context (e.g., component, request ID).
	WithKV(key string, value any) Logger
	// GetAllKV returns all persistent key-value pairs for this logger.
	// Use to inspect logger context.
	GetAllKV() []any
	// WithName returns a logger with a specific name (e.g., module or component).
	// Use to identify the source of logs.
	WithName(name string) Logger
	// Name returns the logger's name.
	Name() string
	// AddCallerSkip returns a logger that skips extra stack frames when reporting log source.
	// Use when wrapping the logger in helpers; returns itself if unsupported.
	AddCallerSkip(skip int) Logger
}

// Level represents the severity level of a log message.
// It can be used to filter log output based on importance.
type Level string

const (
	// LevelDebug is the most verbose level, used for debugging purposes.
	LevelDebug Level = "debug"
	// LevelInfo is used for informational messages.
	LevelInfo Level = "info"
	// LevelWarn is used for warning messages that indicate potential issues.
	LevelWarn Level = "warn"
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
