package database

import (
	"encoding/json"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"gorm.io/gorm"
)

type RPCMessageType int

const (
	RPCMessageTypeRequest  RPCMessageType = 1
	RPCMessageTypeResponse RPCMessageType = 2
	RPCMessageTypeEvent    RPCMessageType = 3
)

// RPCRecord represents an RPC message in the database
type RPCRecord struct {
	ID        uint           `gorm:"primaryKey"`
	ReqID     uint64         `gorm:"column:req_id;not null"`
	MsgType   RPCMessageType `gorm:"column:msg_type;not null"` // 1 for request, 2 for response, 3 for event
	Method    string         `gorm:"column:method;type:varchar(255);not null"`
	Payload   []byte         `gorm:"column:params;type:text;not null"`
	Timestamp uint64         `gorm:"column:timestamp;not null"`
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
func (s *RPCStore) StoreMessage(sender string, req *rpc.Payload, reqSigs []sign.Signature, resBytes []byte, resSigs []sign.Signature) error {
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return err
	}

	msg := &RPCRecord{
		ReqID:     req.RequestID,
		MsgType:   RPCMessageTypeRequest, // TODO: record response and events as well
		Method:    req.Method,
		Payload:   paramsBytes,
		Timestamp: req.Timestamp,
	}

	return s.db.Create(msg).Error
}

// GetRPCHistory retrieves RPC history for a specific user with pagination
func (s *RPCStore) GetRPCHistory(userWallet string, options *ListOptions) ([]RPCRecord, error) {
	query := applyListOptions(s.db, "timestamp", SortTypeDescending, options)
	var rpcHistory []RPCRecord
	err := query.Where("sender = ?", userWallet).Find(&rpcHistory).Error
	return rpcHistory, err
}
