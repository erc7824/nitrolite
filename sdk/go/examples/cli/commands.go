package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
  wallet                        Show your wallet address
  import wallet                 Setup wallet (import existing or generate new)
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
  balances [wallet]             Get user balances (defaults to your wallet)
  transactions [wallet]         Get transaction history (defaults to your wallet)

LOW-LEVEL STATE MANAGEMENT (Base Client)
  state [wallet] <asset>        Get latest state (wallet defaults to yours)
  home-channel [wallet] <asset> Get home channel (wallet defaults to yours)
  escrow-channel <channel_id>   Get escrow channel by ID

LOW-LEVEL APP SESSIONS (Base Client)
  app-sessions                  List app sessions

OTHER
  exit                          Exit the CLI

EXAMPLES
  import wallet
  import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
  deposit 80002 usdc 100
  transfer 0x1234... usdc 50
  balances              # Uses your imported wallet
  balances 0x1234...    # Check another wallet
  state usdc            # Get your state for USDC
  chains`)
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
		signer, err := sign.NewEthereumRawSigner(privateKey)
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
	nodeConfig, err := o.client.GetConfig(ctx)
	if err == nil {
		fmt.Printf("\n📡 Node Info\n")
		fmt.Printf("   Address:   %s\n", nodeConfig.NodeAddress)
		fmt.Printf("   Version:   %s\n", nodeConfig.NodeVersion)
		fmt.Printf("   Chains:    %d\n", len(nodeConfig.Blockchains))
	}
}

// ============================================================================
// Wallet Commands
// ============================================================================

func (o *Operator) showWallet(ctx context.Context) {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		fmt.Println("❌ No wallet imported")
		fmt.Println("💡 Use 'import wallet' to setup your wallet")
		return
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		fmt.Printf("❌ Failed to get wallet address: %v\n", err)
		return
	}

	address := signer.PublicKey().Address().String()

	fmt.Println("🔑 Your Wallet")
	fmt.Println("==============")
	fmt.Printf("Address: %s\n", address)
}

// ============================================================================
// Import Commands
// ============================================================================

func (o *Operator) importWallet(ctx context.Context) {
	fmt.Println("🔑 Wallet Setup")
	fmt.Println("===============")
	fmt.Println()
	fmt.Println("Choose an option:")
	fmt.Println("  1. Import existing private key")
	fmt.Println("  2. Generate new wallet")
	fmt.Println()
	fmt.Print("Enter choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)

	var privateKey string
	var signer sign.Signer
	var err error

	switch choice {
	case "1":
		// Import existing key
		fmt.Println()
		fmt.Println("📥 Import Existing Wallet")
		fmt.Print("Enter private key (with or without 0x prefix): ")
		fmt.Scanln(&privateKey)

		privateKey = strings.TrimSpace(privateKey)
		if privateKey == "" {
			fmt.Println("❌ Private key cannot be empty")
			return
		}

		// Validate by creating signer
		signer, err = sign.NewEthereumRawSigner(privateKey)
		if err != nil {
			fmt.Printf("❌ Invalid private key: %v\n", err)
			return
		}

	case "2":
		// Generate new wallet
		fmt.Println()
		fmt.Println("🆕 Generate New Wallet")
		privateKey, err = generatePrivateKey()
		if err != nil {
			fmt.Printf("❌ Failed to generate private key: %v\n", err)
			return
		}

		signer, err = sign.NewEthereumRawSigner(privateKey)
		if err != nil {
			fmt.Printf("❌ Failed to create signer: %v\n", err)
			return
		}

		fmt.Println()
		fmt.Println("⚠️  IMPORTANT: Save your private key securely!")
		fmt.Println("=====================================")
		fmt.Printf("Private Key: %s\n", privateKey)
		fmt.Println("=====================================")
		fmt.Println()
		fmt.Print("Type 'I have saved my private key' to continue: ")

		var confirmation string
		fmt.Scanln(&confirmation)
		// Read the full line
		if confirmation == "" {
			fmt.Println("❌ You must confirm that you saved the private key")
			return
		}

	default:
		fmt.Println("❌ Invalid choice")
		return
	}

	// Save to storage
	if err := o.store.SetPrivateKey(privateKey); err != nil {
		fmt.Printf("❌ Failed to save private key: %v\n", err)
		return
	}

	fmt.Printf("✅ Wallet setup completed successfully\n")
	fmt.Printf("📍 Address: %s\n", signer.PublicKey().Address().String())

	if choice == "2" {
		fmt.Println()
		fmt.Println("💡 Tips:")
		fmt.Println("   - Store your private key in a secure location")
		fmt.Println("   - Never share your private key with anyone")
		fmt.Println("   - Consider using a hardware wallet for large amounts")
	}
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

	fmt.Printf("✅ RPC imported for chain %d\n", chainID)
}

// ============================================================================
// High-Level Operations (Smart Client)
// ============================================================================

func (o *Operator) deposit(ctx context.Context, chainIDStr, asset, amountStr string) {
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

	txHash, err := o.client.Deposit(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("❌ Deposit failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Deposit successful!\n")
	fmt.Printf("📝 Transaction: %s\n", txHash)
}

func (o *Operator) withdraw(ctx context.Context, chainIDStr, asset, amountStr string) {
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

	txHash, err := o.client.Withdraw(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("❌ Withdrawal failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Withdrawal successful!\n")
	fmt.Printf("📝 Transaction: %s\n", txHash)
}

func (o *Operator) transfer(ctx context.Context, recipient, asset, amountStr string) {
	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Printf("📤 Transferring %s %s to %s...\n", amount.String(), asset, recipient)

	txID, err := o.client.Transfer(ctx, recipient, asset, amount)
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
	err := o.client.Ping(ctx)
	if err != nil {
		fmt.Printf("❌ Failed: %v\n", err)
		return
	}
	fmt.Println("✅ Pong!")
}

func (o *Operator) nodeInfo(ctx context.Context) {
	config, err := o.client.GetConfig(ctx)
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
	chains, err := o.client.GetBlockchains(ctx)
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

	assets, err := o.client.GetAssets(ctx, chainID)
	if err != nil {
		fmt.Printf("❌ Failed to list assets: %v\n", err)
		return
	}

	if chainID != nil {
		fmt.Printf("Assets on Chain %d (%d)\n", *chainID, len(assets))
	} else {
		fmt.Printf("All Supported Assets (%d)\n", len(assets))
	}
	fmt.Println("==========================")

	for _, asset := range assets {
		fmt.Printf("• %s (%s)\n", asset.Name, asset.Symbol)
		fmt.Printf("  Decimals:  %d\n", asset.Decimals)
		fmt.Printf("  Tokens:    %d connected\n", len(asset.Tokens))

		// Show token details
		if len(asset.Tokens) > 0 {
			if chainID != nil {
				// When filtering by chain, show detailed info for each token
				for _, token := range asset.Tokens {
					fmt.Printf("    • Chain %d: %s\n", token.BlockchainID, token.Address)
					fmt.Printf("      Decimals: %d\n", token.Decimals)
				}
			} else {
				// When showing all assets, list chains with their token details
				for _, token := range asset.Tokens {
					fmt.Printf("    • Chain %d: %s (decimals: %d)\n", token.BlockchainID, token.Address, token.Decimals)
				}
			}
		}
		fmt.Println()
	}
}

// ============================================================================
// User Queries (Base Client)
// ============================================================================

func (o *Operator) getBalances(ctx context.Context, wallet string) {
	balances, err := o.client.GetBalances(ctx, wallet)
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

func (o *Operator) getHomeChannel(ctx context.Context, wallet, asset string) {
	channel, err := o.client.GetHomeChannel(ctx, wallet, asset)
	if err != nil {
		fmt.Printf("❌ Failed to get home channel: %v\n", err)
		return
	}

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

	fmt.Printf("📡 Home Channel for %s (%s)\n", wallet, asset)
	fmt.Println("==========================================")
	fmt.Printf("Channel ID:  %s\n", channel.ChannelID)
	fmt.Printf("Type:        %s\n", typeStr)
	fmt.Printf("Status:      %s\n", statusStr)
	fmt.Printf("Version:     %d\n", channel.StateVersion)
	fmt.Printf("Nonce:       %d\n", channel.Nonce)
	fmt.Printf("Chain ID:    %d\n", channel.BlockchainID)
	fmt.Printf("Token:       %s\n", channel.TokenAddress)
	fmt.Printf("Challenge:   %d seconds\n", channel.ChallengeDuration)
}

func (o *Operator) getEscrowChannel(ctx context.Context, escrowChannelID string) {
	channel, err := o.client.GetEscrowChannel(ctx, escrowChannelID)
	if err != nil {
		fmt.Printf("❌ Failed to get escrow channel: %v\n", err)
		return
	}

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

	fmt.Printf("📡 Escrow Channel %s\n", escrowChannelID)
	fmt.Println("==========================================")
	fmt.Printf("Channel ID:  %s\n", channel.ChannelID)
	fmt.Printf("User Wallet: %s\n", channel.UserWallet)
	fmt.Printf("Type:        %s\n", typeStr)
	fmt.Printf("Status:      %s\n", statusStr)
	fmt.Printf("Version:     %d\n", channel.StateVersion)
	fmt.Printf("Nonce:       %d\n", channel.Nonce)
	fmt.Printf("Chain ID:    %d\n", channel.BlockchainID)
	fmt.Printf("Token:       %s\n", channel.TokenAddress)
	fmt.Printf("Challenge:   %d seconds\n", channel.ChallengeDuration)
}

func (o *Operator) listTransactions(ctx context.Context, wallet string) {
	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	txs, meta, err := o.client.GetTransactions(ctx, wallet, opts)
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
	state, err := o.client.GetLatestState(ctx, wallet, asset, false)
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

// ============================================================================
// Low-Level App Sessions (Base Client)
// ============================================================================

func (o *Operator) listAppSessions(ctx context.Context) {
	sessions, meta, err := o.client.GetAppSessions(ctx, nil)
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
// Helper Methods
// ============================================================================

// generatePrivateKey generates a new Ethereum private key
func generatePrivateKey() (string, error) {
	// Generate new ECDSA private key
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Convert to hex string
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	return privateKeyHex, nil
}
