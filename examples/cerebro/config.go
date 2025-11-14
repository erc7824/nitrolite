package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

type OperatorConfig struct {
	BrokerAddress common.Address
	Blockchains   []BlockchainConfig
	Wallet        sign.Signer
	Signer        sign.Signer
}

func (c OperatorConfig) GetBlockchainByID(chainID uint32) *BlockchainConfig {
	for _, blockchain := range c.Blockchains {
		if blockchain.ID == chainID {
			return &blockchain
		}
	}
	return nil
}

func (c OperatorConfig) GetAssetSymbols() []string {
	var symbols []string
	var alreadyAdded = make(map[string]bool)
	for _, network := range c.Blockchains {
		for _, asset := range network.Assets {
			if !alreadyAdded[asset.Symbol] {
				symbols = append(symbols, asset.Symbol)
				alreadyAdded[asset.Symbol] = true
			}
		}
	}
	return symbols
}

type BlockchainConfig struct {
	ID                 uint32
	AdjudicatorAddress common.Address
	CustodyAddress     common.Address
	Assets             []ChainAssetConfig
}

func (c BlockchainConfig) GetAssetBySymbol(symbol string) *ChainAssetConfig {
	for _, asset := range c.Assets {
		if asset.Symbol == symbol {
			return &asset
		}
	}
	return nil
}

func (c BlockchainConfig) HasEnabledAssets() bool {
	for _, asset := range c.Assets {
		if asset.IsEnabled() {
			return true
		}
	}
	return false
}

func (c BlockchainConfig) HasDisabledAssets() bool {
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

	ChannelID          string
	ChannelResizing    bool
	ChannelParticipant string
	RawChannelBalance  *big.Int
}

func (c ChainAssetConfig) IsEnabled() bool {
	return c.ChannelID != ""
}
