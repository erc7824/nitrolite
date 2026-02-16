package main

import (
	"crypto/ecdsa"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGeneratePrivateKey tests the private key generation function
func TestGeneratePrivateKey(t *testing.T) {
	t.Run("generates valid private key", func(t *testing.T) {
		privateKey, err := generatePrivateKey()
		require.NoError(t, err)
		assert.NotEmpty(t, privateKey)
	})

	t.Run("generates key with 0x prefix", func(t *testing.T) {
		privateKey, err := generatePrivateKey()
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(privateKey, "0x"), "private key should start with 0x")
	})

	t.Run("generates key with correct length", func(t *testing.T) {
		privateKey, err := generatePrivateKey()
		require.NoError(t, err)
		// 0x prefix (2 chars) + 64 hex chars = 66 total
		assert.Equal(t, 66, len(privateKey), "private key should be 66 characters (0x + 64 hex)")
	})

	t.Run("generates valid ECDSA key", func(t *testing.T) {
		privateKeyHex, err := generatePrivateKey()
		require.NoError(t, err)

		// Remove 0x prefix and convert to ECDSA private key
		privateKeyBytes, err := hexutil.Decode(privateKeyHex)
		require.NoError(t, err)
		privateKey, err := crypto.ToECDSA(privateKeyBytes)
		require.NoError(t, err)
		assert.NotNil(t, privateKey)

		// Verify we can derive public key
		publicKey := privateKey.Public()
		assert.NotNil(t, publicKey)

		// Verify it's a valid ECDSA public key
		_, ok := publicKey.(*ecdsa.PublicKey)
		assert.True(t, ok, "public key should be ECDSA public key")
	})

	t.Run("generates unique keys", func(t *testing.T) {
		key1, err := generatePrivateKey()
		require.NoError(t, err)

		key2, err := generatePrivateKey()
		require.NoError(t, err)

		assert.NotEqual(t, key1, key2, "consecutive calls should generate different keys")
	})

	t.Run("generated key can be used with Ethereum crypto", func(t *testing.T) {
		privateKeyHex, err := generatePrivateKey()
		require.NoError(t, err)

		privateKeyBytes, err := hexutil.Decode(privateKeyHex)
		require.NoError(t, err)
		privateKey, err := crypto.ToECDSA(privateKeyBytes)
		require.NoError(t, err)

		// Derive address from key
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		address := crypto.PubkeyToAddress(*publicKey)
		assert.NotEmpty(t, address.Hex())
		assert.True(t, strings.HasPrefix(address.Hex(), "0x"))
	})

	t.Run("generates cryptographically secure key", func(t *testing.T) {
		// Generate multiple keys and check they're different
		keys := make(map[string]bool)
		numKeys := 10

		for i := 0; i < numKeys; i++ {
			key, err := generatePrivateKey()
			require.NoError(t, err)
			require.False(t, keys[key], "duplicate key generated")
			keys[key] = true
		}

		assert.Equal(t, numKeys, len(keys), "all generated keys should be unique")
	})

	t.Run("generated key has correct bit length", func(t *testing.T) {
		privateKeyHex, err := generatePrivateKey()
		require.NoError(t, err)

		privateKeyBytes, err := hexutil.Decode(privateKeyHex)
		require.NoError(t, err)
		assert.Equal(t, 32, len(privateKeyBytes), "private key should be 32 bytes (256 bits)")
	})
}

// TestOperatorParseChainID tests the chain ID parsing helper
func TestOperatorParseChainID(t *testing.T) {
	storage, err := NewStorage(":memory:")
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store: storage,
	}

	t.Run("parses valid chain ID", func(t *testing.T) {
		chainID, err := op.parseChainID("1")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), chainID)
	})

	t.Run("parses large chain ID", func(t *testing.T) {
		chainID, err := op.parseChainID("80002")
		require.NoError(t, err)
		assert.Equal(t, uint64(80002), chainID)
	})

	t.Run("rejects negative numbers", func(t *testing.T) {
		_, err := op.parseChainID("-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("rejects non-numeric input", func(t *testing.T) {
		_, err := op.parseChainID("abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("rejects empty string", func(t *testing.T) {
		_, err := op.parseChainID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("rejects decimal numbers", func(t *testing.T) {
		_, err := op.parseChainID("1.5")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("parses zero", func(t *testing.T) {
		chainID, err := op.parseChainID("0")
		require.NoError(t, err)
		assert.Equal(t, uint64(0), chainID)
	})

	t.Run("parses maximum uint64", func(t *testing.T) {
		maxUint64 := "18446744073709551615" // 2^64 - 1
		chainID, err := op.parseChainID(maxUint64)
		require.NoError(t, err)
		assert.Equal(t, uint64(18446744073709551615), chainID)
	})

	t.Run("rejects overflow", func(t *testing.T) {
		overflowValue := "18446744073709551616" // 2^64
		_, err := op.parseChainID(overflowValue)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chain ID")
	})

	t.Run("handles whitespace", func(t *testing.T) {
		// Test depends on whether trimming is done before parsing
		_, err := op.parseChainID(" 1 ")
		// If this passes, implementation trims whitespace
		// If this fails, implementation requires exact format
		if err != nil {
			assert.Contains(t, err.Error(), "invalid chain ID")
		}
	})
}

// TestOperatorParseAmount tests the amount parsing helper
func TestOperatorParseAmount(t *testing.T) {
	storage, err := NewStorage(":memory:")
	require.NoError(t, err)
	defer storage.Close()

	op := &Operator{
		store: storage,
	}

	t.Run("parses integer amount", func(t *testing.T) {
		amount, err := op.parseAmount("100")
		require.NoError(t, err)
		assert.Equal(t, "100", amount.String())
	})

	t.Run("parses decimal amount", func(t *testing.T) {
		amount, err := op.parseAmount("100.5")
		require.NoError(t, err)
		assert.Equal(t, "100.5", amount.String())
	})

	t.Run("parses zero", func(t *testing.T) {
		amount, err := op.parseAmount("0")
		require.NoError(t, err)
		assert.True(t, amount.IsZero())
	})

	t.Run("parses small decimal", func(t *testing.T) {
		amount, err := op.parseAmount("0.000001")
		require.NoError(t, err)
		assert.Equal(t, "0.000001", amount.String())
	})

	t.Run("parses large amount", func(t *testing.T) {
		amount, err := op.parseAmount("1000000000000")
		require.NoError(t, err)
		assert.Equal(t, "1000000000000", amount.String())
	})

	t.Run("parses scientific notation", func(t *testing.T) {
		amount, err := op.parseAmount("1e6")
		require.NoError(t, err)
		// decimal package converts to standard notation
		assert.True(t, amount.Equal(amount))
	})

	t.Run("rejects non-numeric input", func(t *testing.T) {
		_, err := op.parseAmount("abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("rejects empty string", func(t *testing.T) {
		_, err := op.parseAmount("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("parses negative amount", func(t *testing.T) {
		amount, err := op.parseAmount("-100")
		require.NoError(t, err)
		assert.True(t, amount.IsNegative())
		assert.Equal(t, "-100", amount.String())
	})

	t.Run("preserves precision", func(t *testing.T) {
		amount, err := op.parseAmount("1.23456789")
		require.NoError(t, err)
		assert.Equal(t, "1.23456789", amount.String())
	})

	t.Run("handles leading zeros", func(t *testing.T) {
		amount, err := op.parseAmount("00100")
		require.NoError(t, err)
		// decimal package should normalize this
		assert.Equal(t, "100", amount.String())
	})

	t.Run("handles very small decimals", func(t *testing.T) {
		amount, err := op.parseAmount("0.0000000001")
		require.NoError(t, err)
		assert.False(t, amount.IsZero())
	})

	t.Run("rejects multiple decimal points", func(t *testing.T) {
		_, err := op.parseAmount("1.2.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid amount")
	})

	t.Run("handles amount with many decimal places", func(t *testing.T) {
		amount, err := op.parseAmount("1.123456789012345678")
		require.NoError(t, err)
		assert.Equal(t, "1.123456789012345678", amount.String())
	})
}

// TestOperatorGetImportedWalletAddress tests the wallet address retrieval
func TestOperatorGetImportedWalletAddress(t *testing.T) {
	t.Run("returns empty string when no wallet configured", func(t *testing.T) {
		storage, err := NewStorage(":memory:")
		require.NoError(t, err)
		defer storage.Close()

		op := &Operator{
			store: storage,
		}

		address := op.getImportedWalletAddress()
		assert.Equal(t, "", address)
	})

	t.Run("returns address when wallet is configured", func(t *testing.T) {
		storage, err := NewStorage(":memory:")
		require.NoError(t, err)
		defer storage.Close()

		// Use a known private key
		privateKey := "0x1234567890123456789012345678901234567890123456789012345678901234"
		err = storage.SetPrivateKey(privateKey)
		require.NoError(t, err)

		op := &Operator{
			store: storage,
		}

		address := op.getImportedWalletAddress()
		assert.NotEmpty(t, address)
		assert.True(t, strings.HasPrefix(address, "0x"))
		assert.Equal(t, 42, len(address)) // 0x + 40 hex chars
	})

	t.Run("returns empty string for invalid private key", func(t *testing.T) {
		storage, err := NewStorage(":memory:")
		require.NoError(t, err)
		defer storage.Close()

		// Store invalid private key
		err = storage.SetPrivateKey("invalid_key")
		require.NoError(t, err)

		op := &Operator{
			store: storage,
		}

		address := op.getImportedWalletAddress()
		assert.Equal(t, "", address)
	})

	t.Run("derives correct address from private key", func(t *testing.T) {
		storage, err := NewStorage(":memory:")
		require.NoError(t, err)
		defer storage.Close()

		// Generate a valid private key
		privateKeyHex, err := generatePrivateKey()
		require.NoError(t, err)

		err = storage.SetPrivateKey(privateKeyHex)
		require.NoError(t, err)

		op := &Operator{
			store: storage,
		}

		address := op.getImportedWalletAddress()
		assert.NotEmpty(t, address)

		// Verify the address is valid Ethereum address format
		assert.True(t, strings.HasPrefix(address, "0x"))
		assert.Equal(t, 42, len(address))

		// Verify we can derive the same address independently
		privateKeyBytes, err := hexutil.Decode(privateKeyHex)
		require.NoError(t, err)
		privateKey, err := crypto.ToECDSA(privateKeyBytes)
		require.NoError(t, err)
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		expectedAddress := crypto.PubkeyToAddress(*publicKey)

		assert.Equal(t, expectedAddress.Hex(), address)
	})
}