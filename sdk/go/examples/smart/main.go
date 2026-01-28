package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
	"github.com/shopspring/decimal"
)

func main() {
	// Configuration - Set these via environment variables or hardcode for testing
	wsURL := getEnv("CLEARNODE_WS", "wss://clearnode.example.com/ws")
	privateKey := getEnv("PRIVATE_KEY", "") // Your private key (with 0x prefix)
	polygonRPC := getEnv("POLYGON_RPC", "https://polygon-amoy.g.alchemy.com/v2/YOUR_KEY")

	if privateKey == "" {
		log.Fatal("Please set PRIVATE_KEY environment variable")
	}

	fmt.Println("Clearnode Smart Client Example")
	fmt.Println("===============================")
	fmt.Println()

	// Create signer from private key
	signer, err := sign.NewEthereumSigner(privateKey)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	fmt.Printf("Signer address: %s\n", signer.PublicKey().Address().String())
	fmt.Println()

	// Create smart client with blockchain RPC configuration
	client, err := sdk.NewSmartClient(
		wsURL,
		signer,
		sdk.WithBlockchainRPC(80002, polygonRPC), // Polygon Amoy testnet
	)
	if err != nil {
		log.Fatalf("Failed to create smart client: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to clearnode!")
	fmt.Println()

	ctx := context.Background()

	// Example 1: Check initial balance
	fmt.Println("1. Checking initial balance...")
	balances, err := client.GetBalances(ctx, signer.PublicKey().Address().String())
	if err != nil {
		log.Printf("Warning: Could not get balances: %v", err)
	} else {
		fmt.Printf("Current balances:\n")
		for _, balance := range balances {
			fmt.Printf("  %s: %s\n", balance.Asset, balance.Balance.String())
		}
	}
	fmt.Println()

	// Example 2: Deposit to channel
	// This will automatically:
	// - Create a channel if it doesn't exist
	// - Or checkpoint to existing channel
	fmt.Println("2. Depositing to channel...")
	depositAmount := decimal.NewFromInt(100)
	txHash, err := client.Deposit(ctx, 80002, "usdc", depositAmount)
	if err != nil {
		log.Printf("Deposit failed: %v", err)
		log.Println("Make sure you have:")
		log.Println("  - USDC tokens in your wallet")
		log.Println("  - Approved the contract to spend USDC")
		log.Println("  - Enough MATIC for gas")
	} else {
		fmt.Printf("✓ Deposit successful!\n")
		fmt.Printf("  Transaction: %s\n", txHash)
		fmt.Printf("  Amount: %s USDC\n", depositAmount.String())
	}
	fmt.Println()

	// Example 3: Transfer to another wallet
	fmt.Println("3. Transferring to another wallet...")
	recipientWallet := "0x0000000000000000000000000000000000000001" // Replace with real address
	transferAmount := decimal.NewFromInt(50)
	txID, err := client.Transfer(ctx, recipientWallet, "usdc", transferAmount)
	if err != nil {
		log.Printf("Transfer failed: %v", err)
		log.Println("Make sure you have sufficient balance in your channel")
	} else {
		fmt.Printf("✓ Transfer successful!\n")
		fmt.Printf("  Transaction ID: %s\n", txID)
		fmt.Printf("  Recipient: %s\n", recipientWallet)
		fmt.Printf("  Amount: %s USDC\n", transferAmount.String())
	}
	fmt.Println()

	// Example 4: Withdraw from channel
	fmt.Println("4. Withdrawing from channel...")
	withdrawAmount := decimal.NewFromInt(25)
	txHash, err = client.Withdraw(ctx, 80002, "usdc", withdrawAmount)
	if err != nil {
		log.Printf("Withdraw failed: %v", err)
		log.Println("Make sure you have:")
		log.Println("  - Sufficient balance in your channel")
		log.Println("  - Enough MATIC for gas")
	} else {
		fmt.Printf("✓ Withdrawal successful!\n")
		fmt.Printf("  Transaction: %s\n", txHash)
		fmt.Printf("  Amount: %s USDC\n", withdrawAmount.String())
	}
	fmt.Println()

	// Example 5: Check final balance
	fmt.Println("5. Checking final balance...")
	balances, err = client.GetBalances(ctx, signer.PublicKey().Address().String())
	if err != nil {
		log.Printf("Warning: Could not get balances: %v", err)
	} else {
		fmt.Printf("Final balances:\n")
		for _, balance := range balances {
			fmt.Printf("  %s: %s\n", balance.Asset, balance.Balance.String())
		}
	}
	fmt.Println()

	// Example 6: View transaction history
	fmt.Println("6. Transaction history...")
	txs, meta, err := client.GetTransactions(ctx, signer.PublicKey().Address().String(), nil)
	if err != nil {
		log.Printf("Warning: Could not get transactions: %v", err)
	} else {
		fmt.Printf("Found %d transactions (total: %d)\n", len(txs), meta.TotalCount)
		for i, tx := range txs {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(txs)-5)
				break
			}
			fmt.Printf("  %s: %s %s (%s → %s)\n",
				tx.TxType.String(), tx.Amount.String(), tx.Asset,
				truncateAddr(tx.FromAccount), truncateAddr(tx.ToAccount))
		}
	}
	fmt.Println()

	fmt.Println("All operations completed!")
	fmt.Println()
	fmt.Println("Note: This example demonstrates the Smart Client API.")
	fmt.Println("The SDK automatically handles:")
	fmt.Println("  - Channel creation/detection")
	fmt.Println("  - State building and signing")
	fmt.Println("  - Blockchain interactions")
	fmt.Println("  - Error handling and validation")
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to truncate addresses for display
func truncateAddr(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	return addr[:6] + "..." + addr[len(addr)-4:]
}
