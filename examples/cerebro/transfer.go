package main

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func (o *Operator) handleTransfer(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: transfer <token_symbol>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	assetSymbol := args[1]
	getLedgerBalancesRes, err := o.clearnode.GetLedgerBalances(o.config.Wallet.PublicKey().Address().String())
	if err != nil {
		fmt.Printf("Failed to get ledger balances: %s\n", err.Error())
		return
	}
	assetBalance := decimal.New(0, 0)
	for _, balance := range getLedgerBalancesRes.LedgerBalances {
		if balance.Asset == assetSymbol {
			assetBalance = balance.Amount
			break
		}
	}

	fmt.Println("What destination format do you want to use?")
	destinationFormatSuggestions := []prompt.Suggest{
		{Text: "tag", Description: "Transfer by user tag"},
		{Text: "address", Description: "Transfer by user address"},
	}
	destinationFormat := o.readSelectionArg("destination", destinationFormatSuggestions)

	var transferByTag bool
	switch destinationFormat {
	case "tag":
		transferByTag = true
	case "address":
		transferByTag = false
	default:
		fmt.Printf("Unknown destination format: %s. Use 'tag' or 'address'.\n", destinationFormat)
		return
	}

	var destinationValue string
	if transferByTag {
		fmt.Println("What is the user tag you want to transfer to?")
		destinationValue = o.readExtraArg("user_tag")
		if destinationValue == "" {
			fmt.Println("User Tag cannot be empty.")
			return
		}
	} else {
		fmt.Println("What is the user address you want to transfer to?")
		destinationValue = o.readExtraArg("user_address")
		if !common.IsHexAddress(destinationValue) {
			fmt.Println("Invalid address format. Please provide a valid Ethereum address.")
			return
		}
	}

	fmt.Printf("Your current balance for asset %s is: %s\n",
		assetSymbol, assetBalance.String())
	fmt.Printf("How much %s do you want to transfer?\n", assetSymbol)
	amountStr := o.readExtraArg("amount")
	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	_, err = o.clearnode.Transfer(transferByTag, destinationValue, assetSymbol, amount)
	if err != nil {
		fmt.Printf("Transfer failed: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully transferred %s %s to %s.\n",
		amount.String(), assetSymbol, destinationValue)
}
