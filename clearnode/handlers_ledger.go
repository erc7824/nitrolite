package main

// handlers_ledger.go contains the RPC handler functions for ledger-related operations.
// This includes the handler for subscribing to ledger updates, which allows clients
// to receive real-time notifications when ledger entries are recorded.
//
// The handlers in this file integrate with the ledger publishing system implemented
// in ledger_publisher.go and ws_ledger.go.

import (
	"time"
)

// HandleSubscribeLedger creates a subscription for ledger updates
func HandleSubscribeLedger(rpc *RPCMessage) (*RPCMessage, error) {
	// Currently, we don't need any parameters for subscription
	// Could be extended later to support filtering by wallet, asset, etc.

	response := map[string]interface{}{
		"subscribed": true,
		"timestamp":  time.Now().UnixMilli(),
	}

	rpcResponse := CreateResponse(rpc.Req.RequestID, rpc.Req.Method, []any{response}, time.Now())
	return rpcResponse, nil
}
