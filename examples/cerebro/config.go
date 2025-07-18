package main

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/erc7824/nitrolite/examples/cerebro/unisig"
)

type OperatorConfig struct {
	BrokerAddress common.Address
	Networks      []NetworkConfig
	Wallet        unisig.Signer
	Signer        unisig.Signer
}

func (c OperatorConfig) GetNetworkByName(name string) *NetworkConfig {
	for _, network := range c.Networks {
		if network.ChainName == name {
			return &network
		}
	}
	return nil
}

func (c OperatorConfig) GetSymbolsOfEnabledAssets() []string {
	var symbols []string
	var alreadyAdded = make(map[string]bool)
	for _, network := range c.Networks {
		for _, asset := range network.Assets {
			if asset.IsEnabled() && !alreadyAdded[asset.Symbol] {
				symbols = append(symbols, asset.Symbol)
				alreadyAdded[asset.Symbol] = true
			}
		}
	}
	return symbols
}

type NetworkConfig struct {
	ChainName          string
	ChainID            uint32
	AdjudicatorAddress common.Address
	CustodyAddress     common.Address
	Assets             []ChainAssetConfig
}

func (c NetworkConfig) GetAssetBySymbol(symbol string) *ChainAssetConfig {
	for _, asset := range c.Assets {
		if asset.Symbol == symbol {
			return &asset
		}
	}
	return nil
}

func (c NetworkConfig) HasEnabledAssets() bool {
	for _, asset := range c.Assets {
		if asset.IsEnabled() {
			return true
		}
	}
	return false
}

func (c NetworkConfig) HasDisabledAssets() bool {
	for _, asset := range c.Assets {
		if !asset.IsEnabled() {
			return true
		}
	}
	return false
}

type ChainAssetConfig struct {
	Token    common.Address
	Symbol   string
	Decimals uint8

	ChannelID      string
	ChannelBalance string
}

func (c ChainAssetConfig) IsEnabled() bool {
	return c.ChannelID != ""
}
