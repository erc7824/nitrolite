package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// HandleLedgerSubscription sets up a subscription to ledger updates
func (h *UnifiedWSHandler) HandleLedgerSubscription(conn *websocket.Conn, rpc *RPCMessage, subscriberID string) (*RPCMessage, error) {
	// Setup ping handler to keep connection alive
	conn.SetPingHandler(func(message string) error {
		err := conn.WriteControl(
			websocket.PongMessage,
			[]byte(message),
			time.Now().Add(5*time.Second),
		)
		if err != nil {
			log.Printf("Error sending pong to %s: %v", subscriberID, err)
		}
		return nil
	})

	// Register the subscriber
	h.ledgerPublisher.Subscribe(subscriberID, conn)

	// We don't need a separate goroutine here since LedgerPublisher now has
	// its own connection health checks

	// Return success response
	response := map[string]interface{}{
		"subscribed": true,
		"timestamp":  time.Now().UnixMilli(),
	}

	return CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now()), nil
}

// SendLedgerUpdate sends a single ledger entry update to subscribers through the publisher
func (h *UnifiedWSHandler) PublishLedgerEntry(entry *Entry) {
	if h.ledgerPublisher != nil {
		h.ledgerPublisher.PublishEntry(entry)
	}
}
