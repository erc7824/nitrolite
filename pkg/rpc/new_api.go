package rpc

import "github.com/shopspring/decimal"

// PaginationParams represents pagination request parameters
type PaginationParams struct {
	Offset *uint   `json:"offset,omitempty"` // Pagination offset (number of items to skip)
	Limit  *uint   `json:"limit,omitempty"`  // Number of items to return
	Sort   *string `json:"sort,omitempty"`   // Sort order (asc/desc)
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       uint `json:"page"`        // Current page number
	PerPage    uint `json:"per_page"`    // Number of items per page
	TotalCount uint `json:"total_count"` // Total number of items
	PageCount  uint `json:"page_count"`  // Total number of pages
}

// BalanceEntry represents balance for a specific asset
type BalanceEntry struct {
	Asset  string          `json:"asset"`  // Asset symbol
	Amount decimal.Decimal `json:"amount"` // Balance amount
}

// BlockchainInfo describes a supported blockchain network.
type BlockchainInfo struct {
	// ID is the network's chain identifier
	ID uint32 `json:"chain_id"`
	// Name is the human-readable name of the blockchain
	Name string `json:"name"` // TODO: add to SDK
	// CustodyAddress is the custody contract address
	CustodyAddress string `json:"custody_address"`
	// AdjudicatorAddress is the adjudicator contract address
	AdjudicatorAddress string `json:"adjudicator_address"`
}
