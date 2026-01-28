package node_v1

import (
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/rpc"
)

func mapBlockchainV1(blockchain core.Blockchain) rpc.BlockchainInfoV1 {
	return rpc.BlockchainInfoV1{
		Name:            blockchain.Name,
		BlockchainID:    blockchain.ID,
		ContractAddress: blockchain.ContractAddress,
	}
}

func mapAssetV1(asset core.Asset) rpc.AssetV1 {
	tokens := []rpc.TokenV1{}
	for _, token := range asset.Tokens {
		tokens = append(tokens, mapTokenV1(token))
	}

	return rpc.AssetV1{
		Name:     asset.Name,
		Symbol:   asset.Symbol,
		Decimals: asset.Decimals,
		Tokens:   tokens,
	}
}

func mapTokenV1(token core.Token) rpc.TokenV1 {
	return rpc.TokenV1{
		Name:         token.Name,
		Symbol:       token.Symbol,
		Address:      token.Address,
		BlockchainID: token.BlockchainID,
		Decimals:     token.Decimals,
	}
}
