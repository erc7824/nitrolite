package clearnet

import (
	"encoding/json"
	"fmt"

	"github.com/erc7824/nitrolite/examples/cerebro/unisig"
)

type RPCRequest struct {
	Req RPCData            `json:"req"`
	Sig []unisig.Signature `json:"sig"`
}

type RPCResponse struct {
	Res RPCData            `json:"res"`
	Sig []unisig.Signature `json:"sig"`
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64 `json:"request_id" validate:"required"`
	Method    string `json:"method" validate:"required"`
	Params    any    `json:"params" validate:"required"`
	Timestamp uint64 `json:"ts" validate:"required"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCData) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 4 {
		return fmt.Errorf("invalid RPCData: expected 4 elements in array")
	}

	// Element 0: uint64 RequestID
	if err := json.Unmarshal(rawArr[0], &m.RequestID); err != nil {
		return fmt.Errorf("invalid request_id: %w", err)
	}
	// Element 1: string Method
	if err := json.Unmarshal(rawArr[1], &m.Method); err != nil {
		return fmt.Errorf("invalid method: %w", err)
	}
	// Element 2: []any Params
	if err := json.Unmarshal(rawArr[2], &m.Params); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	// Element 3: uint64 Timestamp
	if err := json.Unmarshal(rawArr[3], &m.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	return nil
}

// MarshalJSON for RPCData always emits the array‐form [RequestID, Method, Params, Timestamp].
func (m RPCData) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
}
