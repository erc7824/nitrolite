package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type WSNotifier struct {
	notify func(userID string, method string, params ...any)
}

func NewWSNotifier(notifyFunc func(userID string, method string, params ...any)) *WSNotifier {
	return &WSNotifier{
		notify: notifyFunc,
	}
}

func (n *WSNotifier) Notify(notifications ...*Notification) {
	for _, notification := range notifications {
		if notification != nil {
			n.notify(notification.userID, notification.eventType.String(), notification.data)
			notification.logger.Info(fmt.Sprintf("%s notification sent", notification.eventType), "userID", notification.userID, "data", notification.data)
		}
	}
}

type Notification struct {
	logger    Logger
	userID    string
	eventType EventType
	data      any
}

type EventType string

const (
	BalanceUpdateEventType EventType = "bu"
	ChannelUpdateEventType EventType = "cu"
	TransferEventType      EventType = "transfer"
)

func (e EventType) String() string {
	return string(e)
}

// NewBalanceNotification fetches the balance for a given wallet and creates a notification
func NewBalanceNotification(logger Logger, wallet string, db *gorm.DB) *Notification {
	senderAddress := common.HexToAddress(wallet)
	senderAccountID := NewAccountID(wallet)
	balances, err := GetWalletLedger(db, senderAddress).GetBalances(senderAccountID)
	if err != nil {
		logger.Error("error getting balances", "userID", wallet, "error", err)
		return nil
	}
	return &Notification{
		logger:    logger,
		userID:    wallet,
		eventType: BalanceUpdateEventType,
		data:      balances,
	}
}

// NewChannelNotification creates a notification for a channel update event
func NewChannelNotification(logger Logger, channel Channel) *Notification {
	return &Notification{
		logger:    logger,
		userID:    channel.Wallet,
		eventType: ChannelUpdateEventType,
		data: ChannelResponse{
			ChannelID:   channel.ChannelID,
			Participant: channel.Participant,
			Status:      channel.Status,
			Token:       channel.Token,
			RawAmount:   channel.RawAmount.BigInt(),
			ChainID:     channel.ChainID,
			Adjudicator: channel.Adjudicator,
			Challenge:   channel.Challenge,
			Nonce:       channel.Nonce,
			Version:     channel.Version,
			CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   channel.UpdatedAt.Format(time.RFC3339),
		},
	}
}

// NewTransferNotification creates a notification for a transfer event
func NewTransferNotification(logger Logger, wallet string, transferredAllocations []TransactionResponse) *Notification {
	return &Notification{
		logger:    logger,
		userID:    wallet,
		eventType: TransferEventType,
		data:      transferredAllocations,
	}
}
