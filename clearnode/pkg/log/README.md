# Logger Package

This package provides a structured, context-aware logging system designed for distributed tracing support. It emphasizes explicit logger dependency injection and context propagation rather than global state.

## Core Design Principles

1. **No Global State**: Loggers should be explicitly initialized and passed through constructors or contexts.
2. **Context Propagation**: Use contexts to pass loggers across API boundaries.
3. **Structured Logging**: Always use key-value pairs for structured information rather than formatted strings.
4. **Tracing Integration**: Spans automatically receive log events when a logger is stored in a context with an active span.
5. **Pluggable Implementation**: The Logger interface allows for different implementations (Zap, Noop, Span-aware).

## Components

### Logger Interface

The core `Logger` interface provides structured logging methods:

```go
type Logger interface {
    Debug(msg string, keysAndValues ...any)
    Info(msg string, keysAndValues ...any)
    Warn(msg string, keysAndValues ...any)
    Error(msg string, keysAndValues ...any)
    Fatal(msg string, keysAndValues ...any)
    WithKV(key string, value any) Logger
    GetAllKV() []any
    WithName(name string) Logger
    Name() string
    AddCallerSkip(skip int) Logger
}
```

### Available Implementations

1. **ZapLogger**: Production-ready logger based on Uber's zap library
2. **NoopLogger**: Discards all log messages (useful for testing)
3. **SpanLogger**: Decorator that records log events to both a wrapped logger and a span

## Usage Guide

### Basic Initialization

```go
// Initialize ZapLogger with configuration
conf := log.Config{
    Format: "json",    // "console", "logfmt", or "json"
    Level:  log.LevelInfo,  // LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal
    Output: "stderr",  // "stdout", "stderr", or file path
}

logger := log.NewZapLogger(conf)

// For testing or when logging should be disabled
logger := log.NewNoopLogger()
```

### Constructor Injection (Recommended)

Pass the logger explicitly to your components:

```go
type Service struct {
    config Config
    logger log.Logger
}

func NewService(config Config, logger log.Logger) *Service {
    // Enhance logger with service info
    serviceLogger := logger.WithName("service").WithKV("serviceID", config.ServiceID)
    
    return &Service{
        config: config,
        logger: serviceLogger,
    }
}

// Use the injected logger
func (s *Service) DoSomething() {
    s.logger.Info("Operation started", "operation", "something")
    // ...
}
```

### Context Propagation

For request-scoped operations with automatic span integration:

```go
// In your HTTP handler or request entry point
func (s *Service) HandleRequest(ctx context.Context, req Request) {
    // Create a request-scoped logger
    logger := s.logger.WithKV("requestID", req.ID)
    
    // Store it in the context
    // If ctx has a valid OpenTelemetry span, the logger will be wrapped
    // with a SpanLogger that records events to both the logger and span
    ctx = log.SetContextLogger(ctx, logger)
    
    // Process the request
    result, err := s.processRequest(ctx, req)
    // ...
}

// In downstream functions
func (s *Service) processRequest(ctx context.Context, req Request) (Result, error) {
    // Extract logger from context (returns NoopLogger if none found)
    logger := log.FromContext(ctx)
    
    logger.Debug("Processing request", "id", req.ID)
    // ...
}
```

### Working with OpenTelemetry Spans

When a context contains a valid OpenTelemetry span, `SetContextLogger` automatically wraps the logger with a `SpanLogger`:

```go
// Start a span
ctx, span := tracer.Start(ctx, "operation")
defer span.End()

// Set logger in context - automatically creates SpanLogger
ctx = log.SetContextLogger(ctx, logger)

// All logs will be recorded to both the logger output and the span
log.FromContext(ctx).Info("Operation info", "key", "value")
// This creates a span event with the log data
```

## Best Practices

### Use Structured Logging

Always use key-value pairs instead of formatted strings:

```go
// GOOD
logger.Info("User logged in", "userID", user.ID, "method", "oauth")

// BAD
logger.Info(fmt.Sprintf("User %s logged in using %s", user.ID, "oauth"))
```

### Contextual Enrichment

Use `WithKV()` and `WithName()` to create derived loggers with additional context:

```go
// Base component logger
logger := log.NewZapLogger(config)

// Enrich with component name
orderLogger := logger.WithName("orderbook")

// Further enrich with specific context
btcLogger := orderLogger.WithKV("market", "BTCUSD")
```

### Level Usage Guidelines

* **Fatal**: Critical failures requiring application termination
* **Error**: Operation failures that don't require termination
* **Warn**: Unexpected conditions that deserve attention
* **Info**: Significant business events or state changes
* **Debug**: Detailed information for debugging

### Error Handling

Log errors at business boundaries with appropriate context:

```go
// Low-level function - returns error without logging
func readConfig(path string) (Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return Config{}, fmt.Errorf("reading config: %w", err)
    }
    // ...
}

// Business function - logs with context
func (s *Service) Initialize(ctx context.Context) error {
    logger := log.FromContext(ctx)
    
    config, err := readConfig(s.configPath)
    if err != nil {
        logger.Error("Failed to initialize service", 
            "configPath", s.configPath,
            "error", err)
        return err
    }
    
    logger.Info("Service initialized", "config", config.Name)
    return nil
}
```

### Key-Value Pairs

Ensure key-value pairs are properly formatted:

```go
// GOOD - Even number of arguments with string keys
logger.Info("Message", "key1", "value1", "key2", 42)

// BAD - Odd number of arguments (missing value)
logger.Info("Message", "key1", "value1", "key2") // "MISSING" will be appended

// BAD - Non-string key
logger.Info("Message", 123, "value") // Will use "invalidKeysAndValues"
```

## Testing

For unit tests, use the `NoopLogger` or create a test logger:

```go
func TestService(t *testing.T) {
    // Option 1: Use noop logger
    logger := log.NewNoopLogger()
    
    // Option 2: Use real logger for debugging
    logger := log.NewZapLogger(log.Config{
        Format: "console",
        Level:  log.LevelDebug,
        Output: "stdout",
    })
    
    service := NewService(testConfig, logger)
    // ... run tests
}
```

## Configuration

The `Config` struct supports environment variables:

```go
type Config struct {
    Format string `env:"LOG_FORMAT" env-default:"console"`
    Level  Level  `env:"LOG_LEVEL" env-default:"info"`
    Output string `env:"LOG_OUTPUT" env-default:"stderr"`
}
```

Environment variables:
- `LOG_FORMAT`: Output format (console, logfmt, json)
- `LOG_LEVEL`: Minimum log level (debug, info, warn, error, fatal)
- `LOG_OUTPUT`: Output destination (stderr, stdout, or file path)

## Migration from Global Loggers

When migrating from global logger patterns:

1. Replace global logger variables with constructor injection
2. Update function signatures to accept logger or context parameters
3. Use `SetContextLogger` and `FromContext` for cross-cutting concerns
4. Replace string formatting with structured key-value logging

## Performance Considerations

For high-frequency logging paths:

1. Pre-create loggers with common fields using `WithKV()`
2. Use appropriate log levels to reduce output volume
3. Consider using `NoopLogger` in performance-critical paths where logging is optional

## OpenTelemetry Integration

The package provides seamless integration with OpenTelemetry tracing:

- `SpanLogger`: Automatically created when setting a logger in a context with an active span
- `OtelSpanEventRecorder`: Records log events as span events with appropriate attributes
- Error and Fatal levels are recorded as span errors with proper status

This ensures that logs are correlated with traces, providing better observability in distributed systems.