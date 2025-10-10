package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCheckpoint(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		state := UnsignedState{
			Intent:  StateIntent(1),
			Version: 5,
			Data:    "test-data",
			Allocations: []Allocation{{
				Participant:  "0xUser123",
				TokenAddress: "0xToken456",
				RawAmount:    decimal.NewFromInt(1000),
			}},
		}
		userSig := Signature{1, 2, 3}
		serverSig := Signature{4, 5, 6}
		channelId := common.HexToHash("0xchannel1")

		err := CreateCheckpoint(db, channelId, 1, state, userSig, serverSig)
		require.NoError(t, err)

		var action BlockchainAction
		err = db.Where("channel_id = ?", channelId).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeCheckpoint, action.Type)
		assert.Equal(t, channelId, action.ChannelID)
		assert.Equal(t, uint32(1), action.ChainID)
		assert.Equal(t, StatusPending, action.Status)
		assert.Equal(t, 0, action.Retries)
		assert.Empty(t, action.Error)
		assert.Empty(t, action.TxHash)
		assert.False(t, action.CreatedAt.IsZero())
		assert.False(t, action.UpdatedAt.IsZero())

		var data CheckpointData
		err = json.Unmarshal([]byte(action.Data), &data)
		require.NoError(t, err)
		assert.Equal(t, state, data.State)
		assert.Equal(t, userSig, data.UserSig)
		assert.Equal(t, serverSig, data.ServerSig)
	})

	t.Run("Database error", func(t *testing.T) {
		channelId := common.HexToHash("0xchannel1")

		db, cleanup := setupTestDB(t)
		defer cleanup()

		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.Close()

		err = CreateCheckpoint(db, channelId, 1, UnsignedState{}, Signature{}, Signature{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database is closed")
	})
}

func TestBlockchainAction_Fail(t *testing.T) {
	channelId := common.HexToHash("0xchannel1")

	db, cleanup := setupTestDB(t)
	defer cleanup()

	action := &BlockchainAction{
		Type:      ActionTypeCheckpoint,
		ChannelID: channelId,
		ChainID:   1,
		Data:      []byte{1},
		Status:    StatusPending,
		Retries:   2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(action).Error)

	err := action.Fail(db, "test error")
	require.NoError(t, err)

	assert.Equal(t, StatusFailed, action.Status)
	assert.Equal(t, "test error", action.Error)
	assert.Equal(t, 3, action.Retries)

	var dbAction BlockchainAction
	err = db.First(&dbAction, action.ID).Error
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, dbAction.Status)
	assert.Equal(t, "test error", dbAction.Error)
	assert.Equal(t, 3, dbAction.Retries)
}

func TestBlockchainAction_Complete(t *testing.T) {
	channelId := common.HexToHash("0xchannel1")

	db, cleanup := setupTestDB(t)
	defer cleanup()

	action := &BlockchainAction{
		Type:      ActionTypeCheckpoint,
		ChannelID: channelId,
		ChainID:   1,
		Data:      []byte{1},
		Status:    StatusPending,
		Error:     "previous error",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(action).Error)

	txHash := common.HexToHash("0xabcdef1234567890")
	err := action.Complete(db, txHash)
	require.NoError(t, err)

	assert.Equal(t, StatusCompleted, action.Status)
	assert.Equal(t, txHash, action.TxHash)
	assert.Empty(t, action.Error)

	var dbAction BlockchainAction
	err = db.First(&dbAction, action.ID).Error
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, dbAction.Status)
	assert.Equal(t, txHash, dbAction.TxHash)
	assert.Empty(t, dbAction.Error)
}

func TestBlockchainAction_TableName(t *testing.T) {
	action := BlockchainAction{}
	assert.Equal(t, "blockchain_actions", action.TableName())
}

func TestCheckpointData_Serialization(t *testing.T) {
	original := CheckpointData{
		State: UnsignedState{
			Intent:  StateIntent(2),
			Version: 10,
			Data:    "test-data",
			Allocations: []Allocation{{
				Participant:  "0xUser1",
				TokenAddress: "0xToken1",
				RawAmount:    decimal.NewFromInt(5000),
			}},
		},
		UserSig:   Signature{0x11, 0x22, 0x33},
		ServerSig: Signature{0x44, 0x55, 0x66},
	}

	bytes, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshaled CheckpointData
	err = json.Unmarshal(bytes, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, original, unmarshaled)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, BlockchainActionType("checkpoint"), ActionTypeCheckpoint)
	assert.Equal(t, BlockchainActionStatus("pending"), StatusPending)
	assert.Equal(t, BlockchainActionStatus("completed"), StatusCompleted)
	assert.Equal(t, BlockchainActionStatus("failed"), StatusFailed)
}
