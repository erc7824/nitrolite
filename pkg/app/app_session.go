package app

import "github.com/shopspring/decimal"

// AppSessionInfo represents information about an application session
type AppSessionInfo struct {
	AppSessionID string           `json:"app_session_id"`         // A unique application session identifier
	Status       string           `json:"status"`                 // Session status (open/closed)
	Participants []AppParticipant `json:"participants"`           // List of participant wallet addresses with weights
	SessionData  *string          `json:"session_data,omitempty"` // JSON stringified session data
	Quorum       uint             `json:"quorum"`                 // Quorum required for operations
	Version      uint             `json:"version"`                // Current version of the session state
	Nonce        uint             `json:"nonce"`                  // Nonce for the session
	Allocations  []AppAllocation  `json:"allocations"`            // List of allocations in the app state
}

// AppParticipant represents definition for an app participant
type AppParticipant struct {
	WalletAddress   string `json:"wallet_address"`   // Participant's wallet address
	SignatureWeight uint   `json:"signature_weight"` // Signature weight for the participant
}

// AppDefinition represents definition for an app session
type AppDefinition struct {
	Application  string           `json:"application"`  // Application identifier from an app registry
	Participants []AppParticipant `json:"participants"` // List of participants in the app session
	Quorum       uint             `json:"quorum"`       // Quorum required for the app session
	Nonce        uint             `json:"nonce"`        // A unique number to prevent replay attacks
}

// AppAllocation represents allocation of assets to a participant in an app session
type AppAllocation struct {
	Participant string          `json:"participant"` // Participant's wallet address
	Asset       string          `json:"asset"`       // Asset symbol
	Amount      decimal.Decimal `json:"amount"`      // Amount allocated to the participant
}

// AppStateUpdate represents the current state of an application session
type AppStateUpdate struct {
	AppSessionID string          `json:"app_session_id"` // A unique application session identifier
	Intent       string          `json:"intent"`         // The intent of the app session update (operate, deposit, withdraw)
	Version      uint            `json:"version"`        // Version of the app state
	Allocations  []AppAllocation `json:"allocations"`    // List of allocations in the app state
	SessionData  string          `json:"session_data"`   // JSON stringified session data
}
