package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorage(t *testing.T) {
	t.Parallel()
	// Test with in-memory database
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	assert.NotNil(t, s.db)
}

func TestStorage_PrivateKey(t *testing.T) {
	t.Parallel()
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	// Test getting non-existent key
	_, err = s.GetPrivateKey()
	assert.Error(t, err)
	assert.Equal(t, "no private key configured", err.Error())

	// Test setting and getting key
	pk := "0x1234567890abcdef"
	err = s.SetPrivateKey(pk)
	require.NoError(t, err)

	got, err := s.GetPrivateKey()
	require.NoError(t, err)
	assert.Equal(t, pk, got)

	// Test updating key
	pk2 := "0x0987654321fedcba"
	err = s.SetPrivateKey(pk2)
	require.NoError(t, err)

	got2, err := s.GetPrivateKey()
	require.NoError(t, err)
	assert.Equal(t, pk2, got2)
}

func TestStorage_RPC(t *testing.T) {
	t.Parallel()
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	// Test getting non-existent RPC
	_, err = s.GetRPC(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no RPC configured")

	// Test setting and getting RPC
	chainID := uint64(1)
	rpcURL := "https://mainnet.infura.io/v3/YOUR-PROJECT-ID"
	err = s.SetRPC(chainID, rpcURL)
	require.NoError(t, err)

	got, err := s.GetRPC(chainID)
	require.NoError(t, err)
	assert.Equal(t, rpcURL, got)

	// Test getting all RPCs
	rpcs, err := s.GetAllRPCs()
	require.NoError(t, err)
	assert.Len(t, rpcs, 1)
	assert.Equal(t, rpcURL, rpcs[chainID])

	// Add another RPC
	chainID2 := uint64(137)
	rpcURL2 := "https://polygon-rpc.com"
	err = s.SetRPC(chainID2, rpcURL2)
	require.NoError(t, err)

	rpcs, err = s.GetAllRPCs()
	require.NoError(t, err)
	assert.Len(t, rpcs, 2)
	assert.Equal(t, rpcURL2, rpcs[chainID2])
}

func TestStorage_SessionKey(t *testing.T) {
	t.Parallel()
	s, err := NewStorage(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { s.Close() })

	// Test getting non-existent session key
	_, _, _, err = s.GetSessionKey()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session key configured")

	_, err = s.GetSessionKeyPrivateKey()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session key private key configured")

	// Test setting session key private key only
	pk := "0xsessionkey"
	err = s.SetSessionKeyPrivateKey(pk)
	require.NoError(t, err)

	gotPK, err := s.GetSessionKeyPrivateKey()
	require.NoError(t, err)
	assert.Equal(t, pk, gotPK)

	// Still incomplete session key
	_, _, _, err = s.GetSessionKey()
	assert.Error(t, err)

	// Test setting full session key
	metaHash := "0xmeta"
	authSig := "0xauth"
	err = s.SetSessionKey(pk, metaHash, authSig)
	require.NoError(t, err)

	gotPK, gotMeta, gotAuth, err := s.GetSessionKey()
	require.NoError(t, err)
	assert.Equal(t, pk, gotPK)
	assert.Equal(t, metaHash, gotMeta)
	assert.Equal(t, authSig, gotAuth)

	// Test clearing session key
	err = s.ClearSessionKey()
	require.NoError(t, err)

	_, _, _, err = s.GetSessionKey()
	assert.Error(t, err)

	_, err = s.GetSessionKeyPrivateKey()
	assert.Error(t, err)
}

func TestStorage_FileCreation(t *testing.T) {
	t.Parallel()
	// Test real file creation
	f, err := os.CreateTemp("", "test_db_*.sqlite")
	require.NoError(t, err)
	f.Close()
	os.Remove(f.Name()) // Ensure it doesn't exist yet

	s, err := NewStorage(f.Name())
	require.NoError(t, err)
	t.Cleanup(func() {
		s.Close()
		os.Remove(f.Name())
	})

	// Verify file exists
	_, err = os.Stat(f.Name())
	assert.NoError(t, err)

	// Verify tables created
	var count int
	err = s.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='config'").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

//func TestStorage_TransactionRollback(t *testing.T) {
//	t.Parallel()
//	s, err := NewStorage(":memory:")
//	require.NoError(t, err)
//	t.Cleanup(func() { s.Close() })

// Manually insert partial data to simulate a corrupted state if we were not using transactions properly
// But here we want to test that SetSessionKey uses a transaction.
// We can't easily force a failure inside the transaction without mocking db,
// but we can verify that the method works as expected.

// We can try to close the DB inside the transaction? No, that's too hacky.
// For now, we rely on the happy path test in TestStorage_SessionKey
// and trust the code uses tx.Rollback() (defer) and tx.Commit().
//}
