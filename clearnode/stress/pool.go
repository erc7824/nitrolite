package stress

import (
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

// CreateClientPool opens up to n WebSocket connections to the clearnode.
// It tolerates individual connection failures and returns whatever connections
// succeeded. Returns an error only if zero connections could be established.
func CreateClientPool(wsURL, privateKey string, n int) ([]*sdk.Client, error) {
	ethMsgSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}
	stateSigner, err := core.NewChannelDefaultSigner(ethMsgSigner)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel signer: %w", err)
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	opts := []sdk.Option{
		sdk.WithErrorHandler(func(_ error) {}),
	}

	clients := make([]*sdk.Client, 0, n)
	var lastErr error
	failed := 0
	for i := 0; i < n; i++ {
		client, err := sdk.NewClient(wsURL, stateSigner, txSigner, opts...)
		if err != nil {
			lastErr = err
			failed++
			fmt.Printf("\r  Connections: %d/%d (failed: %d)  ", len(clients), n, failed)
			time.Sleep(50 * time.Millisecond)
			continue
		}
		clients = append(clients, client)
		if (i+1)%10 == 0 || i+1 == n {
			fmt.Printf("\r  Connections: %d/%d  ", len(clients), i+1)
		}
		// Pace connection attempts to avoid overwhelming the server
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Println()

	if len(clients) == 0 {
		return nil, fmt.Errorf("failed to open any connections: %w", lastErr)
	}

	if len(clients) < n {
		fmt.Printf("WARNING: Only %d/%d connections established\n", len(clients), n)
	}

	return clients, nil
}

// CloseClientPool closes all clients in the pool.
func CloseClientPool(clients []*sdk.Client) {
	for _, c := range clients {
		c.Close()
	}
}
