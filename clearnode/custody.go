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
}

// NewCustody initializes the Ethereum client and custody contract wrapper.
func NewCustody(signer *Signer, db *gorm.DB, sendBalanceUpdate func(string), sendChannelUpdate func(Channel), infuraURL, custodyAddressStr, adjudicatorAddr string, chain uint32) (*Custody, error) {
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
	}, nil
}

// ListenEvents initializes event listening for the custody contract
func (c *Custody) ListenEvents(ctx context.Context) {
	// TODO: store processed events in a database
	listenEvents(ctx, c.client, c.custodyAddr, c.chainID, 0, c.handleBlockChainEvent)
}

// Join calls the join method on the custody contract
func (c *Custody) Join(channelID string, lastStateData []byte) error {
	// Convert string channelID to bytes32
	channelIDBytes := common.HexToHash(channelID)

	// The broker will always join as participant with index 1 (second participant)
	index := big.NewInt(1)

	sig, err := c.signer.NitroSign(lastStateData)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w", err)
	}

	c.transactOpts.GasPrice = gasPrice.Add(gasPrice, gasPrice)
	// Call the join method on the custody contract
	tx, err := c.custody.Join(c.transactOpts, channelIDBytes, index, sig)
	if err != nil {
		return fmt.Errorf("failed to join channel: %w", err)
	}
	log.Println("TxHash:", tx.Hash().Hex())

	return nil
}

// handleBlockChainEvent processes different event types received from the blockchain
func (c *Custody) handleBlockChainEvent(l types.Log) {
	log.Printf("Received event: %+v\n", l)

	eventID := l.Topics[0]
	switch eventID {
	case custodyAbi.Events["Created"].ID:
		ev, err := c.custody.ParseCreated(l)
		log.Printf("[Created] Event data: %+v\n", ev)
		if err != nil {
			log.Println("error parsing Created event:", err)
			return
		}

		if len(ev.Channel.Participants) < 2 {
			log.Println("[Created] Error: not enough participants in the channel")
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
			log.Printf("[Created] Error: broker amount should be 0, got %d\n", brokerAmount)
			return
		}

		if challenge < 3600 {
			log.Printf("[Created] Error: invalid challenge period: %d\n", challenge)
			return
		}

		if adjudicator != c.adjudicatorAddress {
			log.Printf("[Created] Error: unsupported adjudicator %s, expected %s\n", adjudicator.Hex(), c.adjudicatorAddress.Hex())
			return
		}

		// Check if channel was created with the broker.
		if broker != c.signer.GetAddress() {
			log.Printf("participantB %s is not Broker %s\n", broker, c.signer.GetAddress().Hex())
			return
		}

		// Check if there is already existing open channel with the broker
		existingOpenChannel, err := CheckExistingChannels(c.db, participantSigner, tokenAddress, c.chainID)
		if err != nil {
			log.Printf("[Created] Error checking channels in database: %v", err)
			return
		}

		if existingOpenChannel != nil {
			log.Printf("[Created] An open channel with broker already exists: %s", existingOpenChannel.ChannelID)
			return
		}

		err = AddSigner(c.db, wallet, participantSigner)
		if err != nil {
			log.Printf("[Created] Error recording signer in database: %v", err)
			return
		}

		channelID := common.BytesToHash(ev.ChannelId[:]).Hex()
		ch, err := CreateChannel(
			c.db,
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
			log.Printf("[ChannelCreated] Error creating/updating channel in database: %v", err)
			return
		}

		encodedState, err := nitrolite.EncodeState(ev.ChannelId, nitrolite.IntentINITIALIZE, big.NewInt(0), ev.Initial.Data, ev.Initial.Allocations)
		if err != nil {
			log.Printf("[ChannelCreated] Error encoding state hash: %v", err)
			return
		}

		if err := c.Join(channelID, encodedState); err != nil {
			log.Printf("[ChannelCreated] Error joining channel: %v", err)
			return
		}

		c.sendChannelUpdate(ch)

		log.Printf("[ChannelCreated] Successfully initiated join for channel %s on chain %d", channelID, c.chainID)

	case custodyAbi.Events["Joined"].ID:
		ev, err := c.custody.ParseJoined(l)
		if err != nil {
			log.Println("error parsing ChannelJoined event:", err)
			return
		}
		log.Printf("Joined event data: %+v\n", ev)

		var channel Channel
		channelID := common.BytesToHash(ev.ChannelId[:]).Hex()
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
			log.Printf("Joined channel with ID: %s", channelID)

			asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
			if err != nil {
				return fmt.Errorf("DB error fetching asset: %w", err)
			}

			if asset == nil {
				return fmt.Errorf("Asset not found in database for token: %s", channel.Token)
			}

			tokenAmount := decimal.NewFromBigInt(big.NewInt(int64(channel.Amount)), -int32(asset.Decimals))

			ledger := GetWalletLedger(tx, channel.Wallet)
			if err := ledger.Record(channel.Wallet, asset.Symbol, tokenAmount); err != nil {
				log.Printf("[Joined] Error recording balance update for wallet: %v", err)
				return err
			}

			return nil
		})
		if err != nil {
			log.Printf("[Joined] Error joining: %v", err)
			return
		}
		c.sendBalanceUpdate(channel.Wallet)
		c.sendChannelUpdate(channel)

	case custodyAbi.Events["Closed"].ID:
		ev, err := c.custody.ParseClosed(l)
		if err != nil {
			log.Println("error parsing ChannelClosed event:", err)
			return
		}
		log.Printf("Closed event data: %+v\n", ev)

		var channel Channel
		channelID := common.BytesToHash(ev.ChannelId[:]).Hex()
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

			if asset == nil {
				return fmt.Errorf("Asset not found in database for token: %s", channel.Token)
			}

			finalAllocation := ev.FinalState.Allocations[0].Amount
			tokenAmount := decimal.NewFromBigInt(finalAllocation, -int32(asset.Decimals))

			ledger := GetWalletLedger(tx, channel.Wallet)
			if err := ledger.Record(channel.Wallet, asset.Symbol, tokenAmount.Neg()); err != nil {
				log.Printf("[Closed] Error recording balance update for participant: %v", err)
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

			log.Printf("Closed channel with ID: %s", channelID)

			return nil
		})
		if err != nil {
			log.Printf("[Closed] Error closing channel: %v", err)
			return
		}
		c.sendBalanceUpdate(channel.Wallet)
		c.sendChannelUpdate(channel)

	case custodyAbi.Events["Resized"].ID:
		ev, err := c.custody.ParseResized(l)
		if err != nil {
			log.Println("error parsing Resized event:", err)
			return
		}
		log.Printf("Resized event data: %+v\n", ev)

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
				return fmt.Errorf("[Resized] Error saving channel in database: %w", err)
			}

			resizeAmount := ev.DeltaAllocations[0] // Participant deposits or withdraws.
			if resizeAmount.Cmp(big.NewInt(0)) != 0 {
				asset, err := GetAssetByToken(tx, channel.Token, c.chainID)
				if err != nil {
					return fmt.Errorf("DB error fetching asset: %w", err)
				}

				if asset == nil {
					return fmt.Errorf("Asset not found in database for token: %s", channel.Token)
				}

				amount := decimal.NewFromBigInt(resizeAmount, -int32(asset.Decimals))
				ledger := GetWalletLedger(tx, channel.Wallet)
				if err := ledger.Record(channel.Wallet, asset.Symbol, amount); err != nil {
					log.Printf("[Resized] Error recording balance update for participant: %v", err)
					return err
				}
			}

			return nil
		})

		if err != nil {
			log.Printf("[Resized] Error resizing channel: %v", err)
			return
		}

		c.sendBalanceUpdate(channel.Wallet)
		c.sendChannelUpdate(channel)
	default:
		log.Println("Unknown event ID:", eventID.Hex())
	}
}

// UpdateBalanceMetrics fetches the broker's account information from the smart contract and updates metrics
func (c *Custody) UpdateBalanceMetrics(ctx context.Context, tokens []common.Address, metrics *Metrics) {
	if metrics == nil {
		logger.Errorw("Metrics not initialized for custody client", "network", c.chainID)
		return
	}

	brokerAddr := c.signer.GetAddress()
	for _, token := range tokens {
		// Create a call opts with the provided context
		callOpts := &bind.CallOpts{
			Context: ctx,
		}

		logger.Infow("Fetching account info", "network", c.chainID, "token", token.Hex(), "broker", brokerAddr.Hex())
		// Call getAccountInfo on the custody contract
		info, err := c.custody.GetAccountInfo(callOpts, brokerAddr, token)
		if err != nil {
			logger.Errorw("Failed to get account info", "network", c.chainID, "token", token.Hex(), "error", err)
			continue
		}

		metrics.BrokerBalanceAvailable.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
			"token":   token.Hex(),
		}).Set(float64(info.Available.Int64()))

		metrics.BrokerChannelCount.With(prometheus.Labels{
			"network": fmt.Sprintf("%d", c.chainID),
			"token":   token.Hex(),
		}).Set(float64(info.ChannelCount.Int64()))

		logger.Infow("Updated contract balance metrics", "network", c.chainID, "token", token.Hex(), "available", info.Available.String(), "channels", info.ChannelCount.String())
	}
}
