package main

import (
	"encoding/json"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// RPCRecord represents an RPC message in the database
type RPCRecord struct {
	ID        uint           `gorm:"primaryKey"`
	Sender    string         `gorm:"column:sender;type:varchar(255);not null"`
	ReqID     uint64         `gorm:"column:req_id;not null"`
	Method    string         `gorm:"column:method;type:varchar(255);not null"`
	Params    []byte         `gorm:"column:params;type:text;not null"`
	Timestamp uint64         `gorm:"column:timestamp;not null"`
	ReqSig    pq.StringArray `gorm:"type:text[];column:req_sig;"`
	Response  []byte         `gorm:"column:response;type:text;not null"`
	ResSig    pq.StringArray `gorm:"type:text[];column:res_sig;"`
}

// TableName specifies the table name for the RPCMessageDB model
func (RPCRecord) TableName() string {
	return "rpc_store"
}

// RPCStore handles RPC message storage and retrieval
type RPCStore struct {
	db *gorm.DB
}

// NewRPCStore creates a new RPCStore instance
func NewRPCStore(db *gorm.DB) *RPCStore {
	return &RPCStore{db: db}
}

// StoreMessage stores an RPC message in the database
func (s *RPCStore) StoreMessage(sender string, req *RPCData, reqSig []string, resBytes []byte, resSig []string) error {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return err
	}

	msg := &RPCRecord{
		ReqID:     req.RequestID,
		Sender:    sender,
		Method:    req.Method,
		Params:    paramsBytes,
		Response:  resBytes,
		ReqSig:    reqSig,
		ResSig:    resSig,
		Timestamp: req.Timestamp,
	}

	return s.db.Create(msg).Error
}

// GetMessages retrieves RPC messages from the database with pagination
func (s *RPCStore) GetMessages(limit int, offset int) (messages []RPCRecord, total int64, err error) {
	// Get total count
	if err := s.db.Model(&RPCRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated messages
	err = s.db.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, total, err
}

// GetMessageByID retrieves a specific RPC message by its request ID
func (s *RPCStore) GetMessageByID(reqID uint64) (*RPCRecord, error) {
	var message RPCRecord
	err := s.db.Where("req_id = ?", reqID).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}
