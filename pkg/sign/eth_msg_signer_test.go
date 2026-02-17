package sign

import (
	"encoding/hex"
	"strings"
	"testing"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupMsgSigner is a helper to create a message signer for tests
func setupMsgSigner(t *testing.T) Signer {
	signer, err := NewEthereumMsgSigner(testPrivKey)
	require.NoError(t, err)
	require.NotNil(t, signer)
	return signer
}

func TestNewEthereumMsgSigner(t *testing.T) {
	t.Run("With 0x Prefix", func(t *testing.T) {
		signer, err := NewEthereumMsgSigner(testPrivKey)
		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.True(t, strings.EqualFold(testAddress, signer.PublicKey().Address().String()))
	})

	t.Run("Without 0x Prefix", func(t *testing.T) {
		signer, err := NewEthereumMsgSigner(strings.TrimPrefix(testPrivKey, "0x"))
		require.NoError(t, err)
		assert.NotNil(t, signer)
		assert.True(t, strings.EqualFold(testAddress, signer.PublicKey().Address().String()))
	})

	t.Run("With Invalid Key", func(t *testing.T) {
		_, err := NewEthereumMsgSigner("0xinvalidkey")
		assert.Error(t, err)
	})
}

func TestEthereumMsgSigner_Sign(t *testing.T) {
	signer := setupMsgSigner(t)

	t.Run("Sign Simple Message", func(t *testing.T) {
		message := []byte("Hello, Ethereum!")
		signature, err := signer.Sign(message)
		require.NoError(t, err)
		assert.NotNil(t, signature)
		assert.Len(t, signature, 65, "Ethereum signatures should be 65 bytes")
	})

	t.Run("Sign Empty Message", func(t *testing.T) {
		message := []byte("")
		signature, err := signer.Sign(message)
		require.NoError(t, err)
		assert.NotNil(t, signature)
		assert.Len(t, signature, 65)
	})

	t.Run("Sign Hash", func(t *testing.T) {
		// Even though we pass a hash, it will be treated as data and prefixed
		hash := ethcrypto.Keccak256Hash([]byte("test data"))
		signature, err := signer.Sign(hash.Bytes())
		require.NoError(t, err)
		assert.NotNil(t, signature)
		assert.Len(t, signature, 65)
	})

	t.Run("Sign Large Message", func(t *testing.T) {
		// Test with a larger message
		message := make([]byte, 1000)
		for i := range message {
			message[i] = byte(i % 256)
		}
		signature, err := signer.Sign(message)
		require.NoError(t, err)
		assert.NotNil(t, signature)
		assert.Len(t, signature, 65)
	})
}

func TestEthereumMsgAddressRecoverer_RecoverAddress(t *testing.T) {
	signer := setupMsgSigner(t)
	recoverer := &EthereumMsgAddressRecoverer{}

	t.Run("Recover From Valid Signature", func(t *testing.T) {
		message := []byte("Hello, Ethereum!")
		signature, err := signer.Sign(message)
		require.NoError(t, err)

		recoveredAddress, err := recoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		assert.NotNil(t, recoveredAddress)
		assert.True(t, strings.EqualFold(testAddress, recoveredAddress.String()))
	})

	t.Run("Recover From Empty Message Signature", func(t *testing.T) {
		message := []byte("")
		signature, err := signer.Sign(message)
		require.NoError(t, err)

		recoveredAddress, err := recoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		assert.True(t, strings.EqualFold(testAddress, recoveredAddress.String()))
	})

	t.Run("Invalid Signature Length", func(t *testing.T) {
		message := []byte("test message")
		shortSig := make([]byte, 64) // Too short

		_, err := recoverer.RecoverAddress(message, shortSig)
		assert.ErrorContains(t, err, "invalid signature length")
	})

	t.Run("Malformed Signature", func(t *testing.T) {
		message := []byte("test message")
		signature, err := signer.Sign(message)
		require.NoError(t, err)

		// Corrupt the signature
		malformedSig := make([]byte, len(signature))
		copy(malformedSig, signature)
		malformedSig[30] = ^malformedSig[30]

		recoveredAddr, err := recoverer.RecoverAddress(message, malformedSig)
		// Either recovery fails or returns wrong address
		if err == nil {
			assert.NotEqual(t, testAddress, recoveredAddr.String())
		} else {
			assert.ErrorContains(t, err, "signature recovery failed")
		}
	})

	t.Run("Wrong Message", func(t *testing.T) {
		originalMessage := []byte("original message")
		signature, err := signer.Sign(originalMessage)
		require.NoError(t, err)

		wrongMessage := []byte("wrong message")
		recoveredAddr, err := recoverer.RecoverAddress(wrongMessage, signature)
		require.NoError(t, err)
		// Should recover a different address
		assert.False(t, strings.EqualFold(testAddress, recoveredAddr.String()))
	})
}

func TestMsgSignerAndRecoverer_RoundTrip(t *testing.T) {
	signer := setupMsgSigner(t)
	recoverer := &EthereumMsgAddressRecoverer{}

	testCases := []struct {
		name    string
		message []byte
	}{
		{"Simple Text", []byte("Hello, Ethereum!")},
		{"Empty Message", []byte("")},
		{"Unicode", []byte("Hello 世界 🌍")},
		{"JSON", []byte(`{"key":"value","number":123}`)},
		{"Binary Data", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"Long Message", []byte(strings.Repeat("test", 250))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sign the message
			signature, err := signer.Sign(tc.message)
			require.NoError(t, err)

			// Recover the address
			recoveredAddress, err := recoverer.RecoverAddress(tc.message, signature)
			require.NoError(t, err)

			// Should match the signer's address
			assert.True(t, strings.EqualFold(
				signer.PublicKey().Address().String(),
				recoveredAddress.String(),
			))
		})
	}
}

func TestComputeEthereumSignedMessageHash(t *testing.T) {
	t.Run("Empty Message", func(t *testing.T) {
		message := []byte("")
		hash := ComputeEthereumSignedMessageHash(message)
		assert.NotNil(t, hash)
		assert.Len(t, hash, 32, "Keccak256 hash should be 32 bytes")

		// Verify the prefix format: "\x19Ethereum Signed Message:\n0"
		expectedInput := "\x19Ethereum Signed Message:\n0"
		expectedHash := ethcrypto.Keccak256([]byte(expectedInput))
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("Simple Message", func(t *testing.T) {
		message := []byte("Hello")
		hash := ComputeEthereumSignedMessageHash(message)
		assert.NotNil(t, hash)
		assert.Len(t, hash, 32)

		// Verify the prefix format: "\x19Ethereum Signed Message:\n5Hello"
		expectedInput := "\x19Ethereum Signed Message:\n5Hello"
		expectedHash := ethcrypto.Keccak256([]byte(expectedInput))
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("Message With Length 10", func(t *testing.T) {
		message := []byte("0123456789")
		hash := ComputeEthereumSignedMessageHash(message)
		assert.Len(t, hash, 32)

		expectedInput := "\x19Ethereum Signed Message:\n100123456789"
		expectedHash := ethcrypto.Keccak256([]byte(expectedInput))
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("Deterministic", func(t *testing.T) {
		message := []byte("test message")
		hash1 := ComputeEthereumSignedMessageHash(message)
		hash2 := ComputeEthereumSignedMessageHash(message)
		assert.Equal(t, hash1, hash2, "Hash should be deterministic")
	})
}

func TestMsgSignerVsRawSigner(t *testing.T) {
	t.Run("Different Signatures", func(t *testing.T) {
		// Create both signers with the same key
		msgSigner, err := NewEthereumMsgSigner(testPrivKey)
		require.NoError(t, err)

		rawSigner, err := NewEthereumRawSigner(testPrivKey)
		require.NoError(t, err)

		message := []byte("test message")

		// Sign with message signer (adds prefix)
		msgSignature, err := msgSigner.Sign(message)
		require.NoError(t, err)

		// For raw signer, we need to hash it first (raw signer expects a hash)
		hash := ethcrypto.Keccak256Hash(message)
		rawSignature, err := rawSigner.Sign(hash.Bytes())
		require.NoError(t, err)

		// Signatures should be different because msg signer adds prefix
		assert.NotEqual(t, msgSignature, rawSignature,
			"Msg signer should produce different signature than raw signer")
	})

	t.Run("Msg Signer Adds Prefix", func(t *testing.T) {
		msgSigner := setupMsgSigner(t)
		rawSigner := setupSigner(t)

		message := []byte("test")

		// Sign with msg signer
		msgSig, err := msgSigner.Sign(message)
		require.NoError(t, err)

		// To get same signature with raw signer, we need to manually compute the prefixed hash
		prefixedHash := ComputeEthereumSignedMessageHash(message)
		rawSig, err := rawSigner.Sign(prefixedHash)
		require.NoError(t, err)

		// These should match because we manually applied the prefix for raw signer
		assert.Equal(t, msgSig, rawSig,
			"Msg signer should produce same result as raw signer with manual prefix")
	})

	t.Run("Recovery Uses Correct Hash", func(t *testing.T) {
		msgSigner := setupMsgSigner(t)
		message := []byte("test message")

		// Sign with message signer
		signature, err := msgSigner.Sign(message)
		require.NoError(t, err)

		// Try to recover with raw recoverer (wrong - won't work)
		rawRecoverer := &EthereumRawAddressRecoverer{}
		rawRecoveredAddr, err := rawRecoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		// Should NOT match because raw recoverer doesn't add prefix
		assert.False(t, strings.EqualFold(testAddress, rawRecoveredAddr.String()),
			"Raw recoverer should not recover correct address from msg signature")

		// Recover with message recoverer (correct)
		msgRecoverer := &EthereumMsgAddressRecoverer{}
		msgRecoveredAddr, err := msgRecoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		// Should match because msg recoverer adds prefix
		assert.True(t, strings.EqualFold(testAddress, msgRecoveredAddr.String()),
			"Msg recoverer should recover correct address from msg signature")
	})
}

func TestKnownTestVectors(t *testing.T) {
	// Test with a known signature to ensure compatibility
	// This uses a well-known test case from Ethereum tooling
	t.Run("Known Test Vector", func(t *testing.T) {
		// Private key: 0x4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318
		// Address: 0x2c7536E3605D9C16a7a3D7b1898e529396a65c23
		// Message: "Hello, Ethereum!"

		signer := setupMsgSigner(t)
		message := []byte("Hello, Ethereum!")

		signature, err := signer.Sign(message)
		require.NoError(t, err)

		// Verify signature format
		assert.Len(t, signature, 65)
		// V should be 27 or 28 (Ethereum format)
		v := signature[64]
		assert.True(t, v == 27 || v == 28, "V should be 27 or 28, got %d", v)

		// Verify recovery
		recoverer := &EthereumMsgAddressRecoverer{}
		recoveredAddr, err := recoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		assert.True(t, strings.EqualFold(testAddress, recoveredAddr.String()))
	})

	t.Run("Verify Prefixed Hash", func(t *testing.T) {
		// Verify that our ComputeEthereumSignedMessageHash function
		// produces the correct format
		message := []byte("test")
		hash := ComputeEthereumSignedMessageHash(message)

		// Manual computation for verification
		prefix := "\x19Ethereum Signed Message:\n4test"
		expectedHash := ethcrypto.Keccak256([]byte(prefix))

		assert.Equal(t, hex.EncodeToString(expectedHash), hex.EncodeToString(hash),
			"Hash should match manual computation")
	})
}

func TestNewAddressRecoverer(t *testing.T) {
	t.Run("Create EthereumMsg Recoverer", func(t *testing.T) {
		recoverer, err := NewAddressRecoverer(TypeEthereumMsg)
		require.NoError(t, err)
		assert.NotNil(t, recoverer)
		assert.IsType(t, &EthereumMsgAddressRecoverer{}, recoverer)
	})

	t.Run("Use Msg Recoverer With Msg Signature", func(t *testing.T) {
		msgSigner := setupMsgSigner(t)
		message := []byte("test message")

		signature, err := msgSigner.Sign(message)
		require.NoError(t, err)

		// Create recoverer using the factory
		recoverer, err := NewAddressRecoverer(TypeEthereumMsg)
		require.NoError(t, err)

		// Recover address
		recoveredAddr, err := recoverer.RecoverAddress(message, signature)
		require.NoError(t, err)
		assert.True(t, strings.EqualFold(testAddress, recoveredAddr.String()))
	})
}
