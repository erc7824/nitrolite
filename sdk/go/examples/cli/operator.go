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
	wsURL      string
	store      *Storage
	baseClient *sdk.Client
	sdkClient  *sdk.SDKClient
	exitCh     chan struct{}
}

func NewOperator(wsURL string, store *Storage) (*Operator, error) {
	// Create base client
	baseClient, err := sdk.NewClient(wsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	op := &Operator{
		wsURL:      wsURL,
		store:      store,
		baseClient: baseClient,
		exitCh:     make(chan struct{}),
	}

	// Monitor WebSocket connection - exit if connection is lost
	go func() {
		<-baseClient.WaitCh()
		fmt.Println("\n⚠️  WebSocket connection lost. Exiting...")
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
			{Text: "wallet", Description: "🔑 Show your wallet address"},
			{Text: "import", Description: "Import wallet or blockchain RPC"},

			// High-level operations
			{Text: "deposit", Description: "💰 Deposit funds to channel"},
			{Text: "withdraw", Description: "💸 Withdraw funds from channel"},
			{Text: "transfer", Description: "📤 Transfer funds to another wallet"},

			// Node information
			{Text: "ping", Description: "🏓 Ping the Clearnode server"},
			{Text: "node", Description: "🖥️  Get node information"},
			{Text: "chains", Description: "⛓️  List supported blockchains"},
			{Text: "assets", Description: "💎 List supported assets"},

			// User queries
			{Text: "balances", Description: "💵 Get user balances"},
			{Text: "channels", Description: "📡 List user channels"},
			{Text: "transactions", Description: "📋 Get transaction history"},

			// State management
			{Text: "state", Description: "📊 Get latest state"},
			{Text: "states", Description: "📚 Get state history"},
			{Text: "submit-state", Description: "🔧 Build and submit state interactively"},

			// App sessions (Base Client - Low-level)
			{Text: "app-sessions", Description: "🎮 List app sessions"},

			{Text: "exit", Description: "👋 Exit the CLI"},
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
		case "node":
			return []prompt.Suggest{
				{Text: "info", Description: "Get node configuration"},
			}
		}
	}

	// Third level - chain IDs for import rpc
	if len(args) < 4 {
		switch args[0] {
		case "import":
			if args[1] == "rpc" {
				return o.getChainSuggestions()
			}
		case "deposit", "withdraw":
			return o.getChainSuggestions()
		case "transfer", "balances", "channels", "transactions", "state", "states":
			// These need wallet address
			return nil
		case "assets":
			return o.getChainSuggestions()
		}
	}

	// Fourth level - assets
	if len(args) < 5 {
		switch args[0] {
		case "deposit", "withdraw", "transfer":
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
			fmt.Println("❌ Usage: import <wallet|rpc> ...")
			return
		}
		switch args[1] {
		case "wallet":
			o.importWallet(ctx)
		case "rpc":
			if len(args) < 4 {
				fmt.Println("❌ Usage: import rpc <chain_id> <rpc_url>")
				return
			}
			o.importRPC(ctx, args[2], args[3])
		default:
			fmt.Printf("❌ Unknown import type: %s\n", args[1])
		}

	// High-level operations
	case "deposit":
		if len(args) < 4 {
			fmt.Println("❌ Usage: deposit <chain_id> <asset> <amount>")
			return
		}
		o.deposit(ctx, args[1], args[2], args[3])
	case "withdraw":
		if len(args) < 4 {
			fmt.Println("❌ Usage: withdraw <chain_id> <asset> <amount>")
			return
		}
		o.withdraw(ctx, args[1], args[2], args[3])
	case "transfer":
		if len(args) < 4 {
			fmt.Println("❌ Usage: transfer <recipient_address> <asset> <amount>")
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
		if len(args) < 2 {
			fmt.Println("❌ Usage: balances <wallet_address>")
			return
		}
		o.getBalances(ctx, args[1])
	case "channels":
		if len(args) < 2 {
			fmt.Println("❌ Usage: channels <wallet_address>")
			return
		}
		o.listChannels(ctx, args[1])
	case "transactions":
		if len(args) < 2 {
			fmt.Println("❌ Usage: transactions <wallet_address>")
			return
		}
		o.listTransactions(ctx, args[1])

	// State management (low-level)
	case "state":
		if len(args) < 3 {
			fmt.Println("❌ Usage: state <wallet_address> <asset>")
			return
		}
		o.getLatestState(ctx, args[1], args[2])
	case "states":
		if len(args) < 3 {
			fmt.Println("❌ Usage: states <wallet_address> <asset>")
			return
		}
		o.getStates(ctx, args[1], args[2])
	case "submit-state":
		o.interactiveSubmitState(ctx)

	// App sessions
	case "app-sessions":
		o.listAppSessions(ctx)

	case "exit":
		fmt.Println("👋 Exiting...")
		close(o.exitCh)
	default:
		fmt.Printf("❌ Unknown command: %s (type 'help' for available commands)\n", args[0])
	}
}

func (o *Operator) getChainSuggestions() []prompt.Suggest {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	chains, err := o.baseClient.GetBlockchains(ctx)
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

	assets, err := o.baseClient.GetAssets(ctx, nil)
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

func (o *Operator) ensureSmartClient(ctx context.Context) error {
	if o.sdkClient != nil {
		return nil
	}

	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		return fmt.Errorf("no wallet imported (use 'import wallet' first)")
	}

	// Create stateSigner
	txSigner, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}

	// Create signer
	stateSigner, err := sign.NewEthereumMsgSigner(privateKey)
	if err != nil {
		return fmt.Errorf("failed to create signer: %w", err)
	}

	// Get all RPCs
	rpcs, err := o.store.GetAllRPCs()
	if err != nil {
		return fmt.Errorf("failed to get RPCs: %w", err)
	}

	// Create smart client with all configured RPCs
	opts := []sdk.Option{}
	for chainID, rpcURL := range rpcs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, rpcURL))
	}

	sdkClient, err := sdk.NewSDKClient(o.wsURL, stateSigner, txSigner, opts...)
	if err != nil {
		return fmt.Errorf("failed to create smart client: %w", err)
	}

	o.sdkClient = sdkClient

	// Monitor SDK client connection - exit if connection is lost
	go func() {
		<-sdkClient.WaitCh()
		fmt.Println("\n⚠️  WebSocket connection lost. Exiting...")
		select {
		case <-o.exitCh:
			// Already closed
		default:
			close(o.exitCh)
		}
	}()

	return nil
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
