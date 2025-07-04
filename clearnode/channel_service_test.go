package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelService_ResizeChannel(t *testing.T) {
	t.Run("SuccessfulAllocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenResize", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with initial amount 1000
		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResize",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger with 1500 USDC (enough for resize)
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(1500)))

		// Verify initial balance
		initialBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), initialBalance)

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.Version+1, response.Version)

		// New channel amount should be initial + 200
		expected := initialRawAmount.Add(decimal.NewFromInt(200))
		assert.Equal(t, 0, response.Allocations[0].RawAmount.Cmp(expected.BigInt()), "Allocated amount mismatch")
		assert.Equal(t, 0, response.Allocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		var unchangedChannel Channel
		require.NoError(t, db.Where("channel_id = ?", ch.ChannelID).First(&unchangedChannel).Error)
		assert.Equal(t, initialRawAmount, unchangedChannel.RawAmount) // Should remain unchanged
		assert.Equal(t, ch.Version, unchangedChannel.Version)         // Should remain unchanged
		assert.Equal(t, ChannelStatusOpen, unchangedChannel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), finalBalance) // Should remain unchanged
	})

	t.Run("SuccessfulDeallocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenResize2", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		initialRawAmount := decimal.NewFromInt(1000)
		ch := Channel{
			ChannelID:   "0xChanResize2",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromInt(500)))

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(-300),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.NoError(t, err)

		// Channel amount should decrease
		expected := initialRawAmount.Sub(decimal.NewFromInt(300))
		assert.Equal(t, 0, response.Allocations[0].RawAmount.Cmp(expected.BigInt()), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenClosed", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanClosed",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusClosed,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xChanClosed is not open: closed")
	})

	t.Run("ErrorChannelJoining", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenJoining", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanJoining",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusJoining,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xChanJoining is not open: joining")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xToken", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChan",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		asset := Asset{Token: "0xTokenInsufficient", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanInsufficient",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund with very small amount (0.000001 USDC), but try to allocate 200 raw units
		// This will create insufficient balance when converted to raw units
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, "usdc", decimal.NewFromFloat(0.000001)))

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(200),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient unified balance")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		asset := Asset{Token: "0xTokenSig", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		ch := Channel{
			ChannelID:   "0xChanSig",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(1000),
			Version:     1,
		}
		require.NoError(t, db.Create(&ch).Error)

		service := NewChannelService(db, &signer)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   big.NewInt(100),
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{} // Empty signers

		_, err = service.RequestResize(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}

func TestChannelService_CloseChannel(t *testing.T) {
	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialRawAmount := decimal.NewFromInt(500)
		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			"usdc",
			rawToDecimal(initialRawAmount.BigInt(), asset.Decimals),
		))

		service := NewChannelService(db, &signer)
		params := &CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestClose(LoggerFromContext(context.Background()), params, rpcSigners)
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.Version+1, response.Version)

		// Final allocation should send full balance to destination
		assert.Equal(t, 0, response.FinalAllocations[0].RawAmount.Cmp(initialRawAmount.BigInt()), "Primary allocation mismatch")
		assert.Equal(t, 0, response.FinalAllocations[1].RawAmount.Cmp(big.NewInt(0)), "Broker allocation should be zero")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()
		userAccountID := NewAccountID(userAddress.Hex())

		// Create asset
		asset := Asset{Token: "0xTokenClose", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create channel with amount 500
		initialRawAmount := decimal.NewFromInt(500)

		// Seed other challenged channel
		require.NoError(t, db.Create(&Channel{
			ChannelID:   "0xChanChallenged",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusChallenged,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}).Error)

		ch := Channel{
			ChannelID:   "0xChanClose",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   initialRawAmount,
			Version:     2,
		}
		require.NoError(t, db.Create(&ch).Error)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			"usdc",
			rawToDecimal(initialRawAmount.BigInt(), asset.Decimals),
		))

		service := NewChannelService(db, &signer)
		params := &CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		_, err = service.RequestClose(LoggerFromContext(context.Background()), params, rpcSigners)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}
