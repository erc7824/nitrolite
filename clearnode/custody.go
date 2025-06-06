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
	custodyAbi *abi.ABI
)

// Custody implements the BlockchainClient interface using the Custody contract
type Custody struct {
	client             *ethclient.Client
	custody            *nitrolite.Custody
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
func NewCustody(signer *Signer, db *gorm.DB, sendBalanceUpdate func(string), sendChannelUpdate func(Channel), infuraURL, custodyAddressStr, adjudicatorAddr string, chain uint32, logger Logger) (*Custody, error) {
	custodyAddress := common.HexToAddress(custodyAddressStr)
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

	custody, err := nitrolite.NewCustody(custodyAddress, client)
	if err != nil {
		return nil, fmt.Errorf("failed to bind custody contract: %w", err)
	}

	return &Custody{
		client:             client,
		custody:            custody,
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
		logger := logger.With("event", "Created")
		ev, err := c.custody.ParseCreated(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
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

	case custodyAbi.Events["Joined"].ID:
		logger := logger.With("event", "Joined")
		ev, err := c.custody.ParseJoined(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		channelID := common.Hash(ev.ChannelId).Hex()
		logger.Debug("parsed event", "channelId", channelID, "index", ev.Index)

		var channel Channel
		err = c.db.Transaction(func(tx *gorm.DB) error {
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

	case custodyAbi.Events["Closed"].ID:
		ev, err := c.custody.ParseClosed(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		channelID := common.Hash(ev.ChannelId).Hex()
		logger.Debug("parsed event", "channelId", channelID, "final", ev.FinalState)

		var channel Channel
		err = c.db.Transaction(func(tx *gorm.DB) error {
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

	case custodyAbi.Events["Resized"].ID:
		ev, err := c.custody.ParseResized(l)
		if err != nil {
			logger.Warn("error parsing event", "error", err)
			return
		}
		channelID := common.Hash(ev.ChannelId).Hex()
		logger.Debug("parsed event", "channelId", channelID, "deltaAllocations", ev.DeltaAllocations)

		var channel Channel
		err = c.db.Transaction(func(tx *gorm.DB) error {
			channelID := common.BytesToHash(ev.ChannelId[:]).Hex()
			result := c.db.Where("channel_id = ?", channelID).First(&channel)
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
			if err := c.db.Save(&channel).Error; err != nil {
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
	default:
		logger.Warn("unknown event", "eventID", eventID.Hex())
	}
}

// UpdateBalanceMetrics fetches the broker's account information from the smart contract and updates metrics
func (c *Custody) UpdateBalanceMetrics(ctx context.Context, assets []Asset, metrics *Metrics) {
	logger := LoggerFromContext(ctx)

	if metrics == nil {
		logger.Error("metrics not initialized for custody client", "network", c.chainID)
		return
	}

	brokerAddr := c.signer.GetAddress()
	for _, asset := range assets {
		// Create a call opts with the provided context
		callOpts := &bind.CallOpts{
			Context: ctx,
		}

		logger.Debug("fetching account info", "network", c.chainID, "token", asset.Token, "asset", asset.Symbol, "broker", brokerAddr.Hex())
		// Call getAccountInfo on the custody contract
		tokenAddr := common.HexToAddress(asset.Token)
		info, err := c.custody.GetAccountInfo(callOpts, brokerAddr, tokenAddr)
		if err != nil {
			logger.Error("failed to get account info", "network", c.chainID, "token", asset.Token, "error", err)
			continue
		}

		availableBalance := decimal.NewFromBigInt(info.Available, -int32(asset.Decimals))

		metrics.BrokerBalanceAvailable.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
			"token":   asset.Token,
			"asset":   asset.Symbol,
		}).Set(availableBalance.InexactFloat64())

		metrics.BrokerChannelCount.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
		}).Set(float64(info.ChannelCount.Int64()))

		logger.Info("updated contract balance metrics", "network", c.chainID, "available", availableBalance.String(), "channels", info.ChannelCount.String())

		// Fetch broker wallet balances
		walletBalance := decimal.Zero

		if asset.Token == "0x0000000000000000000000000000000000000000" {
			walletBalanceRaw, err := c.client.BalanceAt(context.TODO(), brokerAddr, nil)
			if err != nil {
				logger.Error("failed to get base_asset balance", "network", c.chainID, "token", asset.Token, "error", err)
				continue
			}
			walletBalance = decimal.NewFromBigInt(walletBalanceRaw, -int32(asset.Decimals))

		} else {
			caller, err := nitrolite.NewErc20(tokenAddr, c.client)
			if err != nil {
				logger.Error("failed to initialize erc20 caller", "network", c.chainID, "token", asset.Token, "error", err)
				continue
			}

			walletBalanceRaw, err := caller.BalanceOf(callOpts, brokerAddr)
			if err != nil {
				logger.Error("failed to get erc20 balance", "network", c.chainID, "token", asset.Token, "error", err)
				continue
			}
			walletBalance = decimal.NewFromBigInt(walletBalanceRaw, -int32(asset.Decimals))
		}

		metrics.BrokerWalletBalance.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
			"token":   asset.Token,
			"asset":   asset.Symbol,
		}).Set(walletBalance.InexactFloat64())

		logger.Info("updated erc20 balance metrics", "network", c.chainID, "token", asset.Token, "asset", asset.Symbol, "balance", walletBalance.String())
	}
}
