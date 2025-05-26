package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEIPSignature(t *testing.T) {
	ok, err := VerifyEip712Data("0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7", "0cdc6ba4-71ce-41f7-9683-f64b48542e68", "c2a0a867ce869aeccfc4883d4c2688067313bb578c8205e3c44a6d454c93890065ef0fbff2b7946f4c4866b57e63bd4b3b71123373e1101b02de24f016b47a501c")
	assert.True(t, ok)
	assert.NoError(t, err)
}
