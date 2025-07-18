package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/examples/cerebro/clearnet"
	"github.com/erc7824/nitrolite/examples/cerebro/custody"
)

func (o *Operator) handleOpenChannel(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: open channel <chain_name> <token_symbol>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	chainName := args[2]
	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		fmt.Printf("Chain %s is not supported by the broker.\n", chainName)
		return
	}

	assetSymbol := args[3]
	asset := network.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on %s.\n", assetSymbol, chainName)
		return
	}
	if asset.IsEnabled() {
		fmt.Printf("Channel is already opened for asset %s on %s.\n", assetSymbol, chainName)
		return
	}

	chainRPCDTOs, err := o.store.GetChainRPCs(network.ChainID)
	if err != nil {
		fmt.Printf("Failed to get RPCs for chain %s: %s\n", chainName, err.Error())
		return
	}
	if len(chainRPCDTOs) == 0 {
		fmt.Printf("No RPCs found for chain %s. Please import an RPC first.\n", chainName)
		return
	}

	chainRPCIndex := -1
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
		if chainIDFromRPC.Uint64() != uint64(network.ChainID) {
			continue
		}

		chainRPCIndex = i
		break
	}
	if chainRPCIndex < 0 {
		fmt.Printf("No valid RPC found for chain %s. Please import a valid RPC first.\n", chainName)
		return
	}

	fmt.Printf("Opening custody channel on %s...\n", chainName)

	if err := o.custody.OpenChannel(
		o.config.Wallet, o.config.Signer,
		network.ChainID, chainRPCDTOs[chainRPCIndex].URL,
		network.CustodyAddress,
		network.AdjudicatorAddress,
		o.config.BrokerAddress,
		asset.Token,
		0,
	); err != nil {
		fmt.Printf("Failed to open custody channel on %s: %s\n", chainName, err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPCDTOs[chainRPCIndex].URL); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully opened custody channel on %s!\n", chainName)
}

func (o *Operator) handleCloseChannel(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: close channel <chain_name> <token_symbol>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	chainName := args[2]
	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		fmt.Printf("Chain %s is not supported by the broker.\n", chainName)
		return
	}

	assetSymbol := args[3]
	asset := network.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on %s.\n", assetSymbol, chainName)
		return
	}
	if !asset.IsEnabled() {
		fmt.Printf("There are no opened channels for %s on %s.\n", assetSymbol, chainName)
		return
	}

	closureRes, err := o.clearnode.RequestChannelClosure(o.config.Wallet.Address(), asset.ChannelID)
	if err != nil {
		fmt.Printf("Failed to request channel closure for chain %s: %s\n", chainName, err.Error())
		return
	}

	chainRPC, err := o.getChainRPC(network.ChainID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainName, err.Error())
		return
	}

	fmt.Printf("Closing custody channel on %s...\n", chainName)

	allocations := make([]custody.Allocation, len(closureRes.FinalAllocations))
	for i, alloc := range closureRes.FinalAllocations {
		allocations[i] = convertAllocationRes(alloc)
	}

	brokerSig, err := convertSignatureRes(closureRes.Signature)
	if err != nil {
		fmt.Printf("Failed to convert broker signature: %s\n", err.Error())
		return
	}

	if err := o.custody.CloseChannel(
		o.config.Wallet, o.config.Signer,
		network.ChainID, chainRPC,
		network.CustodyAddress,
		common.HexToHash(closureRes.ChannelID),
		new(big.Int).SetUint64(closureRes.Version),
		allocations,
		brokerSig,
	); err != nil {
		fmt.Printf("Failed to close custody channel on %s: %s\n", chainName, err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPC); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	unlockedAmount := decimal.NewFromBigInt(allocations[0].Amount, -int32(asset.Decimals))
	fmt.Printf("Successfully closed custody channel on %s with unlocked %s %s!\n", chainName, unlockedAmount, strings.ToUpper(asset.Symbol))
}

func (o *Operator) handleResizeChannel(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: resize channel <chain_name> <token_symbol>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	chainName := args[2]
	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		fmt.Printf("Chain %s is not supported by the broker.\n", chainName)
		return
	}

	assetSymbol := args[3]
	asset := network.GetAssetBySymbol(assetSymbol)
	if asset == nil {
		fmt.Printf("Asset %s is not supported on %s.\n", assetSymbol, chainName)
		return
	}
	if !asset.IsEnabled() {
		fmt.Printf("There are no opened channels for %s on %s.\n", assetSymbol, chainName)
		return
	}

	chainRPC, err := o.getChainRPC(network.ChainID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainName, err.Error())
		return
	}

	custodyBalance, err := o.custody.GetLedgerBalance(
		network.ChainID, chainRPC,
		network.CustodyAddress, o.config.Wallet.Address(), asset.Token)
	if err != nil {
		fmt.Printf("Failed to get balance for asset %s on %s: %s\n", assetSymbol, chainName, err.Error())
		return
	}
	decCustodyBalance := decimal.NewFromBigInt(custodyBalance, -int32(asset.Decimals))

	channelBalance, ok := new(big.Int).SetString(asset.ChannelBalance, 10)
	if !ok {
		fmt.Printf("Failed to parse channel balance for asset %s on %s.\n", assetSymbol, chainName)
		return
	}
	decChannelBalance := decimal.NewFromBigInt(channelBalance, -int32(asset.Decimals))

	unifiedBalances, err := o.clearnode.GetLedgerBalances()
	if err != nil {
		fmt.Printf("Failed to get ledger balances: %s\n", err.Error())
		return
	}

	decUnifiedBalance := decimal.New(0, 0)
	for _, balance := range unifiedBalances {
		if balance.Asset == asset.Symbol {
			decUnifiedBalance = balance.Amount
			break
		}
	}

	fmt.Printf("Your current balances for asset %s:\n", assetSymbol)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Type", "Value"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"On Custody Ledger", decCustodyBalance.String()})
	t.AppendRow(table.Row{"On Channel", decChannelBalance.String()})
	t.AppendRow(table.Row{"Unified On Clearnode", decUnifiedBalance.String()})
	t.Render()

	fmt.Printf("How much %s do you want to resize (+)into/(-)out channel?", assetSymbol)
	resizeAmountStr := o.readExtraArg("resize_amount")

	decResizeAmount, err := decimal.NewFromString(resizeAmountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if decResizeAmount.GreaterThan(decCustodyBalance) {
		fmt.Printf("You cannot resize more than your current balance of %s %s.\n",
			decCustodyBalance.String(), assetSymbol)
		return
	}
	resizeAmount := decResizeAmount.Shift(int32(asset.Decimals))

	fmt.Printf("How much %s do you want to allocate to channel?\n", assetSymbol)
	fmt.Println("That's the amount moved from unified balance to channel.")
	allocateAmountStr := o.readExtraArg("allocate_amount")

	decAllocateAmount, err := decimal.NewFromString(allocateAmountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if decAllocateAmount.GreaterThan(decUnifiedBalance) {
		fmt.Printf("You cannot allocate more than your current unified balance of %s %s.\n",
			decUnifiedBalance.String(), assetSymbol)
		return
	}
	allocateAmount := decAllocateAmount.Shift(int32(asset.Decimals))

	if newChannelAmount := new(big.Int).Add(new(big.Int).Add(channelBalance, allocateAmount.BigInt()), resizeAmount.BigInt()); newChannelAmount.Cmp(big.NewInt(0)) < 0 {
		fmt.Printf("New channel amount must not be negative after resize: %s\n", newChannelAmount)
		return
	}

	resizeRes, err := o.clearnode.RequestChannelResize(o.config.Wallet.Address(), asset.ChannelID, allocateAmount, resizeAmount)
	if err != nil {
		fmt.Printf("Failed to request channel resize on %s: %s\n", chainName, err.Error())
		return
	}

	fmt.Printf("Resizing custody channel on %s...\n", chainName)

	stateData, err := hexutil.Decode(resizeRes.StateData)
	if err != nil {
		fmt.Printf("Failed to decode state data: %s\n", err.Error())
		return
	}

	allocations := make([]custody.Allocation, len(resizeRes.Allocations))
	for i, alloc := range resizeRes.Allocations {
		allocations[i] = convertAllocationRes(alloc)
	}

	brokerSig, err := convertSignatureRes(resizeRes.Signature)
	if err != nil {
		fmt.Printf("Failed to convert broker signature: %s\n", err.Error())
		return
	}

	if err := o.custody.Resize(
		o.config.Wallet, o.config.Signer,
		network.ChainID, chainRPC,
		network.CustodyAddress,
		common.HexToHash(resizeRes.ChannelID),
		new(big.Int).SetUint64(resizeRes.Version),
		stateData,
		allocations,
		brokerSig,
	); err != nil {
		fmt.Printf("Failed to resize custody channel on %s: %s\n", chainName, err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPC); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully resized custody channel on %s!\n", chainName)
}

func convertAllocationRes(a clearnet.AllocationRes) custody.Allocation {
	return custody.Allocation{
		Destination: common.HexToAddress(a.Destination),
		Token:       common.HexToAddress(a.Token),
		Amount:      a.Amount.BigInt(),
	}
}

func convertSignatureRes(sig clearnet.SignatureRes) (custody.Signature, error) {
	r, err := hexutil.Decode(sig.R)
	if err != nil {
		return custody.Signature{}, fmt.Errorf("failed to decode R: %w", err)
	}

	s, err := hexutil.Decode(sig.S)
	if err != nil {
		return custody.Signature{}, fmt.Errorf("failed to decode S: %w", err)
	}

	return custody.Signature{
		V: sig.V,
		R: [32]byte(r),
		S: [32]byte(s),
	}, nil
}
