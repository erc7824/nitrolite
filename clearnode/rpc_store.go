package main

import (
	"encoding/json"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
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
func (s *RPCStore) StoreMessage(sender string, req *RPCData, reqSigs []Signature, resBytes []byte, resSigs []Signature) error {
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
		ReqSig:    nitrolite.SignaturesToStrings(reqSigs),
		ResSig:    nitrolite.SignaturesToStrings(resSigs),
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
