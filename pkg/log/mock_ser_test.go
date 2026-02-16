package log_test

// MockSpanEventRecorder is a test double for SpanEventRecorder interface.
// It captures the last recorded event and tracks whether an error was recorded.
// This is useful for testing components that interact with span event recording.
type MockSpanEventRecorder struct {
	traceID           string
	spanID            string
	hasErr            bool
	lastEventMetadata []any
}

// NewMockSpanEventRecorder creates a new mock with the specified trace and span IDs.
func NewMockSpanEventRecorder(traceID, spanID string) *MockSpanEventRecorder {
	return &MockSpanEventRecorder{
		traceID: traceID,
		spanID:  spanID,
	}
}

// TraceID returns the configured trace ID.
func (ser *MockSpanEventRecorder) TraceID() string {
	return ser.traceID
}

// SpanID returns the configured span ID.
func (ser *MockSpanEventRecorder) SpanID() string {
	return ser.spanID
}

// RecordEvent captures the event name and metadata for verification in tests.
// The event is stored with "msg" key prepended to the metadata.
func (ser *MockSpanEventRecorder) RecordEvent(name string, keysAndValues ...any) {
	ser.lastEventMetadata = append([]any{"msg", name}, keysAndValues...)
}

// RecordError captures the error event and sets the hasErr flag to true.
// The error is stored with "msg" key prepended to the metadata.
func (ser *MockSpanEventRecorder) RecordError(name string, keysAndValues ...any) {
	ser.hasErr = true
	ser.lastEventMetadata = append([]any{"msg", name}, keysAndValues...)
}

// LastEventMetadata returns the metadata from the most recently recorded event.
// This is useful for asserting that the correct data was recorded.
func (ser *MockSpanEventRecorder) LastEventMetadata() []any {
	return ser.lastEventMetadata
}

// HasError returns true if RecordError was called at least once.
// This helps verify error recording behavior in tests.
func (ser *MockSpanEventRecorder) HasError() bool {
	return ser.hasErr
}
