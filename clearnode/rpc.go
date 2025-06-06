package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RPCMessage represents a complete message in the RPC protocol, including data and signatures
type RPCMessage struct {
	Req          *RPCData `json:"req,omitempty" validate:"required_without=Res,excluded_with=Res"`
	Res          *RPCData `json:"res,omitempty" validate:"required_without=Req,excluded_with=Req"`
	AppSessionID string   `json:"sid,omitempty"`
	Sig          []string `json:"sig"`
}

// ParseRPCMessage parses a JSON string into an RPCMessage
func ParseRPCMessage(data []byte) (RPCMessage, error) {
	var req RPCMessage
	if err := json.Unmarshal(data, &req); err != nil {
		return RPCMessage{}, fmt.Errorf("failed to parse request: %w", err)
	}
	return req, nil
}

// GetRequestSignersMap returns map with request signers public adresses
func (r RPCMessage) GetRequestSignersMap() (map[string]struct{}, error) {
	recoveredAddresses := make(map[string]struct{}, len(r.Sig))
	for _, sigHex := range r.Sig {
		recovered, err := RecoverAddress(r.Req.rawBytes, sigHex)
		if err != nil {
			return nil, err
		}
		recoveredAddresses[recovered] = struct{}{}
	}

	return recoveredAddresses, nil
}

// GetRequestSignersArray returns array of RPCMessage signers
// We first call GetRequestSignersMap in order to make sure, that user
// did not submit several copies of the same signature to reach desired quorum
func (r RPCMessage) GetRequestSignersArray() ([]string, error) {
	signersMap, err := r.GetRequestSignersMap()
	if err != nil {
		return nil, err
	}
	signers := make([]string, len(signersMap))
	for signer, _ := range signersMap {
		signers = append(signers, signer)
	}

	return signers, nil
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64
	Method    string
	Params    []any
	Timestamp uint64
	rawBytes  []byte
}

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCData) UnmarshalJSON(data []byte) error {
	var rawArr []json.RawMessage
	if err := json.Unmarshal(data, &rawArr); err != nil {
		return fmt.Errorf("error reading RPCData as array: %w", err)
	}
	if len(rawArr) != 4 {
		return errors.New("invalid RPCData: expected 4 elements in array")
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

	// Store raw bytes for signature verification
	m.rawBytes = data

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

// CreateResponse is unchanged. It simply constructs an RPCMessage with a "res" array.
func CreateResponse(id uint64, method string, responseParams []any) *RPCMessage {
	return &RPCMessage{
		Res: &RPCData{
			RequestID: id,
			Method:    method,
			Params:    responseParams,
			Timestamp: uint64(time.Now().UnixMilli()),
		},
		Sig: []string{},
	}
}
