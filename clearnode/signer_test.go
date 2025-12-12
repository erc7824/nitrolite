package main

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEIPSignature_EOA(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	walletAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

	allowances := []Allowance{
		{
			Asset:  "usdc",
			Amount: "123.45",
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
				{Name: "expires_at", Type: "uint64"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "string"},
			},
		},
		PrimaryType: "Policy",
		Domain:      apitypes.TypedDataDomain{Name: "Yellow App Store"},
		Message: map[string]interface{}{
			"challenge":   "a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
			"scope":       "console",
			"wallet":      walletAddress,
			"session_key": "0x6966978ce78df3228993aa46984eab6d68bbe195",
			"expires_at":  big.NewInt(1748608702),
			"allowances":  convertedAllowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(td)
	assert.NoError(t, err)
	sigBytes, err := crypto.Sign(hash, privKey)
	assert.NoError(t, err)

	swBlockchainClient, err := ethclient.Dial("wss://0xrpc.io/sep")
	require.NoError(t, err)

	ok, err := VerifyEip712Signature(
		swBlockchainClient,
		walletAddress,
		"a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd",
		"0x6966978ce78df3228993aa46984eab6d68bbe195",
		"Yellow App Store",
		allowances,
		"console",
		uint64(1748608702),
		sigBytes,
	)

	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestEIPSignature_SW(t *testing.T) {
	walletAddress := "0xC37F852cbEdA26153B08eFd41ba9Dc086f59Ce30"
	sigBytes := hexutil.MustDecode("0x80183b55dcd2d1a833fb4f1cb48a0769ae20de340933266fa18464e4cf2228c34861955736d2ed4be87372f12fbc3ccbf53bda6b8f0bb5b9300bf6ff9cfaab6d1b")

	swBlockchainClient, err := ethclient.Dial("wss://0xrpc.io/sep")
	require.NoError(t, err)

	ok, err := VerifyEip712Signature(
		swBlockchainClient,
		walletAddress,
		"530d020c-a7df-4157-a7ce-9bd79f1bc500",
		"0x93A47127014e84E0846467f05571f58e6b68Cd8A",
		"clearnode",
		[]Allowance{},
		"yellow.com",
		uint64(1765624413),
		sigBytes,
	)

	assert.NoError(t, err)
	assert.True(t, ok)
}
