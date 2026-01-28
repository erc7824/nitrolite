package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/erc7824/nitrolite/pkg/core"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

func main() {
	// Default clearnode WebSocket URL
	// Adjust this to match your clearnode instance
	wsURL := "wss://clearnode-v1-rc.yellow.org/ws"

	// Create client with default settings
	client, err := sdk.NewClient(wsURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to clearnode!")
	fmt.Println()

	// You can also create a client with custom options
	// Uncomment to use custom configuration:
	/*
		client, err := sdk.NewClient(
			wsURL,
			sdk.WithHandshakeTimeout(10*time.Second),
			sdk.WithPingInterval(3*time.Second),
			sdk.WithErrorHandler(func(err error) {
				log.Printf("Connection error: %v", err)
			}),
		)
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()
	*/

	// Create a context for requests
	ctx := context.Background()

	// 1. Ping the server
	fmt.Println("Testing connection...")
	if err := client.Ping(ctx); err != nil {
		log.Fatalf("Ping failed: %v", err)
	}
	fmt.Println("✓ Ping successful")
	fmt.Println()

	// 2. Get node configuration
	fmt.Println("Fetching node configuration...")
	config, err := client.GetConfig(ctx)
	if err != nil {
		log.Fatalf("GetConfig failed: %v", err)
	}
	fmt.Printf("✓ Node Address: %s\n", config.NodeAddress)
	fmt.Printf("✓ Node Version: %s\n", config.NodeVersion)
	fmt.Printf("✓ Supported Networks: %d\n", len(config.Blockchains))
	fmt.Println()

	// 3. Get list of supported blockchains
	fmt.Println("Fetching supported blockchains...")
	blockchains, err := client.GetBlockchains(ctx)
	if err != nil {
		log.Fatalf("GetBlockchains failed: %v", err)
	}
	fmt.Println("✓ Supported Blockchains:")
	for _, bc := range blockchains {
		fmt.Printf("  • %s\n", bc.Name)
		fmt.Printf("    Chain ID: %d\n", bc.ID)
		fmt.Printf("    Contract: %s\n", bc.ContractAddress)
		fmt.Println()
	}

	// 4. Get supported assets
	fmt.Println("Fetching supported assets...")
	assets, err := client.GetAssets(ctx, nil)
	if err != nil {
		log.Fatalf("GetAssets failed: %v", err)
	}
	fmt.Printf("✓ Found %d assets\n", len(assets))
	for _, asset := range assets {
		fmt.Printf("  • %s (%s)\n", asset.Name, asset.Symbol)
		fmt.Printf("    Tokens: %d implementations\n", len(asset.Tokens))
		if len(asset.Tokens) > 0 {
			fmt.Printf("    Example: %s on chain %d\n", asset.Tokens[0].Address, asset.Tokens[0].BlockchainID)
		}
		fmt.Println()
	}

	// 5. Get user balances (example with a wallet address)
	// Note: Replace with actual wallet address to see real balances
	fmt.Println("Example: Getting user balances...")
	exampleWallet := "0x0000000000000000000000000000000000000000"
	balances, err := client.GetBalances(ctx, exampleWallet)
	if err != nil {
		// This might fail if the wallet doesn't exist, which is expected for this example
		fmt.Printf("  (GetBalances example: %v)\n", err)
	} else {
		fmt.Printf("✓ Found %d balances for wallet\n", len(balances))
		for _, balance := range balances {
			fmt.Printf("  • %s: %s\n", balance.Asset, balance.Balance.String())
		}
	}
	fmt.Println()

	// 6. Get user channels (example with a wallet address)
	fmt.Println("Example: Getting user channels...")
	channels, meta, err := client.GetChannels(ctx, exampleWallet, nil)
	if err != nil {
		// This might fail if the wallet doesn't exist, which is expected for this example
		fmt.Printf("  (GetChannels example: %v)\n", err)
	} else {
		fmt.Printf("✓ Found %d channels (page %d of %d, total: %d)\n",
			len(channels), meta.Page, meta.PageCount, meta.TotalCount)
		for _, channel := range channels {
			channelType := "unknown"
			if channel.Type == core.ChannelTypeHome {
				channelType = "home"
			} else if channel.Type == core.ChannelTypeEscrow {
				channelType = "escrow"
			}

			channelStatus := "unknown"
			switch channel.Status {
			case core.ChannelStatusVoid:
				channelStatus = "void"
			case core.ChannelStatusOpen:
				channelStatus = "open"
			case core.ChannelStatusChallenged:
				channelStatus = "challenged"
			case core.ChannelStatusClosed:
				channelStatus = "closed"
			}

			fmt.Printf("  • Channel %s (%s)\n", channel.ChannelID, channelType)
			fmt.Printf("    Status: %s, Chain: %d\n", channelStatus, channel.BlockchainID)
		}
	}
	fmt.Println()

	// 7. Get user transactions (example with a wallet address)
	fmt.Println("Example: Getting user transactions...")
	limit := uint32(5)
	txOpts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}
	txs, txMeta, err := client.GetTransactions(ctx, exampleWallet, txOpts)
	if err != nil {
		// This might fail if the wallet doesn't exist, which is expected for this example
		fmt.Printf("  (GetTransactions example: %v)\n", err)
	} else {
		fmt.Printf("✓ Found %d transactions (showing %d per page, total: %d)\n",
			len(txs), txMeta.PerPage, txMeta.TotalCount)
		for _, tx := range txs {
			fmt.Printf("  • %s: %s → %s\n", tx.TxType.String(), tx.FromAccount, tx.ToAccount)
			fmt.Printf("    Amount: %s %s, Created: %s\n",
				tx.Amount.String(), tx.Asset, tx.CreatedAt.Format(time.RFC3339))
		}
	}
	fmt.Println()

	// 8. Example with timeout context
	fmt.Println("Testing with timeout context...")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctxWithTimeout); err != nil {
		log.Printf("Ping with timeout failed: %v", err)
	} else {
		fmt.Println("✓ Ping with timeout successful")
	}
	fmt.Println()

	// 9. Monitor connection status (non-blocking)
	go func() {
		<-client.WaitCh()
		log.Println("Connection closed")
	}()

	fmt.Println("All operations completed successfully!")
	fmt.Println("\nNote: Some examples may show errors if used with an empty/example wallet address.")
	fmt.Println("Replace 'exampleWallet' with a real wallet address to see actual data.")
}
