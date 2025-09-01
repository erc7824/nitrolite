package main

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedAsset(t *testing.T, db *gorm.DB, token string, chainID uint32, symbol string, decimals uint8) Asset {
	asset := Asset{Token: token, ChainID: chainID, Symbol: symbol, Decimals: decimals}
	require.NoError(t, db.Create(&asset).Error)
	return asset
}

func seedChannel(t *testing.T, db *gorm.DB, channelID, participant, wallet, token string, chainID uint32, rawAmount decimal.Decimal, version uint64, status ChannelStatus) Channel {
	ch := Channel{
		ChannelID:   channelID,
		Participant: participant,
		Wallet:      wallet,
		Status:      status,
		Token:       token,
		ChainID:     chainID,
		RawAmount:   rawAmount,
		State: UnsignedState{
			Version: version,
		},
	}
	require.NoError(t, db.Create(&ch).Error)
	return ch
}

func getCreateChannelParams(chainID uint32, token string, amount decimal.Decimal) *CreateChannelParams {
	return &CreateChannelParams{
		ChainID: chainID,
		Token:   token,
		Amount:  &amount,
	}
}

func getResizeChannelParams(channelID string, allocateAmount *decimal.Decimal, resizeAmount *decimal.Decimal, destination string) *ResizeChannelParams {
	return &ResizeChannelParams{
		ChannelID:        channelID,
		AllocateAmount:   allocateAmount,
		ResizeAmount:     resizeAmount,
		FundsDestination: destination,
	}
}

func getCloseChannelParams(channelID, destination string) *CloseChannelParams {
	return &CloseChannelParams{
		ChannelID:        channelID,
		FundsDestination: destination,
	}
}

func TestChannelService(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := Signer{privateKey: key}
	userAddress := signer.GetAddress()
	userAccountID := NewAccountID(userAddress.Hex())

	rpcSigners := map[string]struct{}{
		userAddress.Hex(): {},
	}

	tokenAddress := "0x1234567890123456789012345678901234567890"
	tokenSymbol := "usdc"
	channelID := "0xDefaultChannelID"
	channelAmountRaw := decimal.NewFromInt(1000)
	chainID := uint32(137)

	networks := map[uint32]*NetworkConfig{
		137: {
			Name:               "polygon",
			ChainID:            chainID,
			InfuraURL:          "https://polygon-mainnet.infura.io/v3/test",
			CustodyAddress:     "0x2e189bd6f6FD3EB59fd97FcA03251d93Af4E522a",
			AdjudicatorAddress: "0xdadB0d80178819F2319190D340ce9A924f783711",
		},
	}

	t.Run("RequestResize_Success", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)

		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusOpen)

		// Fund participant ledger with 1500 USDC (enough for resize)
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, tokenSymbol, decimal.NewFromInt(1500)))

		initialBalance, err := ledger.Balance(userAccountID, tokenSymbol)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), initialBalance)

		service := NewChannelService(db, nil, &signer)

		allocateAmount := decimal.NewFromInt(200)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		response, err := service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Validate response
		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.State.Version+1, response.State.Version)

		// New channel amount should be initial + 200
		expected := channelAmountRaw.Add(decimal.NewFromInt(200))
		assert.Equal(t, 0, response.State.Allocations[0].RawAmount.Cmp(expected), "Allocated amount mismatch")
		assert.Equal(t, 0, response.State.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")

		// Verify channel state in database remains unchanged (no update until blockchain confirmation)
		channel, err := GetChannelByID(db, ch.ChannelID)
		require.NoError(t, err)
		require.NotNil(t, channel)
		assert.Equal(t, channelAmountRaw, channel.RawAmount)
		assert.Equal(t, ch.State.Version, channel.State.Version)
		assert.Equal(t, ChannelStatusOpen, channel.Status)

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, tokenSymbol)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(1500), finalBalance) // Should remain unchanged
	})

	t.Run("RequestResize_SuccessfulDeallocation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusOpen)

		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, tokenSymbol, decimal.NewFromInt(500)))

		service := NewChannelService(db, nil, &signer)

		allocateAmount := decimal.NewFromInt(-300)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		response, err := service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Channel amount should decrease
		expected := channelAmountRaw.Sub(decimal.NewFromInt(300))
		assert.Equal(t, 0, response.State.Allocations[0].RawAmount.Cmp(expected), "Decreased amount mismatch")

		// Verify ledger balance remains unchanged (no update until blockchain confirmation)
		finalBalance, err := ledger.Balance(userAccountID, tokenSymbol)
		require.NoError(t, err)
		assert.Equal(t, decimal.NewFromInt(500), finalBalance) // Should remain unchanged
	})

	t.Run("RequestResize_ErrorInvalidChannelID", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		allocateAmount := decimal.NewFromInt(100)
		params := getResizeChannelParams("0xNonExistentChannel", &allocateAmount, nil, userAddress.Hex())
		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "channel 0xNonExistentChannel not found")
	})

	t.Run("RequestResize_ErrorChannelClosed", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusClosed)
		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		allocateAmount := decimal.NewFromInt(100)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "channel 0xDefaultChannelID is not open: closed")
	})

	t.Run("RequestResize_ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusChallenged)
		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		allocateAmount := decimal.NewFromInt(100)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("RequestResize_ErrorInsufficientFunds", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusOpen)

		// Fund with very small amount (0.000001 USDC), but try to allocate 200 raw units
		// This will create insufficient balance when converted to raw units
		ledger := GetWalletLedger(db, userAddress)
		require.NoError(t, ledger.Record(userAccountID, tokenSymbol, decimal.NewFromFloat(0.000001)))

		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		allocateAmount := decimal.NewFromInt(200)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "insufficient unified balance")
	})

	t.Run("RequestResize_ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusOpen)
		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		allocateAmount := decimal.NewFromInt(100)
		params := getResizeChannelParams(ch.ChannelID, &allocateAmount, nil, userAddress.Hex())
		rpcSigners := map[string]struct{}{} // Empty signers map

		_, err = service.RequestResize(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("RequestClose_Success", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 2, ChannelStatusOpen)

		// Fund participant ledger so that raw units match channel.Amount
		require.NoError(t, GetWalletLedger(db, userAddress).Record(
			userAccountID,
			tokenSymbol,
			rawToDecimal(channelAmountRaw.BigInt(), asset.Decimals),
		))

		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		params := getCloseChannelParams(ch.ChannelID, userAddress.Hex())
		response, err := service.RequestClose(params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		assert.Equal(t, ch.ChannelID, response.ChannelID)
		assert.Equal(t, ch.State.Version+1, response.State.Version)

		// Final allocation should send full balance to destination
		assert.Equal(t, 0, response.State.Allocations[0].RawAmount.Cmp(channelAmountRaw), "Primary allocation mismatch")
		assert.Equal(t, 0, response.State.Allocations[1].RawAmount.Cmp(decimal.Zero), "Broker allocation should be zero")
	})

	t.Run("RequestClose_ErrorOtherChallengedChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		ch := seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 2, ChannelStatusChallenged)
		service := NewChannelService(db, map[uint32]*NetworkConfig{}, &signer)

		params := getCloseChannelParams(ch.ChannelID, userAddress.Hex())
		_, err = service.RequestClose(params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "has challenged channels")
	})

	t.Run("RequestCreate_Success", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		amount := decimal.NewFromInt(1000000) // 1 USDC in raw units
		params := getCreateChannelParams(chainID, asset.Token, amount)
		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		// Validate response structure
		assert.NotEmpty(t, response.ChannelID, "Channel ID should not be empty")
		assert.NotNil(t, response.State, "State should not be nil")

		// Verify state structure
		assert.Equal(t, StateIntent(StateIntentInitialize), response.State.Intent, "Intent should be INITIALIZE (1)")
		assert.Equal(t, uint64(0), response.State.Version, "Version should be 0")
		assert.Len(t, response.State.Allocations, 2, "Should have 2 allocations")
		assert.NotEmpty(t, response.StateSignature, "Should have 1 signature")

		// Verify allocations
		assert.Equal(t, userAddress.Hex(), response.State.Allocations[0].Participant, "First allocation should be for user")
		assert.Equal(t, asset.Token, response.State.Allocations[0].TokenAddress, "Token address should match")
		assert.Equal(t, amount, response.State.Allocations[0].RawAmount, "Amount should match")

		assert.Equal(t, signer.GetAddress().Hex(), response.State.Allocations[1].Participant, "Second allocation should be for broker")
		assert.Equal(t, asset.Token, response.State.Allocations[1].TokenAddress, "Token address should match")
		assert.True(t, response.State.Allocations[1].RawAmount.IsZero(), "Broker allocation should be zero")
		assert.Equal(t, 2, len(response.Channel.Participants), "Expected 2 participants")
		assert.Equal(t, networks[chainID].AdjudicatorAddress, response.Channel.Adjudicator, "Adjudicator address should match")
		assert.Equal(t, uint64(3600), response.Channel.Challenge, "Challenge should match")
		assert.NotEqual(t, uint64(0), response.Channel.Nonce, "Nonce should not be 0")
	})

	t.Run("RequestCreate_ZeroAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		amount := decimal.Zero // Zero amount channel
		params := getCreateChannelParams(chainID, asset.Token, amount)
		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		assert.True(t, response.State.Allocations[0].RawAmount.IsZero(), "User allocation should be zero")
		assert.True(t, response.State.Allocations[1].RawAmount.IsZero(), "Broker allocation should be zero")
	})

	t.Run("RequestCreate_ErrorInvalidSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		amount := decimal.NewFromInt(1000000)
		params := getCreateChannelParams(chainID, asset.Token, amount)
		rpcSigners := map[string]struct{}{} // Empty signers map

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("RequestCreate_ErrorExistingOpenChannel", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		_ = seedChannel(t, db, channelID, userAddress.Hex(), userAddress.Hex(), asset.Token, chainID, channelAmountRaw, 1, ChannelStatusOpen)
		service := NewChannelService(db, networks, &signer)

		amount := decimal.NewFromInt(1000000)
		params := getCreateChannelParams(chainID, asset.Token, amount)
		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "an open channel with broker already exists")
	})

	t.Run("RequestCreate_ErrorUnsupportedToken", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		// Don't create any assets in the database
		service := NewChannelService(db, networks, &signer)

		amount := decimal.NewFromInt(1000000)
		params := getCreateChannelParams(chainID, "0xUnsupportedToken1234567890123456789012", amount)

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "token not supported")
	})

	t.Run("RequestCreate_ErrorUnsupportedChainID", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		// Create asset for unsupported chain ID to pass asset check first
		asset := seedAsset(t, db, tokenAddress, 999, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		amount := decimal.NewFromInt(1000000)
		params := getCreateChannelParams(999, asset.Token, amount)
		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "unsupported chain ID")
	})

	t.Run("RequestCreate_LargeAmount", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		largeAmount := decimal.NewFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(30), nil), 0) // 10^30
		params := getCreateChannelParams(chainID, asset.Token, largeAmount)
		response, err := service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.NoError(t, err)

		assert.Equal(t, largeAmount, response.State.Allocations[0].RawAmount, "Large amount should be preserved")
	})

	t.Run("RequestCreate_ErrorDifferentUserSignature", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		t.Cleanup(cleanup)

		asset := seedAsset(t, db, tokenAddress, chainID, tokenSymbol, 6)
		service := NewChannelService(db, networks, &signer)

		// Create a different user
		differentKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		differentSigner := Signer{privateKey: differentKey}
		differentAddress := differentSigner.GetAddress()

		amount := decimal.NewFromInt(1000000)
		params := getCreateChannelParams(chainID, asset.Token, amount)

		// Use different user's signature but pass userAddress as wallet
		rpcSigners := map[string]struct{}{
			differentAddress.Hex(): {},
		}

		_, err = service.RequestCreate(userAddress, params, rpcSigners, LoggerFromContext(context.Background()))
		require.Error(t, err)

		assert.Contains(t, err.Error(), "invalid signature")
	})
}
