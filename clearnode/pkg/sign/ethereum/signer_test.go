package ethereum

import (
	"math/big"
	"strings"
	"testing"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPrivKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	testAddress = "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"
)

// setupSigner is a helper to create a signer for tests
func setupSigner(t *testing.T) sign.Signer {
	signer, err := NewEthereumSigner(testPrivKey)
	require.NoError(t, err)
	require.NotNil(t, signer)
	return signer
}

func TestSigner(t *testing.T) {
	t.Run("Initialisation", func(t *testing.T) {
		t.Run("With 0x Prefix", func(t *testing.T) {
			signer, err := NewEthereumSigner(testPrivKey)
			require.NoError(t, err)
			assert.True(t, strings.EqualFold(testAddress, signer.Address().String()))
		})

		t.Run("Without 0x Prefix", func(t *testing.T) {
			signer, err := NewEthereumSigner(strings.TrimPrefix(testPrivKey, "0x"))
			require.NoError(t, err)
			assert.True(t, strings.EqualFold(testAddress, signer.Address().String()))
		})

		t.Run("With Invalid Key", func(t *testing.T) {
			_, err := NewEthereumSigner("0xinvalidkey")
			assert.Error(t, err)
		})
	})

	t.Run("Getters", func(t *testing.T) {
		signer := setupSigner(t)
		pubKey := signer.PublicKey()
		pubKeyBytes := pubKey.Bytes()

		assert.True(t, strings.EqualFold(testAddress, signer.Address().String()))
		assert.Len(t, signer.PrivateKey().Bytes(), 32)
		assert.Len(t, pubKeyBytes, 65)
		assert.Equal(t, byte(0x04), pubKeyBytes[0])
		assert.True(t, strings.EqualFold(testAddress, pubKey.Address().String()))
	})
}

func TestSignAndRecover(t *testing.T) {
	t.Run("Message", func(t *testing.T) {
		signer := setupSigner(t)
		message := []byte("test message for signing")

		signature, err := signer.PrivateKey().Sign(message)
		require.NoError(t, err)

		recoveredAddress, err := RecoverAddress(message, signature)
		require.NoError(t, err)

		assert.True(t, strings.EqualFold(signer.Address().String(), recoveredAddress))
	})

	t.Run("EIP-712", func(t *testing.T) {
		typedData := apitypes.TypedData{
			Types: apitypes.Types{
				"EIP712Domain": {{Name: "name", Type: "string"}},
				"Policy": {
					{Name: "challenge", Type: "string"}, {Name: "scope", Type: "string"},
					{Name: "wallet", Type: "address"}, {Name: "application", Type: "address"},
					{Name: "participant", Type: "address"}, {Name: "expire", Type: "uint256"},
					{Name: "allowances", Type: "Allowance[]"},
				},
				"Allowance": {{Name: "asset", Type: "string"}, {Name: "amount", Type: "uint256"}},
			},
			PrimaryType: "Policy",
			Domain:      apitypes.TypedDataDomain{Name: "Yellow App Store"},
			Message: map[string]interface{}{
				"challenge": "a9d5b4fd-ef30-4bb6-b9b6-4f2778f004fd", "scope": "console",
				"wallet": "0x21f7d1f35979b125f6f7918fc16cb9101e5882d7", "application": "0x21f7d1f35979b125f6f7918fc16cb9101e5882d7",
				"participant": "0x6966978ce78df3228993aa46984eab6d68bbe195", "expire": "1748608702",
				"allowances": []map[string]interface{}{{"asset": "usdc", "amount": big.NewInt(0)}},
			},
		}
		expectedAddr := "0x21F7D1F35979B125f6F7918fC16Cb9101e5882d7"
		precomputedSig := "0xe758880bc3d75e9433e0f50c9b40712b0bcf90f437a8c42ba6f8a5a3d144a5ce4b10b020c4eb728323daad49c4cc6329eaa45e3ea88c95d76e55b982f6a0a8741b"

		signature, err := hexutil.Decode(precomputedSig)
		require.NoError(t, err)

		recoveredSigner, err := RecoverAddressEIP712(typedData, signature)
		require.NoError(t, err)

		assert.True(t, strings.EqualFold(expectedAddr, recoveredSigner))
	})
}

func TestRecoveryErrors(t *testing.T) {
	signer := setupSigner(t)
	message := []byte("some data to sign")
	signature, err := signer.PrivateKey().Sign(message)
	require.NoError(t, err)

	// This is a minimal but structurally valid TypedData object for testing error paths.
	validTypedData := apitypes.TypedData{
		Types:       apitypes.Types{"EIP712Domain": {{Name: "name", Type: "string"}}, "Test": {{Name: "message", Type: "string"}}},
		PrimaryType: "Test",
		Domain:      apitypes.TypedDataDomain{Name: "Test Domain"},
		Message:     map[string]interface{}{"message": "hello"},
	}

	t.Run("Invalid Signature Length", func(t *testing.T) {
		shortSig := signature[:64]

		_, err := RecoverAddress(message, shortSig)
		assert.ErrorContains(t, err, "invalid signature length")

		_, err = RecoverAddressEIP712(validTypedData, shortSig)
		assert.ErrorContains(t, err, "invalid signature length")
	})

	t.Run("Malformed Signature", func(t *testing.T) {
		malformedSig := make([]byte, len(signature))
		copy(malformedSig, signature)
		malformedSig[30] = ^malformedSig[30] // Invert some bytes

		recoveredAddr, err := RecoverAddress(message, malformedSig)
		if err == nil {
			assert.NotEqual(t, signer.Address().String(), recoveredAddr)
		} else {
			assert.ErrorContains(t, err, "signature recovery failed")
		}

		recoveredAddr712, err := RecoverAddressEIP712(validTypedData, malformedSig)
		if err == nil {
			assert.NotEqual(t, signer.Address().String(), recoveredAddr712)
		} else {
			assert.ErrorContains(t, err, "signature recovery failed")
		}
	})
}
