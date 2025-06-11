package main

import (
	"testing"
	"time"
)

func TestRPCMessageValidate(t *testing.T) {
	rpcMsg := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "testMethod",
			Params:    []any{"param1", 2},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []string{"0x1234567890abcdef"},
	}

	if err := validate.Struct(rpcMsg); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	rpcMsg.Req.Method = ""
	if err := validate.Struct(rpcMsg); err == nil {
		t.Error("expected error for empty method, got nil")
	}

	rpcMsg.Req = nil
	if err := validate.Struct(rpcMsg); err == nil {
		t.Error("expected error for empty method, got nil")
	}
}
