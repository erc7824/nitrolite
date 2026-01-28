package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/examples/cerebro/custody"
)

func (o *Operator) handleDepositCustody(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: deposit custody <chain_name> <token_symbol>")
		return
	}
	chainIDStr := args[2]

	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		fmt.Printf("Invalid chain ID: %s.\n", chainIDStr)
		return
	}

	blockchain := o.config.GetBlockchainByID(uint32(chainID.Uint64()))
	if blockchain == nil {
		fmt.Printf("Unknown chain: %s.\n", chainIDStr)
		return
	}

	assetSymbol := args[3]
	asset := blockchain.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on %s.\n", assetSymbol, chainID.String())
		return
	}

	chainRPC, err := o.getChainRPC(blockchain.ID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	walletAddress := common.HexToAddress(o.config.Wallet.PublicKey().Address().String())
	rawTokenBalance, err := custody.GetTokenBalance(blockchain.ID, chainRPC, asset.Token, walletAddress)
	if err != nil {
		fmt.Printf("Failed to get token balance for asset %s on chain %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}

	tokenBalance := decimal.NewFromBigInt(rawTokenBalance, -int32(asset.Decimals))
	fmt.Printf("Your current balance for asset %s on chain %s is: %s\n",
		assetSymbol, chainID.String(), fmtDec(tokenBalance))

	fmt.Printf("How much %s do you want to deposit?\n", assetSymbol)
	amountStr := o.readExtraArg("deposit_amount")

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		fmt.Printf("Amount must be greater than zero: %s\n", fmtDec(amount))
		return
	}
	rawAmount := amount.Shift(int32(asset.Decimals)).BigInt()

	if tokenBalance.Cmp(amount) < 0 {
		fmt.Printf("You do not have enough %s to deposit. Available: %s, Required: %s\n",
			assetSymbol, fmtDec(tokenBalance), amount)
		return
	}

	if err := custody.ApproveAllowance(o.config.Wallet, blockchain.ID, chainRPC,
		asset.Token, blockchain.CustodyAddress, rawAmount); err != nil {
		fmt.Printf("Failed to approve allowance for %s on chain %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}

	if err := o.custody.Deposit(
		o.config.Wallet,
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress, asset.Token,
		rawAmount,
	); err != nil {
		fmt.Printf("Failed to deposit %s on chain %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}

	fmt.Printf("Successfully deposited %s %s to custody on chain %s.\n",
		fmtDec(amount), assetSymbol, chainID.String())
}

func (o *Operator) handleWithdrawCustody(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: withdraw custody <chain_name> <token_symbol>")
		return
	}
	chainIDStr := args[2]

	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		fmt.Printf("Invalid chain ID: %s.\n", chainIDStr)
		return
	}

	blockchain := o.config.GetBlockchainByID(uint32(chainID.Uint64()))
	if blockchain == nil {
		fmt.Printf("Unknown chain: %s.\n", chainIDStr)
		return
	}

	assetSymbol := args[3]
	asset := blockchain.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on %s.\n", assetSymbol, chainID.String())
		return
	}

	chainRPC, err := o.getChainRPC(blockchain.ID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	walletAddress := common.HexToAddress(o.config.Wallet.PublicKey().Address().String())
	rawCustodyBalance, err := o.custody.GetLedgerBalance(
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress, walletAddress, asset.Token)
	if err != nil {
		fmt.Printf("Failed to get custody balance for asset %s on %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}
	if rawCustodyBalance == nil || rawCustodyBalance.Cmp(new(big.Int)) <= 0 {
		fmt.Printf("Insufficient custody balance for asset %s on %s.\n", assetSymbol, chainID.String())
		return
	}

	custodyBalance := decimal.NewFromBigInt(rawCustodyBalance, -int32(asset.Decimals))
	fmt.Printf("Your current custody balance for asset %s on %s is: %s\n",
		assetSymbol, chainID.String(), fmtDec(custodyBalance))

	fmt.Printf("How much %s do you want to withdraw?\n", assetSymbol)
	amountStr := o.readExtraArg("withdraw_amount")

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		fmt.Printf("Amount must be greater than zero: %s\n", fmtDec(amount))
		return
	}
	if amount.GreaterThan(custodyBalance) {
		fmt.Printf("You cannot withdraw more than your current custody balance of %s %s.\n",
			fmtDec(custodyBalance), assetSymbol)
		return
	}

	rawAmount := amount.Shift(int32(asset.Decimals)).BigInt()
	if err := o.custody.Withdraw(
		o.config.Wallet,
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress, asset.Token,
		rawAmount,
	); err != nil {
		fmt.Printf("Failed to withdraw %s on %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}

	fmt.Printf("Successfully withdrawn %s %s on %s.\n",
		fmtDec(amount), assetSymbol, chainID.String())
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
