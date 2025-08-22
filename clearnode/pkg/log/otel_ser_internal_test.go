package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

// Test_kvToOtelAttributes tests the conversion of key-value pairs to OpenTelemetry attributes.
// It verifies:
// 1. Empty input handling
// 2. Correct type conversion for various Go types (string, int, bool, etc.)
// 3. Handling of odd number of elements (missing values)
// 4. Error handling for non-string keys
func Test_kvToOtelAttributes(t *testing.T) {
	tests := []struct {
		name           string
		keysAndValues  []any
		expectedOutput []attribute.KeyValue
	}{
		{
			name:           "empty input",
			keysAndValues:  []any{},
			expectedOutput: []attribute.KeyValue{},
		},
		{
			name:          "even number of elements",
			keysAndValues: []any{"key1", "value1", "key2", 42, "key3", true},
			expectedOutput: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.Int("key2", 42),
				attribute.Bool("key3", true),
			},
		},
		{
			name:          "odd number of elements",
			keysAndValues: []any{"key1", "value1", "key2"},
			expectedOutput: []attribute.KeyValue{
				attribute.String("key1", "value1"),
				attribute.String("key2", "MISSING"),
			},
		},
		{
			name:          "non-string key",
			keysAndValues: []any{123, "value1", "key2", 42},
			expectedOutput: []attribute.KeyValue{
				attribute.String("invalidKeysAndValues", "[123 value1 key2 42]"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kvToOtelAttributes(tt.keysAndValues...)
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// Test_toInt64 tests the conversion of various numeric types to int64.
// It ensures all integer and floating-point types are correctly converted,
// and non-numeric types return 0.
func Test_toInt64(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
	}{
		{input: int(42), expected: 42},
		{input: int8(42), expected: 42},
		{input: int16(42), expected: 42},
		{input: int32(42), expected: 42},
		{input: int64(42), expected: 42},
		{input: uint(42), expected: 42},
		{input: uint8(42), expected: 42},
		{input: uint16(42), expected: 42},
		{input: uint32(42), expected: 42},
		{input: uint64(42), expected: 42},
		{input: float32(42.0), expected: 42},
		{input: float64(42.0), expected: 42},
		{input: "not a number", expected: 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := toInt64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_toFloat64 tests the conversion of various numeric types to float64.
// It ensures all integer and floating-point types are correctly converted,
// and non-numeric types return 0.0.
func Test_toFloat64(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
	}{
		{input: int(42), expected: 42.0},
		{input: int8(42), expected: 42.0},
		{input: int16(42), expected: 42.0},
		{input: int32(42), expected: 42.0},
		{input: int64(42), expected: 42.0},
		{input: uint(42), expected: 42.0},
		{input: uint8(42), expected: 42.0},
		{input: uint16(42), expected: 42.0},
		{input: uint32(42), expected: 42.0},
		{input: uint64(42), expected: 42.0},
		{input: float32(42.5), expected: 42.5},
		{input: float64(42.5), expected: 42.5},
		{input: "not a number", expected: 0.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := toFloat64(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
