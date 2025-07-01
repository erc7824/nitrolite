package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

func (o *Operator) handleImport(args []string) {
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
		fmt.Printf("\nError reading password: %v\n", err)
		return
	}

	pkeyDTO, err := o.store.AddPrivateKey(args[2], string(privateKeyHex), isSigner)
	if err != nil {
		fmt.Printf("Failed to import private key: %s\n", err.Error())
		return
	}
	fmt.Printf("Private key imported successfully: %s (%s)\n", pkeyDTO.Name, pkeyDTO.Address)
}

func (o *Operator) handleAuthenticate(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: authenticate <wallet> <signer>")
		return
	}

	walletPKey, err := o.store.GetPrivateKeyByName(args[1])
	if err != nil {
		fmt.Printf("Failed to retrieve wallet private key: %s\n", err.Error())
		return
	}
	wallet, err := NewSigner(walletPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create wallet signer: %s\n", err.Error())
		return
	}

	signerPKey, err := o.store.GetPrivateKeyByName(args[2])
	if err != nil {
		fmt.Printf("Failed to retrieve signer private key: %s\n", err.Error())
		return
	}
	signer, err := NewSigner(signerPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create signer: %s\n", err.Error())
		return
	}

	if err := o.clearnode.Authenticate(wallet, signer); err != nil {
		fmt.Printf("\nAuthentication failed: %s\n", err.Error())
		return
	}

	fmt.Println("Authentication successful!")
}

func (o *Operator) handleListChains() {
	signer := o.clearnode.Signer()
	if signer == nil {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	channels, err := o.clearnode.GetChannels(signer.Address().Hex(), "open")
	if err != nil {
		fmt.Printf("Failed to fetch channels: %s\n", err.Error())
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "ID", "Enabled"})
	t.AppendSeparator()

	for _, network := range o.brokerConf.Networks {
		enabled := false
		for _, channel := range channels {
			if channel.ChainID == network.ChainID {
				enabled = true
				break
			}
		}

		t.AppendRow([]interface{}{network.Name, network.ChainID, enabled})
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
