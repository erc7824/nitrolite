// Package log provides a structured, context-aware logging system with distributed tracing support.
//
// The package is designed around explicit dependency injection and context propagation,
// avoiding global state and encouraging clean, testable code.
//
// # Core Types
//
// The package centers around the Logger interface, which provides structured logging methods:
//
//	type Logger interface {
//	    Debug(msg string, keysAndValues ...any)
//	    Info(msg string, keysAndValues ...any)
//	    Warn(msg string, keysAndValues ...any)
//	    Error(msg string, keysAndValues ...any)
//	    Fatal(msg string, keysAndValues ...any)
//	    WithKV(key string, value any) Logger
//	    GetAllKV() []any
//	    WithName(name string) Logger
//	    Name() string
//	    AddCallerSkip(skip int) Logger
//	}
//
// Three implementations are provided:
//
//   - ZapLogger: A production-ready logger based on Uber's zap library
//   - NoopLogger: A logger that discards all messages (useful for testing)
//   - SpanLogger: A decorator that records logs to both a wrapped logger and a trace span
//
// # Basic Usage
//
// Create a logger and use it directly:
//
//	conf := log.Config{
//	    Format: "json",
//	    Level:  log.LevelInfo,
//	    Output: "stderr",
//	}
//	logger := log.NewZapLogger(conf)
//	logger.Info("Application started", "version", "1.0.0")
//
// # Context Integration
//
// The package provides context-aware logging with automatic span integration:
//
//	// Store logger in context
//	ctx = log.SetContextLogger(ctx, logger)
//
//	// Retrieve logger from context
//	logger := log.FromContext(ctx)
//
// When SetContextLogger is called with a context containing a valid OpenTelemetry span,
// the logger is automatically wrapped with a SpanLogger that records events to both
// the logger output and the trace span.
//
// # Structured Logging
//
// All logging methods accept key-value pairs for structured data:
//
//	logger.Info("User action",
//	    "userID", user.ID,
//	    "action", "login",
//	    "ip", request.RemoteAddr,
//	)
//
// # Logger Enrichment
//
// Create derived loggers with additional context:
//
//	// Add a name hierarchy
//	serviceLogger := logger.WithName("auth-service")
//
//	// Add persistent key-value pairs
//	userLogger := serviceLogger.WithKV("userID", userID)
//
// # OpenTelemetry Integration
//
// The package seamlessly integrates with OpenTelemetry tracing. When a logger is set
// in a context with an active span, log events are automatically recorded as span events:
//
//	ctx, span := tracer.Start(ctx, "operation")
//	defer span.End()
//
//	ctx = log.SetContextLogger(ctx, logger)
//	log.FromContext(ctx).Info("Operation started") // Recorded in both log and span
//
// Error and Fatal level logs are recorded as span errors with appropriate status codes.
//
// # Using AddCallerSkip for Helper Functions
//
// When you wrap logging calls in helper functions, use AddCallerSkip(1) to ensure
// the log output reports the correct source line from your application code,
// not the helper itself.
//
//	func handleError(logger log.Logger, err error) {
//	    // Skip this helper frame so the log points to the real caller
//	    logger.AddCallerSkip(1).Error("operation failed", "err", err)
//	}
//
//	func doSomething(logger log.Logger) {
//	    err := someOperation()
//	    if err != nil {
//	        handleError(logger, err) // Log will point here, not inside handleError
//	    }
//	}
//
// # Testing
//
// For unit tests, use NoopLogger to avoid log output:
//
//	func TestSomething(t *testing.T) {
//	    logger := log.NewNoopLogger()
//	    service := NewService(logger)
//	    // ... test service
//	}
//
// # Environment Configuration
//
// The Config struct supports environment variables:
//
//   - LOG_FORMAT: Output format (console, logfmt, json)
//   - LOG_LEVEL: Minimum log level (debug, info, warn, error, fatal)
//   - LOG_OUTPUT: Output destination (stderr, stdout, or file path)
package log
