package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// TODO: Move out from the memory store
// checkChainId connects to an RPC endpoint and verifies it returns the expected chain ID.
// This ensures the RPC URL points to the correct blockchain network.
// The function uses a 5-second timeout for the connection and chain ID query.
func checkChainId(blockchainRPC string, expectedChainID uint32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, blockchainRPC)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID from blockchain RPC: %w", err)
	}

	if uint32(chainID.Uint64()) != expectedChainID {
		return fmt.Errorf("unexpected chain ID from blockchain RPC: got %d, want %d", chainID.Uint64(), expectedChainID)
	}

	return nil
}
