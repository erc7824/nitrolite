package main

// ledger_publisher.go implements a publish-subscribe mechanism for ledger updates.
// It provides real-time notifications to clients when ledger entries are recorded.
// 
// The implementation features:
// - Thread-safe subscription management
// - Connection health monitoring with automatic cleanup
// - Efficient broadcasting to all subscribers
// - Support for both authenticated and anonymous subscribers
// - Proper error handling and resource management

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

// LedgerPublisher manages subscriptions to ledger updates
type LedgerPublisher struct {
	subscribers      map[string]*websocket.Conn
	subscribersMutex sync.RWMutex
	signer          *Signer
	cleanupTicker   *time.Ticker
	stopChan        chan struct{}
}

// NewLedgerPublisher creates a new ledger publisher
func NewLedgerPublisher(signer *Signer) *LedgerPublisher {
	publisher := &LedgerPublisher{
		subscribers:    make(map[string]*websocket.Conn),
		signer:         signer,
		cleanupTicker:  time.NewTicker(30 * time.Second),
		stopChan:       make(chan struct{}),
	}
	
	// Start the background health check
	go publisher.runHealthChecks()
	
	return publisher
}

// Subscribe adds a new subscriber to ledger updates
func (p *LedgerPublisher) Subscribe(subscriberID string, conn *websocket.Conn) {
	p.subscribersMutex.Lock()
	defer p.subscribersMutex.Unlock()
	
	// Check if already subscribed
	if existingConn, exists := p.subscribers[subscriberID]; exists {
		if existingConn != conn {
			// Close old connection if it's different
			existingConn.WriteControl(
				websocket.CloseMessage, 
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "replaced by new connection"),
				time.Now().Add(5 * time.Second),
			)
			log.Printf("Replacing existing subscription for %s", subscriberID)
		} else {
			log.Printf("Client %s already subscribed", subscriberID)
			return
		}
	}
	
	// Setup ping handler to keep connection alive
	conn.SetPingHandler(func(message string) error {
		err := conn.WriteControl(
			websocket.PongMessage, 
			[]byte(message),
			time.Now().Add(5 * time.Second),
		)
		if err != nil {
			log.Printf("Error sending pong to %s: %v", subscriberID, err)
		}
		return nil
	})
	
	p.subscribers[subscriberID] = conn
	log.Printf("Client %s subscribed to ledger updates", subscriberID)
}

// Unsubscribe removes a subscriber from ledger updates
func (p *LedgerPublisher) Unsubscribe(subscriberID string) {
	p.subscribersMutex.Lock()
	delete(p.subscribers, subscriberID)
	p.subscribersMutex.Unlock()
	log.Printf("Client %s unsubscribed from ledger updates", subscriberID)
}

// Stop shuts down the publisher and releases resources
func (p *LedgerPublisher) Stop() {
	close(p.stopChan)
	if p.cleanupTicker != nil {
		p.cleanupTicker.Stop()
	}
	
	// Close all subscriber connections
	p.subscribersMutex.Lock()
	for id, conn := range p.subscribers {
		conn.WriteControl(
			websocket.CloseMessage, 
			websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
			time.Now().Add(5 * time.Second),
		)
		log.Printf("Closing connection for subscriber: %s", id)
	}
	p.subscribers = make(map[string]*websocket.Conn)
	p.subscribersMutex.Unlock()
}

// runHealthChecks periodically checks connections and removes dead ones
func (p *LedgerPublisher) runHealthChecks() {
	for {
		select {
		case <-p.cleanupTicker.C:
			p.checkConnections()
		case <-p.stopChan:
			return
		}
	}
}

// checkConnections verifies all subscriber connections are healthy
func (p *LedgerPublisher) checkConnections() {
	deadSubscribers := []string{}
	
	p.subscribersMutex.Lock()
	for id, conn := range p.subscribers {
		// Send ping to check if connection is alive
		err := conn.WriteControl(
			websocket.PingMessage, 
			[]byte{}, 
			time.Now().Add(5 * time.Second),
		)
		if err != nil {
			log.Printf("Detected dead connection for %s: %v", id, err)
			deadSubscribers = append(deadSubscribers, id)
		}
	}
	
	// Remove dead connections
	for _, id := range deadSubscribers {
		delete(p.subscribers, id)
		log.Printf("Removed unresponsive subscriber: %s", id)
	}
	p.subscribersMutex.Unlock()
}

// PublishEntry broadcasts a ledger entry to all subscribers
func (p *LedgerPublisher) PublishEntry(entry *Entry) {
	response := LedgerEntryResponse{
		ID:          entry.ID,
		AccountID:   entry.AccountID,
		AccountType: entry.AccountType,
		Asset:       entry.AssetSymbol,
		Participant: entry.Wallet,
		Credit:      entry.Credit,
		Debit:       entry.Debit,
		CreatedAt:   entry.CreatedAt,
	}

	// Create RPC response
	rpcResponse := CreateResponse(uint64(time.Now().UnixMilli()), "ledger_update", []any{response}, time.Now())

	// Sign the response
	resBytes, err := json.Marshal(rpcResponse.Res)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		return
	}
	
	signature, err := p.signer.Sign(resBytes)
	if err != nil {
		log.Printf("Error signing response: %v", err)
		return
	}
	
	// Convert the signature to hexadecimal format
	hexSignature := hexutil.Encode(signature)
	rpcResponse.Sig = []string{hexSignature}

	// Marshal the response
	responseData, err := json.Marshal(rpcResponse)
	if err != nil {
		log.Printf("Error marshaling ledger update: %v", err)
		return
	}

	// Broadcast to all subscribers
	p.subscribersMutex.RLock()
	defer p.subscribersMutex.RUnlock()

	deadSubscribers := []string{}
	for subscriberID, conn := range p.subscribers {
		// Set a write deadline to prevent blocking on slow clients
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

		err := conn.WriteMessage(websocket.TextMessage, responseData)
		if err != nil {
			log.Printf("Error sending ledger update to %s: %v", subscriberID, err)
			deadSubscribers = append(deadSubscribers, subscriberID)
			continue
		}

		// Reset the write deadline
		conn.SetWriteDeadline(time.Time{})
	}
	
	// Clean up any dead subscribers outside the read lock
	if len(deadSubscribers) > 0 {
		go func(toRemove []string) {
			for _, id := range toRemove {
				p.Unsubscribe(id)
			}
		}(deadSubscribers)
	}
}