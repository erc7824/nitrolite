package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"
)

func (o *Operator) handleListChains() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Name", "ID", "Asset", "RPCs", "Last Used"})
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
			t.AppendRow(table.Row{network.ChainName, network.ChainID, asset.Symbol, numRPCs, lastUsed.Format(time.RFC3339)})
		}
	}
	t.SetColumnConfigs(
		[]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 2, AutoMerge: true},
		},
	)
	t.Render()
}

func (o *Operator) handleListChannels() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Chain", "Asset", "ID", "Balance"})
	t.AppendSeparator()

	for _, network := range o.config.Networks {
		for _, asset := range network.Assets {
			channelID := "N/A"
			channelBalance := decimal.NewFromInt(0)
			if asset.ChannelID != "" {
				channelID = asset.ChannelID
				channelBalance = decimal.NewFromBigInt(asset.RawChannelBalance, -int32(asset.Decimals))
			}

			t.AppendRow(table.Row{network.ChainName, asset.Symbol, channelID, channelBalance.StringFixed(2)})
		}
	}
	t.SetColumnConfigs(
		[]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 2, AutoMerge: true},
		},
	)
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
