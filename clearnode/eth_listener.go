package main

import (
	"context"
	"math/big"
	"sync/atomic"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ipfs/go-log/v2"
)

const (
	maxBackOffCount = 5
)

type Ethereum interface {
	CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error)
	CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	PendingCodeAt(ctx context.Context, contract common.Address) ([]byte, error)
	PendingCallContract(ctx context.Context, call ethereum.CallMsg) ([]byte, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, call ethereum.CallMsg) (gas uint64, err error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	FilterLogs(ctx context.Context, query ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(ctx context.Context, query ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	TransactionByHash(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
}

func init() {
	log.SetAllLoggers(log.LevelDebug)
	log.SetLogLevel("base-event-listener", "debug")

	var err error
	custodyAbi, err = nitrolite.CustodyMetaData.GetAbi()
	if err != nil {
		panic(err)
	}

	balanceCheckerAbi, err = nitrolite.BalanceCheckerMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

type LogHandler func(ctx context.Context, l types.Log)

// listenEvents listens for blockchain events and processes them with the provided handler
func listenEvents(
	ctx context.Context,
	client bind.ContractBackend,
	contractAddress common.Address,
	chainID uint32,
	lastBlock uint64,
	handler LogHandler,
	logger Logger,
) {
	var backOffCount atomic.Uint64
	var currentCh chan types.Log
	var eventSubscription event.Subscription

	logger.Info("starting listening events", "chainID", chainID, "contractAddress", contractAddress.String())
	for {
		if eventSubscription == nil {
			waitForBackOffTimeout(logger, int(backOffCount.Load()))

			currentCh = make(chan types.Log, 100)

			watchFQ := ethereum.FilterQuery{
				Addresses: []common.Address{contractAddress},
			}
			eventSub, err := client.SubscribeFilterLogs(ctx, watchFQ, currentCh)
			if err != nil {
				logger.Error("failed to subscribe on events", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
				backOffCount.Add(1)
				continue
			}

			eventSubscription = eventSub
			logger.Info("watching events", "chainID", chainID, "contractAddress", contractAddress.String())
			backOffCount.Store(0)
		}

		select {
		case eventLog := <-currentCh:
			lastBlock = eventLog.BlockNumber
			logger.Debug("received new event", "chainID", chainID, "contractAddress", contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)
			handler(ctx, eventLog)
		case err := <-eventSubscription.Err():
			if err != nil {
				logger.Error("event subscription error", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
				eventSubscription.Unsubscribe()
			} else {
				logger.Debug("subscription closed, resubscribing", "chainID", chainID, "contractAddress", contractAddress.String())
			}

			eventSubscription = nil
		}
	}
}

// waitForBackOffTimeout implements exponential backoff between retries
func waitForBackOffTimeout(logger Logger, backOffCount int) {
	if backOffCount > maxBackOffCount {
		logger.Fatal("back off limit reached, exiting", "backOffCollisionCount", backOffCount)
		return
	}

	if backOffCount > 0 {
		logger.Info("backing off before subscribing on contract events", "backOffCollisionCount", backOffCount)
		<-time.After(time.Duration(2^backOffCount-1) * time.Second)
	}
}
