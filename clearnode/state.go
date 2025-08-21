package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type StateIntent uint8

const (
	StateIntentOperate    uint8 = 0 // Operate the state application
	StateIntentInitialize uint8 = 1 // Initial funding state
	StateIntentResize     uint8 = 2 // Resize state
	StateIntentFinalize   uint8 = 3 // Final closing state
)

type UnsignedState struct {
	Intent      StateIntent  `json:"intent"`
	Version     uint64       `json:"version"`
	Data        string       `json:"state_data"`
	Allocations []Allocation `json:"allocations"`
}

// Value implements driver.Valuer interface for database storage
func (u UnsignedState) Value() (driver.Value, error) {
	return json.Marshal(u)
}

// Scan implements sql.Scanner interface for database retrieval
func (u *UnsignedState) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into UnsignedState", value)
	}

	return json.Unmarshal(bytes, u)
}

type Allocation struct {
	Participant  string          `json:"destination"`
	TokenAddress string          `json:"token"`
	RawAmount    decimal.Decimal `json:"amount"`
}
