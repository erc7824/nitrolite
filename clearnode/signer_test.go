package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestEIPSignature(t *testing.T) {
	allowances := []Allowance{
		{
			Asset:  "usdc",
			Amount: "0",
		},
	}
	signature, err := hexutil.Decode("0xe758880bc3d75e9433e0f50c9b40712b0bcf90f437a8c42ba6f8a5a3d144a5ce4b10b020c4eb728323daad49c4cc6329eaa45e3ea88c95d76e55b982f6a0a8741b")
	assert.NoError(t, err)

	recoveredSigner, err := RecoverAddressFromEip712Signature(
		"0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
		"a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
		"0x6966978ce78df3228993aa46984eab6d68bbe195",
		"Yellow App Store",
		allowances,
		"console",
		"0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
		"1748608702",
		signature,
	)

	assert.Equal(t, recoveredSigner, "0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7")
	assert.NoError(t, err)
}
