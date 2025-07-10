package clearnet

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
)

func (c *ClearnodeClient) handleEvent(event RPCData) {
	if !c.printEvents {
		return
	}

	switch event.Method {
	case "assets":
		c.handleAssetsEvent(event)
	case "channels":
		c.handleChannelsEvent(event)
	case "bu":
		c.handleBalancesEvent(event)
	default:
		fmt.Printf("Unknown event method: %s\n", event.Method)
	}
}

type ChannelRes struct {
	ChannelID   string   `json:"channel_id"`
	Participant string   `json:"participant"`
	Status      string   `json:"status"`
	Token       string   `json:"token"`
	RawAmount   *big.Int `json:"amount"` // Total amount in the channel (user + broker)
	ChainID     uint32   `json:"chain_id"`
	Adjudicator string   `json:"adjudicator"`
	Challenge   uint64   `json:"challenge"`
	UpdatedAt   string   `json:"updated_at"`
}

func (c *ClearnodeClient) handleChannelsEvent(event RPCData) {
	if len(event.Params) < 1 {
		fmt.Println("Invalid channels event format")
		return
	}

	channelsData, err := json.Marshal(event.Params[0])
	if err != nil {
		fmt.Printf("Failed to marshal channels data: %s\n", err.Error())
		return
	}

	channels := make([]ChannelRes, 0)
	if err := json.Unmarshal(channelsData, &channels); err != nil {
		fmt.Printf("Failed to parse channels: %s\n", err.Error())
		return
	}

	if len(channels) == 0 {
		return
	}

	fmt.Printf("Active Channels:\n")
	for _, channel := range channels {
		fmt.Printf("* Channel ID: %s\n", channel.ChannelID)
		fmt.Printf("  Chain ID: %d\n", channel.ChainID)
		// fmt.Printf("  Participant: %s\n", channel.Participant)
		fmt.Printf("  Status: %s\n", channel.Status)
		fmt.Printf("  Token: %s\n", channel.Token)
		fmt.Printf("  Amount: %s\n", channel.RawAmount.String())
		// fmt.Printf("  Adjudicator: %s\n", channel.Adjudicator)
		// fmt.Printf("  Challenge: %d\n", channel.Challenge)
		// fmt.Printf("  Updated At: %s\n", channel.UpdatedAt)
	}
	fmt.Println()
}

func (c *ClearnodeClient) handleBalancesEvent(event RPCData) {
	if len(event.Params) < 1 {
		fmt.Println("Invalid channels event format")
		return
	}

	channelsData, err := json.Marshal(event.Params[0])
	if err != nil {
		fmt.Printf("Failed to marshal channels data: %s\n", err.Error())
		return
	}

	balances := make([]BalanceRes, 0)
	if err := json.Unmarshal(channelsData, &balances); err != nil {
		fmt.Printf("Failed to parse channels: %s\n", err.Error())
		return
	}

	if len(balances) == 0 {
		return
	}

	fmt.Printf("Balances' update:\n")
	for _, balance := range balances {
		fmt.Printf("* Asset: %s, Amount: %s\n", balance.Asset, balance.Amount.String())
	}
	fmt.Println()
}

type AssetRes struct {
	Token    string `json:"token"`
	ChainID  uint32 `json:"chain_id"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
}

func (c *ClearnodeClient) handleAssetsEvent(event RPCData) {
	if len(event.Params) < 1 {
		fmt.Println("Invalid assets event format")
		return
	}

	assetsData, err := json.Marshal(event.Params[0])
	if err != nil {
		fmt.Printf("Failed to marshal assets data: %s\n", err.Error())
		return
	}

	assets := make([]AssetRes, 0)
	if err := json.Unmarshal(assetsData, &assets); err != nil {
		fmt.Printf("Failed to parse assets: %s\n", err.Error())
		return
	}

	assetsMap := make(map[string]map[uint32]AssetRes)
	assetSymbols := []string{}
	assetChainIDs := []uint32{}
	for _, asset := range assets {
		if _, exists := assetsMap[asset.Symbol]; !exists {
			assetsMap[asset.Symbol] = make(map[uint32]AssetRes)
			assetSymbols = append(assetSymbols, asset.Symbol)
		}
		chainAssetMap := assetsMap[asset.Symbol]

		if _, exists := chainAssetMap[asset.ChainID]; !exists {
			chainAssetMap[asset.ChainID] = asset
		}
		assetChainIDs = append(assetChainIDs, asset.ChainID)
	}
	sort.Strings(assetSymbols)
	assetSymbols = UniqueSortedSlice(assetSymbols)
	sort.Slice(assetChainIDs, func(i, j int) bool {
		return assetChainIDs[i] < assetChainIDs[j]
	})
	assetChainIDs = UniqueSortedSlice(assetChainIDs)

	fmt.Printf("Supported Assets:\n")
	for _, symbol := range assetSymbols {
		fmt.Printf("* %s:\n", symbol)
		chainAssets := assetsMap[symbol]
		for _, chainID := range assetChainIDs {
			if asset, exists := chainAssets[chainID]; exists {
				fmt.Printf("  * Chain %d: %s (%d decimals)\n", chainID, asset.Token, asset.Decimals)
			} else {
				fmt.Printf("  * Chain %d: Not available\n", chainID)
			}
		}
	}
	fmt.Println()
}

func UniqueSortedSlice[T comparable](input []T) []T {
	if len(input) == 0 {
		return []T{}
	}

	result := []T{input[0]}
	for i := 1; i < len(input); i++ {
		if input[i] != input[i-1] {
			result = append(result, input[i])
		}
	}
	return result
}
