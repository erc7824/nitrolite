package rpc

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_Next(t *testing.T) {
	t.Parallel()

	t.Run("Missed", func(t *testing.T) {
		valueMap := make(map[string]bool)
		ctx := &Context{
			handlers: []Handler{
				func(c *Context) {
					valueMap["step1"] = true
				},
				func(c *Context) {
					valueMap["step2"] = true
				},
			},
		}
		ctx.Next()

		assert.True(t, valueMap["step1"], "First handler should have been executed")
		assert.False(t, valueMap["step2"], "Second handler should not have been executed")
		assert.Len(t, ctx.handlers, 1, "One handler should remain")
	})

	t.Run("Correct", func(t *testing.T) {
		valueMap := make(map[string]bool)
		ctx := &Context{
			handlers: []Handler{
				func(c *Context) {
					valueMap["step1"] = true
					c.Next()
				},
				func(c *Context) {
					valueMap["step2"] = true
				},
			},
		}
		ctx.Next()

		assert.True(t, valueMap["step1"], "First handler should have been executed")
		assert.True(t, valueMap["step2"], "Second handler should have been executed")
		assert.Len(t, ctx.handlers, 0, "No handlers should remain")
	})
}

func TestContext_Succeed(t *testing.T) {
	t.Parallel()

	ctx := &Context{
		Request: Request{
			Req: Payload{
				RequestID: 1,
			},
		},
	}

	method := "method1"
	params := Params{
		"key": json.RawMessage(`"value"`),
	}
	ctx.Succeed(method, params)

	assert.Equal(t, ctx.Request.Req.RequestID, ctx.Response.Res.RequestID, "Response RequestID should match Request RequestID")
	assert.Equal(t, method, ctx.Response.Res.Method, "Response Method should match the expected method")
	assert.Equal(t, params, ctx.Response.Res.Params, "Response Params should match the expected params")
	assert.Empty(t, ctx.Response.Sig, "Response Sig should be empty")
}

func TestContext_Fail(t *testing.T) {
	t.Parallel()

	t.Run("With rpc.Error", func(t *testing.T) {
		ctx := &Context{
			Request: Request{
				Req: Payload{
					RequestID: 2,
				},
			},
		}

		rpcErr := Errorf("RPC error occurred")
		ctx.Fail(rpcErr, "This message should be ignored")

		assert.Equal(t, ctx.Request.Req.RequestID, ctx.Response.Res.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, "RPC error occurred", ctx.Response.Error().Error(), "Response Message should match the rpc.Error message")
		assert.Empty(t, ctx.Response.Sig, "Response Sig should be empty")
	})

	t.Run("With standard error and fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Request{
				Req: Payload{
					RequestID: 3,
				},
			},
		}

		stdErr := assert.AnError
		fallbackMessage := "A standard error occurred"
		ctx.Fail(stdErr, fallbackMessage)

		assert.Equal(t, ctx.Request.Req.RequestID, ctx.Response.Res.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, fallbackMessage, ctx.Response.Error().Error(), "Response Message should match the fallback message")
		assert.Empty(t, ctx.Response.Sig, "Response Sig should be empty")
	})

	t.Run("With nil error and fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Request{
				Req: Payload{
					RequestID: 4,
				},
			},
		}

		fallbackMessage := "An error occurred"
		ctx.Fail(nil, fallbackMessage)

		assert.Equal(t, ctx.Request.Req.RequestID, ctx.Response.Res.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, fallbackMessage, ctx.Response.Error().Error(), "Response Message should match the fallback message")
		assert.Empty(t, ctx.Response.Sig, "Response Sig should be empty")
	})

	t.Run("With nil error and empty fallback message", func(t *testing.T) {
		ctx := &Context{
			Request: Request{
				Req: Payload{
					RequestID: 5,
				},
			},
		}

		ctx.Fail(nil, "")

		assert.Equal(t, ctx.Request.Req.RequestID, ctx.Response.Res.RequestID, "Response RequestID should match Request RequestID")
		assert.Equal(t, defaultNodeErrorMessage, ctx.Response.Error().Error(), "Response Message should match the default error message")
		assert.Empty(t, ctx.Response.Sig, "Response Sig should be empty")
	})
}

func TestContext_GetRawResponse(t *testing.T) {
	t.Parallel()

	ctx := &Context{
		Signer: sign.NewMockSigner("signer1"),
		Request: Request{
			Req: Payload{
				RequestID: 6,
			},
		},
	}

	method := "method2"
	params := Params{
		"key": json.RawMessage(`"value2"`),
	}
	ctx.Succeed(method, params)

	rawResponse, err := ctx.GetRawResponse()
	assert.NoError(t, err, "GetRawResponse should not return an error")
	assert.NotEmpty(t, rawResponse, "Raw response should not be empty")

	var responseMsg Response
	err = json.Unmarshal(rawResponse, &responseMsg)
	assert.NoError(t, err, "Unmarshalling raw response should not return an error")

	assert.Equal(t, ctx.Request.Req.RequestID, responseMsg.Res.RequestID, "Response RequestID should match Request RequestID")
	assert.Equal(t, method, responseMsg.Res.Method, "Response Method should match the expected method")
	assert.Equal(t, params, responseMsg.Res.Params, "Response Params should match the expected params")
	require.Len(t, responseMsg.Sig, 1, "Response Sig should contain one signature")

	sig := responseMsg.Sig[0]
	assert.True(t, strings.HasSuffix(string(sig), "-signed-by-signer1"), "Signature should end with the expected suffix")
}

func TestSafeStorage(t *testing.T) {
	t.Parallel()

	storage := NewSafeStorage()

	key := "testKey"
	value := "testValue"

	storage.Set(key, value)
	retrievedValue, ok := storage.Get(key)
	require.True(t, ok, "Key should exist in storage")
	require.Equal(t, value, retrievedValue, "Retrieved value should match the set value")

	// Test concurrent access
	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	go func() {
		defer wg.Done()

		for range 100 {
			storage.Set(key, value)
		}
	}()

	go func() {
		defer wg.Done()

		for range 100 {
			value, ok := storage.Get(key)
			assert.True(t, ok, "Key should exist in storage")
			assert.Equal(t, value, value, "Value should match the set value")
		}
	}()
}
