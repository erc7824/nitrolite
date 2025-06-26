package main

import (
	"math/big"
	"time"

	"github.com/shopspring/decimal"
)

// SendBalanceUpdate sends balance updates to the client
func (r *RPCRouter) SendBalanceUpdate(sender string) {
	balances, err := GetWalletLedger(r.DB, sender).GetBalances(sender)
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

type AppDepositUpdate struct {
	sessionID   string
	participant string
	allocations []TransferAllocation
}

// SendAppDepositUpdate sends a deposit update to all participants in the app session
func (r *RPCRouter) SendDepositUpdate(appSession AppSession, appDepositUpdate AppDepositUpdate) {
	for _, participant := range appSession.ParticipantWallets {
		walletAddress := participant
		if wallet := GetWalletBySigner(participant); wallet != "" {
			walletAddress = wallet
		}

		for _, allocation := range appDepositUpdate.allocations {
			depositUpdate := struct {
				SessionID      string          `json:"session_id"`
				AssetID        string          `json:"asset"`
				DepositAmount  decimal.Decimal `json:"deposit_amount"`
				UpdatedBalance decimal.Decimal `json:"updated_balance"`
				AppVersion     uint64          `json:"app_version"`
			}{
				SessionID:     appSession.SessionID,
				AssetID:       allocation.AssetSymbol,
				DepositAmount: allocation.Amount,
				AppVersion:    appSession.Version,
			}
			r.Node.Notify(walletAddress, "app_deposit", depositUpdate)
			r.lg.Info("deposit update sent",
				"userID", walletAddress,
				"participant", participant,
				"sessionID", appSession.SessionID,
				"version", appSession.Version,
				"asset", depositUpdate.AssetID,
				"depositAmount", depositUpdate.DepositAmount.String(),
				"new_balance", depositUpdate.UpdatedBalance.String(),
			)
		}
	}
}
