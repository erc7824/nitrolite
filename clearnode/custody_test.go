package main

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

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

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
)

var tokenAddress = "0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"

func newTestCommonAddress(s string) common.Address {
	return common.Address(newTestCommonHash(s).Bytes()[:common.AddressLength])
}

func newTestCommonHash(s string) common.Hash {
	return crypto.Keccak256Hash([]byte(s))
}

func setupMockCustody(t *testing.T) (*Custody, *gorm.DB, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)

	rawKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	signer := &Signer{privateKey: rawKey}

	logger := NewLoggerIPFS("custody_test")

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

	assetsCfg := &AssetsConfig{
		Assets: []AssetConfig{
			{
				Symbol: "usdc",
				Tokens: []TokenConfig{
					{
						BlockchainID: uint32(chainID.Int64()),
						Address:      tokenAddress,
						Symbol:       "usdc",
						Decimals:     6,
					},
				},
			},
		},
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
		adjudicatorAddress: newTestCommonAddress("0xAdjudicatorAddress"),
		assetsCfg:          assetsCfg,
		wsNotifier:         NewWSNotifier(func(userID string, method string, params RPCDataParams) {}, logger),
		logger:             logger,
	}

	return custody, db, cleanup
}

func createMockLog(eventID common.Hash) types.Log {
	return types.Log{
		Address:     newTestCommonAddress("0xCustodyContractAddress"),
		Topics:      []common.Hash{eventID},
		Data:        []byte{},
		TxHash:      newTestCommonHash("0xTransactionHash"),
		BlockNumber: 12345678,
		Index:       0,
	}
}

func createMockCreatedEvent(t *testing.T, signer *Signer, token string, amount *big.Int) (*types.Log, *nitrolite.CustodyCreated) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}
	walletAddr := newTestCommonAddress("0xWallet123")
	participantAddr := newTestCommonAddress("0xParticipant1")

	channel := nitrolite.Channel{
		Participants: []common.Address{participantAddr, signer.GetAddress()},
		Adjudicator:  newTestCommonAddress("0xAdjudicatorAddress"),
		Challenge:    3600,
		Nonce:        12345,
	}

	allocation := []nitrolite.Allocation{
		{
			Destination: participantAddr,
			Token:       common.HexToAddress(token),
			Amount:      amount,
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
		Sigs:        [][]byte{},
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

func createMockClosedEvent(t *testing.T, signer *Signer, token string, amount *big.Int) (*types.Log, *nitrolite.CustodyClosed) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	participantAddr := common.HexToAddress("0xParticipant1")
	allocation := []nitrolite.Allocation{
		{
			Destination: participantAddr,
			Token:       common.HexToAddress(token),
			Amount:      amount,
		},
		{
			Destination: signer.GetAddress(),
			Token:       common.HexToAddress(token),
			Amount:      big.NewInt(0),
		},
	}

	finalState := nitrolite.State{
		Intent:      2,
		Version:     big.NewInt(2),
		Data:        []byte{},
		Allocations: allocation,
		Sigs:        [][]byte{},
	}

	event := &nitrolite.CustodyClosed{
		ChannelId:  channelID,
		FinalState: finalState,
	}

	log := createMockLog(custodyAbi.Events["Closed"].ID)

	return &log, event
}

func createMockChallengedEvent(t *testing.T, signer *Signer, token string, amount *big.Int) (*types.Log, *nitrolite.CustodyChallenged) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	participantAddr := newTestCommonAddress("0xParticipant1")
	allocation := []nitrolite.Allocation{
		{
			Destination: participantAddr,
			Token:       common.HexToAddress(token),
			Amount:      amount,
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
		Sigs:        [][]byte{},
	}

	event := &nitrolite.CustodyChallenged{
		ChannelId:  channelID,
		State:      state,
		Expiration: big.NewInt(time.Now().Add(1 * time.Hour).Unix()),
	}

	log := createMockLog(custodyAbi.Events["Challenged"].ID)

	return &log, event
}

func createMockResizedEvent(t *testing.T, amount *big.Int) (*types.Log, *nitrolite.CustodyResized) {
	t.Helper()

	channelID := [32]byte{1, 2, 3, 4}

	deltaAllocations := []*big.Int{
		amount,
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
	uint256Max := new(big.Int)
	uint256Max.Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
	uint256MaxMinus1 := new(big.Int).Sub(uint256Max, big.NewInt(1))

	testCases := []struct {
		name        string
		amount      *big.Int
		description string
	}{
		{
			name:        "Normal Amount",
			amount:      big.NewInt(1000000),
			description: "Test with normal amount of 1,000,000",
		},
		{
			name:        "Max Uint256 - 1",
			amount:      uint256MaxMinus1,
			description: "Test with uint256 max - 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			custody, db, cleanup := setupMockCustody(t)
			defer cleanup()

			channelIDBytes := [32]byte{1, 2, 3, 4}
			walletAddr := common.HexToAddress("0xWallet123")
			channelStruct := nitrolite.Channel{
				Participants: []common.Address{common.HexToAddress("0xParticipant1"), custody.signer.GetAddress()},
				Adjudicator:  newTestCommonAddress("0xAdjudicatorAddress"),
				Challenge:    3600,
				Nonce:        12345,
			}

			allocation := []nitrolite.Allocation{
				{
					Destination: common.HexToAddress("0xParticipant1"),
					Token:       common.HexToAddress(tokenAddress),
					Amount:      tc.amount,
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
				Sigs:        [][]byte{},
			}

			mockEvent := &nitrolite.CustodyCreated{
				ChannelId: channelIDBytes,
				Wallet:    walletAddr,
				Channel:   channelStruct,
				Initial:   initialState,
			}

			capturedNotifications := make(map[string][]Notification)
			custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {
				if capturedNotifications[userID] == nil {
					capturedNotifications[userID] = make([]Notification, 0)
				}
				capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
					userID:    userID,
					eventType: EventType(method),
					data:      params,
				})
			}

			logger := custody.logger
			custody.handleCreated(logger, mockEvent)

			channelIDStr := common.Hash(mockEvent.ChannelId).Hex()
			var dbChannel Channel
			dbErr := db.Where("channel_id = ?", channelIDStr).First(&dbChannel).Error
			require.NoError(t, dbErr)

			assert.Equal(t, dbChannel.ChannelID, channelIDStr)
			assert.Equal(t, dbChannel.Wallet, mockEvent.Wallet.Hex())
			assert.Equal(t, dbChannel.Participant, mockEvent.Channel.Participants[0].Hex())
			assert.Equal(t, dbChannel.Nonce, mockEvent.Channel.Nonce)
			assert.Equal(t, dbChannel.Challenge, mockEvent.Channel.Challenge)
			assert.Equal(t, dbChannel.Adjudicator, mockEvent.Channel.Adjudicator.Hex())
			assert.Equal(t, dbChannel.RawAmount, decimal.NewFromBigInt(tc.amount, 0))
			assert.Equal(t, dbChannel.Token, tokenAddress)
			assert.Equal(t, dbChannel.Status, ChannelStatusOpen)

			var entries []Entry
			entriesErr := db.Where("wallet = ?", mockEvent.Wallet.Hex()).Find(&entries).Error
			require.NoError(t, entriesErr)
			assert.NotEmpty(t, entries)

			assertNotifications(t, capturedNotifications, mockEvent.Wallet.Hex(), 2)

			assert.Equal(t, dbChannel.ChainID, uint32(custody.chainID))
			assert.False(t, dbChannel.CreatedAt.IsZero())
			assert.False(t, dbChannel.UpdatedAt.IsZero())

			assert.WithinDuration(t, time.Now(), dbChannel.CreatedAt, 2*time.Second)
			assert.WithinDuration(t, time.Now(), dbChannel.UpdatedAt, 2*time.Second)

			walletLedger := GetWalletLedger(db, mockEvent.Wallet)
			balance, err := walletLedger.Balance(NewAccountID(mockEvent.Wallet.Hex()), "usdc")
			require.NoError(t, err)

			assert.Equal(t, tc.amount.String(), balance.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(6))).String()) // 6 decimals for USDC default test token
		})
	}
}

func TestHandleClosedEvent(t *testing.T) {
	t.Run("Success Smaller Final Amount", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		finalAmount := big.NewInt(500000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialRawAmountDecimal := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		err = ledger.Record(walletAccountID, asset.Symbol, initialRawAmountDecimal, nil)
		require.NoError(t, err)

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, finalAmount)

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Closed")
		custody.handleClosed(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Equal(t, decimal.NewFromInt(0), updatedChannel.RawAmount, "Amount should be zero after closing")
		assert.Greater(t, updatedChannel.State.Version, initialChannel.State.Version, "Version should be incremented")

		var entries []Entry
		err = db.Where("wallet = ?", walletAddr.Hex()).Find(&entries).Error
		require.NoError(t, err)
		assert.NotEmpty(t, entries)

		assertNotifications(t, capturedNotifications, walletAddr.Hex(), 2)
		assert.Equal(t, ChannelUpdateEventType, capturedNotifications[walletAddr.Hex()][1].eventType)

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)

		assert.Equal(t, walletBalance.String(), "0.5") // Final amount

		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after closing")

		// Verify transaction was recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", walletAddr.Hex(), channelID).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 1, "Should have 1 withdrawal transaction recorded")

		tx := transactions[0]
		assert.Equal(t, TransactionTypeWithdrawal, tx.Type, "Transaction type should be withdrawal")
		assert.Equal(t, walletAddr.Hex(), tx.FromAccount, "From account should be wallet address")
		assert.Equal(t, channelID, tx.ToAccount, "To account should be channel ID")
		assert.Equal(t, asset.Symbol, tx.AssetSymbol, "Asset symbol should match")

		finalAmountDecimal := decimal.NewFromBigInt(finalAmount, 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		assert.True(t, finalAmountDecimal.Equal(tx.Amount), "Transaction amount should match final amount")
		assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
	})

	t.Run("Success Equal Final Amount", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		finalAmount := big.NewInt(1000000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		// Set up initial wallet balance
		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		// Initial balance in wallet
		err = ledger.Record(walletAccountID, asset.Symbol, initialAmountDecimal, nil)
		require.NoError(t, err)

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, finalAmount)

		logger := custody.logger.With("event", "Closed")
		custody.handleClosed(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Equal(t, decimal.NewFromInt(0), updatedChannel.RawAmount, "Amount should be zero after closing")

		// Check final wallet balance
		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)

		// Wallet should have initial balance
		assert.True(t, walletBalance.Equal(decimal.Zero))

		// Channel balance should be zero
		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after closing")

		// Verify transaction was recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", walletAddr.Hex(), channelID).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 1, "Should have 1 withdrawal transaction recorded")

		tx := transactions[0]
		assert.Equal(t, TransactionTypeWithdrawal, tx.Type, "Transaction type should be withdrawal")
		assert.Equal(t, walletAddr.Hex(), tx.FromAccount, "From account should be wallet address")
		assert.Equal(t, channelID, tx.ToAccount, "To account should be channel ID")
		assert.Equal(t, asset.Symbol, tx.AssetSymbol, "Asset symbol should match")

		finalAmountDecimal := decimal.NewFromBigInt(finalAmount, 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		assert.True(t, finalAmountDecimal.Equal(tx.Amount), "Transaction amount should match final amount")
		assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
	})

	t.Run("Close Resizing Channel With Escrow", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		escrowAmount := decimal.NewFromInt(300000) // Amount locked in channel escrow during resize
		finalAmount := big.NewInt(500000)          // Final amount after close

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusResizing, // Channel is in resizing state
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialWalletBalanceDecimal := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		escrowAmountDecimal := escrowAmount.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		// Set up wallet balance
		require.NoError(t, ledger.Record(walletAccountID, asset.Symbol, initialWalletBalanceDecimal, nil))
		// Set up channel escrow balance (simulating funds locked during resize)
		require.NoError(t, ledger.Record(channelAccountID, asset.Symbol, escrowAmountDecimal, nil))

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, finalAmount)

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {
			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		logger := custody.logger.With("event", "Closed")
		custody.handleClosed(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		// Verify channel is closed
		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Equal(t, decimal.NewFromInt(0), updatedChannel.RawAmount, "Amount should be zero after closing")

		// Verify channel escrow balance is now zero (unlocked)
		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel escrow balance should be zero after closing")

		// Verify wallet balance
		finalAmountDecimal := decimal.NewFromBigInt(finalAmount, 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		expectedWalletBalance := initialWalletBalanceDecimal.Add(escrowAmountDecimal).Sub(finalAmountDecimal)

		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, walletBalance.Equal(expectedWalletBalance),
			"Wallet balance should be %s (initial %s + escrow %s - final %s), got %s",
			expectedWalletBalance, initialWalletBalanceDecimal, escrowAmountDecimal, finalAmountDecimal, walletBalance)

		// Verify transactions were recorded
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? OR to_account = ?", channelID, channelID).Order("created_at ASC").Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 2, "Should have 2 transactions: escrow unlock and withdrawal")

		// First transaction should be escrow unlock
		unlockTx := transactions[0]
		assert.Equal(t, TransactionTypeEscrowUnlock, unlockTx.Type, "First transaction should be escrow unlock")
		assert.Equal(t, channelID, unlockTx.FromAccount, "Escrow unlock from account should be channel ID")
		assert.Equal(t, walletAddr.Hex(), unlockTx.ToAccount, "Escrow unlock to account should be wallet address")
		assert.Equal(t, asset.Symbol, unlockTx.AssetSymbol, "Asset symbol should match")
		assert.True(t, escrowAmountDecimal.Equal(unlockTx.Amount), "Unlock amount should match escrow amount")

		// Second transaction should be withdrawal
		withdrawalTx := transactions[1]
		assert.Equal(t, TransactionTypeWithdrawal, withdrawalTx.Type, "Second transaction should be withdrawal")
		assert.Equal(t, walletAddr.Hex(), withdrawalTx.FromAccount, "Withdrawal from account should be wallet address")
		assert.Equal(t, channelID, withdrawalTx.ToAccount, "Withdrawal to account should be channel ID")
		assert.Equal(t, asset.Symbol, withdrawalTx.AssetSymbol, "Asset symbol should match")
		assert.True(t, finalAmountDecimal.Equal(withdrawalTx.Amount), "Withdrawal amount should match final amount")

		// Verify notifications
		assertNotifications(t, capturedNotifications, walletAddr.Hex(), 2)
	})
}

func TestHandleChallengedEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		amount := decimal.NewFromInt(1000000)

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
			RawAmount:   amount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, amount.BigInt())

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Challenged")
		custody.handleChallenged(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusChallenged, updatedChannel.Status)
		assert.Equal(t, uint64(1), updatedChannel.State.Version, "Version should be updated to match event")
		assert.Equal(t, initialChannel.RawAmount, updatedChannel.RawAmount, "Amount should not change")

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		assertNotifications(t, capturedNotifications, walletAddr, 1)
		assert.Equal(t, ChannelUpdateEventType, capturedNotifications[walletAddr][0].eventType)
	})
}

func TestHandleResizedEvent(t *testing.T) {
	t.Run("Positive Resize", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		deltaAmount := decimal.NewFromInt(500000) // Increase
		expectedAmount := decimal.NewFromInt(1500000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusResizing,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialAmountDecimal := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		err = ledger.Record(walletAccountID, asset.Symbol, initialAmountDecimal, nil)
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, deltaAmount.BigInt())

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		beforeUpdate := time.Now()
		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status, "Status should remain open")
		assert.Equal(t, expectedAmount, updatedChannel.RawAmount, "Amount should be increased by deltaAmount")
		assert.Greater(t, updatedChannel.State.Version, initialChannel.State.Version, "Version should be incremented")

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		assertNotifications(t, capturedNotifications, walletAddr.Hex(), 2)
		assert.Equal(t, ChannelUpdateEventType, capturedNotifications[walletAddr.Hex()][1].eventType)

		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)

		deltaAmountDecimal := deltaAmount.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		expectedWalletBalance := initialAmountDecimal.Add(deltaAmountDecimal)

		assert.True(t, expectedWalletBalance.Equal(walletBalance),
			"Wallet balance should be %s after resize, got %s", expectedWalletBalance, walletBalance)

		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.IsZero(), "Channel balance should be zero after resize (funds moved to wallet)")

		// Verify transaction was recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", channelID, walletAddr.Hex()).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 1, "Should have 1 deposit transaction recorded")

		tx := transactions[0]
		assert.Equal(t, TransactionTypeDeposit, tx.Type, "Transaction type should be deposit")
		assert.Equal(t, channelID, tx.FromAccount, "From account should be channel ID")
		assert.Equal(t, walletAddr.Hex(), tx.ToAccount, "To account should be wallet address")
		assert.Equal(t, asset.Symbol, tx.AssetSymbol, "Asset symbol should match")
		assert.True(t, deltaAmountDecimal.Equal(tx.Amount), "Transaction amount should match delta amount")
		assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
	})

	t.Run("Negative Resize", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		deltaAmount := decimal.NewFromInt(-300000) // Decrease
		expectedAmount := decimal.NewFromInt(700000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusResizing,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialWalletBalance := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		err = ledger.Record(walletAccountID, asset.Symbol, initialWalletBalance, nil)
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, deltaAmount.BigInt())

		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status)
		assert.Equal(t, expectedAmount, updatedChannel.RawAmount)
		assert.Greater(t, updatedChannel.State.Version, initialChannel.State.Version)

		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)

		deltaAmountDecimal := deltaAmount.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		assert.True(t, initialWalletBalance.Equal(walletBalance),
			"Wallet balance should stay %s after resize event, got %s", initialWalletBalance, walletBalance)

		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalance.Equal(deltaAmountDecimal), "Channel escrow balance should be %s after resize withdrawal, got %s", deltaAmountDecimal, channelBalance)

		// Verify transaction was recorded to the database
		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", walletAddr.Hex(), channelID).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 1, "Should have 1 withdrawal transaction recorded")

		tx := transactions[0]
		assert.Equal(t, TransactionTypeWithdrawal, tx.Type, "Transaction type should be withdrawal")
		assert.Equal(t, walletAddr.Hex(), tx.FromAccount, "From account should be wallet address")
		assert.Equal(t, channelID, tx.ToAccount, "To account should be channel ID")
		assert.Equal(t, asset.Symbol, tx.AssetSymbol, "Asset symbol should match")
		assert.True(t, deltaAmountDecimal.Abs().Equal(tx.Amount), "Transaction amount should match delta amount")
		assert.False(t, tx.CreatedAt.IsZero(), "CreatedAt should be set")
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
			RawAmount: decimal.NewFromInt(1000000),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockResizedEvent(t, big.NewInt(500000))

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params RPCDataParams) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)

		// Event should be ignored, and no callbacks should be called
		assert.Equal(t, 0, len(capturedNotifications), "No notifications should be sent")

		// Initial channel should remain unmodified
		var checkChannel Channel
		err = db.Where("channel_id = ?", initialChannel.ChannelID).First(&checkChannel).Error
		require.NoError(t, err)
		assert.Equal(t, initialChannel.RawAmount, checkChannel.RawAmount, "Amount of other channel should not change")
	})

	t.Run("Negative Resize On Non-Resizing Channel", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		initialRawAmount := decimal.NewFromInt(1000000)
		deltaAmount := decimal.NewFromInt(-300000) // Decrease (withdrawal)
		expectedAmount := decimal.NewFromInt(700000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		channelAccountID := NewAccountID(channelID)
		walletAddr := newTestCommonAddress("0xWallet123")
		walletAccountID := NewAccountID(walletAddr.Hex())
		participantAddr := newTestCommonAddress("0xParticipant1")

		// Create a channel that is NOT in resizing state
		initialChannel := Channel{
			ChannelID:   channelID,
			Wallet:      walletAddr.Hex(),
			Participant: participantAddr.Hex(),
			Status:      ChannelStatusOpen,
			Token:       tokenAddress,
			ChainID:     custody.chainID,
			RawAmount:   initialRawAmount,
			Nonce:       12345,
			State: UnsignedState{
				Version: 1,
			},
			Challenge:   3600,
			Adjudicator: newTestCommonAddress("0xAdjudicatorAddress").Hex(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		asset, ok := custody.assetsCfg.GetAssetTokenByAddressAndChainID(tokenAddress, custody.chainID)
		require.True(t, ok)

		ledger := GetWalletLedger(db, walletAddr)
		initialWalletBalance := decimal.NewFromBigInt(initialRawAmount.BigInt(), 0).Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))

		err = ledger.Record(walletAccountID, asset.Symbol, initialWalletBalance, nil)
		require.NoError(t, err)

		// Verify initial channel escrow balance is zero
		channelBalanceBefore, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, channelBalanceBefore.IsZero(), "Channel escrow should start at zero")

		_, mockEvent := createMockResizedEvent(t, deltaAmount.BigInt())

		logger := custody.logger.With("event", "Resized")
		custody.handleResized(logger, mockEvent)

		// Verify channel was updated despite being in wrong state
		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusOpen, updatedChannel.Status, "Status should be open")
		assert.Equal(t, expectedAmount, updatedChannel.RawAmount, "Channel amount should be decreased")
		assert.Greater(t, updatedChannel.State.Version, initialChannel.State.Version, "Version should be incremented")

		// Verify wallet balance hasn't changed
		walletBalance, err := ledger.Balance(walletAccountID, asset.Symbol)
		require.NoError(t, err)
		assert.True(t, initialWalletBalance.Equal(walletBalance),
			"Wallet balance should stay %s, got %s", initialWalletBalance, walletBalance)

		// Verify channel escrow balance is now NEGATIVE in this scenario
		deltaAmountDecimal := deltaAmount.Div(decimal.NewFromInt(10).Pow(decimal.NewFromInt(int64(asset.Token.Decimals))))
		channelBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		require.NoError(t, err)

		assert.True(t, channelBalance.Equal(deltaAmountDecimal),
			"Channel escrow balance should be %s (negative!), got %s", deltaAmountDecimal, channelBalance)
		assert.True(t, channelBalance.IsNegative(),
			"Channel escrow balance should be negative")

		var transactions []LedgerTransaction
		err = db.Where("from_account = ? AND to_account = ?", walletAddr.Hex(), channelID).Find(&transactions).Error
		require.NoError(t, err)
		assert.Len(t, transactions, 1, "Should have 1 withdrawal transaction recorded")

		tx := transactions[0]
		assert.Equal(t, TransactionTypeWithdrawal, tx.Type, "Transaction type should be withdrawal")
		assert.Equal(t, walletAddr.Hex(), tx.FromAccount, "From account should be wallet address")
		assert.Equal(t, channelID, tx.ToAccount, "To account should be channel ID")
		assert.Equal(t, asset.Symbol, tx.AssetSymbol, "Asset symbol should match")
		assert.True(t, deltaAmountDecimal.Abs().Equal(tx.Amount), "Transaction amount should match delta amount")
	})
}

func TestHandleEventWithInvalidChannel(t *testing.T) {
	t.Run("Invalid Channel For Closed", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, big.NewInt(500000))

		logger := custody.logger.With("event", "Closed")
		// Should not panic when channel doesn't exist
		custody.handleClosed(logger, mockEvent)
	})

	t.Run("Invalid Channel For Challenged", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, big.NewInt(500000))

		logger := custody.logger.With("event", "Challenged")
		// Should not panic when channel doesn't exist
		custody.handleChallenged(logger, mockEvent)
	})

	t.Run("Invalid Channel For Resized", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockResizedEvent(t, big.NewInt(500000))

		logger := custody.logger.With("event", "Resized")
		// Should not panic when channel doesn't exist
		custody.handleResized(logger, mockEvent)
	})
}

func TestChallengeHandling(t *testing.T) {
	channelID := common.HexToHash("0x0000000000000000000000001234567890abcdef1234567890abcdef12345678")
	initialState := UnsignedState{
		Intent:  StateIntent(StateIntentOperate),
		Version: 5,
		Data:    "data",
		Allocations: []Allocation{
			{
				Participant:  "0xUser123456789",
				TokenAddress: "0xToken123456789",
				RawAmount:    decimal.NewFromInt(1000),
			},
			{
				Participant:  "0xBroker123456789",
				TokenAddress: "0xToken123456789",
				RawAmount:    decimal.NewFromInt(500),
			},
		},
	}

	userSig := Signature{1, 2, 3}
	serverSig := Signature{4, 5, 6}

	t.Run("Challenge with older state creates checkpoint action", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		channel, err := CreateChannel(
			db,
			channelID.Hex(),
			"0xWallet123456789",
			"0xParticipant123456789",
			1,
			3600,
			"0xAdjudicator123456789",
			custody.chainID,
			"0xToken123456789",
			decimal.NewFromInt(1500),
			initialState,
		)
		require.NoError(t, err)

		channel.UserStateSignature = &userSig
		channel.ServerStateSignature = &serverSig
		require.NoError(t, db.Save(&channel).Error)

		challengedEvent := &nitrolite.CustodyChallenged{
			ChannelId: [32]byte(channelID),
			State: nitrolite.State{
				Intent:  0,
				Version: big.NewInt(3), // Older version - should trigger checkpoint
				Data:    []byte("attack-data"),
				Allocations: []nitrolite.Allocation{
					{
						Destination: common.HexToAddress("0xUser123456789"),
						Token:       common.HexToAddress("0xToken123456789"),
						Amount:      big.NewInt(2000),
					},
					{
						Destination: common.HexToAddress("0xBroker123456789"),
						Token:       common.HexToAddress("0xToken123456789"),
						Amount:      big.NewInt(0),
					},
				},
			},
			Raw: types.Log{
				TxHash: common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111"),
				Index:  0,
			},
		}

		custody.handleChallenged(custody.logger, challengedEvent)

		// Verify channel is marked as challenged
		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID.Hex()).First(&updatedChannel).Error
		require.NoError(t, err)
		assert.Equal(t, ChannelStatusChallenged, updatedChannel.Status)

		// Verify checkpoint action was created
		var action BlockchainAction
		err = db.Where("channel_id = ? AND action_type = ?", channelID, ActionTypeCheckpoint).First(&action).Error
		require.NoError(t, err)

		assert.Equal(t, ActionTypeCheckpoint, action.Type)
		assert.Equal(t, channelID, action.ChannelID)
		assert.Equal(t, custody.chainID, action.ChainID)
		assert.Equal(t, StatusPending, action.Status)
		assert.Equal(t, 0, action.Retries)

		// Verify checkpoint data is correct
		var checkpointData CheckpointData
		err = json.Unmarshal([]byte(action.Data), &checkpointData)
		require.NoError(t, err)
		assert.Equal(t, initialState, checkpointData.State)
		assert.Equal(t, userSig, checkpointData.UserSig)
		assert.Equal(t, serverSig, checkpointData.ServerSig)
	})

	t.Run("Challenge with same version - no checkpoint needed", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		channel, err := CreateChannel(
			db,
			channelID.Hex(),
			"0xWallet123456789",
			"0xParticipant123456789",
			1,
			3600,
			"0xAdjudicator123456789",
			custody.chainID,
			"0xToken123456789",
			decimal.NewFromInt(1500),
			initialState,
		)
		require.NoError(t, err)

		channel.UserStateSignature = &userSig
		channel.ServerStateSignature = &serverSig
		require.NoError(t, db.Save(&channel).Error)

		challengedEvent := &nitrolite.CustodyChallenged{
			ChannelId: [32]byte(channelID),
			State: nitrolite.State{
				Intent:  0,
				Version: big.NewInt(5), // Same version - no checkpoint needed
				Data:    []byte("valid-data"),
				Allocations: []nitrolite.Allocation{
					{
						Destination: common.HexToAddress("0xUser123456789"),
						Token:       common.HexToAddress("0xToken123456789"),
						Amount:      big.NewInt(1000),
					},
					{
						Destination: common.HexToAddress("0xBroker123456789"),
						Token:       common.HexToAddress("0xToken123456789"),
						Amount:      big.NewInt(500),
					},
				},
			},
			Raw: types.Log{
				TxHash: common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222"),
				Index:  0,
			},
		}

		custody.handleChallenged(custody.logger, challengedEvent)

		// Verify channel is marked as challenged
		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID.Hex()).First(&updatedChannel).Error
		require.NoError(t, err)
		assert.Equal(t, ChannelStatusChallenged, updatedChannel.Status)

		// Verify NO checkpoint action was created
		var count int64
		err = db.Model(&BlockchainAction{}).Where("channel_id = ? AND action_type = ?", channelID.Hex(), ActionTypeCheckpoint).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count, "No checkpoint action should be created for same version")
	})
}
