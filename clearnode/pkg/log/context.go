package log

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type contextKey struct{}

var loggerContextKey = contextKey{}

// SetContextLogger attaches the provided logger to the context.
// If the context contains a valid OpenTelemetry span, the logger is wrapped with a SpanLogger
// that automatically records log events to the span. If logger is nil, a NoopLogger is used.
func SetContextLogger(ctx context.Context, lg Logger) context.Context {
	if lg == nil {
		lg = NewNoopLogger()
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		// If there's no valid span, we don't need to create a spanLogger.
		return context.WithValue(ctx, loggerContextKey, lg)
	}

	ser := NewOtelSpanEventRecorder(span)
	lg = NewSpanLogger(lg, ser)

	return context.WithValue(ctx, loggerContextKey, lg)
}

// FromContext retrieves the logger stored in the context.
// If no logger is found in the context, it returns a NoopLogger as a safe default.
func FromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerContextKey).(Logger); ok {
		return l
	}
	return NewNoopLogger()
}
