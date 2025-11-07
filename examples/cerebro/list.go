package main

import (
	"fmt"
	"os"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/shopspring/decimal"
)

const (
	initialQueryOffset    = uint32(0)
	defaultPageSize       = uint32(10)
	ledgerEntriesPageSize = uint32(100)
)

func (o *Operator) handleListChains() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Asset", "RPCs", "Last Used"})
	t.AppendSeparator()

	for _, network := range o.config.Blockchains {
		chainRPCDTOs, err := o.store.GetChainRPCs(network.ID)
		if err != nil {
			fmt.Printf("Failed to get RPCs for chain %d: %s\n", network.ID, err.Error())
			continue
		}

		numRPCs := len(chainRPCDTOs)
		lastUsed := time.Unix(0, 0)
		if numRPCs > 0 {
			lastUsed = chainRPCDTOs[0].LastUsedAt
		}

		for _, asset := range network.Assets {
			t.AppendRow(table.Row{network.ID, asset.Symbol, numRPCs, lastUsed.Format(time.RFC3339)})
		}
	}
	t.SetColumnConfigs(
		[]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
			{Number: 3, AutoMerge: true},
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
	t.AppendHeader(table.Row{"ChainID", "Asset", "ID", "Balance"})
	t.AppendSeparator()

	for _, network := range o.config.Blockchains {
		for _, asset := range network.Assets {
			channelID := "N/A"
			channelBalance := decimal.NewFromInt(0)
			if asset.ChannelID != "" {
				channelID = asset.ChannelID
				channelBalance = decimal.NewFromBigInt(asset.RawChannelBalance, -int32(asset.Decimals))
			}

			t.AppendRow(table.Row{network.ID, asset.Symbol, channelID, fmtDec(channelBalance)})
		}
	}
	t.SetColumnConfigs(
		[]table.ColumnConfig{
			{Number: 1, AutoMerge: true},
		},
	)
	t.Render()
}

func (o *Operator) handleListAppSessions() {
	fmt.Println("Specify participant filter for app sessions (required):")
	participant := o.readExtraArg("participant")
	if !common.IsHexAddress(participant) {
		fmt.Println("Invalid address format. Please provide a valid Ethereum address.")
		return
	}

	fmt.Println("Specify status filter for app sessions (leave empty for no filter):")
	statusSuggestions := []prompt.Suggest{
		{Text: "open", Description: "List only open sessions"},
		{Text: "closed", Description: "List only closed sessions"},
	}
	statusFilter := o.readSelectionArg("status", statusSuggestions)

	offset := initialQueryOffset
	for {
		res, err := o.clearnode.GetAppSessions(participant, statusFilter, offset, defaultPageSize)
		if err != nil {
			fmt.Printf("Failed to get app sessions: %s\n", err.Error())
			return
		}

		if len(res.AppSessions) == 0 {
			fmt.Println("No more app sessions found.")
			break
		}

		fmt.Printf("App Sessions for participant %s:\n", participant)
		for _, session := range res.AppSessions {
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)

			t.AppendRow(table.Row{"ID", session.AppSessionID, session.AppSessionID, session.AppSessionID}, table.RowConfig{AutoMerge: true})
			t.AppendRow(table.Row{"NAME", "", "PROTOCOL", session.Protocol}, table.RowConfig{AutoMerge: true})
			t.AppendRow(table.Row{"NONCE", session.Nonce, "CHALLENGE", session.Challenge}, table.RowConfig{AutoMerge: true})
			t.AppendRow(table.Row{"VERSION", session.Version, "STATUS", session.Status}, table.RowConfig{AutoMerge: true})
			t.AppendRow(table.Row{"CREATED_AT", session.CreatedAt, "UPDATED_AT", session.UpdatedAt}, table.RowConfig{AutoMerge: true})
			t.AppendSeparator()

			if len(session.ParticipantWallets) > len(session.Weights) {
				fmt.Printf("Warning: Mismatched participant wallets and weights in session %s\n", session.AppSessionID)
				session.Weights = append(session.Weights, make([]int64, len(session.ParticipantWallets)-len(session.Weights))...)
			}

			for i := range session.ParticipantWallets {
				t.AppendRow(table.Row{fmt.Sprintf("PARTICIPANT_%d", i+1), session.ParticipantWallets[i], fmt.Sprintf("WEIGHT_%d", i+1), fmt.Sprintf("%d/%d", session.Weights[i], session.Quorum)}, table.RowConfig{AutoMerge: true})
			}

			t.Render()
		}

		fmt.Println("Do you want to see the next page?")
		actionSuggestions := []prompt.Suggest{
			{Text: "yes", Description: "Show next page"},
			{Text: "no", Description: "Quit listing"},
		}
		showNextPage := o.readSelectionArg("answer", actionSuggestions) == "yes"
		if !showNextPage {
			break
		}

		offset += defaultPageSize
	}
}

func (o *Operator) handleListLedgerBalances() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	res, err := o.clearnode.GetLedgerBalances(o.config.Wallet.PublicKey().Address().String())
	if err != nil {
		fmt.Printf("Failed to get ledger balances: %s\n", err.Error())
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Asset", "Balance"})
	t.AppendSeparator()

	for _, balance := range res.LedgerBalances {
		t.AppendRow(table.Row{balance.Asset, fmtDec(balance.Amount)})
	}
	t.Render()
}

func (o *Operator) handleListLedgerTransactions() {
	if !o.isUserAuthenticated() {
		fmt.Println("Not authenticated. Please authenticate first.")
		return
	}

	offset := initialQueryOffset
	for {
		res, err := o.clearnode.GetLedgerTransactions(o.config.Wallet.PublicKey().Address().String(), "", offset, defaultPageSize)
		if err != nil {
			fmt.Printf("Failed to get ledger transactions: %s\n", err.Error())
			return
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "Type", "From", "To", "Asset", "Amount", "Timestamp"})
		t.AppendSeparator()

		if len(res.LedgerTransactions) == 0 {
			fmt.Println("No more ledger transactions found.")
			break
		}

		for _, tx := range res.LedgerTransactions {
			t.AppendRow(table.Row{tx.Id, tx.TxType, tx.FromAccount, tx.ToAccount, tx.Asset, fmtDec(tx.Amount), tx.CreatedAt.Format(time.RFC3339)})
		}
		t.Render()

		fmt.Println("Do you want to see the next page?")
		actionSuggestions := []prompt.Suggest{
			{Text: "yes", Description: "Show next page"},
			{Text: "no", Description: "Quit listing"},
		}
		showNextPage := o.readSelectionArg("answer", actionSuggestions) == "yes"
		if !showNextPage {
			break
		}

		offset += defaultPageSize
	}
}

func (o *Operator) handleListLedgerEntries() {
	fmt.Println("Specify wallet filter (required):")
	wallet := o.readExtraArg("wallet")
	if !common.IsHexAddress(wallet) {
		fmt.Println("Invalid address format. Please provide a valid Ethereum address.")
		return
	}

	fmt.Println("Specify account ID (required):")
	fmt.Println(`> How to choose?
> If you want to see ledger entries associated with a wallet's unified balance, paste here the same wallet address.
> If you want to see ledger entries related to allocations in a specific app session of the participant having specified wallet address, paste here the corresponding app session ID.
> If you want to see ledger entries associated with wallet's channel escrow, paste here the channel ID.`)
	accountID := o.readExtraArg("account_id")

	fmt.Println("Specify asset symbol (required):")
	assetSuggestions := o.getAssetSuggestions("", 0)
	assetSymbol := o.readSelectionArg("asset_symbol", assetSuggestions)

	res, err := o.clearnode.GetLedgerEntries(wallet, accountID, assetSymbol, 0, ledgerEntriesPageSize)
	if err != nil {
		fmt.Printf("Failed to get ledger entries: %s\n", err.Error())
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Wallet", "Account", "Asset", "Debit", "Credit", "Timestamp"})
	t.AppendSeparator()

	totalCredit := decimal.NewFromInt(0)
	totalDebit := decimal.NewFromInt(0)
	for _, tx := range res.LedgerEntries {
		t.AppendRow(table.Row{tx.ID, tx.Participant, tx.AccountID, tx.Asset, fmtDec(tx.Debit), fmtDec(tx.Credit), tx.CreatedAt.Format(time.RFC3339)})
		totalCredit = totalCredit.Add(tx.Credit)
		totalDebit = totalDebit.Add(tx.Debit)
	}
	t.AppendFooter(table.Row{"", "", "", "TOTAL", fmtDec(totalDebit), fmtDec(totalCredit), ""})
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
