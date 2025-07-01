package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
)

type Operator struct {
	clearnode  *ClearnodeClient
	store      *Storage
	brokerConf *BrokerConfig
}

func NewOperator(clearnode *ClearnodeClient, store *Storage) (*Operator, error) {
	brokerConf, err := clearnode.GetConfig()
	if err != nil {
		return nil, err
	}

	return &Operator{
		clearnode:  clearnode,
		store:      store,
		brokerConf: brokerConf,
	}, nil
}

func (operator *Operator) Complete(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(operator.complete(d), d.GetWordBeforeCursor(), true)
}

func (o *Operator) complete(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) < 2 {
		return []prompt.Suggest{
			{Text: "import", Description: "Import a wallet or signer"},
			{Text: "list", Description: "List available chains, wallets, or signers"},
			{Text: "authenticate", Description: "Authenticate to the Clearnode using your wallet private key"},
			{Text: "exit", Description: "Exit the application"},
		}
	}

	if len(args) < 3 {
		switch args[0] {
		case "import":
			return []prompt.Suggest{
				{Text: "wallet", Description: "Import a wallet using its private key"},
				{Text: "signer", Description: "Import a signer using its private key"},
			}
		case "list":
			return []prompt.Suggest{
				{Text: "chains", Description: "List all available chains"},
				{Text: "wallets", Description: "List all imported wallets"},
				{Text: "signers", Description: "List all imported signers"},
			}
		case "authenticate":
			return o.getWalletSuggestions()
		default:
			return nil // No suggestions for other commands
		}
	}

	if len(args) < 4 {
		switch args[0] {
		case "list":
			return o.getWalletSuggestions()
		case "authenticate":
			return o.getSignerSuggestions()
		default:
			return nil // No suggestions for other commands
		}
	}

	return nil // No suggestions for commands with more than 3 arguments
}

func (o *Operator) Execute(s string) {
	args := strings.Split(s, " ")
	if len(args) == 0 {
		// No command provided
		return
	}

	switch args[0] {
	case "authenticate":
		o.handleAuthenticate(args)
	case "list":
		if len(args) < 2 {
			fmt.Println("Usage: list <chains|wallets|signers>")
			return
		}

		switch args[1] {
		case "chains":
			o.handleListChains()
		case "wallets", "signers":
			o.handleListPKeys(args)
		default:
			fmt.Printf("Unknown list type: %s. Use 'chains', 'wallets', or 'signers'.\n", args[1])
			return
		}
	case "import":
		o.handleImport(args)
	case "exit":
		fmt.Println("Exiting...")
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", s)
	}
}

func (o *Operator) getWalletSuggestions() []prompt.Suggest {
	walletDTOs, err := o.store.GetPrivateKeys(false)
	if err != nil {
		fmt.Printf("Failed to fetch wallets: %s\n", err.Error())
		return nil
	}

	s := make([]prompt.Suggest, 0, len(walletDTOs))
	for _, wallet := range walletDTOs {
		s = append(s, prompt.Suggest{
			Text:        wallet.Name,
			Description: fmt.Sprintf("Wallet with address %s", wallet.Address),
		})
	}
	return s
}

func (o *Operator) getSignerSuggestions() []prompt.Suggest {
	signerDTOs, err := o.store.GetPrivateKeys(true)
	if err != nil {
		fmt.Printf("Failed to fetch signers: %s\n", err.Error())
		return nil
	}

	s := make([]prompt.Suggest, 0, len(signerDTOs))
	for _, signer := range signerDTOs {
		s = append(s, prompt.Suggest{
			Text:        signer.Name,
			Description: fmt.Sprintf("Signer with address %s", signer.Address),
		})
	}
	return s
}
