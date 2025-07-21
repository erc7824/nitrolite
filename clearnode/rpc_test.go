package main

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestRPCMessageValidate(t *testing.T) {
	validate := getValidator()
	rpcMsg := &RPCMessage{
		Req: &RPCData{
			RequestID: 1,
			Method:    "testMethod",
			Params:    []any{"param1", 2},
			Timestamp: uint64(time.Now().Unix()),
		},
		Sig: []Signature{Signature([]byte("0x1234567890abcdef"))},
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

func TestRPCParamsValidate(t *testing.T) {
	validate := getValidator()

	params := struct {
		TestDecimalToBigInt decimal.Decimal `validate:"bigint"`
		TestStringToBigInt  string          `validate:"bigint"`
	}{
		TestDecimalToBigInt: decimal.RequireFromString("-1234567890"),
		TestStringToBigInt:  "-1234567890",
	}

	if err := validate.Struct(params); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	params.TestDecimalToBigInt = decimal.RequireFromString("123.456")

	if err := validate.Struct(params); err == nil {
		t.Error("expected error for decimal value, got nil")
	}
}
