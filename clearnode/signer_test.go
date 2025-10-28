package main

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEIPSignature(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	walletAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

	allowances := []Allowance{
		{
			Asset:  "usdc",
			Amount: "0",
		},
	}
	convertedAllowances := convertAllowances(allowances)

	td := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {{Name: "name", Type: "string"}},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "session_key", Type: "address"},
				{Name: "expire", Type: "uint256"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "uint256"},
			},
		},
		PrimaryType: "Policy",
		Domain:      apitypes.TypedDataDomain{Name: "Yellow App Store"},
		Message: map[string]interface{}{
			"challenge":   "a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
			"scope":       "console",
			"wallet":      walletAddress,
			"session_key": "0x6966978ce78df3228993aa46984eab6d68bbe195",
			"expire":      "1748608702",
			"allowances":  convertedAllowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(td)
	assert.NoError(t, err)
	sigBytes, err := crypto.Sign(hash, privKey)
	assert.NoError(t, err)

	recoveredSigner, err := RecoverAddressFromEip712Signature(
		walletAddress,
		"a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
		"0x6966978ce78df3228993aa46984eab6d68bbe195",
		"Yellow App Store",
		allowances,
		"console",
		"1748608702",
		sigBytes,
	)

	assert.Equal(t, recoveredSigner, walletAddress)
	assert.NoError(t, err)
}
