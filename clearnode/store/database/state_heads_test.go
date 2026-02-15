package database

import (
	"testing"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStoreUserState_WithTransaction verifies that StoreUserState works correctly within a transaction
func TestStoreUserState_WithTransaction(t *testing.T) {
	t.Run("Success - Store new state in transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create home channel first
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Store state within transaction
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			// GetLastUserState should create initial head and lock it
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state := core.State{
				ID:            "state123",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       1,
				HomeChannelID: &homeChannelID,
				Transition: core.Transition{
					Type:      core.TransitionTypeHomeDeposit,
					AccountID: homeChannelID,
					Amount:    decimal.NewFromInt(1000),
				},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(1000),
					UserNetFlow: decimal.NewFromInt(1000),
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				UserSig: &userSig,
				NodeSig: &nodeSig,
			}

			return txStore.StoreUserState(state)
		})
		require.NoError(t, err)

		// Verify state was stored in history
		var dbState State
		err = db.Where("id = ?", "state123").First(&dbState).Error
		require.NoError(t, err)
		assert.Equal(t, "state123", dbState.ID)
		assert.Equal(t, uint64(1), dbState.Version)

		// Verify head was created
		var head StateHead
		err = db.Where("user_wallet = ? AND asset = ?", "0xuser123", "USDC").First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(1), head.Version)
		assert.NotNil(t, head.HistoryID)
		assert.Equal(t, "state123", *head.HistoryID)
		assert.NotNil(t, head.LastSignedStateID)
		assert.Equal(t, "state123", *head.LastSignedStateID)
	})

	t.Run("Error - StoreUserState without transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		state := core.State{
			ID:         "state456",
			Asset:      "USDC",
			UserWallet: "0xuser123",
			Epoch:      1,
			Version:    1,
			Transition: core.Transition{},
			HomeLedger: core.Ledger{
				UserBalance: decimal.NewFromInt(1000),
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
		}

		// Should fail because not in transaction
		err := store.StoreUserState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be called within a transaction")
	})
}

// TestHeadTable_LastSignedStateTracking verifies last_signed_state_id tracking
func TestHeadTable_LastSignedStateTracking(t *testing.T) {
	t.Run("Success - Last signed state ID updates correctly", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Store signed state (version 1)
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state1 := core.State{
				ID:            "state1",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       1,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(1000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				UserSig: &userSig,
				NodeSig: &nodeSig,
			}
			return txStore.StoreUserState(state1)
		})
		require.NoError(t, err)

		// Verify last_signed_state_id is set to state1
		var head StateHead
		err = db.Where("user_wallet = ? AND asset = ?", "0xuser123", "USDC").First(&head).Error
		require.NoError(t, err)
		assert.NotNil(t, head.LastSignedStateID)
		assert.Equal(t, "state1", *head.LastSignedStateID)

		// Store unsigned state (version 2)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state2 := core.State{
				ID:            "state2",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       2,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(2000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				// No signatures
			}
			return txStore.StoreUserState(state2)
		})
		require.NoError(t, err)

		// Verify last_signed_state_id still points to state1
		err = db.Where("user_wallet = ? AND asset = ?", "0xuser123", "USDC").First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(2), head.Version) // Head version should be 2
		assert.NotNil(t, head.LastSignedStateID)
		assert.Equal(t, "state1", *head.LastSignedStateID) // Last signed still state1

		// Store another signed state (version 3)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state3 := core.State{
				ID:            "state3",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       3,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(3000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				UserSig: &userSig,
				NodeSig: &nodeSig,
			}
			return txStore.StoreUserState(state3)
		})
		require.NoError(t, err)

		// Verify last_signed_state_id now points to state3
		err = db.Where("user_wallet = ? AND asset = ?", "0xuser123", "USDC").First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(3), head.Version)
		assert.NotNil(t, head.LastSignedStateID)
		assert.Equal(t, "state3", *head.LastSignedStateID)
	})
}

// TestGetLastUserState_SignedFallback verifies the signed state fallback logic
func TestGetLastUserState_SignedFallback(t *testing.T) {
	t.Run("Success - Returns signed head when head is signed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Store signed state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state := core.State{
				ID:            "state1",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       1,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(1000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				UserSig: &userSig,
				NodeSig: &nodeSig,
			}
			return txStore.StoreUserState(state)
		})
		require.NoError(t, err)

		// GetLastUserState with signed=true should return the head
		result, err := store.GetLastUserState("0xuser123", "USDC", true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, uint64(1), result.Version)
	})

	t.Run("Success - Returns last_signed_state_id when head is unsigned", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"
		userSig := "0xusersig"
		nodeSig := "0xnodesig"

		// Create channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Store signed state (version 1)
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state1 := core.State{
				ID:            "state1",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       1,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(1000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				UserSig: &userSig,
				NodeSig: &nodeSig,
			}
			return txStore.StoreUserState(state1)
		})
		require.NoError(t, err)

		// Store unsigned state (version 2)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state2 := core.State{
				ID:            "state2",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       2,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(2000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
				// No signatures
			}
			return txStore.StoreUserState(state2)
		})
		require.NoError(t, err)

		// GetLastUserState with signed=true should return state1 (from history)
		result, err := store.GetLastUserState("0xuser123", "USDC", true)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state1", result.ID)
		assert.Equal(t, uint64(1), result.Version)
		assert.True(t, result.HomeLedger.UserBalance.Equal(decimal.NewFromInt(1000)))

		// GetLastUserState with signed=false should return state2 (from head)
		result, err = store.GetLastUserState("0xuser123", "USDC", false)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "state2", result.ID)
		assert.Equal(t, uint64(2), result.Version)
		assert.True(t, result.HomeLedger.UserBalance.Equal(decimal.NewFromInt(2000)))
	})
}

// TestOptimisticLocking verifies the optimistic locking mechanism
func TestOptimisticLocking(t *testing.T) {
	t.Run("Error - Concurrent modification detected", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create channel
		homeChannel := core.Channel{
			ChannelID:    homeChannelID,
			UserWallet:   "0xuser123",
			Type:         core.ChannelTypeHome,
			BlockchainID: 1,
			TokenAddress: "0xtoken123",
			Status:       core.ChannelStatusOpen,
		}
		require.NoError(t, store.CreateChannel(homeChannel))

		// Store initial state
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			_, err := txStore.GetLastUserState("0xuser123", "USDC", false)
			require.NoError(t, err)

			state1 := core.State{
				ID:            "state1",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       1,
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(1000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
			}
			return txStore.StoreUserState(state1)
		})
		require.NoError(t, err)

		// Try to store a state with wrong previous version (simulating concurrent modification)
		err = store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			// Try to store version 3 when current version is 1
			// This simulates someone else having stored version 2
			state3 := core.State{
				ID:            "state3",
				Asset:         "USDC",
				UserWallet:    "0xuser123",
				Epoch:         1,
				Version:       3, // Skipped version 2
				HomeChannelID: &homeChannelID,
				Transition:    core.Transition{},
				HomeLedger: core.Ledger{
					UserBalance: decimal.NewFromInt(3000),
					UserNetFlow: decimal.Zero,
					NodeBalance: decimal.Zero,
					NodeNetFlow: decimal.Zero,
				},
			}
			return txStore.StoreUserState(state3)
		})

		// Should fail due to optimistic locking (version mismatch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "concurrent modification")
	})
}

// TestGetLastUserState_InitialHeadCreation verifies initial head creation
func TestGetLastUserState_InitialHeadCreation(t *testing.T) {
	t.Run("Success - Creates initial head in transaction", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Call GetLastUserState in transaction for non-existent state
		var result *core.State
		err := store.ExecuteInTransaction(func(txStore DatabaseStore) error {
			var err error
			result, err = txStore.GetLastUserState("0xnewuser", "USDC", false)
			return err
		})
		require.NoError(t, err)

		// Should return nil (no state yet)
		assert.Nil(t, result)

		// But head should be created with version 0
		var head StateHead
		err = db.Where("user_wallet = ? AND asset = ?", "0xnewuser", "USDC").First(&head).Error
		require.NoError(t, err)
		assert.Equal(t, uint64(0), head.Version)
		assert.Equal(t, uint64(0), head.Epoch)
	})
}
