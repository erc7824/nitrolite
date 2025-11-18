package main

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
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
	if asset.IsEnabled() {
		fmt.Printf("Channel is already opened for asset %s on %s: %s.\n", assetSymbol, chainID.String(), asset.ChannelID)
		return
	}

	creationRes, err := o.clearnode.RequestChannelCreation(blockchain.ID, asset.Token.Hex())
	if err != nil {
		fmt.Printf("Failed to request channel creation for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	if creationRes.Channel == nil {
		fmt.Printf("Response from Clearnode doesn't have a channel data: %v\n", creationRes)
		return
	}

	chainRPC, err := o.getChainRPC(blockchain.ID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	fmt.Printf("Opening custody channel on %s...\n", chainID.String())

	participantSigner := o.config.Signer
	if participants := creationRes.Channel.Participants; len(participants) > 0 && participants[0] == o.config.Wallet.PublicKey().Address().String() {
		participantSigner = o.config.Wallet
	}
	channelID, err := o.custody.OpenChannel(
		o.config.Wallet, participantSigner,
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress,
		blockchain.AdjudicatorAddress,
		o.config.BrokerAddress,
		asset.Token,
		creationRes.Channel.Challenge,
		creationRes.Channel.Nonce,
		creationRes.StateSignature,
	)
	if err != nil {
		fmt.Printf("Failed to open custody channel on %s: %s\n", chainID.String(), err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPC); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully opened custody channel (%s) on %s!\n", channelID, chainID.String())
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
	if !asset.IsEnabled() {
		fmt.Printf("There are no opened channels for %s on %s.\n", assetSymbol, chainID.String())
		return
	}

	closureRes, err := o.clearnode.RequestChannelClosure(o.config.Wallet.PublicKey().Address(), asset.ChannelID)
	if err != nil {
		fmt.Printf("Failed to request channel closure for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	chainRPC, err := o.getChainRPC(blockchain.ID)
	if err != nil {
		fmt.Printf("Failed to get RPC for chain %s: %s\n", chainID.String(), err.Error())
		return
	}

	fmt.Printf("Closing custody channel (%s) on %s...\n", asset.ChannelID, chainID.String())

	allocations := make([]custody.Allocation, len(closureRes.State.Allocations))
	for i, alloc := range closureRes.State.Allocations {
		allocations[i] = convertAllocationRes(alloc)
	}

	participantSigner := o.config.Signer
	if asset.ChannelParticipant == o.config.Wallet.PublicKey().Address().String() {
		participantSigner = o.config.Wallet
	}
	if err := o.custody.CloseChannel(
		o.config.Wallet, participantSigner,
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress,
		common.HexToHash(closureRes.ChannelID),
		new(big.Int).SetUint64(closureRes.State.Version),
		allocations,
		closureRes.StateSignature,
	); err != nil {
		fmt.Printf("Failed to close custody channel on %s: %s\n", chainID.String(), err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPC); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	unlockedAmount := decimal.NewFromBigInt(allocations[0].Amount, -int32(asset.Decimals))
	fmt.Printf("Successfully closed custody channel (%s) on %s with unlocked %s %s!\n", asset.ChannelID, chainID.String(), fmtDec(unlockedAmount), strings.ToUpper(asset.Symbol))
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
	if !asset.IsEnabled() {
		fmt.Printf("There are no opened channels for %s on %s.\n", assetSymbol, chainID.String())
		return
	}
	if asset.ChannelResizing {
		fmt.Printf("Channel for asset %s on %s is already being resized.\n", assetSymbol, chainID.String())
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
		fmt.Printf("Failed to get balance for asset %s on %s: %s\n", assetSymbol, chainID.String(), err.Error())
		return
	}
	custodyBalance := decimal.NewFromBigInt(rawCustodyBalance, -int32(asset.Decimals))

	channelBalance := decimal.NewFromBigInt(asset.RawChannelBalance, -int32(asset.Decimals))

	getLedgerBalancesRes, err := o.clearnode.GetLedgerBalances(o.config.Wallet.PublicKey().Address().String())
	if err != nil {
		fmt.Printf("Failed to get ledger balances: %s\n", err.Error())
		return
	}

	unifiedBalance := decimal.New(0, 0)
	for _, balance := range getLedgerBalancesRes.LedgerBalances {
		if balance.Asset == asset.Symbol {
			unifiedBalance = balance.Amount
			break
		}
	}

	fmt.Printf("Your current balances for asset %s:\n", assetSymbol)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Type", "Value"})
	t.AppendSeparator()
	t.AppendRow(table.Row{"On Custody Ledger", fmtDec(custodyBalance)})
	t.AppendRow(table.Row{"On Channel", fmtDec(channelBalance)})
	t.AppendRow(table.Row{"Unified On Clearnode", fmtDec(unifiedBalance)})
	t.Render()

	fmt.Printf("How much %s do you want to resize (+)into/(-)out channel?\n", assetSymbol)
	fmt.Println("That's the amount moved between custody ledger and channel.")
	resizeAmountStr := o.readExtraArg("resize_amount")

	resizeAmount, err := decimal.NewFromString(resizeAmountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if resizeAmount.GreaterThan(custodyBalance) {
		fmt.Printf("You cannot resize more than your current balance of %s %s.\n",
			fmtDec(custodyBalance), assetSymbol)
		return
	}
	rawResizeAmount := resizeAmount.Shift(int32(asset.Decimals))

	fmt.Printf("How much %s do you want to allocate (+)into/(-)out channel?\n", assetSymbol)
	fmt.Println("That's the amount moved between unified balance and channel.")
	allocateAmountStr := o.readExtraArg("allocate_amount")

	allocateAmount, err := decimal.NewFromString(allocateAmountStr)
	if err != nil {
		fmt.Printf("Invalid amount format: %s\n", err.Error())
		return
	}

	if allocateAmount.GreaterThan(unifiedBalance) {
		fmt.Printf("You cannot allocate more than your current unified balance of %s %s.\n",
			fmtDec(unifiedBalance), assetSymbol)
		return
	}
	rawAllocateAmount := allocateAmount.Shift(int32(asset.Decimals))

	if newChannelBalance := channelBalance.Add(allocateAmount).Add(resizeAmount); newChannelBalance.LessThan(decimal.Zero) {
		fmt.Printf("New channel amount must not be negative after resize: %s\n", fmtDec(newChannelBalance))
		return
	}

	resizeRes, err := o.clearnode.RequestChannelResize(o.config.Wallet.PublicKey().Address(), asset.ChannelID, rawAllocateAmount, rawResizeAmount)
	if err != nil {
		fmt.Printf("Failed to request channel resize on %s: %s\n", chainID.String(), err.Error())
		return
	}

	fmt.Printf("Resizing custody channel (%s) on %s...\n", asset.ChannelID, chainID.String())

	stateData, err := hexutil.Decode(resizeRes.State.Data)
	if err != nil {
		fmt.Printf("Failed to decode state data: %s\n", err.Error())
		return
	}

	allocations := make([]custody.Allocation, len(resizeRes.State.Allocations))
	for i, alloc := range resizeRes.State.Allocations {
		allocations[i] = convertAllocationRes(alloc)
	}

	participantSigner := o.config.Signer
	if asset.ChannelParticipant == o.config.Wallet.PublicKey().Address().String() {
		participantSigner = o.config.Wallet
	}
	if err := o.custody.Resize(
		o.config.Wallet, participantSigner,
		blockchain.ID, chainRPC,
		blockchain.CustodyAddress,
		common.HexToHash(resizeRes.ChannelID),
		new(big.Int).SetUint64(resizeRes.State.Version),
		stateData,
		allocations,
		resizeRes.StateSignature,
	); err != nil {
		fmt.Printf("Failed to resize custody channel on %s: %s\n", chainID.String(), err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPC); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully resized custody channel (%s) on %s!\n", asset.ChannelID, chainID.String())
}

func convertAllocationRes(a rpc.StateAllocation) custody.Allocation {
	return custody.Allocation{
		Destination: common.HexToAddress(a.Participant),
		Token:       common.HexToAddress(a.TokenAddress),
		Amount:      a.RawAmount.BigInt(),
	}
}
