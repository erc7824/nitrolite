package main

import (
	"math/big"
	"time"
)

// SendBalanceUpdate sends balance updates to the client
func (r *RPCRouter) SendBalanceUpdate(sender string) {
	balances, err := GetWalletLedger(r.DB, sender).GetBalances(sender)
	if err != nil {
		r.lg.Error("error getting balances", "sender", sender, "error", err)
		return
	}

	r.Node.Notify(sender, "bu", balances)
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
}
