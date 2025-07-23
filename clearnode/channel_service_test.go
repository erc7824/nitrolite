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
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, nil, &signer)
		allocateAmount := decimal.NewFromInt(200)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.Version+1, response.Version)

		// New channel amount should be initial + 200
		expected := initialRawAmount.Add(decimal.NewFromInt(200))
		assert.Equal(t, 0, response.Allocations[0].RawAmount.Cmp(expected), "Allocated amount mismatch")
		assert.Equal(t, 0, response.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")

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
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, nil, &signer)
		allocateAmount := decimal.NewFromInt(-300)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Channel amount should decrease
		expected := initialRawAmount.Sub(decimal.NewFromInt(300))
		assert.Equal(t, 0, response.Allocations[0].RawAmount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, "usdc")
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("ErrorInvalidChannelID", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		allocateAmount := decimal.NewFromInt(100)
		params := &ResizeChannelParams{
			ChannelID:        "0xNonExistentChannel",
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xNonExistentChannel not found")
	})

	t.Run("ErrorChannelClosed", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		allocateAmount := decimal.NewFromInt(100)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel 0xChanClosed is not open: closed")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		allocateAmount := decimal.NewFromInt(100)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		allocateAmount := decimal.NewFromInt(200)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{userAddress.Hex(): {}}

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient unified balance")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		allocateAmount := decimal.NewFromInt(100)
		params := &ResizeChannelParams{
			ChannelID:        ch.ChannelID,
			AllocateAmount:   &allocateAmount,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{} // Empty signers

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}

func TestChannelService_CloseChannel(t *testing.T) {
	t.Run("SuccessfulCloseChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		params := &CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestClose(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.Version+1, response.Version)

		// Final allocation should send full balance to destination
		assert.Equal(t, 0, response.FinalAllocations[0].RawAmount.Cmp(initialRawAmount), "Primary allocation mismatch")
		assert.Equal(t, 0, response.FinalAllocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")
	})

	t.Run("ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

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

		service := NewChannelService(db, map[string]*NetworkConfig{}, &signer)
		params := &CloseChannelParams{
			ChannelID:        ch.ChannelID,
			FundsDestination: userAddress.Hex(),
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		_, err = service.RequestClose(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has challenged channels")
	})
}

func TestChannelService_RequestCreate(t *testing.T) {
	t.Run("SuccessfulCreateChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Set up network configurations
		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000) // 1 USDC in raw units
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Validate response structure
		assert.NotEmpty(t, response.ChannelID, "Channel ID should not be empty")
		assert.NotEmpty(t, response.StateHash, "State hash should not be empty")
		assert.NotNil(t, response.State, "State should not be nil")

		// Verify state structure
		assert.Equal(t, uint8(1), response.State.Intent, "Intent should be INITIALIZE (1)")
		assert.Equal(t, uint64(0), response.State.Version, "Version should be 0")
		assert.Len(t, response.State.Allocations, 2, "Should have 2 allocations")
		assert.Len(t, response.State.Sigs, 1, "Should have 1 signature")

		// Verify allocations
		assert.Equal(t, userAddress.Hex(), response.State.Allocations[0].Participant, "First allocation should be for user")
		assert.Equal(t, asset.Token, response.State.Allocations[0].TokenAddress, "Token address should match")
		assert.Equal(t, amount, response.State.Allocations[0].RawAmount, "Amount should match")

		assert.Equal(t, signer.GetAddress().Hex(), response.State.Allocations[1].Participant, "Second allocation should be for broker")
		assert.Equal(t, asset.Token, response.State.Allocations[1].TokenAddress, "Token address should match")
		assert.True(t, response.State.Allocations[1].RawAmount.IsZero(), "Broker allocation should be zero")
	})

	t.Run("SuccessfulCreateChannelWithZeroAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Set up network configurations
		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.Zero // Zero amount channel
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Verify zero amount is handled correctly
		assert.True(t, response.State.Allocations[0].RawAmount.IsZero(), "User allocation should be zero")
		assert.True(t, response.State.Allocations[1].RawAmount.IsZero(), "Broker allocation should be zero")
	})

	t.Run("ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000)
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{} // Empty signers map

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("ErrorExistingOpenChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Create existing open channel
		existingChannel := Channel{
			ChannelID:   "0xExistingChannel",
			Participant: userAddress.Hex(),
			Wallet:      userAddress.Hex(),
			Status:      ChannelStatusOpen,
			Token:       asset.Token,
			ChainID:     137,
			RawAmount:   decimal.NewFromInt(500),
			Version:     1,
		}
		require.NoError(t, db.Create(&existingChannel).Error)

		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000)
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "an open channel with broker already exists")
	})

	t.Run("ErrorUnsupportedToken", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Don't create any assets in the database

		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000)
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   "0xUnsupportedToken1234567890123456789012",
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token not supported")
	})

	t.Run("ErrorUnsupportedChainID", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset for unsupported chain ID to pass asset check first
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 999, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		// Only provide network config for chain ID 137, not 999
		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000)
		params := &CreateChannelParams{
			ChainID: 999, // Unsupported chain ID
			Token:   asset.Token,
			Amount:  &amount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported chain ID")
	})

	t.Run("ErrorLargeAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		// Test with very large amount that should work
		largeAmount := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil), 0) // 10^30
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &largeAmount,
		}
		rpcSigners := map[string]struct{}{
			userAddress.Hex(): {},
		}

		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Verify large amount is handled correctly
		assert.Equal(t, largeAmount, response.State.Allocations[0].RawAmount, "Large amount should be preserved")
	})

	t.Run("ErrorDifferentUserSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		rawKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		signer := Signer{privateKey: rawKey}
		userAddress := signer.GetAddress()

		// Create a different user
		differentKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		differentSigner := Signer{privateKey: differentKey}
		differentAddress := differentSigner.GetAddress()

		// Create asset
		asset := Asset{Token: "0x1234567890123456789012345678901234567890", ChainID: 137, Symbol: "usdc", Decimals: 6}
		require.NoError(t, db.Create(&asset).Error)

		networks := map[string]*NetworkConfig{
			"137": {
				Name:               "polygon",
				ChainID:            137,
				InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
				CustodyAddress:     "0xCustodyAddress",
				AdjudicatorAddress: "0xAdjudicatorAddress",
			},
		}

		service := NewChannelService(db, networks, &signer)
		amount := decimal.NewFromInt(1000000)
		params := &CreateChannelParams{
			ChainID: 137,
			Token:   asset.Token,
			Amount:  &amount,
		}
		// Use different user's signature but pass userAddress as wallet
		rpcSigners := map[string]struct{}{
			differentAddress.Hex(): {},
		}

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}
