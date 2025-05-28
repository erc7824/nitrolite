package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEIPSignature(t *testing.T) {
	allowances := []Allowance{
		{
			Asset:  "usdc",
			Amount: "100000000000",
		},
	}
	recoveredSigner, err := RecoverAddressFromEip712Signature(
		"0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
		"30322c9f-32f4-4e68-aa9d-42233ff20627",
		"0x6966978ce78df3228993aa46984eab6d68bbe195",
		"Yellow App Store",
		allowances,
		"0x746ec6b9d2de454ef0b0d579c79dc466a75573fb056d9a9505ecaafe7abe7ee92cb0502b6f0be435305b0b7f4eaf4e4d000a2b6c0b7299bfe2e9d8ec8861eaa81c")
	assert.Equal(t, recoveredSigner, "0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7")
	assert.NoError(t, err)
}
