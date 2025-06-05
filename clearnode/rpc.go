// rpc.go

package main

import (
	"encoding/json"
	"fmt"
	"time"
)

// RPCMessage represents a complete message in the RPC protocol, including data and signatures
type RPCMessage struct {
	ReqRaw       json.RawMessage `json:"req,omitempty"`
	Req          *RPCData        `json:"-"`
	Res          *RPCData        `json:"res,omitempty"`
	AppSessionID string          `json:"sid,omitempty"`
	Sig          []string        `json:"sig"`
}

// RPCData represents the common structure for both requests and responses
// Format: [request_id, method, params, ts]
type RPCData struct {
	RequestID uint64
	Method    string
	Params    []any
	Timestamp uint64
}

// ParseRPCMessage parses a JSON string into an RPCMessage
func ParseRPCMessage(data []byte) (*RPCMessage, error) {
	var req RPCMessage
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return &req, nil
}

// CreateResponse creates a response from a request with the given fields
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

// UnmarshalJSON implements the json.Unmarshaler interface for RPCMessage
func (m *RPCMessage) UnmarshalJSON(data []byte) error {
	// 1) First unmarshal into a temporary struct that has fields for
	//    ReqRaw (json.RawMessage), Res, Sig, AppSessionID.
	var aux struct {
		ReqRaw       json.RawMessage `json:"req,omitempty"`
		Res          *RPCData        `json:"res,omitempty"`
		AppSessionID string          `json:"sid,omitempty"`
		Sig          []string        `json:"sig"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal top‐level RPCMessage: %w", err)
	}

	// 2) Save the raw bytes for "req"
	m.ReqRaw = aux.ReqRaw
	m.Res = aux.Res
	m.AppSessionID = aux.AppSessionID
	m.Sig = aux.Sig

	// 3) If there was a "req" array, unmarshal that exact same raw array into RPCData
	if len(aux.ReqRaw) > 0 {
		var rpcdata RPCData
		if err := json.Unmarshal(aux.ReqRaw, &rpcdata); err != nil {
			return fmt.Errorf("failed to unmarshal RPCData from ReqRaw: %w", err)
		}
		m.Req = &rpcdata
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface for RPCData
func (m RPCData) MarshalJSON() ([]byte, error) {
	// Create array representation in the exact order: [RequestID, Method, Params, Timestamp]
	return json.Marshal([]any{
		m.RequestID,
		m.Method,
		m.Params,
		m.Timestamp,
	})
}
