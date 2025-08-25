package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type StateIntent uint8

const (
	StateIntentOperate    StateIntent = 0 // Operate the state application
	StateIntentInitialize StateIntent = 1 // Initial funding state
	StateIntentResize     StateIntent = 2 // Resize state
	StateIntentFinalize   StateIntent = 3 // Final closing state
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

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into UnsignedState", value)
	}

	return json.Unmarshal(bytes, u)
}

type Allocation struct {
	Participant  string          `json:"destination"`
	TokenAddress string          `json:"token"`
	RawAmount    decimal.Decimal `json:"amount"`
}
