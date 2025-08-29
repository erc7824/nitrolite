package sign

import (
	"strings"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testPrivKey = "0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"
	testAddress = "0x2c7536E3605D9C16a7a3D7b1898e529396a65c23"
)

// setupSigner is a helper to create a signer for tests
func setupSigner(t *testing.T) Signer {
	signer, err := NewEthereumSigner(testPrivKey)
	require.NoError(t, err)
	require.NotNil(t, signer)
	return signer
}

func TestEthereumSigner(t *testing.T) {
	t.Run("Initialisation", func(t *testing.T) {
		t.Run("With 0x Prefix", func(t *testing.T) {
			signer, err := NewEthereumSigner(testPrivKey)
			require.NoError(t, err)
			assert.True(t, strings.EqualFold(testAddress, signer.PublicKey().Address().String()))
		})

		t.Run("Without 0x Prefix", func(t *testing.T) {
			signer, err := NewEthereumSigner(strings.TrimPrefix(testPrivKey, "0x"))
			require.NoError(t, err)
			assert.True(t, strings.EqualFold(testAddress, signer.PublicKey().Address().String()))
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

		assert.True(t, strings.EqualFold(testAddress, signer.PublicKey().Address().String()))
		assert.Len(t, pubKeyBytes, 65)
		assert.Equal(t, byte(0x04), pubKeyBytes[0])
		assert.True(t, strings.EqualFold(testAddress, pubKey.Address().String()))
	})
}

func TestSignAndRecover(t *testing.T) {
	t.Run("Message", func(t *testing.T) {
		signer := setupSigner(t)
		message := []byte("test message for signing")
		hash := ethcrypto.Keccak256Hash(message)

		signature, err := signer.Sign(hash.Bytes())
		require.NoError(t, err)

		recoveredAddress, err := RecoverAddressFromHash(hash.Bytes(), signature)
		require.NoError(t, err)

		assert.True(t, strings.EqualFold(signer.PublicKey().Address().String(), recoveredAddress.String()))
	})
}

func TestRecoveryErrors(t *testing.T) {
	signer := setupSigner(t)
	message := []byte("some data to sign")
	msgHash := ethcrypto.Keccak256Hash(message)
	signature, err := signer.Sign(msgHash.Bytes())
	require.NoError(t, err)

	t.Run("Invalid Signature Length", func(t *testing.T) {
		shortSig := signature[:64]

		hash := ethcrypto.Keccak256Hash(message)
		_, err := RecoverAddressFromHash(hash.Bytes(), shortSig)
		assert.ErrorContains(t, err, "invalid signature length")
	})

	t.Run("Malformed Signature", func(t *testing.T) {
		malformedSig := make([]byte, len(signature))
		copy(malformedSig, signature)
		malformedSig[30] = ^malformedSig[30] // Invert some bytes

		hash := ethcrypto.Keccak256Hash(message)
		recoveredAddr, err := RecoverAddressFromHash(hash.Bytes(), malformedSig)
		if err == nil {
			assert.NotEqual(t, signer.PublicKey().Address().String(), recoveredAddr.String())
		} else {
			assert.ErrorContains(t, err, "signature recovery failed")
		}
	})
}
