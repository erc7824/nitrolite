package main

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// SendBalanceUpdate sends balance updates to the client
func (r *RPCRouter) SendBalanceUpdate(sender string) {
	senderAddress := common.HexToAddress(sender)
	senderAccountID := NewAccountID(sender)
	balances, err := GetWalletLedger(r.DB, senderAddress).GetBalances(senderAccountID)
	if err != nil {
		r.lg.Error("error getting balances", "sender", sender, "error", err)
		return
	}

	r.Node.Notify(sender, "bu", balances)
	r.lg.Info("balance update sent", "userID", sender, "balances", balances)
}

// SendChannelUpdate sends a single channel update to the client
func (r *RPCRouter) SendChannelUpdate(channel Channel) {
	channelResponse := ChannelResponse{
		ChannelID:   channel.ChannelID,
		Participant: channel.Participant,
		Status:      channel.Status,
		Token:       channel.Token,
		Amount:      big.NewInt(int64(channel.Amount)),
		ChainID:     channel.ChainID,
		Adjudicator: channel.Adjudicator,
		Challenge:   channel.Challenge,
		Nonce:       channel.Nonce,
		Version:     channel.Version,
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
