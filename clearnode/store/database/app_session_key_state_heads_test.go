package database

import (
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStoreAppSessionKeyState_WithTransaction verifies that StoreAppSessionKeyState works correctly within a transaction
func TestStoreAppSessionKeyState_WithTransaction(t *testing.T) {
	t.Run("Success - Store new session key state in transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store state within transaction
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			// GetLastAppSessionKeyState should work in transaction
			_, err := txStore.GetLastAppSessionKeyState(testUser1, testSessionKey)
			require.NoError(t, err)

			state := app.AppSessionKeyStateV1{
				UserAddress:    testUser1,
				SessionKey:     testSessionKey,
				Version:        1,
				ApplicationIDs: []string{testApp1},
				AppSessionIDs:  []string{testSess1},
				ExpiresAt:      time.Now().Add(24 * time.Hour),
				UserSig:        "0xsig123",
			}

			return txStore.StoreAppSessionKeyState(state)
		})
		require.NoError(t, err)

		// Verify state was stored in history
		var dbState AppSessionKeyStateV1
		err = db.Where("user_address = ? AND session_key = ? AND version = ?",
			testUser1, testSessionKey, 1).First(&dbState).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), dbState.Version)

		// Verify head was created/updated
		var head AppSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), head.Version)
		assert.NotNil(t, head.HistoryID)
		assert.Equal(t, dbState.ID, *head.HistoryID)
	})

	t.Run("Error - StoreAppSessionKeyState without transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := app.AppSessionKeyStateV1{
			UserAddress: testUser1,
			SessionKey:  testSessionKey,
			Version:     1,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			UserSig:     "0xsig",
		}

		// Should fail because not in transaction
		err := store.StoreAppSessionKeyState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be called within a transaction")
	})
}

// TestAppSessionKeyHeadTable_VersionTracking verifies head table properly tracks latest version
func TestAppSessionKeyHeadTable_VersionTracking(t *testing.T) {
	t.Run("Success - Head always reflects latest version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store version 1
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state1 := app.AppSessionKeyStateV1{
				UserAddress:    testUser1,
				SessionKey:     testSessionKey,
				Version:        1,
				ApplicationIDs: []string{testApp1},
				ExpiresAt:      time.Now().Add(24 * time.Hour),
				UserSig:        "0xsig_v1",
			}
			return txStore.StoreAppSessionKeyState(state1)
		})
		require.NoError(t, err)

		// Verify head has version 1
		var head AppSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), head.Version)
		assert.Equal(t, "0xsig_v1", head.UserSig)

		// Store version 2 with different applications
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state2 := app.AppSessionKeyStateV1{
				UserAddress:    testUser1,
				SessionKey:     testSessionKey,
				Version:        2,
				ApplicationIDs: []string{testApp1, testApp2},
				ExpiresAt:      time.Now().Add(24 * time.Hour),
				UserSig:        "0xsig_v2",
			}
			return txStore.StoreAppSessionKeyState(state2)
		})
		require.NoError(t, err)

		// Verify head now has version 2
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(2), head.Version)
		assert.Equal(t, "0xsig_v2", head.UserSig)

		// GetLastAppSessionKeyState should return version 2
		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, uint64(2), result.Version)
		assert.Len(t, result.ApplicationIDs, 2)
	})

	t.Run("Success - Lower version doesn't overwrite head", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store version 5
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state5 := app.AppSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     5,
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				UserSig:     "0xsig_v5",
			}
			return txStore.StoreAppSessionKeyState(state5)
		})
		require.NoError(t, err)

		// Try to store version 3 (lower than current head)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state3 := app.AppSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     3,
				ExpiresAt:   time.Now().Add(24 * time.Hour),
				UserSig:     "0xsig_v3",
			}
			return txStore.StoreAppSessionKeyState(state3)
		})
		require.NoError(t, err)

		// Head should still be version 5 (not overwritten by lower version)
		var head AppSessionKeyHeadV1
		err = db.Where("user_address = ? AND session_key = ?", testUser1, testSessionKey).First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(5), head.Version)
		assert.Equal(t, "0xsig_v5", head.UserSig)
	})
}

// TestAppSessionKeyState_ExpirationHandling verifies expiration is properly handled via heads
func TestAppSessionKeyState_ExpirationHandling(t *testing.T) {
	t.Run("GetLastAppSessionKeyState returns nil for expired head", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store expired state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := app.AppSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
				UserSig:     "0xsig",
			}
			return txStore.StoreAppSessionKeyState(state)
		})
		require.NoError(t, err)

		// GetLastAppSessionKeyState should return nil
		result, err := store.GetLastAppSessionKeyState(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Nil(t, result)

		// GetLastAppSessionKeyVersion should return 0
		version, err := store.GetLastAppSessionKeyVersion(testUser1, testSessionKey)
		require.NoError(t, err)
		assert.Equal(t, uint64(0), version)

		// GetLastAppSessionKeyStates should return empty
		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("GetLastAppSessionKeyState with lock respects expiration", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store expired state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			state := app.AppSessionKeyStateV1{
				UserAddress: testUser1,
				SessionKey:  testSessionKey,
				Version:     1,
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
				UserSig:     "0xsig",
			}
			return txStore.StoreAppSessionKeyState(state)
		})
		require.NoError(t, err)

		// GetLastAppSessionKeyState in transaction should also return nil
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			result, err := txStore.GetLastAppSessionKeyState(testUser1, testSessionKey)
			require.NoError(t, err)
			assert.Nil(t, result)
			return nil
		})
		require.NoError(t, err)
	})
}

// TestAppSessionKeyState_HeadBatchQuery verifies GetLastAppSessionKeyStates uses heads efficiently
func TestAppSessionKeyState_HeadBatchQuery(t *testing.T) {
	t.Run("Success - Batch query returns all non-expired heads", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Store multiple session keys with multiple versions each
		keys := []string{testKeyA, testKeyB, "0xcccccccccccccccccccccccccccccccccccccccc"}
		for i, key := range keys {
			for v := uint64(1); v <= 3; v++ {
				err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
					state := app.AppSessionKeyStateV1{
						UserAddress:    testUser1,
						SessionKey:     key,
						Version:        v,
						ApplicationIDs: []string{testApp1},
						ExpiresAt:      time.Now().Add(24 * time.Hour),
						UserSig:        "0xsig",
					}
					return txStore.StoreAppSessionKeyState(state)
				})
				require.NoError(t, err)
			}

			// Add expired version for last key
			if i == len(keys)-1 {
				err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
					state := app.AppSessionKeyStateV1{
						UserAddress: testUser1,
						SessionKey:  key,
						Version:     4,
						ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
						UserSig:     "0xsig_expired",
					}
					return txStore.StoreAppSessionKeyState(state)
				})
				require.NoError(t, err)
			}
		}

		// Batch query should return only latest non-expired version per key
		results, err := store.GetLastAppSessionKeyStates(testUser1, nil)
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
}
