package main

import (
	"fmt"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/ethereum/go-ethereum/common"

	"github.com/erc7824/nitrolite/examples/cerebro/clearnet"
	"github.com/erc7824/nitrolite/examples/cerebro/custody"
	"github.com/erc7824/nitrolite/examples/cerebro/storage"
)

type Operator struct {
	clearnode *clearnet.ClearnodeClient
	custody   *custody.CustodyClient
	store     *storage.Storage
	config    *OperatorConfig

	exitCh chan struct{}
}

func NewOperator(clearnode *clearnet.ClearnodeClient, store *storage.Storage) (*Operator, error) {
	operator := &Operator{
		clearnode: clearnode,
		custody:   custody.NewCustodyClient(),
		store:     store,
		config:    &OperatorConfig{},
		exitCh:    make(chan struct{}),
	}
	operator.reloadConfig()

	return operator, nil
}

func (operator *Operator) Complete(d prompt.Document) []prompt.Suggest {
	return prompt.FilterHasPrefix(operator.complete(d), d.GetWordBeforeCursor(), true)
}

func (o *Operator) complete(d prompt.Document) []prompt.Suggest {
	args := strings.Split(d.TextBeforeCursor(), " ")

	if len(args) < 2 {
		return []prompt.Suggest{
			{Text: "import", Description: "Import a wallet, signer or chain RPC URL"},
			{Text: "list", Description: "List available chains, wallets, or signers"},
			{Text: "authenticate", Description: "Authenticate to the Clearnode using your wallet private key"},
			{Text: "enable", Description: "Enable a chain for the current wallet"},
			{Text: "disable", Description: "Disable a chain for the current wallet"},
			{Text: "resize", Description: "Resize a channel on a chain"},
			{Text: "deposit", Description: "Deposit assets to a chain"},
			{Text: "withdraw", Description: "Withdraw assets from a chain"},
			{Text: "transfer", Description: "Transfer assets to someone else"},
			{Text: "exit", Description: "Exit the application"},
		}
	}

	if len(args) < 3 {
		switch args[0] {
		case "import":
			return []prompt.Suggest{
				{Text: "wallet", Description: "Import a wallet using its private key"},
				{Text: "signer", Description: "Import a signer using its private key"},
				{Text: "chain-rpc", Description: "Import a chain RPC URL"},
			}
		case "list":
			return []prompt.Suggest{
				{Text: "chains", Description: "List all available chains"},
				{Text: "wallets", Description: "List all imported wallets"},
				{Text: "signers", Description: "List all imported signers"},
			}
		case "authenticate":
			return o.getWalletSuggestions()
		case "enable":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getChainSuggestions(-1) // Suggest only disabled chains
		case "resize", "disable":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getChainSuggestions(1) // Suggest only enabled chains
		case "deposit", "withdraw":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getChainSuggestions(0) // Suggest all chains
		case "transfer":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getAssetSuggestions("", 0) // Suggest all assets for the current wallet
		default:
			return nil // No suggestions for other commands
		}
	}

	if len(args) < 4 {
		switch args[0] {
		case "import":
			switch args[1] {
			case "chain-rpc":
				return o.getChainSuggestions(0) // Suggest all chains for RPC import
			default:
				return nil // No suggestions for other commands
			}
		case "authenticate":
			return o.getSignerSuggestions()
		case "enable":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getAssetSuggestions(args[1], -1) // Suggest only disabled assets for the specified chain
		case "resize", "disable":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getAssetSuggestions(args[1], 1) // Suggest only enabled assets for the specified chain
		case "deposit", "withdraw":
			if !o.isUserAuthenticated() {
				return nil
			}

			return o.getAssetSuggestions(args[1], 0) // Suggest all assets for the specified chain
		default:
			return nil // No suggestions for other commands
		}
	}

	return nil // No suggestions for commands with more than 3 arguments
}

func (o *Operator) Execute(s string) {
	args := strings.Split(s, " ")
	if s == "" || len(args) == 0 {
		// No command provided
		return
	}

	defer o.reloadConfig()

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
		if len(args) < 2 {
			fmt.Println("Usage: import <wallet|signer|chain_rpc> <name>")
			return
		}

		switch args[1] {
		case "wallet", "signer":
			o.handleImportPKey(args)
		case "chain-rpc":
			o.handleImportRPC(args)
		default:
			fmt.Printf("Unknown import type: %s. Use 'wallet', 'signer', or 'chain_rpc'.\n", args[1])
			return
		}
	case "enable":
		o.handleEnableChain(args)
	case "disable":
		o.handleDisableChain(args)
	case "resize":
		o.handleResizeChain(args)
	case "deposit":
		o.handleDepositChain(args)
	case "withdraw":
		o.handleWithdrawChain(args)
	case "transfer":
		o.handleTransfer(args)
	case "exit":
		o.exit()
	default:
		fmt.Printf("Unknown command: %s\n", s)
	}
}

func (o *Operator) Wait() <-chan struct{} {
	// Wait for exit signal
	return o.exitCh
}

func (o *Operator) exit() {
	close(o.exitCh)
}

func (o *Operator) readExtraArg(name string) string {
	promptPrefix := fmt.Sprintf("{%s}>>> ", name)
	return prompt.Input(promptPrefix, emptyCompleter,
		prompt.OptionTitle("Cerebro CLI"),
		prompt.OptionPrefixTextColor(prompt.Yellow),
	)
}

func (o *Operator) reloadConfig() {
	brokerConfig, err := o.clearnode.GetConfig()
	if err != nil {
		fmt.Printf("[Reload] Failed to fetch broker config: %s\n", err.Error())
		return
	}

	assets, err := o.clearnode.GetSupportedAssets()
	if err != nil {
		fmt.Printf("[Reload] Failed to fetch supported assets: %s\n", err.Error())
		return
	}

	channels := []clearnet.ChannelRes{}
	if o.isUserAuthenticated() {
		channels, err = o.clearnode.GetChannels(o.config.Wallet.Address().Hex(), "open")
		if err != nil {
			fmt.Printf("[Reload] Failed to fetch channels: %s\n", err.Error())
			return
		}
	}

	o.config.BrokerAddress = common.HexToAddress(brokerConfig.BrokerAddress)
	o.config.Networks = make([]NetworkConfig, 0, len(brokerConfig.Networks))
	for _, network := range brokerConfig.Networks {
		chainAssets := make([]ChainAssetConfig, 0)
		for _, asset := range assets {
			if asset.ChainID == network.ChainID {
				channelID := ""
				for _, channel := range channels {
					if channel.ChainID == network.ChainID && channel.Token == asset.Token {
						channelID = channel.ChannelID
						break
					}
				}

				chainAssets = append(chainAssets, ChainAssetConfig{
					Token:     common.HexToAddress(asset.Token),
					Symbol:    asset.Symbol,
					Decimals:  asset.Decimals,
					ChannelID: channelID,
				})
			}
		}

		o.config.Networks = append(o.config.Networks, NetworkConfig{
			ChainName:          network.Name,
			ChainID:            network.ChainID,
			CustodyAddress:     common.HexToAddress(network.CustodyAddress),
			AdjudicatorAddress: common.HexToAddress(network.AdjudicatorAddress),
			Assets:             chainAssets,
		})
	}
}

// getChainSuggestions returns a list of chain suggestions based on the filterEnabled parameter.
// filterEnabled can be 0 (all chains), >0 (only enabled chains), or <0 (only disabled chains).
func (o *Operator) getChainSuggestions(filterEnabled int) []prompt.Suggest {
	suggestions := make([]prompt.Suggest, 0)
	for _, network := range o.config.Networks {
		include := filterEnabled == 0 || // Default to including all chains
			(filterEnabled > 0 && network.HasEnabledAssets()) || // Include only chains with enabled assets
			(filterEnabled < 0 && network.HasDisabledAssets()) // Include only chains with disabled assets

		if include {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        network.ChainName,
				Description: fmt.Sprintf("Chain %s (ID: %d)", network.ChainName, network.ChainID),
			})
		}
	}
	return suggestions
}

// getAssetSuggestions returns a list of asset suggestions for a specific chain.
// chainName is the name of the chain, and filterEnabled can be 0 (all assets),
// >0 (only enabled assets), or <0 (only disabled assets).
func (o *Operator) getAssetSuggestions(chainName string, filterEnabled int) []prompt.Suggest {
	if chainName == "" {
		assetSymbols := o.config.GetSymbolsOfEnabledAssets()
		suggestions := make([]prompt.Suggest, len(assetSymbols))
		for i, symbol := range assetSymbols {
			suggestions[i] = prompt.Suggest{
				Text:        symbol,
				Description: fmt.Sprintf("Asset %s", strings.ToUpper(symbol)),
			}
		}

		return suggestions
	}

	network := o.config.GetNetworkByName(chainName)
	if network == nil {
		return nil
	}

	suggestions := make([]prompt.Suggest, 0)
	for _, asset := range network.Assets {
		include := filterEnabled == 0 || // Default to including all assets
			(filterEnabled > 0 && asset.IsEnabled()) || // Include only enabled assets
			(filterEnabled < 0 && !asset.IsEnabled()) // Include only disabled assets

		if include {
			suggestions = append(suggestions, prompt.Suggest{
				Text:        asset.Symbol,
				Description: fmt.Sprintf("%s (%d decimals)", asset.Token.Hex(), asset.Decimals),
			})
		}
	}

	return suggestions
}

// getWalletSuggestions returns a list of wallet suggestions from the storage.
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

// getSignerSuggestions returns a list of signer suggestions from the storage.
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

func (o *Operator) isUserAuthenticated() bool {
	return o.config.Wallet != nil && o.config.Signer != nil
}
