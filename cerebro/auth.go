package main

import (
	"fmt"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

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
	wallet, err := sign.NewEthereumSigner(walletPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create wallet signer: %s\n", err.Error())
		return
	}

	signerPKey, err := o.store.GetPrivateKeyByName(args[2])
	if err != nil {
		fmt.Printf("Failed to retrieve signer private key: %s\n", err.Error())
		return
	}
	signer, err := sign.NewEthereumSigner(signerPKey.PrivateKey)
	if err != nil {
		fmt.Printf("Failed to create signer: %s\n", err.Error())
		return
	}

	if _, err := o.clearnode.Authenticate(wallet, signer); err != nil {
		fmt.Printf("\nAuthentication failed: %s\n", err.Error())
		return
	}

	userTagRes, err := o.clearnode.GetUserTag()
	if err != nil {
		fmt.Printf("Failed to retrieve user tag: %s\n", err.Error())
		return
	}

	o.config.Wallet = wallet
	o.config.Signer = signer
	fmt.Println("Authentication successful!")
	fmt.Printf("Welcome, \"%s\"!\n", userTagRes.Tag)
}
