package protocol_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPayload(t *testing.T) {
	id := uint64(1)
	method := "testMethod"
	params := protocol.Params{
		"param1": json.RawMessage("\"value1\""),
		"param2": json.RawMessage("2"),
	}

	payload := protocol.NewPayload(id, method, params)
	assert.Equal(t, id, payload.RequestID)
	assert.Equal(t, method, payload.Method)
	assert.Equal(t, params, payload.Params)
	assert.LessOrEqual(t, payload.Timestamp, uint64(time.Now().UnixMilli()))
}

func TestPayloadUnmarshalJSON(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected protocol.Payload
		errMsg   string
	}{
		{
			name:  "valid payload",
			input: `[1, "testMethod", {"param1": "value1", "param2": 2}, 1700000000000]`,
			expected: protocol.Payload{
				RequestID: 1,
				Method:    "testMethod",
				Params: protocol.Params{
					"param1": json.RawMessage("\"value1\""),
					"param2": json.RawMessage("2"),
				},
				Timestamp: 1700000000000,
			},
			errMsg: "",
		},
		{
			name:   "wrong number of elements",
			input:  `[1, "testMethod", {"param1": "value1"}]`,
			errMsg: "invalid RPCData: expected 4 elements in array",
		},
		{
			name:   "invalid request_id type",
			input:  `["not a number", "testMethod", {"param1": "value1"}, 1700000000000]`,
			errMsg: "invalid request_id",
		},
		{
			name:   "invalid method type",
			input:  `[1, 123, {"param1": "value1"}, 1700000000000]`,
			errMsg: "invalid method",
		},
		{
			name:   "invalid params type",
			input:  `[1, "testMethod", ["not", "an", "object"], 1700000000000]`,
			errMsg: "invalid params",
		},
		{
			name:   "invalid timestamp type",
			input:  `[1, "testMethod", {"param1": "value1"}, "not a number"]`,
			errMsg: "invalid timestamp",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var payload protocol.Payload
			err := json.Unmarshal([]byte(tc.input), &payload)
			if tc.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, payload)
			}
		})
	}
}

func TestPayloadMarshalJSON(t *testing.T) {
	tcs := []struct {
		name     string
		input    protocol.Payload
		expected string
	}{
		{
			name: "valid payload",
			input: protocol.Payload{
				RequestID: 1,
				Method:    "testMethod",
				Params: protocol.Params{
					"param1": json.RawMessage("\"value1\""),
					"param2": json.RawMessage("2"),
				},
				Timestamp: 1700000000000,
			},
			expected: `[1,"testMethod",{"param1":"value1","param2":2},1700000000000]`,
		},
		{
			name: "empty params",
			input: protocol.Payload{
				RequestID: 2,
				Method:    "anotherMethod",
				Params:    protocol.Params{},
				Timestamp: 1700000001000,
			},
			expected: `[2,"anotherMethod",{},1700000001000]`,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			require.NoError(t, err)
			assert.JSONEq(t, tc.expected, string(data))
		})
	}
}

func TestNewParams(t *testing.T) {
	tcs := []struct {
		name     string
		input    any
		expected protocol.Params
		errMsg   string
	}{
		{
			name: "valid map",
			input: map[string]any{
				"param1": "value1",
				"param2": 2,
			},
			expected: protocol.Params{
				"param1": json.RawMessage("\"value1\""),
				"param2": json.RawMessage("2"),
			},
			errMsg: "",
		},
		{
			name:     "invalid non-map",
			input:    []string{"not", "a", "map"},
			expected: nil,
			errMsg:   "error unmarshalling params: json: cannot unmarshal array into Go value of type protocol.Params",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			params, err := protocol.NewParams(tc.input)
			if tc.errMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, params)
			}
		})
	}
}

func TestParamsTranslate(t *testing.T) {
	type testObj struct {
		Param1 string `json:"param1"`
		Param2 int    `json:"param2"`
	}

	input := protocol.Params{
		"param1": json.RawMessage("\"value1\""),
		"param2": json.RawMessage("2"),
	}

	expectedObj := testObj{
		Param1: "value1",
		Param2: 2,
	}

	objOutput := testObj{}
	err := input.Translate(&objOutput)
	require.NoError(t, err)
	assert.Equal(t, expectedObj, objOutput)

	sliceOutput := []testObj{}
	err = input.Translate(&sliceOutput)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error unmarshalling params: json: cannot unmarshal object into Go value of type []protocol_test.testObj")

	expectedMap := map[string]any{
		"param1": "value1",
		"param2": float64(2), // JSON numbers are unmarshalled into float64
	}

	var mapOutput map[string]any
	err = input.Translate(&mapOutput)
	require.NoError(t, err)
	assert.Equal(t, expectedMap, mapOutput)
}

func TestParamsError(t *testing.T) {
	tcs := []struct {
		name     string
		input    protocol.Params
		expected string
	}{
		{
			name: "with error",
			input: protocol.Params{
				"error": json.RawMessage("\"something went wrong\""),
			},
			expected: "something went wrong",
		},
		{
			name:     "without error",
			input:    protocol.Params{},
			expected: "",
		},
		{
			name: "malformed error",
			input: protocol.Params{
				"error": json.RawMessage("123"), // not a string
			},
			expected: "",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Error()
			if tc.expected == "" {
				assert.Nil(t, err)
			} else {
				require.NotNil(t, err)
				assert.Equal(t, tc.expected, err.Error())
			}
		})
	}
}
