package database

import (
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannel_TableName(t *testing.T) {
	channel := Channel{}
	assert.Equal(t, "channels", channel.TableName())
}

func TestDBStore_CreateChannel(t *testing.T) {
	t.Run("Success - Create home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xhomechannel123").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, "0xhomechannel123", dbChannel.ChannelID)
		assert.Equal(t, "0xuser123", dbChannel.UserWallet)
		assert.Equal(t, core.ChannelTypeHome, dbChannel.Type)
		assert.Equal(t, uint64(1), dbChannel.BlockchainID)
		assert.Equal(t, "0xtoken123", dbChannel.Token)
		assert.Equal(t, uint32(86400), dbChannel.ChallengeDuration)
		assert.Equal(t, uint64(1), dbChannel.Nonce)
		assert.Equal(t, core.ChannelStatusOpen, dbChannel.Status)
		assert.Equal(t, uint64(0), dbChannel.StateVersion)
		assert.False(t, dbChannel.CreatedAt.IsZero())
		assert.False(t, dbChannel.UpdatedAt.IsZero())
	})

	t.Run("Success - Create escrow channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xescrowchannel456",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      137,
			TokenAddress:      "0xtoken456",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xescrowchannel456").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, core.ChannelTypeEscrow, dbChannel.Type)
		assert.Equal(t, uint64(137), dbChannel.BlockchainID)
	})

	t.Run("Error - Duplicate channel ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xchannel789",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Try to create again with same ID
		err = store.CreateChannel(channel)
		assert.Error(t, err)
	})

	t.Run("Success - Create channel with challenge expiration", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		expiresAt := time.Now().Add(24 * time.Hour)
		channel := core.Channel{
			ChannelID:          "0xchannel999",
			UserWallet:         "0xuser123",
			Type:               core.ChannelTypeHome,
			BlockchainID:       1,
			TokenAddress:       "0xtoken123",
			ChallengeDuration:  86400,
			ChallengeExpiresAt: &expiresAt,
			Nonce:              1,
			Status:             core.ChannelStatusChallenged,
			StateVersion:       1,
		}

		err := store.CreateChannel(channel)
		require.NoError(t, err)

		// Verify channel was created
		var dbChannel Channel
		err = db.Where("channel_id = ?", "0xchannel999").First(&dbChannel).Error
		require.NoError(t, err)

		assert.Equal(t, core.ChannelStatusChallenged, dbChannel.Status)
		assert.NotNil(t, dbChannel.ChallengeExpiresAt)
	})
}

func TestDBStore_GetChannelByID(t *testing.T) {
	t.Run("Success - Get existing channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}

		require.NoError(t, store.CreateChannel(channel))

		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "0xhomechannel123", result.ChannelID)
		assert.Equal(t, "0xuser123", result.UserWallet)
		assert.Equal(t, core.ChannelTypeHome, result.Type)
		assert.Equal(t, uint64(1), result.BlockchainID)
		assert.Equal(t, "0xtoken123", result.TokenAddress)
	})

	t.Run("No channel found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetChannelByID("0xnonexistent")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_GetActiveHomeChannel(t *testing.T) {
	t.Run("Success - Get active home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create home channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the home channel
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
		}
		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, homeChannelID, result.ChannelID)
		assert.Equal(t, core.ChannelTypeHome, result.Type)
		assert.Equal(t, core.ChannelStatusOpen, result.Status)
	})

	t.Run("No active home channel - user has no state", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetActiveHomeChannel("0xnonexistent", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - channel is closed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create closed channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the closed channel
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
		}
		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - channel is escrow type", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xescrowchannel123"

		// Create escrow channel (not home)
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeEscrow,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the escrow channel as home
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
		}
		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("No active home channel - state has no home channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create state without home channel
		state := core.State{
			ID:         "state1",
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
		require.NoError(t, store.StoreUserState(state))

		result, err := store.GetActiveHomeChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_CheckOpenChannel(t *testing.T) {
	t.Run("Success - Has open channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create open home channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the channel
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
		}
		require.NoError(t, store.StoreUserState(state))

		hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.True(t, hasOpenChannel)
	})

	t.Run("No open channel - user not found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		hasOpenChannel, err := store.CheckOpenChannel("0xnonexistent", "USDC")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
	})

	t.Run("No open channel - channel is closed", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create closed channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state referencing the closed channel
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
		}
		require.NoError(t, store.StoreUserState(state))

		hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "USDC")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
	})

	t.Run("No open channel - wrong asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		homeChannelID := "0xhomechannel123"

		// Create open home channel
		channel := core.Channel{
			ChannelID:         homeChannelID,
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Create state for USDC
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
		}
		require.NoError(t, store.StoreUserState(state))

		// Check for different asset
		hasOpenChannel, err := store.CheckOpenChannel("0xuser123", "ETH")
		require.NoError(t, err)
		assert.False(t, hasOpenChannel)
	})
}

func TestDBStore_UpdateChannel(t *testing.T) {
	t.Run("Success - Update channel status and version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update channel
		channel.Status = core.ChannelStatusClosed
		channel.StateVersion = 5

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, core.ChannelStatusClosed, result.Status)
		assert.Equal(t, uint64(5), result.StateVersion)
	})

	t.Run("Success - Update challenge expiration", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update with challenge expiration
		expiresAt := time.Now().Add(24 * time.Hour)
		channel.Status = core.ChannelStatusChallenged
		channel.ChallengeExpiresAt = &expiresAt

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, core.ChannelStatusChallenged, result.Status)
		assert.NotNil(t, result.ChallengeExpiresAt)
	})

	t.Run("Success - Update blockchain and token", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xhomechannel123",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusOpen,
			StateVersion:      0,
		}
		require.NoError(t, store.CreateChannel(channel))

		// Update blockchain and token
		channel.BlockchainID = 137
		channel.TokenAddress = "0xnewtoken456"

		err := store.UpdateChannel(channel)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetChannelByID("0xhomechannel123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, uint64(137), result.BlockchainID)
		assert.Equal(t, "0xnewtoken456", result.TokenAddress)
	})

	t.Run("Error - Update non-existent channel", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		channel := core.Channel{
			ChannelID:         "0xnonexistent",
			UserWallet:        "0xuser123",
			Type:              core.ChannelTypeHome,
			BlockchainID:      1,
			TokenAddress:      "0xtoken123",
			ChallengeDuration: 86400,
			Nonce:             1,
			Status:            core.ChannelStatusClosed,
			StateVersion:      1,
		}

		err := store.UpdateChannel(channel)
		require.NoError(t, err) // GORM doesn't return error for update with no rows affected
	})
}
