package log

var _ Logger = SpanLogger{}

// SpanLogger is a logger that wraps another logger and additionally records
// log events to a span using a SpanEventRecorder.
// This allows log messages to be correlated with distributed traces.
type SpanLogger struct {
	lg  Logger
	ser SpanEventRecorder
}

// NewSpanLogger creates a new SpanLogger that wraps the provided logger
// and records events to the given SpanEventRecorder.
// The wrapped logger's caller skip is incremented by 1 to account for the SpanLogger wrapper.
func NewSpanLogger(lg Logger, ser SpanEventRecorder) Logger {
	return &SpanLogger{
		lg:  lg.AddCallerSkip(1), // Skip the spanLogger's own call stack frame
		ser: ser,
	}
}

// Debug logs a debug message to both the wrapped logger and the span.
func (sl SpanLogger) Debug(msg string, keysAndValues ...any) {
	sl.ser.RecordEvent(msg, sl.withLogContext(LevelDebug, keysAndValues)...)
	sl.lg.Debug(msg, sl.withTraceContext(keysAndValues)...)
}

// Info logs an info message to both the wrapped logger and the span.
func (sl SpanLogger) Info(msg string, keysAndValues ...any) {
	sl.ser.RecordEvent(msg, sl.withLogContext(LevelInfo, keysAndValues)...)
	sl.lg.Info(msg, sl.withTraceContext(keysAndValues)...)
}

// Warn logs a warning message to both the wrapped logger and the span.
func (sl SpanLogger) Warn(msg string, keysAndValues ...any) {
	sl.ser.RecordEvent(msg, sl.withLogContext(LevelWarn, keysAndValues)...)
	sl.lg.Warn(msg, sl.withTraceContext(keysAndValues)...)
}

// Error logs an error message to both the wrapped logger and the span.
// The error is recorded as an error event in the span.
func (sl SpanLogger) Error(msg string, keysAndValues ...any) {
	sl.ser.RecordError(msg, sl.withLogContext(LevelError, keysAndValues)...)
	sl.lg.Error(msg, sl.withTraceContext(keysAndValues)...)
}

// Fatal logs a fatal message to both the wrapped logger and the span.
// The error is recorded as an error event in the span.
func (sl SpanLogger) Fatal(msg string, keysAndValues ...any) {
	sl.ser.RecordError(msg, sl.withLogContext(LevelFatal, keysAndValues)...)
	sl.lg.Fatal(msg, sl.withTraceContext(keysAndValues)...)
}

// WithKV returns a new SpanLogger with the key-value pair added to the wrapped logger.
// The SpanEventRecorder remains the same.
func (sl SpanLogger) WithKV(key string, value any) Logger {
	return SpanLogger{
		lg:  sl.lg.WithKV(key, value),
		ser: sl.ser,
	}
}

// GetAllKV returns all key-value pairs from the wrapped logger.
func (sl SpanLogger) GetAllKV() []any {
	return sl.lg.GetAllKV()
}

// WithName returns a new SpanLogger with the given name set on the wrapped logger.
// The SpanEventRecorder remains the same.
func (sl SpanLogger) WithName(name string) Logger {
	return SpanLogger{
		lg:  sl.lg.WithName(name),
		ser: sl.ser,
	}
}

// Name returns the name of the wrapped logger.
func (sl SpanLogger) Name() string {
	return sl.lg.Name()
}

// AddCallerSkip returns a new SpanLogger with increased caller skip on the wrapped logger.
func (sl SpanLogger) AddCallerSkip(skip int) Logger {
	return SpanLogger{
		lg:  sl.lg.AddCallerSkip(skip),
		ser: sl.ser,
	}
}

// withTraceContext appends the current trace and span IDs to the provided slice of key-value pairs.
// It returns a new slice containing the trace information followed by the original keys and values.
// This is useful for ensuring that log entries include trace context for distributed tracing.
func (sl SpanLogger) withTraceContext(keysAndValues []any) []any {
	logKeysAndValues := append([]any{
		"traceId", sl.ser.TraceID(),
		"spanId", sl.ser.SpanID(),
	}, keysAndValues...)

	return logKeysAndValues
}

// withLogContext constructs a combined slice of key-value pairs for logging.
// It prepends the log level and component name, appends all key-value pairs
// from the underlying logger, and finally appends any additional key-value
// pairs provided as input. This ensures that all log entries include consistent
// context information.
//
// Parameters:
//   - level: The log level to associate with the log entry.
//   - keysAndValues: Additional key-value pairs to include in the log entry.
//
// Returns:
//   - A slice of key-value pairs representing the complete log context.
func (sl SpanLogger) withLogContext(level Level, keysAndValues []any) []any {
	fullKeysAndValues := append([]any{
		"level", string(level),
		"component", sl.lg.Name(),
	}, sl.lg.GetAllKV()...)
	fullKeysAndValues = append(fullKeysAndValues, keysAndValues...)

	return fullKeysAndValues
}
