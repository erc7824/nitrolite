package main

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var tokenAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"

func setupMockCustody(t *testing.T) (*Custody, *gorm.DB, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := &Signer{privateKey: rawKey}

	logger := NewLoggerIPFS("custody_test")

	sendBalanceUpdate := func(wallet string) {}
	sendChannelUpdate := func(channel Channel) {}

	if custodyAbi == nil {
		var err error
		custodyAbi, err = nitrolite.CustodyMetaData.GetAbi()
		require.NoError(t, err)
	}

	balance := new(big.Int)
	balance.SetString("10000000000000000000", 10) // 10 eth in wei

	address := signer.GetAddress()
	genesisAlloc := map[common.Address]core.GenesisAccount{
		address: {
			Balance: balance,
		},
	}

	backend := simulated.NewBackend(genesisAlloc)
	client := backend.Client()

	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(signer.GetPrivateKey(), chainID)
	require.NoError(t, err)
	auth.GasPrice = big.NewInt(30000000000)
	auth.GasLimit = uint64(3000000)

	assets := []Asset{
		{Token: tokenAddress, ChainID: uint32(chainID.Int64()), Symbol: "usdc", Decimals: 6},
	}
	for _, asset := range assets {
		require.NoError(t, db.Create(&asset).Error)
	}

	contract, err := nitrolite.NewCustody(common.Address{}, client)
	require.NoError(t, err)

	custody := &Custody{
		db:                 db,
		signer:             signer,
		transactOpts:       auth,
		client:             client,
		custody:            contract,
		chainID:            uint32(chainID.Int64()),
		adjudicatorAddress: common.HexToAddress("0xAdjudicatorAddress"),
		sendBalanceUpdate:  sendBalanceUpdate,
		sendChannelUpdate:  sendChannelUpdate,
		logger:             logger,
	}

	return custody, db, cleanup
}

func createMockLog(eventID common.Hash) types.Log {
	return types.Log{
		Address:     common.HexToAddress("0xCustodyContractAddress"),
		Topics:      []common.Hash{eventID},
		Data:        []byte{},
		TxHash:      common.HexToHash("0xTransactionHash"),
		BlockNumber: 12345678,
		Index:       0,
	}
}

func createMockCreatedEvent(t *testing.T, signer *Signer, token string, amount int64) (*types.Log, *nitrolite.CustodyCreated) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}
	walletAddr := common.HexToAddress("0xWallet123")

	channel := nitrolite.Channel{
		Participants: []common.Address{common.HexToAddress("0xParticipant1"), signer.GetAddress()},
		Adjudicator:  common.HexToAddress("0xAdjudicatorAddress"),
		Challenge:    3600,
		Nonce:        12345,
	}

	allocation := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress("0xParticipant1"),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(amount),
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(0),
		},
	}

	initialState := nitrolite.State{
		Intent:      0,
		Version:     big.NewInt(0),
		Data:        []byte{},
		Allocations: allocation,
		Sigs:        []nitrolite.Signature{},
	}

	event := &nitrolite.CustodyCreated{
		ChannelId: channelID,
		Wallet:    walletAddr,
		Channel:   channel,
		Initial:   initialState,
	}

	log := createMockLog(custodyAbi.Events["Created"].ID)

	return &log, event
}

func createMockJoinedEvent(t *testing.T) (*types.Log, *nitrolite.CustodyJoined) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	event := &nitrolite.CustodyJoined{
		ChannelId: channelID,
		Index:     big.NewInt(1),
	}

	log := createMockLog(custodyAbi.Events["Joined"].ID)

	return &log, event
}

func createMockClosedEvent(t *testing.T, signer *Signer, token string, amount int64) (*types.Log, *nitrolite.CustodyClosed) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	allocation := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress("0xParticipant1"),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(amount),
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(0),
		},
	}

	finalState := nitrolite.State{
		Intent:      2,
		Version:     big.NewInt(1),
		Data:        []byte{},
		Allocations: allocation,
		Sigs:        []nitrolite.Signature{},
	}

	event := &nitrolite.CustodyClosed{
		ChannelId:  channelID,
		FinalState: finalState,
	}

	log := createMockLog(custodyAbi.Events["Closed"].ID)

	return &log, event
}

func createMockChallengedEvent(t *testing.T, signer *Signer, token string, amount int64) (*types.Log, *nitrolite.CustodyChallenged) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	allocation := []nitrolite.Allocation{
		{
			Destination: common.HexToAddress("0xParticipant1"),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(amount),
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(0),
		},
	}

	state := nitrolite.State{
		Intent:      1,
		Version:     big.NewInt(2),
		Data:        []byte{},
		Allocations: allocation,
		Sigs:        []nitrolite.Signature{},
	}

	event := &nitrolite.CustodyChallenged{
		ChannelId:  channelID,
		State:      state,
		Expiration: big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
	}

	log := createMockLog(custodyAbi.Events["Challenged"].ID)

	return &log, event
}

func createMockResizedEvent(t *testing.T, amount int64) (*types.Log, *nitrolite.CustodyResized) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	deltaAllocations := []*big.Int{
		big.NewInt(amount),
		big.NewInt(0),
	}

	event := &nitrolite.CustodyResized{
		ChannelId:        channelID,
		DeltaAllocations: deltaAllocations,
	}

	log := createMockLog(custodyAbi.Events["Resized"].ID)

	return &log, event
}

func TestHandleCreatedEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		amount := int64(1000000)

		channelIDBytes := [32]byte{1, 2, 3, 4}
		walletAddr := common.HexToAddress("0xWallet123")
		channelStruct := nitrolite.Channel{
			Participants: []common.Address{common.HexToAddress("0xParticipant1"), custody.signer.GetAddress()},
			Adjudicator:  common.HexToAddress("0xAdjudicatorAddress"),
			Challenge:    3600,
			Nonce:        12345,
		}

		allocation := []nitrolite.Allocation{
			{
				Destination: common.HexToAddress("0xParticipant1"),
				Token:       common.HexToAddress(tokenAddress),
				Amount:      big.NewInt(amount),
			},
			{
				Destination: custody.signer.GetAddress(),
				Token:       common.HexToAddress(tokenAddress),
				Amount:      big.NewInt(0),
			},
		}

		initialState := nitrolite.State{
			Intent:      0,
			Version:     big.NewInt(0),
			Data:        []byte{},
			Allocations: allocation,
			Sigs:        []nitrolite.Signature{},
		}

		mockEvent := &nitrolite.CustodyCreated{
			ChannelId: channelIDBytes,
			Wallet:    walletAddr,
			Channel:   channelStruct,
			Initial:   initialState,
		}

		var capturedChannel Channel
		custody.sendChannelUpdate = func(ch Channel) {
			capturedChannel = ch
		}

		logger := custody.logger
		custody.handleCreated(logger, mockEvent)

		channelIDStr := common.Hash(mockEvent.ChannelId).Hex()
		var dbChannel Channel
		dbErr := db.Where("channel_id = ?", channelIDStr).First(&dbChannel).Error
		require.NoError(t, dbErr)

		assert.Equal(t, channelIDStr, dbChannel.ChannelID)
		assert.Equal(t, mockEvent.Wallet.Hex(), dbChannel.Wallet)
		assert.Equal(t, mockEvent.Channel.Participants[0].Hex(), dbChannel.Participant)
		assert.Equal(t, mockEvent.Channel.Nonce, dbChannel.Nonce)
		assert.Equal(t, mockEvent.Channel.Challenge, dbChannel.Challenge)
		assert.Equal(t, mockEvent.Channel.Adjudicator.Hex(), dbChannel.Adjudicator)
		assert.Equal(t, uint64(amount), dbChannel.Amount)
		assert.Equal(t, tokenAddress, dbChannel.Token)
		assert.Equal(t, ChannelStatusJoining, dbChannel.Status)

		var signerWallet SignerWallet
		signerErr := db.Where("wallet = ? AND signer = ?", mockEvent.Wallet.Hex(), mockEvent.Channel.Participants[0].Hex()).First(&signerWallet).Error
		require.NoError(t, signerErr)

		var entries []Entry
		entriesErr := db.Where("wallet = ?", mockEvent.Wallet.Hex()).Find(&entries).Error
		require.NoError(t, entriesErr)
		assert.NotEmpty(t, entries)

		assert.Equal(t, channelIDStr, capturedChannel.ChannelID)
		assert.Equal(t, mockEvent.Wallet.Hex(), capturedChannel.Wallet)
		assert.Equal(t, uint32(custody.chainID), dbChannel.ChainID)
		assert.False(t, dbChannel.CreatedAt.IsZero())
		assert.False(t, dbChannel.UpdatedAt.IsZero())

		assert.WithinDuration(t, time.Now(), dbChannel.CreatedAt, 2*time.Second)
		assert.WithinDuration(t, time.Now(), dbChannel.UpdatedAt, 2*time.Second)

		walletLedger := GetWalletLedger(db, mockEvent.Wallet.Hex())
		balance, err := walletLedger.Balance(channelIDStr, "usdc")
		require.NoError(t, err)
		expected := decimal.NewFromInt(amount).Div(decimal.NewFromInt(1000000)) // Adjusted for 6 decimals
		assert.True(t, expected.Equal(balance), "Expected channel balance to be %s, got %s", expected, balance)
	})
}

func TestHandleJoinedEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		amount := uint64(1000000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusJoining,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      amount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, err := GetAssetByToken(db, tokenAddress, custody.chainID)
		require.NoError(t, err)

		ledger := GetWalletLedger(db, walletAddr)
		tokenAmountDecimal := decimal.NewFromInt(int64(amount)).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))
		err = ledger.Record(channelID, asset.Symbol, tokenAmountDecimal)
		require.NoError(t, err)

		_, mockEvent := createMockJoinedEvent(t)

		var balanceUpdateCalled bool
		var capturedWallet string
		custody.sendBalanceUpdate = func(wallet string) {
			balanceUpdateCalled = true
			capturedWallet = wallet
		}

		var channelUpdateCalled bool
		var capturedChannel Channel
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
			capturedChannel = ch
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Joined")
		custody.handleJoined(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status)
		assert.Equal(t, initialChannel.Amount, updatedChannel.Amount, "Amount should not change")
		assert.Equal(t, initialChannel.Nonce, updatedChannel.Nonce, "Nonce should not change")
		assert.Equal(t, initialChannel.Challenge, updatedChannel.Challenge, "Challenge should not change")
		assert.Equal(t, initialChannel.ChainID, updatedChannel.ChainID, "ChainID should not change")
		assert.Equal(t, initialChannel.Token, updatedChannel.Token, "Token should not change")

		var entries []Entry
		err = db.Where("wallet = ?", walletAddr).Find(&entries).Error
		require.NoError(t, err)
		assert.NotEmpty(t, entries)

		assert.True(t, balanceUpdateCalled, "Balance update callback should be called")
		assert.Equal(t, walletAddr, capturedWallet)

		assert.True(t, channelUpdateCalled, "Channel update callback should be called")
		assert.Equal(t, channelID, capturedChannel.ChannelID)
		assert.Equal(t, ChannelStatusOpen, capturedChannel.Status)

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix())
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt))
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		channelBalance, err := ledger.Balance(channelID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after joined event")

		walletBalance, err := ledger.Balance(walletAddr, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, tokenAmountDecimal.Equal(walletBalance),
			"Wallet balance should be %s, got %s", tokenAmountDecimal, walletBalance)
	})

	t.Run("Channel Not Found", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		// Create a different channel ID than the one in the event
		initialChannel := Channel{
			ChannelID: "0xDifferentChannelId",
			Wallet:    "0xWallet123",
			Status:    ChannelStatusJoining,
			Token:     tokenAddress,
			ChainID:   custody.chainID,
			Amount:    1000000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockJoinedEvent(t)

		var balanceUpdateCalled bool
		custody.sendBalanceUpdate = func(wallet string) {
			balanceUpdateCalled = true
		}

		var channelUpdateCalled bool
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
		}

		logger := custody.logger.With("event", "Joined")
		custody.handleJoined(logger, mockEvent)

		// Event should be ignored, and no callbacks should be called
		assert.False(t, balanceUpdateCalled, "Balance update should not be called for non-existent channel")
		assert.False(t, channelUpdateCalled, "Channel update should not be called for non-existent channel")

		// Initial channel should remain unmodified
		var checkChannel Channel
		err = db.Where("channel_id = ?", initialChannel.ChannelID).First(&checkChannel).Error
		require.NoError(t, err)
		assert.Equal(t, ChannelStatusJoining, checkChannel.Status, "Status of other channel should not change")
	})
}

func TestHandleClosedEvent(t *testing.T) {
	t.Run("Success Smaller Final Amount", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialAmount := uint64(1000000)
		finalAmount := int64(500000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      initialAmount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, err := GetAssetByToken(db, tokenAddress, custody.chainID)
		require.NoError(t, err)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromInt(int64(initialAmount)).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))

		err = ledger.Record(walletAddr, asset.Symbol, initialAmountDecimal)
		require.NoError(t, err)

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, finalAmount)

		var balanceUpdateCalled bool
		var capturedWallet string
		custody.sendBalanceUpdate = func(wallet string) {
			balanceUpdateCalled = true
			capturedWallet = wallet
		}

		var channelUpdateCalled bool
		var capturedChannel Channel
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
			capturedChannel = ch
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Closed")
		custody.handleClosed(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Equal(t, uint64(0), updatedChannel.Amount, "Amount should be zero after closing")
		assert.Greater(t, updatedChannel.Version, initialChannel.Version, "Version should be incremented")

		var entries []Entry
		err = db.Where("wallet = ?", walletAddr).Find(&entries).Error
		require.NoError(t, err)
		assert.NotEmpty(t, entries)

		assert.True(t, balanceUpdateCalled, "Balance update callback should be called")
		assert.Equal(t, walletAddr, capturedWallet)

		assert.True(t, channelUpdateCalled, "Channel update callback should be called")
		assert.Equal(t, channelID, capturedChannel.ChannelID)
		assert.Equal(t, ChannelStatusClosed, capturedChannel.Status)

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		walletBalance, err := ledger.Balance(walletAddr, asset.Symbol)
		require.NoError(t, err)

		assert.Equal(t, walletBalance.String(), "0.5") // Final amount

		channelBalance, err := ledger.Balance(channelID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after closing")
	})

	t.Run("Success Equal Final aAmount", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialAmount := uint64(1000000)
		finalAmount := int64(1000000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      initialAmount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		// Set up initial wallet balance
		asset, err := GetAssetByToken(db, tokenAddress, custody.chainID)
		require.NoError(t, err)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromInt(int64(initialAmount)).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))

		// Initial balance in wallet
		err = ledger.Record(walletAddr, asset.Symbol, initialAmountDecimal)
		require.NoError(t, err)

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, finalAmount)

		logger := custody.logger.With("event", "Closed")
		custody.handleClosed(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Equal(t, uint64(0), updatedChannel.Amount, "Amount should be zero after closing")

		// Check final wallet balance
		walletBalance, err := ledger.Balance(walletAddr, asset.Symbol)
		require.NoError(t, err)

		// Wallet should have initial balance
		assert.True(t, walletBalance.Equal(decimal.Zero))

		// Channel balance should be zero
		channelBalance, err := ledger.Balance(channelID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after closing")
	})
}

func TestHandleChallengedEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		amount := uint64(1000000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      amount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, int64(amount))

		var channelUpdateCalled bool
		var capturedChannel Channel
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
			capturedChannel = ch
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Challenged")
		custody.handleChallenged(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusChallenged, updatedChannel.Status)
		assert.Equal(t, uint64(2), updatedChannel.Version, "Version should be updated to match event")
		assert.Equal(t, initialChannel.Amount, updatedChannel.Amount, "Amount should not change")

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		assert.True(t, channelUpdateCalled, "Channel update callback should be called for challenged event")
		assert.Equal(t, channelID, capturedChannel.ChannelID)
		assert.Equal(t, ChannelStatusChallenged, capturedChannel.Status)
	})
}

func TestHandleResizedEvent(t *testing.T) {
	t.Run("Positive Resize", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialAmount := uint64(1000000)
		deltaAmount := int64(500000) // Increase
		expectedAmount := uint64(1500000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      initialAmount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, err := GetAssetByToken(db, tokenAddress, custody.chainID)
		require.NoError(t, err)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromInt(int64(initialAmount)).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))

		err = ledger.Record(walletAddr, asset.Symbol, initialAmountDecimal)
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, deltaAmount)

		var balanceUpdateCalled bool
		var capturedWallet string
		custody.sendBalanceUpdate = func(wallet string) {
			balanceUpdateCalled = true
			capturedWallet = wallet
		}

		var channelUpdateCalled bool
		var capturedChannel Channel
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
			capturedChannel = ch
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status, "Status should remain open")
		assert.Equal(t, expectedAmount, updatedChannel.Amount, "Amount should be increased by deltaAmount")
		assert.Greater(t, updatedChannel.Version, initialChannel.Version, "Version should be incremented")

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		assert.True(t, balanceUpdateCalled, "Balance update callback should be called")
		assert.Equal(t, walletAddr, capturedWallet)

		assert.True(t, channelUpdateCalled, "Channel update callback should be called")
		assert.Equal(t, channelID, capturedChannel.ChannelID)
		assert.Equal(t, expectedAmount, capturedChannel.Amount)

		walletBalance, err := ledger.Balance(walletAddr, asset.Symbol)
		require.NoError(t, err)

		deltaAmountDecimal := decimal.NewFromInt(deltaAmount).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))
		expectedWalletBalance := initialAmountDecimal.Add(deltaAmountDecimal)

		assert.True(t, expectedWalletBalance.Equal(walletBalance),
			"Wallet balance should be %s after resize, got %s", expectedWalletBalance, walletBalance)

		channelBalance, err := ledger.Balance(channelID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after resize (funds moved to wallet)")
	})

	t.Run("Negative Resize", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialAmount := uint64(1000000)
		deltaAmount := int64(-300000) // Decrease
		expectedAmount := uint64(700000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"
		participantAddr := "0xParticipant1"

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr,
			Participant: participantAddr,
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			Amount:      initialAmount,
			Nonce:       12345,
			Version:     1,
			Challenge:   3600,
			Adjudicator: "0xAdjudicatorAddress",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, err := GetAssetByToken(db, tokenAddress, custody.chainID)
		require.NoError(t, err)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromInt(int64(initialAmount)).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))

		err = ledger.Record(walletAddr, asset.Symbol, initialAmountDecimal)
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, deltaAmount)

		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status)
		assert.Equal(t, expectedAmount, updatedChannel.Amount)
		assert.Greater(t, updatedChannel.Version, initialChannel.Version)

		walletBalance, err := ledger.Balance(walletAddr, asset.Symbol)
		require.NoError(t, err)

		deltaAmountDecimal := decimal.NewFromInt(deltaAmount).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Decimals))))
		expectedWalletBalance := initialAmountDecimal.Add(deltaAmountDecimal)

		assert.True(t, expectedWalletBalance.Equal(walletBalance),
			"Wallet balance should be %s after resize, got %s", expectedWalletBalance, walletBalance)

		channelBalance, err := ledger.Balance(channelID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after resize")
	})

	t.Run("Channel Not Found", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		// Create a different channel ID than the one in the event
		initialChannel := Channel{
			ChannelID: "0xDifferentChannelId",
			Wallet:    "0xWallet123",
			Status:    ChannelStatusOpen,
			Token:     tokenAddress,
			ChainID:   custody.chainID,
			Amount:    1000000,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, 500000)

		var balanceUpdateCalled bool
		custody.sendBalanceUpdate = func(wallet string) {
			balanceUpdateCalled = true
		}

		var channelUpdateCalled bool
		custody.sendChannelUpdate = func(ch Channel) {
			channelUpdateCalled = true
		}

		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)

		// Event should be ignored, and no callbacks should be called
		assert.False(t, balanceUpdateCalled, "Balance update should not be called for non-existent channel")
		assert.False(t, channelUpdateCalled, "Channel update should not be called for non-existent channel")

		// Initial channel should remain unmodified
		var checkChannel Channel
		err = db.Where("channel_id = ?", initialChannel.ChannelID).First(&checkChannel).Error
		require.NoError(t, err)
		assert.Equal(t, initialChannel.Amount, checkChannel.Amount, "Amount of other channel should not change")
	})
}

func TestHandleEventWithInvalidChannel(t *testing.T) {
	t.Run("Invalid Channel For Joined", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockJoinedEvent(t)

		logger := custody.logger.With("event", "Joined")
		// Should not panic when channel doesn't exist
		custody.handleJoined(logger, mockEvent)
	})

	t.Run("Invalid Channel For Closed", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, 500000)

		logger := custody.logger.With("event", "Closed")
		// Should not panic when channel doesn't exist
		custody.handleClosed(logger, mockEvent)
	})

	t.Run("Invalid Channel For Challenged", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, 500000)

		logger := custody.logger.With("event", "Challenged")
		// Should not panic when channel doesn't exist
		custody.handleChallenged(logger, mockEvent)
	})

	t.Run("Invalid Channel For Resized", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockResizedEvent(t, 500000)

		logger := custody.logger.With("event", "Resized")
		// Should not panic when channel doesn't exist
		custody.handleResized(logger, mockEvent)
	})
}
