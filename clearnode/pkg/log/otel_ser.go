package log

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var _ SpanEventRecorder = &OtelSpanEventRecorder{}

const (
	// Used when a value is missing for a key in attribute pairs
	missingAttributeValue = "MISSING"
	// Used as the key when an invalid (non-string) key is encountered
	invalidAttributeKey = "invalidKeysAndValues"
)

// OtelSpanEventRecorder is a SpanEventRecorder implementation that records
// events to an OpenTelemetry span. It converts log messages and their
// associated key-value pairs into span events and attributes.
type OtelSpanEventRecorder struct {
	span trace.Span
}

// NewOtelSpanEventRecorder creates a new OtelSpanEventRecorder that will
// record events to the provided OpenTelemetry span.
func NewOtelSpanEventRecorder(span trace.Span) *OtelSpanEventRecorder {
	return &OtelSpanEventRecorder{
		span: span,
	}
}

// TraceID returns the trace ID of the span as a string.
func (ser *OtelSpanEventRecorder) TraceID() string {
	return ser.span.SpanContext().TraceID().String()
}

// SpanID returns the span ID of the span as a string.
func (ser *OtelSpanEventRecorder) SpanID() string {
	return ser.span.SpanContext().SpanID().String()
}

// RecordEvent records an event to the span with the given name and attributes.
// The keysAndValues are converted to OpenTelemetry attributes.
func (ser *OtelSpanEventRecorder) RecordEvent(name string, keysAndValues ...any) {
	ser.span.AddEvent(name, trace.WithAttributes(kvToOtelAttributes(keysAndValues...)...))
}

// RecordError records an error event to the span with the given name and attributes.
// It also sets the span status to error.
func (ser *OtelSpanEventRecorder) RecordError(name string, keysAndValues ...any) {
	ser.span.AddEvent(name, trace.WithAttributes(kvToOtelAttributes(keysAndValues...)...))
	ser.span.SetStatus(codes.Error, name)
}

func kvToOtelAttributes(keysAndValues ...any) []attribute.KeyValue {
	if len(keysAndValues)%2 != 0 {
		keysAndValues = append(keysAndValues, missingAttributeValue)
	}

	attributes := make([]attribute.KeyValue, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		var key string
		s, keyIsStr := keysAndValues[i].(string)
		if keyIsStr {
			key = s
		} else {
			attributes = append(attributes, attribute.String(
				invalidAttributeKey,
				fmt.Sprint(keysAndValues[i:]),
			))
			break
		}

		var keyValue attribute.KeyValue
		switch v := keysAndValues[i+1].(type) {
		case bool:
			keyValue = attribute.Bool(key, v)
		case int:
			keyValue = attribute.Int(key, v)
		case int16, int32, int64, uint8, uint16, uint32:
			keyValue = attribute.Int64(key, toInt64(v))
		case float32, float64:
			keyValue = attribute.Float64(key, toFloat64(v))
		case fmt.Stringer:
			keyValue = attribute.String(key, v.String())
		default:
			keyValue = attribute.String(key, fmt.Sprint(v))
		}

		attributes = append(attributes, keyValue)
	}

	return attributes
}

// toInt64 converts various integer types to int64 for use in attributes.
func toInt64(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

// toFloat64 converts various float types to float64 for use in attributes.
func toFloat64(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
