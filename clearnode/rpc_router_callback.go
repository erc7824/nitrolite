package main

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type BalanceUpdatesResponse struct {
	BalanceUpdates []Balance `json:"balance_updates"`
}

type ChannelsResponse struct {
	Channels []ChannelResponse `json:"channels"`
}

type AssetsResponse struct {
	Assets []AssetResponse `json:"assets"`
}

// SendBalanceUpdate sends balance updates to the client
func (r *RPCRouter) SendBalanceUpdate(destinationWallet string) {
	senderAddress := common.HexToAddress(destinationWallet)
	senderAccountID := NewAccountID(destinationWallet)
	balances, err := GetWalletLedger(r.DB, senderAddress).GetBalances(senderAccountID)
	if err != nil {
		r.lg.Error("error getting balances", "userID", destinationWallet, "error", err)
		return
	}

	r.Node.Notify(destinationWallet, "bu", BalanceUpdatesResponse{BalanceUpdates: balances})
	r.lg.Info("balance update sent", "userID", destinationWallet, "balances", balances)
}

// SendChannelUpdate sends a single channel update to the client
func (r *RPCRouter) SendChannelUpdate(channel Channel) {
	channelResponse := ChannelResponse{
		ChannelID:   channel.ChannelID,
		Participant: channel.Participant,
		Status:      channel.Status,
		Token:       channel.Token,
		RawAmount:   channel.RawAmount,
		ChainID:     channel.ChainID,
		Adjudicator: channel.Adjudicator,
		Challenge:   channel.Challenge,
		Nonce:       channel.Nonce,
		Version:     channel.State.Version,
		CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
	}

	r.Node.Notify(channel.Wallet, "cu", channelResponse)
	r.lg.Info("channel update sent",
		"userID", channel.Wallet,
		"channelID", channel.ChannelID,
		"participant", channel.Participant,
		"status", channel.Status,
	)
}

// TODO: make adequate notifications response/type
// SendTransferNotification sends a transfer notification to the client
func (r *RPCRouter) SendTransferNotification(destinationWallet string, transferredAllocations TransferResponse) {
	r.Node.Notify(destinationWallet, "tr", transferredAllocations)
	r.lg.Info("transfer notification sent", "userID", destinationWallet, "transferred allocations", transferredAllocations)
}
