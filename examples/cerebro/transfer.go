package main

import (
	"fmt"

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
	balances, err := o.clearnode.GetLedgerBalances()
	if err != nil {
		fmt.Printf("Failed to get ledger balances: %s\n", err.Error())
		return
	}
	assetBalance := decimal.New(0, 0)
	for _, balance := range balances {
		if balance.Asset == assetSymbol {
			assetBalance = balance.Amount
			break
		}
	}
	fmt.Printf("Who do you want to transfer %s to?\n", assetSymbol)
	destinationTag := o.readExtraArg("user_tag")
	if destinationTag == "" {
		fmt.Println("User Tag cannot be empty.")
		return
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

	_, err = o.clearnode.Transfer(destinationTag, assetSymbol, amount)
	if err != nil {
		fmt.Printf("Transfer failed: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully transferred %s %s to %s.\n",
		amount.String(), assetSymbol, destinationTag)
}
