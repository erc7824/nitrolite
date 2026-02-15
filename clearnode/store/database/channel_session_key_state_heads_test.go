package database

import (
	"strings"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStoreChannelSessionKeyState_WithTransaction verifies that StoreChannelSessionKeyState works correctly within a transaction
func TestStoreChannelSessionKeyState_WithTransaction(t *testing.T) {
	t.Run("Success - Store new session key state in transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store state within transaction
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      []string{testAsset1, testAsset2},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig123",
			}

			return txStore.StoreChannelSessionKeyState(state)
		})
		require.NoError(t, err)

		// Verify state was stored in history
		var dbState ChannelSessionKeyStateV1
		err = db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&dbState).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), dbState.Version)

		// Verify head was created/updated
		var head ChannelSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), head.Version)
		assert.NotNil(t, head.HistoryID)
		assert.Equal(t, dbState.ID, *head.HistoryID)

		// Verify metadata hash is set correctly
		expectedHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(1, []string{testAsset1, testAsset2}, expiresAt.Unix())
		require.NoError(t, err)
		assert.Equal(t, strings.ToLower(expectedHash.Hex()), head.MetadataHash)
	})

	t.Run("Error - StoreChannelSessionKeyState without transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.ChannelSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			Assets:      []string{testAsset1},
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}

		// Should fail because not in transaction
		err := store.StoreChannelSessionKeyState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be called within a transaction")
	})
}

// TestChannelSessionKeyHeadTable_VersionTracking verifies head table properly tracks latest version
func TestChannelSessionKeyHeadTable_VersionTracking(t *testing.T) {
	t.Run("Success - Head always reflects latest version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store version 1
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state1 := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig_v1",
			}
			return txStore.StoreChannelSessionKeyState(state1)
		})
		require.NoError(t, err)

		// Verify head has version 1
		var head ChannelSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), head.Version)
		assert.Equal(t, "0xsig_v1", head.UserSig)

		// Store version 2 with different assets
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state2 := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     2,
				Assets:      []string{testAsset1, testAsset2},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig_v2",
			}
			return txStore.StoreChannelSessionKeyState(state2)
		})
		require.NoError(t, err)

		// Verify head now has version 2
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(2), head.Version)
		assert.Equal(t, "0xsig_v2", head.UserSig)

		// GetLastChannelSessionKeyStates should return version 2
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, uint64(2), results[0].Version)
		assert.Len(t, results[0].Assets, 2)
	})

	t.Run("Success - Lower version doesn't overwrite head", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store version 5
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state5 := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     5,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig_v5",
			}
			return txStore.StoreChannelSessionKeyState(state5)
		})
		require.NoError(t, err)

		// Try to store version 3 (lower than current head)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state3 := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     3,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig_v3",
			}
			return txStore.StoreChannelSessionKeyState(state3)
		})
		require.NoError(t, err)

		// Head should still be version 5 (not overwritten by lower version)
		var head ChannelSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(5), head.Version)
		assert.Equal(t, "0xsig_v5", head.UserSig)
	})
}

// TestChannelSessionKeyState_ExpirationHandling verifies expiration is properly handled via heads
func TestChannelSessionKeyState_ExpirationHandling(t *testing.T) {
	t.Run("GetLastChannelSessionKeyStates returns empty for expired head", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store expired state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      []string{testAsset1},
				ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
				UserSig:     "0xsig",
			}
			return txStore.StoreChannelSessionKeyState(state)
		})
		require.NoError(t, err)

		// GetLastChannelSessionKeyStates should return empty
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)

		// GetLastChannelSessionKeyVersion should return 0
		version, err := store.GetLastChannelSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), version)
	})

	t.Run("ValidateChannelSessionKeyForAsset rejects expired head", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(-1 * time.Hour).UTC()

		// Store expired state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig",
			}
			return txStore.StoreChannelSessionKeyState(state)
		})
		require.NoError(t, err)

		// Compute metadata hash
		metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(1, []string{testAsset1}, expiresAt.Unix())
		require.NoError(t, err)

		// ValidateChannelSessionKeyForAsset should return false
		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, strings.ToLower(metadataHash.Hex()))
		require.NoError(t, err)
		assert.False(t, valid)
	})
}

// TestChannelSessionKeyState_MetadataHashValidation verifies metadata hash is properly stored and validated
func TestChannelSessionKeyState_MetadataHashValidation(t *testing.T) {
	t.Run("Success - ValidateChannelSessionKeyForAsset uses head table", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		assets := []string{testAsset1, testAsset2}

		// Store state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      assets,
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig",
			}
			return txStore.StoreChannelSessionKeyState(state)
		})
		require.NoError(t, err)

		// Compute expected metadata hash
		metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(1, assets, expiresAt.Unix())
		require.NoError(t, err)

		// Verify head has correct metadata hash
		var head ChannelSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, strings.ToLower(metadataHash.Hex()), head.MetadataHash)

		// ValidateChannelSessionKeyForAsset should use head table
		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, strings.ToLower(metadataHash.Hex()))
		require.NoError(t, err)
		assert.True(t, valid)

		// Also valid for second asset
		valid, err = store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset2, strings.ToLower(metadataHash.Hex()))
		require.NoError(t, err)
		assert.True(t, valid)

		// Invalid for asset not in list
		valid, err = store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset3, strings.ToLower(metadataHash.Hex()))
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("Failure - ValidateChannelSessionKeyForAsset rejects wrong metadata hash", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store state with testAsset1
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsig",
			}
			return txStore.StoreChannelSessionKeyState(state)
		})
		require.NoError(t, err)

		// Try to validate with wrong metadata hash (using testAsset2's hash)
		wrongHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(1, []string{testAsset2}, expiresAt.Unix())
		require.NoError(t, err)

		valid, err := store.ValidateChannelSessionKeyForAsset(testUser1, testSessionKey, testAsset1, strings.ToLower(wrongHash.Hex()))
		require.NoError(t, err)
		assert.False(t, valid)
	})
}

// TestChannelSessionKeyState_HeadBatchQuery verifies GetLastChannelSessionKeyStates uses heads efficiently
func TestChannelSessionKeyState_HeadBatchQuery(t *testing.T) {
	t.Run("Success - Batch query returns all non-expired heads", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store multiple session keys with multiple versions each
		keys := []string{testKeyA, testKeyB, "0xcccccccccccccccccccccccccccccccccccccccc"}
		for i, key := range keys {
			for v := uint64(1); v <= 3; v++ {
				err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
					state := core.ChannelSessionKeyStateV1{
						UserAddress: testUser1,
						SessionKey:  key,
						Version:     v,
						Assets:      []string{testAsset1},
						ExpiresAt:   expiresAt,
						UserSig:     "0xsig",
					}
					return txStore.StoreChannelSessionKeyState(state)
				})
				require.NoError(t, err)
			}

			// Add expired version for last key
			if i == len(keys)-1 {
				err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
					state := core.ChannelSessionKeyStateV1{
						UserAddress: testUser1,
						SessionKey:  key,
						Version:     4,
						Assets:      []string{testAsset1},
						ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
						UserSig:     "0xsig_expired",
					}
					return txStore.StoreChannelSessionKeyState(state)
				})
				require.NoError(t, err)
			}
		}

		// Batch query should return only latest non-expired version per key
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)

		// Should return 2 results (last key is expired, so filtered out)
		// When an expired version is stored that's newer than a non-expired version,
		// the head gets updated to the expired version and is then filtered out by queries
		assert.Len(t, results, 2)

		for _, result := range results {
			// Both returned keys should be version 3
			assert.Equal(t, uint64(3), result.Version)
		}
	})

	t.Run("Success - Batch query properly fetches assets for all heads", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour).UTC()

		// Store multiple keys with different asset sets
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			stateA := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testKeyA,
				Version:     1,
				Assets:      []string{testAsset1},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsigA",
			}
			if err := txStore.StoreChannelSessionKeyState(stateA); err != nil {
				return err
			}

			stateB := core.ChannelSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testKeyB,
				Version:     1,
				Assets:      []string{testAsset1, testAsset2, testAsset3},
				ExpiresAt:   expiresAt,
				UserSig:     "0xsigB",
			}
			return txStore.StoreChannelSessionKeyState(stateB)
		})
		require.NoError(t, err)

		// Batch query should return both with correct assets
		results, err := store.GetLastChannelSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, result := range results {
			if result.SessionKey == testKeyA {
				assert.Len(t, result.Assets, 1)
				assert.Contains(t, result.Assets, testAsset1)
			} else if result.SessionKey == testKeyB {
				assert.Len(t, result.Assets, 3)
				assert.Contains(t, result.Assets, testAsset1)
				assert.Contains(t, result.Assets, testAsset2)
				assert.Contains(t, result.Assets, testAsset3)
			}
		}
	})
}
