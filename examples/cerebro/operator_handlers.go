package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"
	"golang.org/x/term"

	"github.com/erc7824/nitrolite/examples/bridge/clearnet"
	"github.com/erc7824/nitrolite/examples/bridge/custody"
	"github.com/erc7824/nitrolite/examples/bridge/unisig"
)

func (o *Operator) handleImportPKey(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: import <wallet|signer> <name>")
		return
	}

	var isSigner bool
	switch args[1] {
	case "wallet":
		isSigner = false
	case "signer":
		isSigner = true
	default:
		fmt.Printf("Unknown import type: %s. Use 'wallet' or 'signer'.\n", args[1])
		return
	}

	fmt.Println("Paste private key:")
	privateKeyHex, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading key: %v\n", err)
		return
	}

	pkeyDTO, err := o.store.AddPrivateKey(args[2], string(privateKeyHex), isSigner)
	if err != nil {
		fmt.Printf("Failed to import private key: %s\n", err.Error())
		return
	}
	fmt.Printf("Private key imported successfully: %s (%s)\n", pkeyDTO.Name, pkeyDTO.Address)
}

func (o *Operator) handleImportRPC(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: import rpc <chain_id>")
		return
	}

	bigChainID, ok := new(big.Int).SetString(args[2], 10)
	if !ok {
		fmt.Printf("Invalid chain ID: %s\n", args[2])
		return
	}
	chainID := uint32(bigChainID.Uint64())

	fmt.Println("Paste chain RPC URL:")
	rpcURL, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading chain RPC URL: %v\n", err)
		return
	}

	if err := o.store.AddChainRPC(string(rpcURL), chainID); err != nil {
		fmt.Printf("Failed to import chain RPC: %s\n", err.Error())
		return
	}
	fmt.Printf("RPC URL for chain(%d) imported successfully!\n", chainID)
}

func (o *Operator) handleAuthenticate(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: authenticate <wallet> <signer>")
		return
	}
	if o.config.Wallet != nil || o.config.Signer != nil {
		fmt.Println("Already authenticated.")
		return
	}

	walletPKey, err := o.store.GetPrivateKeyByName(args[1])
	if err != nil {
		fmt.Printf("Failed to retrieve wallet private key: %s\n", err.Error())
		return
	}
	wallet, err := unisig.NewEcdsaSigner(walletPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create wallet signer: %s\n", err.Error())
		return
	}

	signerPKey, err := o.store.GetPrivateKeyByName(args[2])
	if err != nil {
		fmt.Printf("Failed to retrieve signer private key: %s\n", err.Error())
		return
	}
	signer, err := unisig.NewEcdsaSigner(signerPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create signer: %s\n", err.Error())
		return
	}

	if err := o.clearnode.Authenticate(wallet, signer); err != nil {
		fmt.Printf("\nAuthentication failed: %s\n", err.Error())
		return
	}

	o.config.Wallet = wallet
	o.config.Signer = signer
	fmt.Println("Authentication successful!")
}

func (o *Operator) handleListChains() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "ID", "Asset", "Enabled", "RPCs", "Last Used"})
	t.AppendSeparator()

	for _, network := range o.config.Networks {
		chainRPCDTOs, err := o.store.GetChainRPCs(network.ChainID)
		if err != nil {
			fmt.Printf("Failed to get RPCs for chain %d: %s\n", network.ChainID, err.Error())
			continue
		}

		numRPCs := len(chainRPCDTOs)
		lastUsed := time.Unix(0, 0)
		if numRPCs > 0 {
			lastUsed = chainRPCDTOs[0].LastUsedAt
		}

		for _, asset := range network.Assets {
			t.AppendRow(table.Row{network.ChainName, network.ChainID, asset.Symbol, asset.IsEnabled(), numRPCs, lastUsed.Format(time.RFC3339)})
		}
	}
	t.Render()
}

func (o *Operator) handleListPKeys(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: list <wallets|signers>")
		return
	}

	var isSigner bool
	switch args[1] {
	case "wallets":
		isSigner = false
	case "signers":
		isSigner = true
	default:
		fmt.Printf("Usage: list <wallets|signers>")
		return
	}

	dtos, err := o.store.GetPrivateKeys(isSigner)
	if err != nil {
		fmt.Printf("Failed to fetch wallets: %s\n", err.Error())
		return
	}
	if len(dtos) == 0 {
		fmt.Println("No keys found.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "Address"})
	t.AppendSeparator()
	for _, dto := range dtos {
		t.AppendRow([]interface{}{dto.Name, dto.Address})
	}
	t.Render()
}

func (o *Operator) handleEnableChain(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: enable <chain_name> <token_symbol>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
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
	if asset.IsEnabled() {
		fmt.Printf("Asset %s on chain %s is already enabled.\n", assetSymbol, chainName)
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

	fmt.Printf("Opening custody channel for %s...\n", chainName)

	if err := o.custody.OpenChannel(
		o.config.Wallet, o.config.Signer,
		network.ChainID, chainRPCDTOs[chainRPCIndex].URL,
		network.CustodyAddress,
		network.AdjudicatorAddress,
		o.config.BrokerAddress,
		asset.Token,
		0,
	); err != nil {
		fmt.Printf("Failed to open custody channel for chain %s: %s\n", chainName, err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPCDTOs[chainRPCIndex].URL); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	fmt.Printf("Successfully opened custody channel for %s!\n", chainName)
}

func (o *Operator) handleDisableChain(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: disable <chain_name>")
		return
	}

	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
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
	if !asset.IsEnabled() {
		fmt.Printf("Asset %s on chain %s is not enabled.\n", assetSymbol, chainName)
		return
	}

	closureRes, err := o.clearnode.RequestChannelClosure(o.config.Wallet.Address(), asset.ChannelID)
	if err != nil {
		fmt.Printf("Failed to request channel closure for chain %s: %s\n", chainName, err.Error())
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

	fmt.Printf("Closing custody channel for %s...\n", chainName)

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
		network.ChainID, chainRPCDTOs[chainRPCIndex].URL,
		network.CustodyAddress,
		common.HexToHash(closureRes.ChannelID),
		new(big.Int).SetUint64(closureRes.Version),
		allocations,
		brokerSig,
	); err != nil {
		fmt.Printf("Failed to close custody channel for chain %s: %s\n", chainName, err.Error())
		return
	}

	if err := o.store.UpdateChainRPCUsage(chainRPCDTOs[chainRPCIndex].URL); err != nil {
		fmt.Printf("Failed to update chain RPC usage: %s\n", err.Error())
		return
	}

	unlockedAmount := decimal.NewFromBigInt(allocations[0].Amount, -int32(asset.Decimals))
	fmt.Printf("Successfully closed custody channel for %s with unlocked %s %s!\n", chainName, unlockedAmount, strings.ToUpper(asset.Symbol))
}

func convertAllocationRes(a clearnet.AllocationRes) custody.Allocation {
	return custody.Allocation{
		Destination: common.HexToAddress(a.Destination),
		Token:       common.HexToAddress(a.Token),
		Amount:      a.Amount,
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
