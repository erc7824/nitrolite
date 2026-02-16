package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStorage tests the creation of a new storage instance
func TestNewStorage(t *testing.T) {
	t.Run("creates storage with valid path", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		require.NotNil(t, storage)
		defer storage.Close()

		// Verify database file was created
		_, err = os.Stat(dbPath)
		assert.NoError(t, err, "database file should exist")
	})

	t.Run("creates tables on initialization", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		// Verify tables exist by attempting to query them
		var count int
		err = storage.db.QueryRow("SELECT COUNT(*) FROM config").Scan(&count)
		assert.NoError(t, err, "config table should exist")

		err = storage.db.QueryRow("SELECT COUNT(*) FROM rpcs").Scan(&count)
		assert.NoError(t, err, "rpcs table should exist")
	})

	t.Run("handles invalid database path gracefully", func(t *testing.T) {
		// Use an invalid path (directory that doesn't exist and can't be created)
		dbPath := "/nonexistent/invalid/path/test.db"

		storage, err := NewStorage(dbPath)
		// Note: SQLite might still open the db, but queries will fail
		// This test documents the behavior
		if err == nil && storage != nil {
			storage.Close()
		}
		// Test passes if it doesn't panic
	})
}

// TestPrivateKeyOperations tests storing and retrieving private keys
func TestPrivateKeyOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("set and get private key", func(t *testing.T) {
		expectedKey := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

		err := storage.SetPrivateKey(expectedKey)
		require.NoError(t, err)

		actualKey, err := storage.GetPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, expectedKey, actualKey)
	})

	t.Run("get private key when not configured", func(t *testing.T) {
		// Use a fresh database
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "empty.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		_, err = storage.GetPrivateKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no private key configured")
	})

	t.Run("update existing private key", func(t *testing.T) {
		key1 := "0x1111111111111111111111111111111111111111111111111111111111111111"
		key2 := "0x2222222222222222222222222222222222222222222222222222222222222222"

		err := storage.SetPrivateKey(key1)
		require.NoError(t, err)

		err = storage.SetPrivateKey(key2)
		require.NoError(t, err)

		actualKey, err := storage.GetPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, key2, actualKey)
	})

	t.Run("handles empty private key", func(t *testing.T) {
		err := storage.SetPrivateKey("")
		require.NoError(t, err)

		key, err := storage.GetPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})
}

// TestRPCOperations tests storing and retrieving RPC URLs
func TestRPCOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("set and get RPC for single chain", func(t *testing.T) {
		chainID := uint64(1)
		rpcURL := "https://mainnet.infura.io/v3/YOUR-PROJECT-ID"

		err := storage.SetRPC(chainID, rpcURL)
		require.NoError(t, err)

		actualURL, err := storage.GetRPC(chainID)
		require.NoError(t, err)
		assert.Equal(t, rpcURL, actualURL)
	})

	t.Run("get RPC when not configured", func(t *testing.T) {
		chainID := uint64(999)

		_, err := storage.GetRPC(chainID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no RPC configured for chain 999")
	})

	t.Run("update existing RPC", func(t *testing.T) {
		chainID := uint64(1)
		url1 := "https://rpc1.example.com"
		url2 := "https://rpc2.example.com"

		err := storage.SetRPC(chainID, url1)
		require.NoError(t, err)

		err = storage.SetRPC(chainID, url2)
		require.NoError(t, err)

		actualURL, err := storage.GetRPC(chainID)
		require.NoError(t, err)
		assert.Equal(t, url2, actualURL)
	})

	t.Run("set and get multiple RPCs", func(t *testing.T) {
		chains := map[uint64]string{
			1:     "https://mainnet.example.com",
			80002: "https://polygon-amoy.example.com",
			84532: "https://base-sepolia.example.com",
		}

		for chainID, rpcURL := range chains {
			err := storage.SetRPC(chainID, rpcURL)
			require.NoError(t, err)
		}

		for chainID, expectedURL := range chains {
			actualURL, err := storage.GetRPC(chainID)
			require.NoError(t, err)
			assert.Equal(t, expectedURL, actualURL)
		}
	})

	t.Run("get all RPCs when empty", func(t *testing.T) {
		// Use a fresh database
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "empty.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		rpcs, err := storage.GetAllRPCs()
		require.NoError(t, err)
		assert.Empty(t, rpcs)
	})

	t.Run("get all RPCs returns all stored RPCs", func(t *testing.T) {
		expectedRPCs := map[uint64]string{
			1:     "https://mainnet.example.com",
			80002: "https://polygon-amoy.example.com",
			84532: "https://base-sepolia.example.com",
		}

		for chainID, rpcURL := range expectedRPCs {
			err := storage.SetRPC(chainID, rpcURL)
			require.NoError(t, err)
		}

		actualRPCs, err := storage.GetAllRPCs()
		require.NoError(t, err)
		assert.Equal(t, expectedRPCs, actualRPCs)
	})
}

// TestSessionKeyPrivateKeyOperations tests session key private key storage
func TestSessionKeyPrivateKeyOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("set and get session key private key", func(t *testing.T) {
		expectedKey := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

		err := storage.SetSessionKeyPrivateKey(expectedKey)
		require.NoError(t, err)

		actualKey, err := storage.GetSessionKeyPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, expectedKey, actualKey)
	})

	t.Run("get session key private key when not configured", func(t *testing.T) {
		// Use a fresh database
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "empty.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		_, err = storage.GetSessionKeyPrivateKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no session key private key configured")
	})

	t.Run("update session key private key", func(t *testing.T) {
		key1 := "0xkey1111111111111111111111111111111111111111111111111111111111111"
		key2 := "0xkey2222222222222222222222222222222222222222222222222222222222222"

		err := storage.SetSessionKeyPrivateKey(key1)
		require.NoError(t, err)

		err = storage.SetSessionKeyPrivateKey(key2)
		require.NoError(t, err)

		actualKey, err := storage.GetSessionKeyPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, key2, actualKey)
	})
}

// TestSessionKeyOperations tests full session key storage (with metadata)
func TestSessionKeyOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("set and get complete session key", func(t *testing.T) {
		expectedPrivateKey := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
		expectedMetadataHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		expectedAuthSig := "0xsignature1234567890abcdef"

		err := storage.SetSessionKey(expectedPrivateKey, expectedMetadataHash, expectedAuthSig)
		require.NoError(t, err)

		actualPrivateKey, actualMetadataHash, actualAuthSig, err := storage.GetSessionKey()
		require.NoError(t, err)
		assert.Equal(t, expectedPrivateKey, actualPrivateKey)
		assert.Equal(t, expectedMetadataHash, actualMetadataHash)
		assert.Equal(t, expectedAuthSig, actualAuthSig)
	})

	t.Run("get session key when not configured", func(t *testing.T) {
		// Use a fresh database
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "empty.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		_, _, _, err = storage.GetSessionKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no session key configured")
	})

	t.Run("update session key atomically", func(t *testing.T) {
		key1 := "0xkey1"
		hash1 := "0xhash1"
		sig1 := "0xsig1"

		err := storage.SetSessionKey(key1, hash1, sig1)
		require.NoError(t, err)

		key2 := "0xkey2"
		hash2 := "0xhash2"
		sig2 := "0xsig2"

		err = storage.SetSessionKey(key2, hash2, sig2)
		require.NoError(t, err)

		actualKey, actualHash, actualSig, err := storage.GetSessionKey()
		require.NoError(t, err)
		assert.Equal(t, key2, actualKey)
		assert.Equal(t, hash2, actualHash)
		assert.Equal(t, sig2, actualSig)
	})

	t.Run("clear session key removes all session key data", func(t *testing.T) {
		err := storage.SetSessionKey("0xkey", "0xhash", "0xsig")
		require.NoError(t, err)

		err = storage.ClearSessionKey()
		require.NoError(t, err)

		_, _, _, err = storage.GetSessionKey()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no session key configured")
	})

	t.Run("clear session key when not set does not error", func(t *testing.T) {
		// Use a fresh database
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "empty.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)
		defer storage.Close()

		err = storage.ClearSessionKey()
		require.NoError(t, err)
	})

	t.Run("set session key does not affect wallet private key", func(t *testing.T) {
		walletKey := "0xwallet_key"
		err := storage.SetPrivateKey(walletKey)
		require.NoError(t, err)

		err = storage.SetSessionKey("0xsession", "0xhash", "0xsig")
		require.NoError(t, err)

		actualWalletKey, err := storage.GetPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, walletKey, actualWalletKey)
	})
}

// TestStorageClose tests database cleanup
func TestStorageClose(t *testing.T) {
	t.Run("close database successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)

		err = storage.Close()
		assert.NoError(t, err)
	})

	t.Run("operations fail after close", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")
		storage, err := NewStorage(dbPath)
		require.NoError(t, err)

		err = storage.Close()
		require.NoError(t, err)

		// Attempt to use closed database
		err = storage.SetPrivateKey("0xtest")
		assert.Error(t, err)
	})
}

// TestConcurrentAccess tests concurrent operations on storage
func TestConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("concurrent RPC writes", func(t *testing.T) {
		done := make(chan bool)

		// Write different RPCs concurrently
		for i := uint64(0); i < 10; i++ {
			chainID := i
			go func() {
				err := storage.SetRPC(chainID, "https://example.com")
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all RPCs were written
		rpcs, err := storage.GetAllRPCs()
		require.NoError(t, err)
		assert.Len(t, rpcs, 10)
	})

	t.Run("concurrent reads and writes", func(t *testing.T) {
		chainID := uint64(1)
		err := storage.SetRPC(chainID, "https://initial.com")
		require.NoError(t, err)

		done := make(chan bool)

		// Start multiple readers
		for i := 0; i < 5; i++ {
			go func() {
				_, err := storage.GetRPC(chainID)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Start multiple writers
		for i := 0; i < 5; i++ {
			url := "https://example.com"
			go func() {
				err := storage.SetRPC(chainID, url)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify final state is consistent
		finalURL, err := storage.GetRPC(chainID)
		require.NoError(t, err)
		assert.NotEmpty(t, finalURL)
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	storage, err := NewStorage(dbPath)
	require.NoError(t, err)
	defer storage.Close()

	t.Run("very long private key", func(t *testing.T) {
		longKey := "0x" + string(make([]byte, 10000))
		err := storage.SetPrivateKey(longKey)
		require.NoError(t, err)

		retrieved, err := storage.GetPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, longKey, retrieved)
	})

	t.Run("very long RPC URL", func(t *testing.T) {
		longURL := "https://example.com/" + string(make([]byte, 10000))
		err := storage.SetRPC(1, longURL)
		require.NoError(t, err)

		retrieved, err := storage.GetRPC(1)
		require.NoError(t, err)
		assert.Equal(t, longURL, retrieved)
	})

	t.Run("chain ID zero", func(t *testing.T) {
		err := storage.SetRPC(0, "https://example.com")
		require.NoError(t, err)

		retrieved, err := storage.GetRPC(0)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", retrieved)
	})

	t.Run("maximum chain ID", func(t *testing.T) {
		maxChainID := uint64(^uint64(0))
		err := storage.SetRPC(maxChainID, "https://example.com")
		require.NoError(t, err)

		retrieved, err := storage.GetRPC(maxChainID)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", retrieved)
	})

	t.Run("special characters in RPC URL", func(t *testing.T) {
		specialURL := "https://example.com/?key=value&special=!@#$%^&*()"
		err := storage.SetRPC(1, specialURL)
		require.NoError(t, err)

		retrieved, err := storage.GetRPC(1)
		require.NoError(t, err)
		assert.Equal(t, specialURL, retrieved)
	})
}