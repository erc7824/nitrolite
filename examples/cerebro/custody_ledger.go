package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/erc7824/nitrolite/examples/bridge/custody"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
)

func (o *Operator) handleDepositChain(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: deposit <chain_name> <token_symbol>")
		return
	}

	chainName := args[1]
	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		fmt.Printf("Chain %s is not supported by the broker.\n", chainName)
		return
	}

	assetSymbol := args[2]
	asset := network.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on chain %s.\n", assetSymbol, chainName)
		return
	}

	chainRPC, err := o.getChainRPC(network.ChainID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainName, err.Error())
		return
	}

	balance, err := o.custody.GetLedgerBalance(
		network.ChainID, chainRPC,
		network.CustodyAddress, o.config.Wallet.Address(), asset.Token)
	if err != nil {
		fmt.Printf("Failed to get balance for asset %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}

	decBalance := decimal.NewFromBigInt(balance, -int32(asset.Decimals))
	fmt.Printf("Your current ledger balance for asset %s on chain %s is: %s\n",
		assetSymbol, chainName, decBalance.String())

	fmt.Printf("How much %s do you want to deposit?\n", assetSymbol)
	amountStr := o.readExtraArg("deposit_amount")

	decAmount, err := decimal.NewFromString(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if decAmount.LessThanOrEqual(decimal.Zero) {
		fmt.Println("Amount must be greater than zero.")
		return
	}
	amount := decAmount.Shift(int32(asset.Decimals)).BigInt()

	tokenBalance, err := custody.GetTokenBalance(network.ChainID, chainRPC, asset.Token, o.config.Wallet.Address())
	if err != nil {
		fmt.Printf("Failed to get token balance for asset %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}

	if tokenBalance.Cmp(amount) < 0 {
		fmt.Printf("Not have enough %s to deposit. Available: %s, Required: %s\n",
			assetSymbol, decimal.NewFromBigInt(tokenBalance, -int32(asset.Decimals)), decAmount.String())
		return
	}

	if err := custody.ApproveAllowance(o.config.Wallet, network.ChainID, chainRPC,
		asset.Token, network.CustodyAddress, amount); err != nil {
		fmt.Printf("Failed to approve allowance for %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}

	if err := o.custody.Deposit(
		o.config.Wallet,
		network.ChainID, chainRPC,
		network.CustodyAddress, asset.Token,
		amount,
	); err != nil {
		fmt.Printf("Failed to deposit %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}

	fmt.Printf("Successfully deposited %s %s on chain %s.\n",
		decAmount.String(), assetSymbol, chainName)
}

func (o *Operator) handleWithdrawChain(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: withdraw <chain_name> <token_symbol>")
		return
	}

	chainName := args[1]
	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		fmt.Printf("Chain %s is not supported by the broker.\n", chainName)
		return
	}

	assetSymbol := args[2]
	asset := network.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on chain %s.\n", assetSymbol, chainName)
		return
	}

	chainRPC, err := o.getChainRPC(network.ChainID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainName, err.Error())
		return
	}

	balance, err := o.custody.GetLedgerBalance(
		network.ChainID, chainRPC,
		network.CustodyAddress, o.config.Wallet.Address(), asset.Token)
	if err != nil {
		fmt.Printf("Failed to get balance for asset %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}
	if balance == nil || balance.Cmp(new(big.Int)) <= 0 {
		fmt.Printf("You have no balance for asset %s on chain %s.\n", assetSymbol, chainName)
		return
	}

	decBalance := decimal.NewFromBigInt(balance, -int32(asset.Decimals))
	fmt.Printf("Your current ledger balance for asset %s on chain %s is: %s\n",
		assetSymbol, chainName, decBalance.String())

	fmt.Printf("How much %s do you want to withdraw?\n", assetSymbol)
	amountStr := o.readExtraArg("withdraw_amount")

	decAmount, err := decimal.NewFromString(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if decAmount.LessThanOrEqual(decimal.Zero) {
		fmt.Println("Amount must be greater than zero.")
		return
	}
	if decAmount.GreaterThan(decBalance) {
		fmt.Printf("You cannot withdraw more than your current balance of %s %s.\n",
			decBalance.String(), assetSymbol)
		return
	}

	amount := decAmount.Shift(int32(asset.Decimals)).BigInt()
	if err := o.custody.Withdraw(
		o.config.Wallet,
		network.ChainID, chainRPC,
		network.CustodyAddress, asset.Token,
		amount,
	); err != nil {
		fmt.Printf("Failed to withdraw %s on chain %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}

	fmt.Printf("Successfully withdrawn %s %s on chain %s.\n",
		decAmount.String(), assetSymbol, chainName)
}

func (o *Operator) getChainRPC(chainID uint32) (string, error) {
	chainRPCDTOs, err := o.store.GetChainRPCs(chainID)
	if err != nil {
		return "", err
	}
	if len(chainRPCDTOs) == 0 {
		return "", fmt.Errorf("no RPCs found for chain ID %d. Please import an RPC first", chainID)
	}

	for i := len(chainRPCDTOs) - 1; i >= 0; i-- {
		ethClient, err := ethclient.Dial(chainRPCDTOs[i].URL)
		if err != nil {
			continue
		}

		// Check if the chain ID matches
		chainIDFromRPC, err := ethClient.ChainID(context.Background())
		if err != nil {
			continue
		}
		if chainIDFromRPC.Uint64() != uint64(chainID) {
			continue
		}

		return chainRPCDTOs[i].URL, nil
	}

	return "", fmt.Errorf("no valid RPC found for chain ID %d", chainID)
}
