package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RPCMessage represents a complete message in the RPC protocol, including data and signatures
type RPCMessage struct {
	ReqRaw       json.RawMessage `json:"req,omitempty" validate:"required_without=Res,excluded_with=Res"` // Keep raw req for signature validation
	Req          *RPCData        `json:"-"`
	Res          *RPCData        `json:"res,omitempty" validate:"required_without=ReqRaw,excluded_with=ReqRaw"`
	AppSessionID string          `json:"sid,omitempty"`
	Sig          []string        `json:"sig"`
}

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCMessage) UnmarshalJSON(data []byte) error {
	var aux struct {
		ReqRaw       json.RawMessage `json:"req,omitempty"`
		Res          *RPCData        `json:"res,omitempty"`
		AppSessionID string          `json:"sid,omitempty"`
		Sig          []string        `json:"sig"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal top‐level RPCMessage: %w", err)
	}

	m.ReqRaw = aux.ReqRaw
	m.Res = aux.Res
	m.AppSessionID = aux.AppSessionID
	m.Sig = aux.Sig

	if len(aux.ReqRaw) > 0 {
		var parsed RPCData
		if err := json.Unmarshal(aux.ReqRaw, &parsed); err != nil {
			return fmt.Errorf("failed to unmarshal RPCData from ReqRaw: %w", err)
		}
		m.Req = &parsed
	}
	return nil
}

// ParseRPCMessage parses a JSON string into an RPCMessage
func ParseRPCMessage(data []byte) (*RPCMessage, error) {
	var req RPCMessage
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return &req, nil
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64
	Method    string
	Params    []any
	Timestamp uint64
}

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
