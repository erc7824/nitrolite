package main

import (
	"context"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ipfs/go-log/v2"
	"github.com/layer-3/clearsync/pkg/debounce"
	"github.com/pkg/errors"
)

var ethLogger = log.Logger("base-event-listener")

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
}

type LogHandler func(ctx context.Context, l types.Log)

// listenEvents listens for blockchain events and processes them with the provided handler
func listenEvents(
	ctx context.Context,
	client bind.ContractBackend,
	contractAddress common.Address,
	chainID uint32,
	blockStep uint64,
	lastBlock uint64,
	lastIndex uint32,
	handler LogHandler,
	logger Logger,
) {
	var backOffCount atomic.Uint64
	var historicalCh, currentCh chan types.Log
	var eventSubscription event.Subscription

	logger.Info("starting listening events", "chainID", chainID, "contractAddress", contractAddress.String())
	for {
		if eventSubscription == nil {
			waitForBackOffTimeout(logger, int(backOffCount.Load()), "event subscription")

			historicalCh = make(chan types.Log, 1)
			currentCh = make(chan types.Log, 100)

			if lastBlock == 0 {
				logger.Info("skipping historical logs fetching", "chainID", chainID, "contractAddress", contractAddress.String())
			} else {
				var header *types.Header
				var err error
				headerCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				err = debounce.Debounce(headerCtx, ethLogger, func(ctx context.Context) error {
					header, err = client.HeaderByNumber(ctx, nil)
					return err
				})
				cancel()
				if err != nil {
					logger.Error("failed to get latest block", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
					backOffCount.Add(1)
					continue
				}

				go ReconcileBlockRange(
					client,
					contractAddress,
					chainID,
					header.Number.Uint64(),
					blockStep,
					lastBlock,
					lastIndex,
					historicalCh,
					logger,
				)
			}

			watchFQ := ethereum.FilterQuery{
				Addresses: []common.Address{contractAddress},
			}
			eventSub, err := client.SubscribeFilterLogs(context.Background(), watchFQ, currentCh)
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
		case eventLog := <-historicalCh:
			logger.Debug("received new event", "chainID", chainID, "contractAddress", contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)
			handler(ctx, eventLog)
		case eventLog := <-currentCh:
			lastBlock = eventLog.BlockNumber
			logger.Debug("received new event", "chainID", chainID, "contractAddress", contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)
			handler(ctx, eventLog)
		case err := <-eventSubscription.Err():
			if err != nil {
				logger.Error("event subscription error", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
				eventSubscription.Unsubscribe()
				// NOTE: do not increment backOffCount here, as connection errors on continuous subscriptions are normal
			} else {
				logger.Debug("subscription closed, resubscribing", "chainID", chainID, "contractAddress", contractAddress.String())
			}

			eventSubscription = nil
		}
	}
}

func ReconcileBlockRange(
	client bind.ContractBackend,
	contractAddress common.Address,
	chainID uint32,
	currentBlock uint64,
	blockStep uint64,
	lastBlock uint64,
	lastIndex uint32,
	historicalCh chan types.Log,
	logger Logger,
) {
	var backOffCount atomic.Uint64
	startBlock := lastBlock
	endBlock := startBlock + blockStep

	for currentBlock > startBlock {
		waitForBackOffTimeout(logger, int(backOffCount.Load()), "reconcile block range")

		// We need to refetch events starting from last known block without adding 1 to it
		// because it's possible that block includes more than 1 event, and some may be still unprocessed.
		//
		// This will cause duplicate key error in logs, but it's completely fine.
		if endBlock > currentBlock {
			endBlock = currentBlock
		}

		fetchFQ := ethereum.FilterQuery{
			Addresses: []common.Address{contractAddress},
			FromBlock: new(big.Int).SetUint64(startBlock),
			ToBlock:   new(big.Int).SetUint64(endBlock),
			// Topics:    topics,
		}

		var logs []types.Log
		var err error
		logsCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		err = debounce.Debounce(logsCtx, ethLogger, func(ctx context.Context) error {
			logs, err = client.FilterLogs(ctx, fetchFQ)
			return err
		})
		cancel()
		if err != nil {
			newStartBlock, newEndBlock, extractErr := extractAdvisedBlockRange(err.Error())
			if extractErr != nil {
				logger.Error("failed to filter logs", "error", err, "chainID", chainID, "contractAddress", contractAddress.String(), "startBlock", startBlock, "endBlock", endBlock)
				backOffCount.Add(1)
				continue
			}
			startBlock, endBlock = newStartBlock, newEndBlock
			logger.Info("retrying with advised block range", "chainID", chainID, "contractAddress", contractAddress.String(), "startBlock", startBlock, "endBlock", endBlock)
			continue // retry with the advised block range
		}
		logger.Info("fetched historical logs", "chainID", chainID, "contractAddress", contractAddress.String(), "count", len(logs), "startBlock", startBlock, "endBlock", endBlock)

		for _, ethLog := range logs {
			// Filter out previously known events
			if ethLog.BlockNumber == lastBlock && ethLog.Index <= uint(lastIndex) {
				logger.Info("skipping previously known event", "chainID", chainID, "contractAddress", contractAddress.String(), "blockNumber", ethLog.BlockNumber, "logIndex", ethLog.Index)
				continue
			}

			historicalCh <- ethLog
		}

		startBlock = endBlock + 1
		endBlock += blockStep
	}
}

// extractAdvisedBlockRange extracts the advised block range from an error message
// when the error indicates too many query results.
// Assumed error format:
// "query returned more than 10000 results. Try with this block range [0x953260, 0x954ED4]."
func extractAdvisedBlockRange(msg string) (startBlock, endBlock uint64, err error) {
	if !strings.Contains(msg, "query returned more than 10000 results") {
		err = errors.New("error message doesn't contain advised block range")
		return
	}

	re := regexp.MustCompile(`\[0x([0-9a-fA-F]+), 0x([0-9a-fA-F]+)\]`)
	match := re.FindStringSubmatch(msg)
	if len(match) != 3 { // Match contains the whole match and two capture groups
		err = errors.New("failed to extract block range from error message")
		return
	}

	startBlock, err = strconv.ParseUint(match[1], 16, 64)
	if err != nil {
		err = errors.Wrap(err, "failed to parse block range from error message")
		return
	}
	endBlock, err = strconv.ParseUint(match[2], 16, 64)
	if err != nil {
		err = errors.Wrap(err, "failed to parse block range from error message")
		return
	}
	return
}

// waitForBackOffTimeout implements exponential backoff between retries
func waitForBackOffTimeout(logger Logger, backOffCount int, originator string) {
	if backOffCount > maxBackOffCount {
		logger.Fatal("back off limit reached, exiting", "originator", originator, "backOffCollisionCount", backOffCount)
		return
	}

	if backOffCount > 0 {
		logger.Info("backing off", "originator", originator, "backOffCollisionCount", backOffCount)
		<-time.After(time.Duration(2^backOffCount-1) * time.Second)
	}
}
