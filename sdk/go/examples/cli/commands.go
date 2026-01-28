package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	sdk "github.com/erc7824/nitrolite/sdk/go"
	"github.com/erc7824/nitrolite/pkg/sign"
	"github.com/shopspring/decimal"
)

// ============================================================================
// Help & Config
// ============================================================================

func (o *Operator) showHelp() {
	fmt.Println(`
🚀 Clearnode CLI - Developer Tool for Clearnode SDK
====================================================

SETUP COMMANDS
  help                          Show this help message
  config                        Show current configuration
  import wallet                 Import private key (interactive)
  import rpc <chain_id> <url>   Import blockchain RPC URL

HIGH-LEVEL OPERATIONS (Smart Client)
  deposit <chain_id> <asset> <amount>          Deposit to channel (auto-create if needed)
  withdraw <chain_id> <asset> <amount>         Withdraw from channel
  transfer <recipient> <asset> <amount>        Transfer to another wallet

NODE INFORMATION (Base Client)
  ping                          Ping the Clearnode server
  node info                     Get node configuration
  chains                        List supported blockchains
  assets [chain_id]             List supported assets (optionally filter by chain)

USER QUERIES (Base Client)
  balances <wallet>             Get user balances
  channels <wallet>             List user channels
  transactions <wallet>         Get transaction history

LOW-LEVEL STATE MANAGEMENT (Base Client)
  state <wallet> <asset>        Get latest state
  states <wallet> <asset>       Get state history

LOW-LEVEL APP SESSIONS (Base Client)
  app-sessions                  List app sessions

ADVANCED STATE MANAGEMENT
  submit-state                  Interactively build and submit a state transition
                                Supports: transfer, deposit, withdrawal, finalize, commit

OTHER
  exit                          Exit the CLI

EXAMPLES
  import wallet
  import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
  deposit 80002 usdc 100
  transfer 0x1234... usdc 50
  balances 0x1234...
  chains
`)
}

func (o *Operator) showConfig(ctx context.Context) {
	fmt.Println("📋 Current Configuration")
	fmt.Println("========================")

	// Private key status
	_, err := o.store.GetPrivateKey()
	if err != nil {
		fmt.Println("🔑 Wallet:     ❌ Not imported")
	} else {
		// Get signer to show address
		privateKey, _ := o.store.GetPrivateKey()
		signer, err := sign.NewEthereumSigner(privateKey)
		if err == nil {
			fmt.Printf("🔑 Wallet:     ✅ Imported (%s)\n", signer.PublicKey().Address().String())
		} else {
			fmt.Println("🔑 Wallet:     ✅ Imported")
		}
	}

	// RPC status
	rpcs, err := o.store.GetAllRPCs()
	if err != nil || len(rpcs) == 0 {
		fmt.Println("🌐 RPCs:       ❌ None configured")
	} else {
		fmt.Printf("🌐 RPCs:       ✅ %d configured\n", len(rpcs))
		for chainID, rpcURL := range rpcs {
			// Truncate URL for display
			displayURL := rpcURL
			if len(displayURL) > 50 {
				displayURL = displayURL[:47] + "..."
			}
			fmt.Printf("   - Chain %d: %s\n", chainID, displayURL)
		}
	}

	// Node info
	nodeConfig, err := o.baseClient.GetConfig(ctx)
	if err == nil {
		fmt.Printf("\n📡 Node Info\n")
		fmt.Printf("   Address:   %s\n", nodeConfig.NodeAddress)
		fmt.Printf("   Version:   %s\n", nodeConfig.NodeVersion)
		fmt.Printf("   Chains:    %d\n", len(nodeConfig.Blockchains))
	}
}

// ============================================================================
// Import Commands
// ============================================================================

func (o *Operator) importWallet(ctx context.Context) {
	fmt.Println("🔑 Import Wallet")
	fmt.Println("================")
	fmt.Print("Enter private key (with or without 0x prefix): ")

	var privateKey string
	fmt.Scanln(&privateKey)

	privateKey = strings.TrimSpace(privateKey)
	if privateKey == "" {
		fmt.Println("❌ Private key cannot be empty")
		return
	}

	// Validate by creating signer
	signer, err := sign.NewEthereumSigner(privateKey)
	if err != nil {
		fmt.Printf("❌ Invalid private key: %v\n", err)
		return
	}

	// Save to storage
	if err := o.store.SetPrivateKey(privateKey); err != nil {
		fmt.Printf("❌ Failed to save private key: %v\n", err)
		return
	}

	// Reset smart client to force recreation with new key
	if o.smartClient != nil {
		o.smartClient.Close()
		o.smartClient = nil
	}

	fmt.Printf("✅ Wallet imported successfully\n")
	fmt.Printf("📍 Address: %s\n", signer.PublicKey().Address().String())
}

func (o *Operator) importRPC(ctx context.Context, chainIDStr, rpcURL string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := o.store.SetRPC(chainID, rpcURL); err != nil {
		fmt.Printf("❌ Failed to save RPC: %v\n", err)
		return
	}

	// Reset smart client to force recreation with new RPC
	if o.smartClient != nil {
		o.smartClient.Close()
		o.smartClient = nil
	}

	fmt.Printf("✅ RPC imported for chain %d\n", chainID)
}

// ============================================================================
// High-Level Operations (Smart Client)
// ============================================================================

func (o *Operator) deposit(ctx context.Context, chainIDStr, asset, amountStr string) {
	if err := o.ensureSmartClient(ctx); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Printf("💰 Depositing %s %s on chain %d...\n", amount.String(), asset, chainID)

	txHash, err := o.smartClient.Deposit(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("❌ Deposit failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Deposit successful!\n")
	fmt.Printf("📝 Transaction: %s\n", txHash)
}

func (o *Operator) withdraw(ctx context.Context, chainIDStr, asset, amountStr string) {
	if err := o.ensureSmartClient(ctx); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Printf("💸 Withdrawing %s %s from chain %d...\n", amount.String(), asset, chainID)

	txHash, err := o.smartClient.Withdraw(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("❌ Withdrawal failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Withdrawal successful!\n")
	fmt.Printf("📝 Transaction: %s\n", txHash)
}

func (o *Operator) transfer(ctx context.Context, recipient, asset, amountStr string) {
	if err := o.ensureSmartClient(ctx); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Printf("📤 Transferring %s %s to %s...\n", amount.String(), asset, recipient)

	txID, err := o.smartClient.Transfer(ctx, recipient, asset, amount)
	if err != nil {
		fmt.Printf("❌ Transfer failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Transfer successful!\n")
	fmt.Printf("📝 Transaction ID: %s\n", txID)
}

// ============================================================================
// Node Information (Base Client)
// ============================================================================

func (o *Operator) ping(ctx context.Context) {
	fmt.Print("🏓 Pinging node... ")
	err := o.baseClient.Ping(ctx)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Pong!")
}

func (o *Operator) nodeInfo(ctx context.Context) {
	config, err := o.baseClient.GetConfig(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get node info: %v\n", err)
		return
	}

	fmt.Println("🖥️  Node Information")
	fmt.Println("===================")
	fmt.Printf("Address:   %s\n", config.NodeAddress)
	fmt.Printf("Version:   %s\n", config.NodeVersion)
	fmt.Printf("Chains:    %d\n", len(config.Blockchains))
	fmt.Println("\nSupported Blockchains:")
	for _, bc := range config.Blockchains {
		fmt.Printf("  • %s (ID: %d)\n", bc.Name, bc.ID)
		fmt.Printf("    Contract: %s\n", bc.ContractAddress)
	}
}

func (o *Operator) listChains(ctx context.Context) {
	chains, err := o.baseClient.GetBlockchains(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to list chains: %v\n", err)
		return
	}

	fmt.Printf("⛓️  Supported Blockchains (%d)\n", len(chains))
	fmt.Println("===========================")
	for _, chain := range chains {
		fmt.Printf("• %s\n", chain.Name)
		fmt.Printf("  Chain ID:  %d\n", chain.ID)
		fmt.Printf("  Contract:  %s\n", chain.ContractAddress)

		// Check if RPC is configured
		_, err := o.store.GetRPC(chain.ID)
		if err == nil {
			fmt.Printf("  RPC:       ✅ Configured\n")
		} else {
			fmt.Printf("  RPC:       ❌ Not configured\n")
		}
		fmt.Println()
	}
}

func (o *Operator) listAssets(ctx context.Context, chainIDStr string) {
	var chainID *uint64
	if chainIDStr != "" {
		parsed, err := o.parseChainID(chainIDStr)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		chainID = &parsed
	}

	assets, err := o.baseClient.GetAssets(ctx, chainID)
	if err != nil {
		fmt.Printf("❌ Failed to list assets: %v\n", err)
		return
	}

	if chainID != nil {
		fmt.Printf("💎 Assets on Chain %d (%d)\n", *chainID, len(assets))
	} else {
		fmt.Printf("💎 All Supported Assets (%d)\n", len(assets))
	}
	fmt.Println("==========================")

	for _, asset := range assets {
		fmt.Printf("• %s (%s)\n", asset.Name, asset.Symbol)
		fmt.Printf("  Decimals:  %d\n", asset.Decimals)
		fmt.Printf("  Tokens:    %d implementations\n", len(asset.Tokens))
		if chainID == nil && len(asset.Tokens) > 0 {
			fmt.Printf("  Chains:    ")
			chainIDs := make(map[uint64]bool)
			for _, token := range asset.Tokens {
				chainIDs[token.BlockchainID] = true
			}
			first := true
			for cid := range chainIDs {
				if !first {
					fmt.Print(", ")
				}
				fmt.Printf("%d", cid)
				first = false
			}
			fmt.Println()
		}
		fmt.Println()
	}
}

// ============================================================================
// User Queries (Base Client)
// ============================================================================

func (o *Operator) getBalances(ctx context.Context, wallet string) {
	balances, err := o.baseClient.GetBalances(ctx, wallet)
	if err != nil {
		fmt.Printf("❌ Failed to get balances: %v\n", err)
		return
	}

	fmt.Printf("💵 Balances for %s\n", wallet)
	fmt.Println("========================================")
	if len(balances) == 0 {
		fmt.Println("No balances found")
		return
	}

	for _, balance := range balances {
		fmt.Printf("• %s: %s\n", balance.Asset, balance.Balance.String())
	}
}

func (o *Operator) listChannels(ctx context.Context, wallet string) {
	channels, meta, err := o.baseClient.GetChannels(ctx, wallet, nil)
	if err != nil {
		fmt.Printf("❌ Failed to list channels: %v\n", err)
		return
	}

	fmt.Printf("📡 Channels for %s (Total: %d)\n", wallet, meta.TotalCount)
	fmt.Println("==============================================")
	if len(channels) == 0 {
		fmt.Println("No channels found")
		return
	}

	for _, channel := range channels {
		typeStr := "unknown"
		if channel.Type == core.ChannelTypeHome {
			typeStr = "Home"
		} else if channel.Type == core.ChannelTypeEscrow {
			typeStr = "Escrow"
		}

		statusStr := "unknown"
		switch channel.Status {
		case core.ChannelStatusVoid:
			statusStr = "Void"
		case core.ChannelStatusOpen:
			statusStr = "Open"
		case core.ChannelStatusChallenged:
			statusStr = "Challenged"
		case core.ChannelStatusClosed:
			statusStr = "Closed"
		}

		fmt.Printf("\n• Channel %s\n", channel.ChannelID)
		fmt.Printf("  Type:      %s\n", typeStr)
		fmt.Printf("  Status:    %s\n", statusStr)
		fmt.Printf("  Chain ID:  %d\n", channel.BlockchainID)
		fmt.Printf("  Token:     %s\n", channel.TokenAddress)
		fmt.Printf("  Challenge: %d seconds\n", channel.ChallengeDuration)
	}
}

func (o *Operator) listTransactions(ctx context.Context, wallet string) {
	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	txs, meta, err := o.baseClient.GetTransactions(ctx, wallet, opts)
	if err != nil {
		fmt.Printf("❌ Failed to list transactions: %v\n", err)
		return
	}

	fmt.Printf("📋 Recent Transactions for %s (Showing %d of %d)\n", wallet, len(txs), meta.TotalCount)
	fmt.Println("==================================================")
	if len(txs) == 0 {
		fmt.Println("No transactions found")
		return
	}

	for _, tx := range txs {
		fmt.Printf("\n• %s\n", tx.TxType.String())
		fmt.Printf("  From:      %s\n", tx.FromAccount)
		fmt.Printf("  To:        %s\n", tx.ToAccount)
		fmt.Printf("  Amount:    %s %s\n", tx.Amount.String(), tx.Asset)
		fmt.Printf("  Created:   %s\n", tx.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}

// ============================================================================
// Low-Level State Management (Base Client)
// ============================================================================

func (o *Operator) getLatestState(ctx context.Context, wallet, asset string) {
	state, err := o.baseClient.GetLatestState(ctx, wallet, asset, false)
	if err != nil {
		fmt.Printf("❌ Failed to get state: %v\n", err)
		return
	}

	fmt.Printf("📊 Latest State for %s (%s)\n", wallet, asset)
	fmt.Println("=====================================")
	fmt.Printf("Version:    %d\n", state.Version)
	fmt.Printf("Epoch:      %d\n", state.Epoch)
	fmt.Printf("State ID:   %s\n", state.ID)
	if state.HomeChannelID != nil {
		fmt.Printf("Channel:    %s\n", *state.HomeChannelID)
	}
	fmt.Printf("\nHome Ledger:\n")
	fmt.Printf("  Chain:      %d\n", state.HomeLedger.BlockchainID)
	fmt.Printf("  Token:      %s\n", state.HomeLedger.TokenAddress)
	fmt.Printf("  User Bal:   %s\n", state.HomeLedger.UserBalance.String())
	fmt.Printf("  Node Bal:   %s\n", state.HomeLedger.NodeBalance.String())
	fmt.Printf("\nTransitions: %d\n", len(state.Transitions))
	for i, t := range state.Transitions {
		fmt.Printf("  %d. %s (Amount: %s)\n", i+1, t.Type.String(), t.Amount.String())
	}
}

func (o *Operator) getStates(ctx context.Context, wallet, asset string) {
	limit := uint32(10)
	opts := &sdk.GetStatesOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	states, meta, err := o.baseClient.GetStates(ctx, wallet, asset, opts)
	if err != nil {
		fmt.Printf("❌ Failed to get states: %v\n", err)
		return
	}

	fmt.Printf("📚 State History for %s (%s) - Showing %d of %d\n", wallet, asset, len(states), meta.TotalCount)
	fmt.Println("=========================================================")
	if len(states) == 0 {
		fmt.Println("No states found")
		return
	}

	for _, state := range states {
		fmt.Printf("\n• Version %d (Epoch %d)\n", state.Version, state.Epoch)
		fmt.Printf("  State ID:      %s\n", state.ID)
		fmt.Printf("  User Balance:  %s\n", state.HomeLedger.UserBalance.String())
		fmt.Printf("  Node Balance:  %s\n", state.HomeLedger.NodeBalance.String())
		fmt.Printf("  Transitions:   %d\n", len(state.Transitions))
		if len(state.Transitions) > 0 {
			lastTransition := state.Transitions[len(state.Transitions)-1]
			fmt.Printf("  Last Action:   %s\n", lastTransition.Type.String())
		}
	}
}

// ============================================================================
// Low-Level App Sessions (Base Client)
// ============================================================================

func (o *Operator) listAppSessions(ctx context.Context) {
	sessions, meta, err := o.baseClient.GetAppSessions(ctx, nil)
	if err != nil {
		fmt.Printf("❌ Failed to list app sessions: %v\n", err)
		return
	}

	fmt.Printf("🎮 App Sessions (Total: %d)\n", meta.TotalCount)
	fmt.Println("============================")
	if len(sessions) == 0 {
		fmt.Println("No app sessions found")
		return
	}

	for _, session := range sessions {
		fmt.Printf("\n• Session %s\n", session.AppSessionID)
		fmt.Printf("  Version:      %d\n", session.Version)
		fmt.Printf("  Nonce:        %d\n", session.Nonce)
		fmt.Printf("  Quorum:       %d\n", session.Quorum)
		fmt.Printf("  Closed:       %v\n", session.IsClosed)
		fmt.Printf("  Participants: %d\n", len(session.Participants))
		fmt.Printf("  Allocations:  %d\n", len(session.Allocations))
	}
}

// ============================================================================
// Advanced State Management
// ============================================================================

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (o *Operator) interactiveSubmitState(ctx context.Context) {
	if err := o.ensureSmartClient(ctx); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Println("🔧 Interactive State Builder")
	fmt.Println("=============================")
	fmt.Println()

	// Step 1: Get wallet address
	var wallet string
	fmt.Print("Enter wallet address (or press Enter to use your own): ")
	fmt.Scanln(&wallet)
	wallet = strings.TrimSpace(wallet)
	if wallet == "" {
		wallet = o.smartClient.GetUserAddress()
		fmt.Printf("Using your wallet: %s\n", wallet)
	}

	// Step 2: Get asset
	var asset string
	fmt.Print("Enter asset symbol (e.g., usdc): ")
	fmt.Scanln(&asset)
	asset = strings.TrimSpace(strings.ToLower(asset))
	if asset == "" {
		fmt.Println("❌ Asset is required")
		return
	}

	// Step 3: Get latest state
	fmt.Printf("\n📊 Fetching latest state for %s (%s)...\n", wallet, asset)
	state, err := o.baseClient.GetLatestState(ctx, wallet, asset, false)
	if err != nil {
		fmt.Printf("❌ Failed to get latest state: %v\n", err)
		return
	}

	// Display current state
	fmt.Println("\n📋 Current State:")
	fmt.Printf("  Version:      %d\n", state.Version)
	fmt.Printf("  Epoch:        %d\n", state.Epoch)
	if state.HomeChannelID != nil {
		fmt.Printf("  Channel:      %s\n", *state.HomeChannelID)
	} else {
		fmt.Println("  Channel:      None (void state)")
	}
	fmt.Printf("  User Balance: %s\n", state.HomeLedger.UserBalance.String())
	fmt.Printf("  Node Balance: %s\n", state.HomeLedger.NodeBalance.String())
	fmt.Printf("  Transitions:  %d\n", len(state.Transitions))

	// Step 4: Choose transition type
	fmt.Println("\n🔀 Available Transition Types:")
	fmt.Println("  1. Transfer Send      - Send funds to another wallet")
	fmt.Println("  2. Home Deposit       - Deposit from blockchain")
	fmt.Println("  3. Home Withdrawal    - Withdraw to blockchain")
	fmt.Println("  4. Finalize           - Finalize the channel")
	fmt.Println("  5. Commit             - Commit funds to an app session")

	var choice string
	fmt.Print("\nSelect transition type (1-5): ")
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)

	// Create next state
	nextState := state.NextState()

	// Step 5: Apply transition based on choice
	switch choice {
	case "1": // Transfer Send
		fmt.Println("\n💸 Transfer Send")
		var recipient string
		fmt.Print("Recipient address: ")
		fmt.Scanln(&recipient)
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			fmt.Println("❌ Recipient is required")
			return
		}

		var amountStr string
		fmt.Print("Amount to transfer: ")
		fmt.Scanln(&amountStr)
		amount, err := o.parseAmount(amountStr)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		transition, err := nextState.ApplyTransferSendTransition(recipient, amount)
		if err != nil {
			fmt.Printf("❌ Failed to apply transfer transition: %v\n", err)
			return
		}
		fmt.Printf("✅ Added transfer transition (TxID: %s)\n", transition.TxID)

	case "2": // Home Deposit
		fmt.Println("\n💰 Home Deposit")
		if nextState.HomeChannelID == nil {
			fmt.Println("❌ Cannot deposit without a home channel. Create one first.")
			return
		}

		var amountStr string
		fmt.Print("Amount to deposit: ")
		fmt.Scanln(&amountStr)
		amount, err := o.parseAmount(amountStr)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		_, err = nextState.ApplyHomeDepositTransition(amount)
		if err != nil {
			fmt.Printf("❌ Failed to apply deposit transition: %v\n", err)
			return
		}
		fmt.Println("✅ Added home deposit transition")

	case "3": // Home Withdrawal
		fmt.Println("\n💸 Home Withdrawal")
		if nextState.HomeChannelID == nil {
			fmt.Println("❌ Cannot withdraw without a home channel")
			return
		}

		var amountStr string
		fmt.Print("Amount to withdraw: ")
		fmt.Scanln(&amountStr)
		amount, err := o.parseAmount(amountStr)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		_, err = nextState.ApplyHomeWithdrawalTransition(amount)
		if err != nil {
			fmt.Printf("❌ Failed to apply withdrawal transition: %v\n", err)
			return
		}
		fmt.Println("✅ Added home withdrawal transition")

	case "4": // Finalize
		fmt.Println("\n🏁 Finalize Channel")
		_, err := nextState.ApplyFinalizeTransition()
		if err != nil {
			fmt.Printf("❌ Failed to apply finalize transition: %v\n", err)
			return
		}
		fmt.Println("✅ Added finalize transition")

	case "5": // Commit (App Session)
		fmt.Println("\n🎮 Commit to App Session")
		if nextState.HomeChannelID == nil {
			fmt.Println("❌ Cannot commit without a home channel")
			return
		}

		var appSessionID string
		fmt.Print("App Session ID: ")
		fmt.Scanln(&appSessionID)
		appSessionID = strings.TrimSpace(appSessionID)
		if appSessionID == "" {
			fmt.Println("❌ App Session ID is required")
			return
		}

		var amountStr string
		fmt.Print("Amount to commit: ")
		fmt.Scanln(&amountStr)
		amount, err := o.parseAmount(amountStr)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}

		transition, err := nextState.ApplyCommitTransition(appSessionID, amount)
		if err != nil {
			fmt.Printf("❌ Failed to apply commit transition: %v\n", err)
			return
		}
		fmt.Printf("✅ Added commit transition (TxID: %s)\n", transition.TxID)

		// Now we need app state update data
		fmt.Println("\n📋 App State Update Information:")

		var versionStr string
		fmt.Print("App session version (new version after this deposit): ")
		fmt.Scanln(&versionStr)
		appVersion, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			fmt.Printf("❌ Invalid version: %v\n", err)
			return
		}

		// Ask for allocations
		fmt.Println("\n💰 Allocations (enter participant allocations, one at a time)")
		fmt.Println("   Format: <participant_address> <amount>")
		fmt.Println("   Type 'done' when finished")

		allocations := []app.AppAllocationV1{}
		for {
			var allocInput string
			fmt.Print("Allocation (or 'done'): ")
			fmt.Scanln(&allocInput)
			allocInput = strings.TrimSpace(allocInput)

			if strings.ToLower(allocInput) == "done" {
				break
			}

			parts := strings.Fields(allocInput)
			if len(parts) != 2 {
				fmt.Println("❌ Invalid format. Use: <participant_address> <amount>")
				continue
			}

			allocAmount, err := decimal.NewFromString(parts[1])
			if err != nil {
				fmt.Printf("❌ Invalid amount: %v\n", err)
				continue
			}

			allocations = append(allocations, app.AppAllocationV1{
				Participant: parts[0],
				Asset:       asset,
				Amount:      allocAmount,
			})
			fmt.Printf("✅ Added allocation: %s -> %s %s\n", parts[0], parts[1], asset)
		}

		if len(allocations) == 0 {
			fmt.Println("❌ At least one allocation is required")
			return
		}

		// Optional session data
		var sessionData string
		fmt.Print("\nSession data (JSON, or press Enter to skip): ")
		fmt.Scanln(&sessionData)
		sessionData = strings.TrimSpace(sessionData)

		// Ask for quorum signatures
		fmt.Println("\n✍️  Quorum Signatures (enter signatures from app session participants)")
		fmt.Println("   Type each signature and press Enter. Type 'done' when finished")

		quorumSigs := []string{}
		for {
			var sig string
			fmt.Print("Signature (or 'done'): ")
			fmt.Scanln(&sig)
			sig = strings.TrimSpace(sig)

			if strings.ToLower(sig) == "done" {
				break
			}

			if sig == "" {
				continue
			}

			quorumSigs = append(quorumSigs, sig)
			fmt.Printf("✅ Added signature: %s...\n", sig[:min(20, len(sig))])
		}

		if len(quorumSigs) == 0 {
			fmt.Println("❌ At least one quorum signature is required")
			return
		}

		// Store app state update data for later
		nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

		// Display next state
		fmt.Println("\n📊 Next State to Submit:")
		fmt.Printf("  State ID:     %s\n", nextState.ID)
		fmt.Printf("  Version:      %d\n", nextState.Version)
		fmt.Printf("  Epoch:        %d\n", nextState.Epoch)
		fmt.Printf("  User Balance: %s\n", nextState.HomeLedger.UserBalance.String())
		fmt.Printf("  Node Balance: %s\n", nextState.HomeLedger.NodeBalance.String())
		fmt.Printf("  Transitions:  %d (commit to %s)\n", len(nextState.Transitions), appSessionID)

		fmt.Println("\n📋 App State Update:")
		fmt.Printf("  App Session:  %s\n", appSessionID)
		fmt.Printf("  Version:      %d\n", appVersion)
		fmt.Printf("  Intent:       deposit\n")
		fmt.Printf("  Allocations:  %d\n", len(allocations))
		for i, alloc := range allocations {
			fmt.Printf("    %d. %s: %s %s\n", i+1, alloc.Participant, alloc.Amount.String(), alloc.Asset)
		}
		fmt.Printf("  Quorum Sigs:  %d\n", len(quorumSigs))

		// Confirm submission
		var confirm string
		fmt.Print("\n❓ Submit this commit state? (yes/no): ")
		fmt.Scanln(&confirm)
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		if confirm != "yes" && confirm != "y" {
			fmt.Println("❌ Submission cancelled")
			return
		}

		// Sign user state
		fmt.Println("\n✍️  Signing user state...")
		sig, err := o.smartClient.SignState(nextState)
		if err != nil {
			fmt.Printf("❌ Failed to sign state: %v\n", err)
			return
		}
		nextState.UserSig = &sig
		fmt.Printf("✅ User state signed: %s\n", sig[:min(20, len(sig))]+"...")

		// Build app state update
		appStateUpdate := app.AppStateUpdateV1{
			AppSessionID: appSessionID,
			Intent:       app.AppStateUpdateIntentDeposit,
			Version:      appVersion,
			Allocations:  allocations,
			SessionData:  sessionData,
		}

		// Submit deposit state
		fmt.Println("📤 Submitting commit state to node...")
		nodeSig, err := o.baseClient.SubmitDepositState(ctx, appStateUpdate, quorumSigs, *nextState)
		if err != nil {
			fmt.Printf("❌ Failed to submit deposit state: %v\n", err)
			return
		}
		nextState.NodeSig = &nodeSig

		fmt.Println("\n✅ Commit state submitted successfully!")
		fmt.Printf("📝 Node signature: %s\n", nodeSig[:min(20, len(nodeSig))]+"...")
		fmt.Printf("🎉 New state version: %d (Epoch: %d)\n", nextState.Version, nextState.Epoch)
		fmt.Printf("🎮 App session updated to version: %d\n", appVersion)
		return // Return early since we handled the submission

	default:
		fmt.Printf("❌ Invalid choice: %s\n", choice)
		return
	}

	// Calculate state ID
	nextState.ID = core.GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

	// Display next state
	fmt.Println("\n📊 Next State to Submit:")
	fmt.Printf("  State ID:     %s\n", nextState.ID)
	fmt.Printf("  Version:      %d\n", nextState.Version)
	fmt.Printf("  Epoch:        %d\n", nextState.Epoch)
	fmt.Printf("  User Balance: %s\n", nextState.HomeLedger.UserBalance.String())
	fmt.Printf("  Node Balance: %s\n", nextState.HomeLedger.NodeBalance.String())
	fmt.Printf("  Transitions:  %d\n", len(nextState.Transitions))

	// Step 6: Confirm submission
	var confirm string
	fmt.Print("\n❓ Submit this state to the node? (yes/no): ")
	fmt.Scanln(&confirm)
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "yes" && confirm != "y" {
		fmt.Println("❌ Submission cancelled")
		return
	}

	// Step 7: Sign and submit
	fmt.Println("\n✍️  Signing state...")
	sig, err := o.smartClient.SignState(nextState)
	if err != nil {
		fmt.Printf("❌ Failed to sign state: %v\n", err)
		return
	}
	nextState.UserSig = &sig
	fmt.Printf("✅ State signed: %s\n", sig[:20]+"...")

	fmt.Println("📤 Submitting state to node...")
	nodeSig, err := o.baseClient.SubmitState(ctx, *nextState)
	if err != nil {
		fmt.Printf("❌ Failed to submit state: %v\n", err)
		return
	}
	nextState.NodeSig = &nodeSig

	fmt.Println("\n✅ State submitted successfully!")
	fmt.Printf("📝 Node signature: %s\n", nodeSig[:20]+"...")
	fmt.Printf("🎉 New state version: %d (Epoch: %d)\n", nextState.Version, nextState.Epoch)
}
