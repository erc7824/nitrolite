package evm

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/erc7824/nitrolite/pkg/log"
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
	logger := log.NewNoopLogger()

	listener := NewListener(
		contractAddress,
		client,
		chainID.Uint64(),
		499, // blockStep
		logger,
		nil, // eventHandler not needed for this test
	)

	// Call reconcileBlockRange with appropriate parameters
	// currentBlock, lastBlock, lastIndex, historicalCh
	listener.reconcileBlockRange(31530000, 31527936, 0, historicalCh)
}
