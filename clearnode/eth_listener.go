package main

import (
	"context"
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

var logger = log.Logger("base-event-listener")

const (
	maxBackOffCount = 5
)

func init() {
	log.SetAllLoggers(log.LevelDebug)
	log.SetLogLevel("base-event-listener", "debug")

	var err error
	custodyAbi, err = nitrolite.CustodyMetaData.GetAbi()
	if err != nil {
		panic(err)
	}
}

type LogHandler func(l types.Log)

// listenEvents listens for blockchain events and processes them with the provided handler
func listenEvents(
	ctx context.Context,
	client bind.ContractBackend,
	contractAddress common.Address,
	chainID uint32,
	lastBlock uint64,
	handler LogHandler,
) {
	var backOffCount atomic.Uint64
	var currentCh chan types.Log
	var eventSubscription event.Subscription

	logger.Infow("starting listening events", "chainID", chainID, "contractAddress", contractAddress.String())
	for {
		if eventSubscription == nil {
			waitForBackOffTimeout(int(backOffCount.Load()))

			currentCh = make(chan types.Log, 100)

			watchFQ := ethereum.FilterQuery{
				Addresses: []common.Address{contractAddress},
			}
			eventSub, err := client.SubscribeFilterLogs(ctx, watchFQ, currentCh)
			if err != nil {
				logger.Errorw("failed to subscribe on events", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
				backOffCount.Add(1)
				continue
			}

			eventSubscription = eventSub
			logger.Infow("watching events", "chainID", chainID, "contractAddress", contractAddress.String())
			backOffCount.Store(0)
		}

		select {
		case eventLog := <-currentCh:
			lastBlock = eventLog.BlockNumber
			logger.Debugw("received new event", "chainID", chainID, "contractAddress", contractAddress.String(), "blockNumber", lastBlock, "logIndex", eventLog.Index)
			handler(eventLog)
		case err := <-eventSubscription.Err():
			if err != nil {
				logger.Errorw("event subscription error", "error", err, "chainID", chainID, "contractAddress", contractAddress.String())
				eventSubscription.Unsubscribe()
			} else {
				logger.Debugw("subscription closed, resubscribing", "chainID", chainID, "contractAddress", contractAddress.String())
			}

			eventSubscription = nil
		}
	}
}

// waitForBackOffTimeout implements exponential backoff between retries
func waitForBackOffTimeout(backOffCount int) {
	if backOffCount > maxBackOffCount {
		logger.Fatalw("back off limit reached, exiting", "backOffCollisionCount", backOffCount)
		return
	}

	if backOffCount > 0 {
		logger.Infow("backing off before subscribing on contract events", "backOffCollisionCount", backOffCount)
		<-time.After(time.Duration(2^backOffCount-1) * time.Second)
	}
}
