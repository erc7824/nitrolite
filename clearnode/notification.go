package main

import (
	"fmt"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type WSNotifier struct {
	notify func(userID string, method string, params RPCDataParams)
	logger Logger
}

func NewWSNotifier(notifyFunc func(userID string, method string, params RPCDataParams), logger Logger) *WSNotifier {
	return &WSNotifier{
		notify: notifyFunc,
		logger: logger,
	}
}

func (n *WSNotifier) Notify(notifications ...*Notification) {
	for _, notification := range notifications {
		if notification != nil {
			n.notify(notification.userID, notification.eventType.String(), notification.data)
			if n.logger != nil {
				n.logger.Info(fmt.Sprintf("%s notification sent", notification.eventType), "userID", notification.userID, "data", notification.data)
			}
		}
	}
}

type Notification struct {
	userID    string
	eventType EventType
	data      any
}

type EventType string

const (
	BalanceUpdateEventType    EventType = "bu"
	ChannelUpdateEventType    EventType = "cu"
	TransferEventType         EventType = "tr"
	AppSessionUpdateEventType EventType = "asu"
)

func (e EventType) String() string {
	return string(e)
}

// NewBalanceNotification fetches the balance for a given wallet and creates a notification
func NewBalanceNotification(wallet string, db *gorm.DB) *Notification {
	balances, _ := GetWalletLedger(db, common.HexToAddress(wallet)).GetBalances(NewAccountID(wallet))
	return &Notification{
		userID:    wallet,
		eventType: BalanceUpdateEventType,
		data:      BalanceUpdatesResponse{BalanceUpdates: balances},
	}
}

// NewChannelNotification creates a notification for a channel update event
func NewChannelNotification(channel Channel) *Notification {
	return &Notification{
		userID:    channel.Wallet,
		eventType: ChannelUpdateEventType,
		data: ChannelResponse{
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
		},
	}
}

// NewTransferNotification creates a notification for a transfer event
func NewTransferNotification(wallet string, transferredAllocations TransferResponse) *Notification {
	return &Notification{
		userID:    wallet,
		eventType: TransferEventType,
		data:      transferredAllocations,
	}
}

// NewAppSessionNotification creates a notification for an app session update event
func NewAppSessionNotification(participant string, appSession AppSession, participantAllocations []AppAllocation) *Notification {
	appSessionData := rpc.AppSession{
		AppSessionID:       appSession.SessionID,
		Status:             string(appSession.Status),
		ParticipantWallets: appSession.ParticipantWallets,
		SessionData:        appSession.SessionData,
		Protocol:           appSession.Protocol,
		Challenge:          appSession.Challenge,
		Weights:            appSession.Weights,
		Quorum:             appSession.Quorum,
		Version:            appSession.Version,
		Nonce:              appSession.Nonce,
		CreatedAt:          appSession.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          appSession.UpdatedAt.Format(time.RFC3339),
	}

	rpcAllocations := make([]rpc.AppAllocation, len(participantAllocations))
	for i, alloc := range participantAllocations {
		rpcAllocations[i] = rpc.AppAllocation{
			Participant: alloc.Participant,
			AssetSymbol: alloc.AssetSymbol,
			Amount:      alloc.Amount,
		}
	}

	return &Notification{
		userID:    participant,
		eventType: AppSessionUpdateEventType,
		data: rpc.AppSessionUpdateNotification{
			AppSession:             appSessionData,
			ParticipantAllocations: rpcAllocations,
		},
	}
}
