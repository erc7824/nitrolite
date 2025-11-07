package main

import (
	"fmt"
	"math/big"
	"syscall"

	"golang.org/x/term"
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
		fmt.Println("Usage: import rpc <chain_name>")
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

	fmt.Println("Paste chain RPC URL:")
	rpcURL, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading chain RPC URL: %v\n", err)
		return
	}

	if err := o.store.AddChainRPC(string(rpcURL), blockchain.ID); err != nil {
		fmt.Printf("Failed to import chain RPC: %s\n", err.Error())
		return
	}
	fmt.Printf("RPC URL for chain %s imported successfully!\n", chainID.String())
}
