package stress

import (
	"fmt"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

// CreateClientPool opens n WebSocket connections to the clearnode.
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
	for i := 0; i < n; i++ {
		client, err := sdk.NewClient(wsURL, stateSigner, txSigner, opts...)
		if err != nil {
			CloseClientPool(clients)
			return nil, fmt.Errorf("failed to open connection %d/%d: %w", i+1, n, err)
		}
		clients = append(clients, client)
	}

	return clients, nil
}

// CloseClientPool closes all clients in the pool.
func CloseClientPool(clients []*sdk.Client) {
	for _, c := range clients {
		c.Close()
	}
}
