package main

import (
	"context"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// runReconcileCli is the entry point for the reconcile command line interface.
// Example: clearnode reconcile 1 1000000 2000000
func runReconcileCli(logger Logger) {
	logger = logger.NewSystem("reconcile")
	if len(os.Args) < 5 {
		logger.Fatal("Usage: clearnode reconcile <network> <block_start> <block_end>")
	}

	networkID, ok := new(big.Int).SetString(os.Args[2], 10)
	if !ok {
		logger.Fatal("Invalid network ID", "value", os.Args[2])
	}
	blockStart, ok := new(big.Int).SetString(os.Args[3], 10)
	if !ok {
		logger.Fatal("Invalid block start", "value", os.Args[3])
	}

	blockEnd, ok := new(big.Int).SetString(os.Args[4], 10)
	if !ok {
		logger.Fatal("Invalid block end value", "value", os.Args[4])
	}

	config, err := LoadConfig(logger)
	if err != nil {
		logger.Fatal("Failed to load configuration", "error", err)
	}

	network, ok := config.networks[uint32(networkID.Uint64())]
	if !ok {
		logger.Fatal("Network is not configured", "network", networkID.Uint64())
	}

	client, err := ethclient.Dial(network.InfuraURL)
	if err != nil {
		logger.Fatal("Failed to connect to Ethereum node", "error", err)
	}

	db, err := ConnectToDB(config.dbConf)
	if err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}

	signer, err := NewSigner(config.privateKeyHex)
	if err != nil {
		logger.Fatal("Failed to initialize signer", "error", err)
	}

	custody, err := NewCustody(
		signer,
		db,
		NewWSNotifier(func(userID, method string, params RPCDataParams) {}, logger),
		network.InfuraURL,
		network.CustodyAddress,
		network.AdjudicatorAddress,
		network.BalanceCHeckerAddress,
		network.ChainID,
		network.BlockStep,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize custody client", "error", err)
	}

	eventCh := make(chan types.Log, 1000)
	go func() {
		ReconcileBlockRange(
			client,
			common.HexToAddress(network.CustodyAddress),
			network.ChainID,
			blockEnd.Uint64(),
			network.BlockStep,
			blockStart.Uint64(),
			0,
			eventCh,
			logger,
		)
		close(eventCh)
	}()

	for event := range eventCh {
		custody.handleBlockChainEvent(context.Background(), event)
	}
}
