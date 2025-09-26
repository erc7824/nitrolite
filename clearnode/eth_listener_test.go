package main

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func TestReconcileBlockRange(t *testing.T) {
	t.Skip("for manual testing only")

	blockchainRPC := "CHANGE_ME"
	contractAddress := common.HexToAddress("CHANGE_ME")

	client, err := ethclient.Dial(blockchainRPC)
	require.NoError(t, err, "Failed to connect to Ethereum client")

	chainID, err := client.ChainID(context.TODO())
	require.NoError(t, err, "Failed to get chain ID")

	historicalCh := make(chan types.Log, 100)
	logger := NewLoggerIPFS("test")
	ReconcileBlockRange(client, contractAddress, uint32(chainID.Uint64()), 31530000, 499, 31527936, 0, historicalCh, logger)
}
