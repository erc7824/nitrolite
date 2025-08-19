package sign

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
)

func TestEIPSignature(t *testing.T) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
			},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "application", Type: "address"},
				{Name: "participant", Type: "address"},
				{Name: "expire", Type: "uint256"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "uint256"},
			}},
		PrimaryType: "Policy",
		Domain: apitypes.TypedDataDomain{
			Name: "Yellow App Store",
		},
		Message: map[string]interface{}{
			"challenge":   "a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
			"scope":       "console",
			"wallet":      "0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
			"application": "0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
			"participant": "0x6966978ce78df3228993aa46984eab6d68bbe195",
			"expire":      "1748608702",
			"allowances": []map[string]interface{}{
				{
					"asset": "usdc",
					"amount": func() *big.Int {
						return big.NewInt(0)
					}(),
				},
			},
		},
	}

	signature, err := hexutil.Decode("0xe758880bc3d75e9433e0f50c9b40712b0bcf90f437a8c42ba6f8a5a3d144a5ce4b10b020c4eb728323daad49c4cc6329eaa45e3ea88c95d76e55b982f6a0a8741b")
	assert.NoError(t, err)

	recoveredSigner, err := RecoverAddressEip712(typedData, signature)

	assert.Equal(t, recoveredSigner, "0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7")
	assert.NoError(t, err)
}
