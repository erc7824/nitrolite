package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
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
	return tx.Create(event).Error
}

func MarshalEvent[T any](event T) ([]byte, error) {
	val := reflect.ValueOf(event)
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct, but got %T", event)
	}

	copyVal := reflect.New(val.Type()).Elem()
	copyVal.Set(val)

	// This is equivalent to `eventCopy.Raw = types.Log{}`.
	rawField := copyVal.FieldByName("Raw")
	if rawField.IsValid() {
		if !rawField.CanSet() {
			return nil, fmt.Errorf("cannot set 'Raw' field on type %s", val.Type())
		}
		zeroValue := reflect.Zero(rawField.Type())
		rawField.Set(zeroValue)
	}
	return json.Marshal(copyVal.Interface())
}

func GetLatestContractEvent(db *gorm.DB, contractAddress string, networkID uint32) (*ContractEvent, error) {
	var ev ContractEvent
	err := db.Where("chain_id = ? AND contract_address = ?", networkID, contractAddress).Order("block_number DESC, log_index DESC").First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &ev, err
}

func IsContractEventPresent(db *gorm.DB, chainID uint32, txHash string, logIndex uint32) (bool, error) {
	var count int64
	err := db.Model(&ContractEvent{}).
		Where("chain_id = ? AND transaction_hash = ? AND log_index = ?", chainID, txHash, logIndex).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}
	return false, nil
}
