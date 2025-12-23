package custody

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

	"github.com/erc7824/nitrolite/clearnode/api"
	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/erc7824/nitrolite/clearnode/store/memory"
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

	db, cleanup := api.SetupTestDB(t)
	signer, _ := sign.NewEthereumSigner("twst-key-seed")

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

	assetsCfg := &memory.AssetsConfig{
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
		wsNotifier:         NewWSNotifier(func(userID string, method string, params rpc.Params) {}, logger),
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

func createMockCreatedEvent(t *testing.T, signer sign.Signer, token string, amount *big.Int) (*types.Log, *nitrolite.CustodyCreated) {
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
			custody.wsNotifier.notify = func(userID string, method string, params rpc.Params) {
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
			assert.Equal(t, dbChannel.UserWallet, mockEvent.Wallet.Hex())
			assert.Equal(t, dbChannel.Nonce, mockEvent.Channel.Nonce)
			assert.Equal(t, dbChannel.Challenge, mockEvent.Channel.Challenge)
			assert.Equal(t, dbChannel.Token, tokenAddress)
			assert.Equal(t, dbChannel.Status, ChannelStatusOpen)

			var entries []Entry
			entriesErr := db.Where("wallet = ?", mockEvent.Wallet.Hex()).Find(&entries).Error
			require.NoError(t, entriesErr)
			assert.NotEmpty(t, entries)

			assertNotifications(t, capturedNotifications, mockEvent.Wallet.Hex(), 2)

			assert.Equal(t, dbChannel.BlockchainID, uint32(custody.chainID))
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

		initialChannel := Channel{
			ChannelID:    channelID,
			UserWallet:   walletAddr.Hex(),
			Status:       ChannelStatusOpen,
			Token:        tokenAddress,
			BlockchainID: custody.chainID,
			Nonce:        12345,
			Challenge:    3600,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
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
		custody.wsNotifier.notify = func(userID string, method string, params rpc.Params) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		beforeUpdate := time.Now()
		logger := custody.logger.WithKV("event", "Closed")
		custody.handleClosed(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)
		assert.Greater(t, updatedChannel.OnChainStateVersion, initialChannel.OnChainStateVersion, "Version should be incremented")

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

		initialChannel := Channel{
			ChannelID:    channelID,
			UserWallet:   walletAddr.Hex(),
			Status:       ChannelStatusOpen,
			Token:        tokenAddress,
			BlockchainID: custody.chainID,
			Nonce:        12345,
			Challenge:    3600,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
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

		logger := custody.logger.WithKV("event", "Closed")
		custody.handleClosed(logger, mockEvent)

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusClosed, updatedChannel.Status)

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
}

func TestHandleChallengedEvent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		custody, db, cleanup := setupMockCustody(t)
		defer cleanup()

		amount := decimal.NewFromInt(1000000)

		channelID := "0x0102030400000000000000000000000000000000000000000000000000000000"
		walletAddr := "0xWallet123"

		initialChannel := Channel{
			ChannelID:  channelID,
			UserWallet: walletAddr,

			Status:       ChannelStatusOpen,
			Token:        tokenAddress,
			BlockchainID: custody.chainID,
			Nonce:        12345,
			Challenge:    3600,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		err := db.Create(&initialChannel).Error
		require.NoError(t, err)

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, amount.BigInt())

		capturedNotifications := make(map[string][]Notification)
		custody.wsNotifier.notify = func(userID string, method string, params rpc.Params) {

			capturedNotifications[userID] = append(capturedNotifications[userID], Notification{
				userID:    userID,
				eventType: EventType(method),
				data:      params,
			})
		}

		beforeUpdate := time.Now()
		logger := custody.logger.WithKV("event", "Challenged")
		custody.handleChallenged(logger, mockEvent)
		afterUpdate := time.Now()

		var updatedChannel Channel
		err = db.Where("channel_id = ?", channelID).First(&updatedChannel).Error
		require.NoError(t, err)

		assert.Equal(t, ChannelStatusChallenged, updatedChannel.Status)
		assert.Equal(t, uint64(1), updatedChannel.BlockchainID, "Version should be updated to match event")

		assert.Equal(t, initialChannel.CreatedAt.Unix(), updatedChannel.CreatedAt.Unix(), "CreatedAt should not change")
		assert.True(t, updatedChannel.UpdatedAt.After(initialChannel.UpdatedAt), "UpdatedAt should increase")
		assert.True(t, updatedChannel.UpdatedAt.After(beforeUpdate) && updatedChannel.UpdatedAt.Before(afterUpdate))

		assertNotifications(t, capturedNotifications, walletAddr, 1)
		assert.Equal(t, ChannelUpdateEventType, capturedNotifications[walletAddr][0].eventType)
	})
}

func TestHandleEventWithInvalidChannel(t *testing.T) {
	t.Run("Invalid Channel For Closed", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockClosedEvent(t, custody.signer, tokenAddress, big.NewInt(500000))

		logger := custody.logger.WithKV("event", "Closed")
		// Should not panic when channel doesn't exist
		custody.handleClosed(logger, mockEvent)
	})

	t.Run("Invalid Channel For Challenged", func(t *testing.T) {
		custody, _, cleanup := setupMockCustody(t)
		defer cleanup()

		_, mockEvent := createMockChallengedEvent(t, custody.signer, tokenAddress, big.NewInt(500000))

		logger := custody.logger.WithKV("event", "Challenged")
		// Should not panic when channel doesn't exist
		custody.handleChallenged(logger, mockEvent)
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

		// channel.UserStateSignature = &userSig
		// channel.ServerStateSignature = &serverSig
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

		// channel.UserStateSignature = &userSig
		// channel.ServerStateSignature = &serverSig
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
