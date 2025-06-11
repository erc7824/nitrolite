package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrEventHasAlreadyBeenProcessed = errors.New("contract event has already been processed")

type ContractEvent struct {
	ID              int64          `gorm:"primary_key;column:id"`
	ContractAddress string         `gorm:"column:contract_address"`
	ChainID         uint32         `gorm:"column:chain_id"`
	Name            string         `gorm:"column:name"`
	BlockNumber     uint64         `gorm:"column:block_number"`
	TransactionHash string         `gorm:"column:transaction_hash"`
	LogIndex        uint32         `gorm:"column:log_index"`
	Data            datatypes.JSON `gorm:"column:data"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
}

func (ContractEvent) TableName() string {
	return "contract_events"
}

func StoreContractEvent(tx *gorm.DB, event *ContractEvent) error {
	// Skip if the event has already been processed
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(event).Error
}

func MarshalCustodyResized(event nitrolite.CustodyResized) ([]byte, error) {
	eventCopy := event
	eventCopy.Raw = types.Log{}

	encodedData, err := json.Marshal(eventCopy)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func MarshalCustodyChallenged(event nitrolite.CustodyChallenged) ([]byte, error) {
	eventCopy := event
	eventCopy.Raw = types.Log{}

	encodedData, err := json.Marshal(eventCopy)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func MarshalCustodyClosed(event nitrolite.CustodyClosed) ([]byte, error) {
	eventCopy := event
	eventCopy.Raw = types.Log{}

	encodedData, err := json.Marshal(eventCopy)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func MarshalCustodyJoined(event nitrolite.CustodyJoined) ([]byte, error) {
	eventCopy := event
	eventCopy.Raw = types.Log{}

	encodedData, err := json.Marshal(eventCopy)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func MarshalCustodyCreated(event nitrolite.CustodyCreated) ([]byte, error) {
	eventCopy := event
	eventCopy.Raw = types.Log{}

	encodedData, err := json.Marshal(eventCopy)
	if err != nil {
		return nil, err
	}

	return encodedData, nil
}

func GetLatestContractEvent(db *gorm.DB, contractAddress string, networkID uint32) (*ContractEvent, error) {
	var ev ContractEvent
	err := db.Where("chain_id = ? AND contract_address = ?", networkID, contractAddress).Order("block_number DESC, log_index DESC").First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &ev, err
}
