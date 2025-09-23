package main

import (
	"context"
	"os"

	"github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
)

// Logger is a logger interface.
type Logger interface {
	// Debug logs a message at debug level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Debug(msg string, keysAndValues ...interface{})
	// Info logs a message at info level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Info(msg string, keysAndValues ...interface{})
	// Warn logs a message at warn level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Warn(msg string, keysAndValues ...interface{})
	// Error logs a message at error level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Error(msg string, keysAndValues ...interface{})
	// Fatal logs a message at fatal level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Fatal(msg string, keysAndValues ...interface{})
	// Trace logs a message at trace level.
	// keysAndValues are treated as key-value pairs (e.g., "key1", value1, "key2", value2).
	Trace(msg string, keysAndValues ...interface{})
	// With returns a new logger with the given key-value pair.
	With(key string, value interface{}) Logger
	// NewSystem returns a new logger with the given name.
	NewSystem(name string) Logger
}

func NewLoggerIPFS(name string) Logger {
	return &ipfsLogger{
		lg:                  log.Logger(name).SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar(),
		commonKeysAndValues: []interface{}{},
	}
}

type ipfsLogger struct {
	lg                  *zap.SugaredLogger
	commonKeysAndValues []interface{}
}

func (l *ipfsLogger) Trace(_ string, _ ...interface{}) {}

func (l *ipfsLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.lg.Debugw(msg, keysAndValues...)
}

func (l *ipfsLogger) Info(msg string, keysAndValues ...interface{}) {
	l.lg.Infow(msg, keysAndValues...)
}

func (l *ipfsLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.lg.Warnw(msg, keysAndValues...)
}

func (l *ipfsLogger) Error(msg string, keysAndValues ...interface{}) {
	l.lg.Errorw(msg, keysAndValues...)
}

func (l *ipfsLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.lg.Fatalw(msg, keysAndValues...)
}

func (l *ipfsLogger) With(key string, value interface{}) Logger {
	return &ipfsLogger{
		lg:                  l.lg.With(key, value),
		commonKeysAndValues: append(l.commonKeysAndValues, key, value),
	}
}

func (l *ipfsLogger) NewSystem(name string) Logger {
	lg := log.Logger(name)
	return &ipfsLogger{
		lg:                  lg.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar().With(l.commonKeysAndValues...),
		commonKeysAndValues: []interface{}{},
	}
}

type loggerContextKey struct{}

// SetContextLogger attaches the provided logger to the context.
func SetContextLogger(ctx context.Context, lg Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, lg)
}

// FromContext retrieves the logger stored in the context.
// If none is found, it returns a noop logger.
func LoggerFromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerContextKey{}).(Logger); ok {
		return l
	}
	return NewLoggerIPFS("noop") // Return a noop logger if none is found
}

func init() {
	logLevel := os.Getenv("CLEARNODE_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info" // Default log level
	}
	zapLevel, err := log.Parse(logLevel)
	if err != nil {
		zapLevel = log.LevelInfo // Fallback to Info level if parsing fails
	}

	log.SetupLogging(log.Config{
		Level:  zapLevel,
		Stderr: true,
	})
}
