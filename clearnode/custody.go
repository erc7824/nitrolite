package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	custodyAbi *abi.ABI
)

var ErrCustodyEventAlreadyProcessed = errors.New("custody event already processed")

type CustodyInterface interface {
	Checkpoint(channelID string, state UnsignedState, userSig, serverSig Signature, proofs []nitrolite.State) (common.Hash, error)
}

var _ CustodyInterface = (*Custody)(nil)

// Custody implements the BlockchainClient interface using the Custody contract
type Custody struct {
	client             Ethereum
	custody            *nitrolite.Custody
	balanceChecker     *nitrolite.BalanceChecker
	db                 *gorm.DB
	custodyAddr        common.Address
	transactOpts       *bind.TransactOpts
	chainID            uint32
	signer             *Signer
	adjudicatorAddress common.Address
	blockStep          uint64
	wsNotifier         *WSNotifier
	logger             Logger
}

// NewCustody initializes the Ethereum client and custody contract wrapper.
func NewCustody(signer *Signer, db *gorm.DB, wsNotifier *WSNotifier, infuraURL, custodyAddressStr, adjudicatorAddr, balanceCheckerAddr string, chain uint32, blockStep uint64, logger Logger) (*Custody, error) {
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain node: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create auth options for transactions.
	auth, err := bind.NewKeyedTransactorWithChainID(signer.GetPrivateKey(), chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction signer: %w", err)
	}

	// TODO: estimate on the go.
	auth.GasPrice = big.NewInt(30000000000) // 30 gwei.
	auth.GasLimit = uint64(3000000)

	custodyAddress := common.HexToAddress(custodyAddressStr)
	custody, err := nitrolite.NewCustody(custodyAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind custody contract: %w", err)
	}

	balanceChecker, err := nitrolite.NewBalanceChecker(common.HexToAddress(balanceCheckerAddr), client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind balance checker contract: %w", err)
	}

	return &Custody{
		client:             client,
		custody:            custody,
		balanceChecker:     balanceChecker,
		db:                 db,
		custodyAddr:        custodyAddress,
		transactOpts:       auth,
		chainID:            uint32(chainID.Int64()),
		signer:             signer,
		adjudicatorAddress: common.HexToAddress(adjudicatorAddr),
		wsNotifier:         wsNotifier,
		blockStep:          blockStep,
		logger:             logger.NewSystem("custody").With("chainID", chainID.Int64()).With("custodyAddress", custodyAddressStr),
	}, nil
}

// ListenEvents initializes event listening for the custody contract
func (c *Custody) ListenEvents(ctx context.Context) {
	ev, err := GetLatestContractEvent(c.db, c.custodyAddr.Hex(), c.chainID)
	if err != nil {
		c.logger.Error("failed to get latest contract event", "error", err)
		return
	}

	var lastBlock uint64
	var lastIndex uint32
	if ev != nil {
		lastBlock = ev.BlockNumber
		lastIndex = ev.LogIndex
	}

	listenEvents(ctx, c.client, c.custodyAddr, c.chainID, c.blockStep, lastBlock, lastIndex, c.handleBlockChainEvent, c.logger)
}

// Checkpoint calls the checkpoint method on the custody contract
func (c *Custody) Checkpoint(channelID string, newState UnsignedState, userSig, serverSig Signature, proofs []nitrolite.State) (common.Hash, error) {
	// Convert string channelID to bytes32
	channelIDBytes := common.HexToHash(channelID)

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	nitroState := nitrolite.State{
		Intent:  uint8(newState.Intent),
		Version: big.NewInt(int64(newState.Version)),
		Data:    []byte(newState.Data),
		Sigs:    [][]byte{userSig, serverSig},
	}

	for _, alloc := range newState.Allocations {
		nitroState.Allocations = append(nitroState.Allocations, nitrolite.Allocation{
			Destination: common.HexToAddress(alloc.Participant),
			Token:       common.HexToAddress(alloc.TokenAddress),
			Amount:      alloc.RawAmount.BigInt(),
		})
	}

	c.transactOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)

	// Call the checkpoint method on the custody contract
	tx, err := c.custody.Checkpoint(c.transactOpts, channelIDBytes, nitroState, proofs)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to checkpoint channel: %w", err)
	}

	return tx.Hash(), nil
}

// handleBlockChainEvent processes different event types received from the blockchain
func (c *Custody) handleBlockChainEvent(ctx context.Context, l types.Log) {
	ctx = SetContextLogger(ctx, c.logger)
	logger := LoggerFromContext(ctx)
	logger.Debug("received event", "blockNumber", l.BlockNumber, "txHash", l.TxHash.String(), "logIndex", l.Index)

	eventID := l.Topics[0]
	switch eventID {
	case custodyAbi.Events["Created"].ID:
		ev, err := c.custody.ParseCreated(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		c.handleCreated(logger, ev)
	case custodyAbi.Events["Joined"].ID:
		return
	case custodyAbi.Events["Challenged"].ID:
		ev, err := c.custody.ParseChallenged(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		c.handleChallenged(logger, ev)
	case custodyAbi.Events["Resized"].ID:
		ev, err := c.custody.ParseResized(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		c.handleResized(logger, ev)
	case custodyAbi.Events["Closed"].ID:
		ev, err := c.custody.ParseClosed(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		c.handleClosed(logger, ev)
	default:
		logger.Warn("unknown event", "eventID", eventID.Hex())
	}
}

func (c *Custody) handleCreated(logger Logger, ev *nitrolite.CustodyCreated) {
	logger = logger.With("event", "Created")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "wallet", ev.Wallet.Hex(), "channel", ev.Channel, "initial", ev.Initial)

	if len(ev.Channel.Participants) != 2 {
		logger.Warn("supported only 2 participants in the channel")
		return
	}

	wallet := ev.Wallet.Hex()
	participantSigner := ev.Channel.Participants[0].Hex()
	nonce := ev.Channel.Nonce
	broker := ev.Channel.Participants[1]
	tokenAddress := ev.Initial.Allocations[0].Token.Hex()
	rawAmount := ev.Initial.Allocations[0].Amount
	adjudicator := ev.Channel.Adjudicator
	challenge := ev.Channel.Challenge

	brokerAmount := ev.Initial.Allocations[1].Amount
	if brokerAmount.Cmp(big.NewInt(0)) != 0 {
		logger.Warn("non-zero broker amount", "amount", brokerAmount)
		return
	}

	if challenge < 3600 {
		logger.Warn("invalid challenge period", "challenge", challenge)
		return
	}

	if adjudicator != c.adjudicatorAddress {
		logger.Warn("unsupported adjudicator", "actual", adjudicator.Hex(), "expected", c.adjudicatorAddress.Hex())
		return
	}

	// Check if channel was created with the broker.
	if broker != c.signer.GetAddress() {
		logger.Warn("channel participant is not the broker", "actual", c.signer.GetAddress().Hex(), "expected", broker.Hex())
		return
	}

	if err := AddSigner(c.db, wallet, participantSigner); err != nil {
		logger.Error("failed to add signer", "error", err)
		return
	}

	var ch Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		if err := c.saveContractEvent(tx, "created", *ev, ev.Raw); err != nil {
			return err
		}

		// Check if there is already existing open channel with the broker
		existingOpenChannel, err := CheckExistingChannels(tx, wallet, tokenAddress, c.chainID)
		if err != nil {
			return err
		}

		if existingOpenChannel != nil {
			return fmt.Errorf("an open channel with broker already exists: %s", existingOpenChannel.ChannelID)
		}

		// Record initial channel state
		state := UnsignedState{
			Intent:  StateIntent(ev.Initial.Intent),
			Version: ev.Initial.Version.Uint64(),
			Data:    string(ev.Initial.Data),
		}
		for _, alloc := range ev.Initial.Allocations {
			state.Allocations = append(state.Allocations, Allocation{
				Participant:  alloc.Destination.Hex(),
				TokenAddress: alloc.Token.Hex(),
				RawAmount:    decimal.NewFromBigInt(alloc.Amount, 0),
			})
		}
		// NOTE: Signatures are not recorded with the initial state.

		ch, err = CreateChannel(
			tx,
			channelID,
			wallet,
			participantSigner,
			nonce,
			challenge,
			adjudicator.Hex(),
			c.chainID,
			tokenAddress,
			decimal.NewFromBigInt(rawAmount, 0),
			state,
		)
		if err != nil {
			return err
		}

		asset, err := GetAssetByToken(tx, tokenAddress, c.chainID)
		if err != nil {
			return fmt.Errorf("failed to get asset: %w", err)
		}
		amount := rawToDecimal(rawAmount, asset.Decimals)

		walletAddress := ev.Wallet
		channelAccountID := NewAccountID(channelID)
		walletAccountID := NewAccountID(walletAddress.Hex())

		ledger := GetWalletLedger(tx, walletAddress)
		if err := ledger.Record(channelAccountID, asset.Symbol, amount); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}
		if err := ledger.Record(channelAccountID, asset.Symbol, amount.Neg()); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}
		ledger = GetWalletLedger(tx, walletAddress)
		if err := ledger.Record(walletAccountID, asset.Symbol, amount); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}

		_, err = RecordLedgerTransaction(tx, TransactionTypeDeposit, channelAccountID, walletAccountID, asset.Symbol, amount)
		if err != nil {
			return fmt.Errorf("failed to record transaction: %w", err)
		}

		logger.Info("handled created event", "channelId", channelID)
		return nil
	})
	if errors.Is(err, ErrCustodyEventAlreadyProcessed) {
		return
	} else if err != nil {
		logger.Error("error creating channel in database", "error", err)
		return
	}

	c.wsNotifier.Notify(NewChannelNotification(ch))
	c.wsNotifier.Notify(NewBalanceNotification(ch.Wallet, c.db))
}

func (c *Custody) handleChallenged(logger Logger, ev *nitrolite.CustodyChallenged) {
	logger = logger.With("event", "Challenged")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		if err := c.saveContractEvent(tx, "challenged", *ev, ev.Raw); err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		challengedVersion := ev.State.Version.Uint64()
		localVersion := channel.State.Version

		logger.Warn("challenge detected", "challengedVersion", challengedVersion, "localVersion", localVersion, "channelId", channelID)

		if challengedVersion < localVersion {
			if channel.UserStateSignature != nil && channel.ServerStateSignature != nil {
				if err := CreateCheckpoint(tx, channelID, c.chainID, channel.State, *channel.UserStateSignature, *channel.ServerStateSignature); err != nil {
					logger.Error("failed to create checkpoint", "error", err)
				} else {
					logger.Info("created checkpoint action", "channelId", channelID, "localVersion", localVersion, "challengedVersion", challengedVersion)
				}
			} else {
				logger.Warn("detected newer local state in db without signatures", "channelId", channelID)
			}
		}
		channel.Status = ChannelStatusChallenged
		channel.UpdatedAt = time.Now()

		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("error saving channel in database: %w", err)
		}

		logger.Info("handled challenged event", "channelId", channelID)
		return nil
	})
	if errors.Is(err, ErrCustodyEventAlreadyProcessed) {
		return
	} else if err != nil {
		logger.Error("failed to update channel", "channelId", channelID, "error", err)
		return
	}

	c.wsNotifier.Notify(NewChannelNotification(channel))
}

func (c *Custody) handleResized(logger Logger, ev *nitrolite.CustodyResized) {
	logger = logger.With("event", "Resized")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "deltaAllocations", ev.DeltaAllocations)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		if err := c.saveContractEvent(tx, "resized", *ev, ev.Raw); err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		newRawAmount := channel.RawAmount.BigInt()
		for _, change := range ev.DeltaAllocations {
			newRawAmount.Add(newRawAmount, change)
		}

		if newRawAmount.Sign() < 0 {
			// TODO: what do we do in this case?
			logger.Error("invalid resize, channel balance cannot be negative", "channelId", channelID)
			return fmt.Errorf("invalid resize, channel balance cannot be negative: %s", newRawAmount.String())
		}

		// Update state allocations
		if len(ev.DeltaAllocations) == 2 && len(channel.State.Allocations) == 2 {
			channel.State.Allocations[0].RawAmount = channel.State.Allocations[0].RawAmount.Add(decimal.NewFromBigInt(ev.DeltaAllocations[0], 0))
			channel.State.Allocations[1].RawAmount = channel.State.Allocations[1].RawAmount.Add(decimal.NewFromBigInt(ev.DeltaAllocations[1], 0))
		}

		channel.RawAmount = decimal.NewFromBigInt(newRawAmount, 0)
		channel.UpdatedAt = time.Now()
		channel.State.Version++
		channel.ServerStateSignature = nil // Reset server signature
		channel.UserStateSignature = nil   // Reset user signature
		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("error saving channel in database: %w", err)
		}

		if len(ev.DeltaAllocations) == 0 {
			return nil
		}

		walletAddress := common.HexToAddress(channel.Wallet)
		channelAccountID := NewAccountID(channelID)
		walletAccountID := NewAccountID(channel.Wallet)
		resizeAmount := ev.DeltaAllocations[0] // Participant deposits or withdraws.
		if resizeAmount.Cmp(big.NewInt(0)) != 0 {
			asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
			if err != nil {
				return fmt.Errorf("failed to get asset: %w", err)
			}

			amount := rawToDecimal(resizeAmount, asset.Decimals)
			// Keep correct order of operation for deposits and withdrawals into the channel.
			if amount.IsPositive() || amount.IsZero() {
				// 1. Deposit into a channel account.
				ledger := GetWalletLedger(tx, walletAddress)
				if err := ledger.Record(channelAccountID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				// 2. Immediately transfer from the channel account into the unified account.
				if err := ledger.Record(channelAccountID, asset.Symbol, amount.Neg()); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				ledger = GetWalletLedger(tx, walletAddress)
				if err := ledger.Record(walletAccountID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for participant: %w", err)
				}
				_, err = RecordLedgerTransaction(tx, TransactionTypeDeposit, channelAccountID, walletAccountID, asset.Symbol, amount)
				if err != nil {
					return fmt.Errorf("failed to record transaction: %w", err)
				}
			} else {
				// 1. Withdraw from the unified account and immediately withdraw from the channel account.
				ledger := GetWalletLedger(tx, walletAddress)
				if err := ledger.Record(walletAccountID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for participant: %w", err)
				}
				if err := ledger.Record(channelAccountID, asset.Symbol, amount.Neg()); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				_, err = RecordLedgerTransaction(tx, TransactionTypeWithdrawal, walletAccountID, channelAccountID, asset.Symbol, amount)
				if err != nil {
					return fmt.Errorf("failed to record transaction: %w", err)
				}
				// 2. Withdraw from the channel account.
				ledger = GetWalletLedger(tx, walletAddress)
				if err := ledger.Record(channelAccountID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
			}
		}

		logger.Info("handled resized event", "channelId", channelID, "newAmount", channel.RawAmount)
		return nil
	})
	if errors.Is(err, ErrCustodyEventAlreadyProcessed) {
		return
	} else if err != nil {
		logger.Error("failed to resize channel", "channelId", channelID, "error", err)
		return
	}

	c.wsNotifier.Notify(
		NewBalanceNotification(channel.Wallet, c.db),
		NewChannelNotification(channel),
	)
}

func (c *Custody) handleClosed(logger Logger, ev *nitrolite.CustodyClosed) {
	logger = logger.With("event", "Closed")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "final", ev.FinalState)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		if err := c.saveContractEvent(tx, "closed", *ev, ev.Raw); err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("channel with ID %s not found", channelID)
			}
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
		if err != nil {
			return fmt.Errorf("failed to get asset: %w", err)
		}

		rawAmount := ev.FinalState.Allocations[0].Amount
		amount := rawToDecimal(rawAmount, asset.Decimals)

		walletAddress := common.HexToAddress(channel.Wallet)
		channelAccountID := NewAccountID(channelID)
		walletAccountID := NewAccountID(channel.Wallet)

		// Transfer from unified account into channel account and then withdraw immediately.
		ledger := GetWalletLedger(tx, walletAddress)
		channelAccountBalance, err := ledger.Balance(channelAccountID, asset.Symbol)
		if err != nil {
			return fmt.Errorf("error fetching channel balance: %w", err)
		}

		// Withdraw from unified balance if channel was not in joining state.
		if channelAccountBalance.IsZero() { // If channel balance is not zero, it means the channel was in joining state.
			if err := ledger.Record(walletAccountID, asset.Symbol, amount.Neg()); err != nil {
				return fmt.Errorf("error recording balance update for participant: %w", err)
			}
			if err := ledger.Record(channelAccountID, asset.Symbol, amount); err != nil {
				return fmt.Errorf("error recording balance update for wallet: %w", err)
			}
			_, err = RecordLedgerTransaction(tx, TransactionTypeWithdrawal, walletAccountID, channelAccountID, asset.Symbol, amount)
			if err != nil {
				return fmt.Errorf("failed to record transaction: %w", err)
			}
		}

		// 2. Withdraw from the channel account.
		if err := ledger.Record(channelAccountID, asset.Symbol, amount.Neg()); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}

		// Update channel state with final state
		if len(ev.FinalState.Allocations) == 2 {
			channel.State.Allocations = []Allocation{
				{
					Participant:  ev.FinalState.Allocations[0].Destination.Hex(),
					TokenAddress: ev.FinalState.Allocations[0].Token.Hex(),
					RawAmount:    decimal.NewFromBigInt(ev.FinalState.Allocations[0].Amount, 0),
				},
				{
					Participant:  ev.FinalState.Allocations[1].Destination.Hex(),
					TokenAddress: ev.FinalState.Allocations[1].Token.Hex(),
					RawAmount:    decimal.NewFromBigInt(ev.FinalState.Allocations[1].Amount, 0),
				},
			}
		}
		channel.State.Version = ev.FinalState.Version.Uint64()
		channel.State.Intent = StateIntent(ev.FinalState.Intent)
		channel.State.Data = string(ev.FinalState.Data)
		channel.ServerStateSignature = nil // Reset server signature
		channel.UserStateSignature = nil   // Reset user signature

		channel.Status = ChannelStatusClosed
		channel.RawAmount = decimal.Zero
		channel.UpdatedAt = time.Now()

		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}

		logger.Info("handled closed event", "channelId", channelID)
		return nil
	})
	if errors.Is(err, ErrCustodyEventAlreadyProcessed) {
		return
	} else if err != nil {
		logger.Error("failed to close channel", "channelId", channelID, "error", err)
		return
	}

	c.wsNotifier.Notify(
		NewBalanceNotification(channel.Wallet, c.db),
		NewChannelNotification(channel),
	)
}

// UpdateBalanceMetrics fetches the broker's account information from the smart contract and updates metrics
func (c *Custody) UpdateBalanceMetrics(ctx context.Context, assets []Asset, metrics *Metrics) {
	logger := LoggerFromContext(ctx)

	if metrics == nil {
		logger.Error("metrics not initialized for custody client", "network", c.chainID)
		return
	}

	callOpts := &bind.CallOpts{Context: ctx}
	brokerAddr := c.signer.GetAddress()

	var tokenAddrs []common.Address
	for _, asset := range assets {
		tokenAddrs = append(tokenAddrs, common.HexToAddress(asset.Token))
	}
	availInfo, err := c.custody.GetAccountsBalances(callOpts, []common.Address{brokerAddr}, tokenAddrs)
	if err != nil {
		logger.Error("failed to get batch account info", "network", c.chainID, "error", err)
		return
	}
	if len(availInfo) == 0 {
		logger.Warn("batch account info is empty", "network", c.chainID)
	} else if len(availInfo[0]) != len(assets) {
		logger.Warn("unexpected batch account info length", "network", c.chainID,
			"expected", len(assets), "got", len(availInfo[0]))
	}

	rawWalletBalances, err := c.balanceChecker.Balances(callOpts, []common.Address{brokerAddr}, tokenAddrs)
	if err != nil {
		logger.Error("failed to get wallet balances", "network", c.chainID, "error", err)
		return
	}
	if len(rawWalletBalances) != len(assets) {
		logger.Warn("unexpected wallet balances length", "network", c.chainID,
			"expected", len(assets), "got", len(rawWalletBalances))
	}

	// Get the native token balance
	rawNativeBalance, err := c.client.BalanceAt(ctx, brokerAddr, nil)
	if err != nil {
		logger.Error("failed to get native asset balance", "network", c.chainID, "error", err)
		return
	}
	rawWalletBalances = append(rawWalletBalances, rawNativeBalance)

	for i, asset := range assets {
		var available decimal.Decimal
		if len(availInfo) > 0 && i < len(availInfo[0]) {
			available = rawToDecimal(availInfo[0][i], asset.Decimals)

			metrics.BrokerBalanceAvailable.With(prometheus.Labels{
				"network": fmt.Sprintf("%d", c.chainID),
				"token":   asset.Token,
				"asset":   asset.Symbol,
			}).Set(available.InexactFloat64())
		}

		walletBalance := rawToDecimal(rawWalletBalances[i], asset.Decimals)
		metrics.BrokerWalletBalance.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
			"token":   asset.Token,
			"asset":   asset.Symbol,
		}).Set(walletBalance.InexactFloat64())

		logger.Debug("metrics updated", "network", c.chainID, "token", asset.Token, "contract_balance", available.String(), "wallet_balance", walletBalance.String())
	}

	openChannels, err := c.custody.GetOpenChannels(callOpts, []common.Address{brokerAddr})
	if err != nil {
		logger.Error("failed to get open channels", "network", c.chainID, "broker", brokerAddr, "error", err)
		return
	}
	if len(openChannels) == 0 {
		logger.Warn("no open channels found", "network", c.chainID, "broker", brokerAddr)
		return
	}
	count := len(openChannels[0])
	metrics.BrokerChannelCount.With(prometheus.Labels{"network": fmt.Sprintf("%d", c.chainID)}).Set(float64(count))
	logger.Debug("open channels metric updated", "network", c.chainID, "channels", count)
}

// saveContractEvent saves a contract event to the database if it has not been processed before.
// It returns ErrCustodyEventAlreadyProcessed if the event was already processed.
func (c *Custody) saveContractEvent(tx *gorm.DB, name string, event any, rawLog types.Log) error {
	alreadyProcessed, err := IsContractEventPresent(tx, c.chainID, rawLog.TxHash.Hex(), uint32(rawLog.Index))
	if err != nil {
		return err
	}

	if alreadyProcessed {
		c.logger.Debug("event already processed", "name", name, "txHash", rawLog.TxHash.Hex(), "logIndex", rawLog.Index)
		return ErrCustodyEventAlreadyProcessed
	}

	eventData, err := MarshalEvent(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data for %s: %w", name, err)
	}

	contractEvent := &ContractEvent{
		ContractAddress: c.custodyAddr.Hex(),
		ChainID:         c.chainID,
		Name:            name,
		BlockNumber:     rawLog.BlockNumber,
		TransactionHash: rawLog.TxHash.Hex(),
		LogIndex:        uint32(rawLog.Index),
		Data:            eventData,
		CreatedAt:       time.Now(),
	}

	return StoreContractEvent(tx, contractEvent)
}

// rawToDecimal converts a raw big.Int amount to a decimal.Decimal with the specified number of decimals.
func rawToDecimal(raw *big.Int, decimals uint8) decimal.Decimal {
	if raw == nil {
		return decimal.Zero
	}
	return decimal.NewFromBigInt(raw, -int32(decimals))
}
