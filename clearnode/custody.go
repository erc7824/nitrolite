package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	custodyAbi        *abi.ABI
	balanceCheckerAbi *abi.ABI
)

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
	sendBalanceUpdate  func(string)
	sendChannelUpdate  func(Channel)
	logger             Logger
}

// NewCustody initializes the Ethereum client and custody contract wrapper.
func NewCustody(signer *Signer, db *gorm.DB, sendBalanceUpdate func(string), sendChannelUpdate func(Channel), infuraURL, custodyAddressStr, adjudicatorAddr, balanceCheckerAddr string, chain uint32, logger Logger) (*Custody, error) {
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
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
	auth.GasPrice = big.NewInt(30000000000) // 20 gwei.
	auth.GasLimit = uint64(3000000)

	custodyAddress := common.HexToAddress(custodyAddressStr)
	custody, err := nitrolite.NewCustody(custodyAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind custody contract: %w", err)
	}

	balanceChecker, err := nitrolite.NewBalanceChecker(common.HexToAddress(balanceCheckerAddr), client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind custody contract: %w", err)
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
		sendBalanceUpdate:  sendBalanceUpdate,
		sendChannelUpdate:  sendChannelUpdate,
		logger:             logger.NewSystem("custody").With("chainID", chainID.Int64()).With("custodyAddress", custodyAddressStr),
	}, nil
}

// ListenEvents initializes event listening for the custody contract
func (c *Custody) ListenEvents(ctx context.Context) {
	// TODO: store processed events in a database
	listenEvents(ctx, c.client, c.custodyAddr, c.chainID, 0, c.handleBlockChainEvent, c.logger)
}

// Join calls the join method on the custody contract
func (c *Custody) Join(channelID string, lastStateData []byte) (common.Hash, error) {
	// Convert string channelID to bytes32
	channelIDBytes := common.HexToHash(channelID)

	// The broker will always join as participant with index 1 (second participant)
	index := big.NewInt(1)

	sig, err := c.signer.NitroSign(lastStateData)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign data: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to suggest gas price: %w", err)
	}

	c.transactOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)
	// Call the join method on the custody contract
	tx, err := c.custody.Join(c.transactOpts, channelIDBytes, index, sig)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to join channel: %w", err)
	}

	return tx.Hash(), nil
}

// handleBlockChainEvent processes different event types received from the blockchain
func (c *Custody) handleBlockChainEvent(ctx context.Context, l types.Log) {
	ctx = SetContextLogger(ctx, c.logger)
	logger := LoggerFromContext(ctx)
	logger.Debug("received event", "blockNumber", l.BlockNumber, "txHahs", l.TxHash.String(), "logIndex", l.Index)

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
		ev, err := c.custody.ParseJoined(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		c.handleJoined(logger, ev)
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

	if len(ev.Channel.Participants) < 2 {
		logger.Warn("not enough participants in the channel")
		return
	}

	wallet := ev.Wallet.Hex()
	participantSigner := ev.Channel.Participants[0].Hex()
	nonce := ev.Channel.Nonce
	broker := ev.Channel.Participants[1]
	tokenAddress := ev.Initial.Allocations[0].Token.Hex()
	tokenAmount := ev.Initial.Allocations[0].Amount.Int64()
	adjudicator := ev.Channel.Adjudicator
	challenge := ev.Channel.Challenge

	brokerAmount := ev.Initial.Allocations[1].Amount.Int64()
	if brokerAmount != 0 {
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
		logger.Warn("participantB is not Broker", "actual", c.signer.GetAddress().Hex(), "expected", broker)
		return
	}

	// Check if there is already existing open channel with the broker
	existingOpenChannel, err := CheckExistingChannels(c.db, participantSigner, tokenAddress, c.chainID)
	if err != nil {
		logger.Error("error checking channels in database", "error", err)
		return
	}

	if existingOpenChannel != nil {
		logger.Error("an open channel with broker already exists", "existingChannelId", existingOpenChannel.ChannelID)
		return
	}

	err = AddSigner(c.db, wallet, participantSigner)
	if err != nil {
		logger.Error("error recording signer in database", "error", err)
		return
	}

	var ch Channel
	err = c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		eventData, err := MarshalCustodyCreated(*ev)
		if err != nil {
			return err
		}

		contractEvent := &ContractEvent{
			ID:              0,
			ContractAddress: c.custodyAddr.Hex(),
			ChainID:         c.chainID,
			Name:            "created",
			BlockNumber:     ev.Raw.BlockNumber,
			TransactionHash: ev.Raw.TxHash.Hex(),
			LogIndex:        uint32(ev.Raw.Index),
			Data:            eventData,
			CreatedAt:       time.Time{},
		}

		err = StoreContractEvent(tx, contractEvent)
		if err != nil {
			return err
		}

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
			uint64(tokenAmount),
		)
		if err != nil {
			return err
		}

		asset, err := GetAssetByToken(tx, tokenAddress, c.chainID)
		if err != nil {
			return fmt.Errorf("DB error fetching asset: %w", err)
		}

		tokenAmount := decimal.NewFromBigInt(big.NewInt(tokenAmount), -int32(asset.Decimals))
		ledger := GetWalletLedger(tx, wallet)
		if err := ledger.Record(channelID, asset.Symbol, tokenAmount); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}

		return nil
	})
	if err != nil {
		logger.Error("error creating channel in database", "error", err)
		return
	}

	encodedState, err := nitrolite.EncodeState(ev.ChannelId, nitrolite.IntentINITIALIZE, big.NewInt(0), ev.Initial.Data, ev.Initial.Allocations)
	if err != nil {
		logger.Error("error encoding state hash", "error", err)
		return
	}

	txHash, err := c.Join(channelID, encodedState)
	if err != nil {
		logger.Error("error joining channel", "error", err)
		return
	}

	c.sendChannelUpdate(ch)

	logger.Info("successfully initiated join for channel", "channelId", channelID, "txHash", txHash.Hex())
}

func (c *Custody) handleJoined(logger Logger, ev *nitrolite.CustodyJoined) {
	logger = logger.With("event", "Joined")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "index", ev.Index)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		eventData, err := MarshalCustodyJoined(*ev)
		if err != nil {
			return err
		}

		contractEvent := &ContractEvent{
			ID:              0,
			ContractAddress: c.custodyAddr.Hex(),
			ChainID:         c.chainID,
			Name:            "joined",
			BlockNumber:     ev.Raw.BlockNumber,
			TransactionHash: ev.Raw.TxHash.Hex(),
			LogIndex:        uint32(ev.Raw.Index),
			Data:            eventData,
			CreatedAt:       time.Time{},
		}

		err = StoreContractEvent(tx, contractEvent)
		if err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return fmt.Errorf("channel with ID %s not found", channelID)
			}
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		// Update the channel status to "open"
		channel.Status = ChannelStatusOpen
		channel.UpdatedAt = time.Now()
		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}

		asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
		if err != nil {
			return fmt.Errorf("DB error fetching asset: %w", err)
		}

		tokenAmount := decimal.NewFromBigInt(big.NewInt(int64(channel.Amount)), -int32(asset.Decimals))
		// Transfer from channel account into user's unified account.
		ledger := GetWalletLedger(tx, channel.Wallet)
		if err := ledger.Record(channelID, asset.Symbol, tokenAmount.Neg()); err != nil {
			log.Printf("[Joined] Error recording balance update for wallet: %v", err)
			return err
		}

		ledger = GetWalletLedger(tx, channel.Wallet)
		if err := ledger.Record(channel.Wallet, asset.Symbol, tokenAmount); err != nil {
			return fmt.Errorf("error recording balance update for wallet: %w", err)
		}

		return nil
	})
	if err != nil {
		logger.Error("failed to join channel", "channelId", channelID, "error", err)
		return
	}
	logger.Info("joined channel", "channelId", channelID)

	c.sendBalanceUpdate(channel.Wallet)
	c.sendChannelUpdate(channel)
}

func (c *Custody) handleChallenged(logger Logger, ev *nitrolite.CustodyChallenged) {
	logger = logger.With("event", "Challenged")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		eventData, err := MarshalCustodyChallenged(*ev)
		if err != nil {
			return err
		}

		contractEvent := &ContractEvent{
			ID:              0,
			ContractAddress: c.custodyAddr.Hex(),
			ChainID:         c.chainID,
			Name:            "challenged",
			BlockNumber:     ev.Raw.BlockNumber,
			TransactionHash: ev.Raw.TxHash.Hex(),
			LogIndex:        uint32(ev.Raw.Index),
			Data:            eventData,
			CreatedAt:       time.Time{},
		}

		err = StoreContractEvent(tx, contractEvent)
		if err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		channel.Status = ChannelStatusChallenged
		channel.UpdatedAt = time.Now()
		channel.Version = ev.State.Version.Uint64()
		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("error saving channel in database: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to update channel", "channelId", channelID, "error", err)
		return
	}
	logger.Info("challenged channel", "channelId", channelID)
	c.sendChannelUpdate(channel)
}

func (c *Custody) handleResized(logger Logger, ev *nitrolite.CustodyResized) {
	logger = logger.With("event", "Resized")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "deltaAllocations", ev.DeltaAllocations)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		eventData, err := MarshalCustodyResized(*ev)
		if err != nil {
			return err
		}

		contractEvent := &ContractEvent{
			ID:              0,
			ContractAddress: c.custodyAddr.Hex(),
			ChainID:         c.chainID,
			Name:            "resized",
			BlockNumber:     ev.Raw.BlockNumber,
			TransactionHash: ev.Raw.TxHash.Hex(),
			LogIndex:        uint32(ev.Raw.Index),
			Data:            eventData,
			CreatedAt:       time.Time{},
		}

		err = StoreContractEvent(tx, contractEvent)
		if err != nil {
			return err
		}

		result := tx.Where("channel_id = ?", channelID).First(&channel)
		if result.Error != nil {
			return fmt.Errorf("error finding channel: %w", result.Error)
		}

		newAmount := int64(channel.Amount)
		for _, change := range ev.DeltaAllocations {
			newAmount += change.Int64()
		}

		channel.Amount = uint64(newAmount)
		channel.UpdatedAt = time.Now()
		channel.Version++
		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("error saving channel in database: %w", err)
		}

		resizeAmount := ev.DeltaAllocations[0] // Participant deposits or withdraws.
		if resizeAmount.Cmp(big.NewInt(0)) != 0 {
			asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
			if err != nil {
				return fmt.Errorf("DB error fetching asset: %w", err)
			}

			amount := decimal.NewFromBigInt(resizeAmount, -int32(asset.Decimals))
			// Keep correct order of operation for deposits and withdrawals into the channel.
			if amount.IsPositive() || amount.IsZero() {
				// 1. Deposit into a channel account.
				ledger := GetWalletLedger(tx, channel.Wallet)
				if err := ledger.Record(channelID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				// 2. Immediately transfer from the channel account into the unified account.
				if err := ledger.Record(channelID, asset.Symbol, amount.Neg()); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				ledger = GetWalletLedger(tx, channel.Wallet)
				if err := ledger.Record(channel.Wallet, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for participant: %w", err)
				}
			} else {
				// 1. Withdraw from the unified account and immediately transfer into the unified account.
				ledger := GetWalletLedger(tx, channel.Wallet)
				if err := ledger.Record(channel.Wallet, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for participant: %w", err)
				}
				if err := ledger.Record(channelID, asset.Symbol, amount.Neg()); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
				// 2. Withdraw from the channel account.
				ledger = GetWalletLedger(tx, channel.Wallet)
				if err := ledger.Record(channelID, asset.Symbol, amount); err != nil {
					return fmt.Errorf("error recording balance update for wallet: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to resize channel", "channelId", channelID, "error", err)
		return
	}
	logger.Info("resized channel", "channelId", channelID, "newAmount", channel.Amount)

	c.sendBalanceUpdate(channel.Wallet)
	c.sendChannelUpdate(channel)
}

func (c *Custody) handleClosed(logger Logger, ev *nitrolite.CustodyClosed) {
	logger = logger.With("event", "Closed")
	channelID := common.Hash(ev.ChannelId).Hex()
	logger.Debug("parsed event", "channelId", channelID, "final", ev.FinalState)

	var channel Channel
	err := c.db.Transaction(func(tx *gorm.DB) error {
		// Save event in DB
		eventData, err := MarshalCustodyClosed(*ev)
		if err != nil {
			return err
		}

		contractEvent := &ContractEvent{
			ID:              0,
			ContractAddress: c.custodyAddr.Hex(),
			ChainID:         c.chainID,
			Name:            "closed",
			BlockNumber:     ev.Raw.BlockNumber,
			TransactionHash: ev.Raw.TxHash.Hex(),
			LogIndex:        uint32(ev.Raw.Index),
			Data:            eventData,
			CreatedAt:       time.Time{},
		}

		err = StoreContractEvent(tx, contractEvent)
		if err != nil {
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
			return fmt.Errorf("DB error fetching asset: %w", err)
		}

		finalAllocation := ev.FinalState.Allocations[0].Amount
		tokenAmount := decimal.NewFromBigInt(finalAllocation, -int32(asset.Decimals))

		// Transfer fron unified account into channel account and then withdraw immidiately.
		ledger := GetWalletLedger(tx, channel.Wallet)
		if err := ledger.Record(channel.Wallet, asset.Symbol, tokenAmount.Neg()); err != nil {
			return fmt.Errorf("error recording balance update for participant: %w", err)
		}
		ledger = GetWalletLedger(tx, channel.Wallet)
		if err := ledger.Record(channelID, asset.Symbol, tokenAmount); err != nil {
			log.Printf("[Closed] Error recording balance update for wallet: %v", err)
			return err
		}
		if err := ledger.Record(channelID, asset.Symbol, tokenAmount.Neg()); err != nil {
			log.Printf("[Closed] Error recording balance update for wallet: %v", err)
			return err
		}

		// Update the channel status to "closed"
		channel.Status = ChannelStatusClosed
		channel.Amount = 0
		channel.UpdatedAt = time.Now()
		channel.Version++
		if err := tx.Save(&channel).Error; err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}

		return nil
	})
	if err != nil {
		logger.Error("failed to close channel", "channelId", channelID, "error", err)
		return
	}
	logger.Info("closed channel", "channelId", channelID)

	c.sendBalanceUpdate(channel.Wallet)
	c.sendChannelUpdate(channel)
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

	walletBalances, err := c.balanceChecker.Balances(callOpts, []common.Address{brokerAddr}, tokenAddrs)
	if err != nil {
		logger.Error("failed to get wallet balances", "network", c.chainID, "error", err)
		return
	}
	if len(walletBalances) != len(assets) {
		logger.Warn("unexpected wallet balances length", "network", c.chainID,
			"expected", len(assets), "got", len(walletBalances))
	}

	// Get the native token balance
	nativeBalance, err := c.client.BalanceAt(ctx, brokerAddr, nil)
	if err != nil {
		logger.Error("failed to get native asset balance", "network", c.chainID, "error", err)
		return
	}
	walletBalances = append(walletBalances, nativeBalance)

	for i, asset := range assets {
		var available decimal.Decimal
		if len(availInfo) > 0 && i < len(availInfo[0]) {
			available = decimal.NewFromBigInt(availInfo[0][i], -int32(asset.Decimals))
			metrics.BrokerBalanceAvailable.With(prometheus.Labels{
				"network": fmt.Sprintf("%d", c.chainID),
				"token":   asset.Token,
				"asset":   asset.Symbol,
			}).Set(available.InexactFloat64())
		}

		walletBalance := decimal.NewFromBigInt(walletBalances[i], -int32(asset.Decimals))
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
