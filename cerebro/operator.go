package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

type Operator struct {
	wsURL  string
	store  *Storage
	client *sdk.Client
	exitCh chan struct{}
}

func NewOperator(wsURL string, store *Storage) (*Operator, error) {
	// Get private key to create full client
	privateKey, err := store.GetPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("no wallet imported (use 'import wallet' first): %w", err)
	}

	// Create signers
	stateSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create state signer: %w", err)
	}

	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx signer: %w", err)
	}

	// Get all RPCs
	rpcs, err := store.GetAllRPCs()
	if err != nil {
		// No RPCs configured is okay, some operations will fail but basic queries work
		rpcs = make(map[uint64]string)
	}

	// Create unified client with all configured RPCs
	opts := []sdk.Option{}
	for chainID, rpcURL := range rpcs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, rpcURL))
	}

	client, err := sdk.NewClient(wsURL, stateSigner, txSigner, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	op := &Operator{
		wsURL:  wsURL,
		store:  store,
		client: client,
		exitCh: make(chan struct{}),
	}

	// Monitor WebSocket connection - exit if connection is lost
	go func() {
		<-client.WaitCh()
		fmt.Println("\nWARNING: WebSocket connection lost. Exiting...")
		select {
		case <-op.exitCh:
			// Already closed
		default:
			close(op.exitCh)
		}
	}()

	return op, nil
}

func (o *Operator) Wait() <-chan struct{} {
	return o.exitCh
}

func (o *Operator) Complete(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(o.complete(d), d.GetWordBeforeCursor(), true)
}

func (o *Operator) complete(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	// First level commands
	if len(args) < 2 {
		return []prompt.Suggest{
			// Setup
			{Text: "help", Description: "Show help information"},
			{Text: "config", Description: "Show current configuration"},
			{Text: "wallet", Description: "Show wallet address"},
			{Text: "import", Description: "Import wallet or blockchain RPC"},
			{Text: "set-home-blockchain", Description: "Set home blockchain for channels"},

			// High-level operations
			{Text: "deposit", Description: "Deposit funds to channel"},
			{Text: "withdraw", Description: "Withdraw funds from channel"},
			{Text: "transfer", Description: "Transfer funds to another wallet"},

			// Node information
			{Text: "ping", Description: "Test node connection"},
			{Text: "node", Description: "Get node information"},
			{Text: "chains", Description: "List supported blockchains"},
			{Text: "assets", Description: "List supported assets"},

			// User queries
			{Text: "balances", Description: "Get user balances"},
			{Text: "transactions", Description: "Get transaction history"},

			// State management
			{Text: "state", Description: "Get latest state"},
			{Text: "home-channel", Description: "Get home channel"},
			{Text: "escrow-channel", Description: "Get escrow channel"},

			// App sessions (Base Client - Low-level)
			{Text: "app-sessions", Description: "List app sessions"},

			{Text: "exit", Description: "Exit the CLI"},
		}
	}

	// Second level
	if len(args) < 3 {
		switch args[0] {
		case "import":
			return []prompt.Suggest{
				{Text: "wallet", Description: "Import private key for signing"},
				{Text: "rpc", Description: "Import blockchain RPC URL"},
			}
		case "set-home-blockchain":
			return o.getAssetSuggestions()
		case "node":
			return []prompt.Suggest{
				{Text: "info", Description: "Get node configuration"},
			}
		}
	}

	// Third level - chain IDs for import rpc, or wallet addresses, or assets
	if len(args) < 4 {
		switch args[0] {
		case "import":
			if args[1] == "rpc" {
				return o.getChainSuggestions()
			}
		case "set-home-blockchain":
			return o.getChainSuggestions()
		case "deposit", "withdraw":
			return o.getChainSuggestions()
		case "balances", "transactions":
			// Suggest wallet address
			return o.getWalletSuggestion()
		case "transfer":
			// For transfer, third arg is recipient (no suggestion)
			return nil
		case "state":
			// If user already typed wallet (or we have 2 args), suggest assets
			// Otherwise suggest wallet address
			if len(args) == 3 {
				return o.getAssetSuggestions()
			}
			// Could be wallet or asset - suggest wallet first
			return o.getWalletSuggestion()
		case "home-channel":
			// If user already typed wallet (or we have 2 args), suggest assets
			// Otherwise suggest wallet address
			if len(args) == 3 {
				return o.getAssetSuggestions()
			}
			// Could be wallet or asset - suggest wallet first
			return o.getWalletSuggestion()
		case "escrow-channel":
			// Escrow channel ID (no suggestion)
			return nil
		case "assets":
			return o.getChainSuggestions()
		}
	}

	// Fourth level - assets or amounts
	if len(args) < 5 {
		switch args[0] {
		case "deposit", "withdraw", "transfer":
			return o.getAssetSuggestions()
		case "state", "home-channel":
			// Asset for state/home-channel commands (when wallet was explicitly provided)
			return o.getAssetSuggestions()
		}
	}

	return nil
}

func (o *Operator) Execute(s string) {
	args := strings.Split(strings.TrimSpace(s), " ")
	if s == "" || len(args) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch args[0] {
	case "help":
		o.showHelp()
	case "config":
		o.showConfig(ctx)
	case "wallet":
		o.showWallet(ctx)
	case "import":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: import <wallet|rpc> ...")
			return
		}
		switch args[1] {
		case "wallet":
			o.importWallet(ctx)
		case "rpc":
			if len(args) < 4 {
				fmt.Println("ERROR: Usage: import rpc <chain_id> <rpc_url>")
				return
			}
			o.importRPC(ctx, args[2], args[3])
		default:
			fmt.Printf("ERROR: Unknown import type: %s\n", args[1])
		}

	case "set-home-blockchain":
		if len(args) < 3 {
			fmt.Println("ERROR: Usage: set-home-blockchain <asset> <chain_id>")
			return
		}
		o.setHomeBlockchain(ctx, args[1], args[2])

	// High-level operations
	case "deposit":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: deposit <chain_id> <asset> <amount>")
			return
		}
		o.deposit(ctx, args[1], args[2], args[3])
	case "withdraw":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: withdraw <chain_id> <asset> <amount>")
			return
		}
		o.withdraw(ctx, args[1], args[2], args[3])
	case "transfer":
		if len(args) < 4 {
			fmt.Println("ERROR: Usage: transfer <recipient_address> <asset> <amount>")
			return
		}
		o.transfer(ctx, args[1], args[2], args[3])

	// Node information
	case "ping":
		o.ping(ctx)
	case "node":
		if len(args) < 2 || args[1] == "info" {
			o.nodeInfo(ctx)
		}
	case "chains":
		o.listChains(ctx)
	case "assets":
		chainID := ""
		if len(args) >= 2 {
			chainID = args[1]
		}
		o.listAssets(ctx, chainID)

	// User queries
	case "balances":
		wallet := ""
		if len(args) >= 2 {
			wallet = args[1]
		} else {
			// Auto-fill with imported wallet
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: balances <wallet_address>")
				fmt.Println("INFO: No wallet configured. Use 'import wallet' first or specify a wallet address.")
				return
			}
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		}
		o.getBalances(ctx, wallet)
	case "transactions":
		wallet := ""
		if len(args) >= 2 {
			wallet = args[1]
		} else {
			// Auto-fill with imported wallet
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: transactions <wallet_address>")
				fmt.Println("INFO: No wallet configured. Use 'import wallet' first or specify a wallet address.")
				return
			}
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		}
		o.listTransactions(ctx, wallet)

	// State management (low-level)
	case "state":
		wallet := ""
		asset := ""
		if len(args) >= 3 {
			wallet = args[1]
			asset = args[2]
		} else if len(args) == 2 {
			// Auto-fill wallet, user provided asset
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: state <wallet_address> <asset>")
				fmt.Println("INFO: No wallet configured. Use 'import wallet' first or specify a wallet address.")
				return
			}
			asset = args[1]
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		} else {
			fmt.Println("ERROR: Usage: state <wallet_address> <asset>")
			fmt.Println("INFO: Or: state <asset> (uses configured wallet)")
			return
		}
		o.getLatestState(ctx, wallet, asset)
	case "home-channel":
		wallet := ""
		asset := ""
		if len(args) >= 3 {
			wallet = args[1]
			asset = args[2]
		} else if len(args) == 2 {
			// Auto-fill wallet, user provided asset
			wallet = o.getImportedWalletAddress()
			if wallet == "" {
				fmt.Println("ERROR: Usage: home-channel <wallet_address> <asset>")
				fmt.Println("INFO: No wallet configured. Use 'import wallet' first or specify a wallet address.")
				return
			}
			asset = args[1]
			fmt.Printf("INFO: Using configured wallet: %s\n", wallet)
		} else {
			fmt.Println("ERROR: Usage: home-channel <wallet_address> <asset>")
			fmt.Println("INFO: Or: home-channel <asset> (uses configured wallet)")
			return
		}
		o.getHomeChannel(ctx, wallet, asset)
	case "escrow-channel":
		if len(args) < 2 {
			fmt.Println("ERROR: Usage: escrow-channel <escrow_channel_id>")
			return
		}
		o.getEscrowChannel(ctx, args[1])

	// App sessions
	case "app-sessions":
		o.listAppSessions(ctx)

	case "exit":
		fmt.Println("Exiting...")
		close(o.exitCh)
	default:
		fmt.Printf("ERROR: Unknown command: %s (type 'help' for available commands)\n", args[0])
	}
}

func (o *Operator) getChainSuggestions() []prompt.Suggest {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chains, err := o.client.GetBlockchains(ctx)
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, len(chains))
	for i, chain := range chains {
		suggestions[i] = prompt.Suggest{
			Text:        fmt.Sprintf("%d", chain.ID),
			Description: fmt.Sprintf("%s (ID: %d)", chain.Name, chain.ID),
		}
	}
	return suggestions
}

func (o *Operator) getAssetSuggestions() []prompt.Suggest {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	assets, err := o.client.GetAssets(ctx, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, len(assets))
	for i, asset := range assets {
		suggestions[i] = prompt.Suggest{
			Text:        asset.Symbol,
			Description: fmt.Sprintf("%s (%d tokens)", asset.Name, len(asset.Tokens)),
		}
	}
	return suggestions
}

func (o *Operator) getWalletSuggestion() []prompt.Suggest {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return nil
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return nil
	}

	address := signer.PublicKey().Address().String()

	return []prompt.Suggest{
		{
			Text:        address,
			Description: "Your imported wallet",
		},
	}
}

func (o *Operator) parseChainID(chainIDStr string) (uint64, error) {
	chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}
	return chainID, nil
}

func (o *Operator) parseAmount(amountStr string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid amount: %s", amountStr)
	}
	return amount, nil
}

func (o *Operator) getImportedWalletAddress() string {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return ""
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return ""
	}

	return signer.PublicKey().Address().String()
}
