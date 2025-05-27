package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRPCStoreNew tests the creation of a new RPCStore instance
func TestRPCStoreNew(t *testing.T) {
	// Set up test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)
	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

// TestRPCStoreStoreMessage tests storing an RPC message in the database
func TestRPCStoreStoreMessage(t *testing.T) {
	// Set up test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)

	// Create test data
	sender := "0xSender123"
	timestamp := uint64(time.Now().Unix())
	reqID := uint64(12345)
	method := "test_method"
	params := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	reqSig := []string{"sig1", "sig2"}
	resBytes := []byte(`{"result": "ok"}`)
	resSig := []string{"resSig1"}

	// Create RPCData
	req := &RPCData{
		RequestID: reqID,
		Method:    method,
		Params:    []any{params},
		Timestamp: timestamp,
	}

	// Store the message
	err := store.StoreMessage(sender, req, reqSig, resBytes, resSig)
	require.NoError(t, err)

	// Verify the message was stored correctly
	var count int64
	err = db.Model(&RPCRecord{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var record RPCRecord
	err = db.First(&record).Error
	require.NoError(t, err)

	assert.Equal(t, sender, record.Sender)
	assert.Equal(t, reqID, record.ReqID)
	assert.Equal(t, method, record.Method)
	assert.Equal(t, timestamp, record.Timestamp)
	assert.ElementsMatch(t, reqSig, record.ReqSig)
	assert.ElementsMatch(t, resSig, record.ResSig)
	assert.Equal(t, resBytes, record.Response)

	// Verify params were stored correctly
	// The params are stored as a JSON array since it comes from req.Params which is []any
	var storedParamsArray []map[string]interface{}
	err = json.Unmarshal(record.Params, &storedParamsArray)
	require.NoError(t, err)
	require.Len(t, storedParamsArray, 1)

	// Extract the first element which is our original map
	storedParams := storedParamsArray[0]
	assert.Equal(t, "value1", storedParams["key1"])
	assert.Equal(t, float64(42), storedParams["key2"]) // JSON unmarshals numbers as float64
}

// TestRPCStoreGetMessages tests retrieving RPC messages with pagination
func TestRPCStoreGetMessages(t *testing.T) {
	// Set up test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)

	// Create test data - 10 messages with different timestamps
	baseTime := time.Now().Unix()

	for i := 0; i < 10; i++ {
		record := &RPCRecord{
			Sender:    "0xSender" + string(rune('A'+i)),
			ReqID:     uint64(1000 + i),
			Method:    "method" + string(rune('A'+i)),
			Params:    []byte(`{"test":"data"}`),
			Response:  []byte(`{"result":"ok"}`),
			ReqSig:    []string{"reqSig" + string(rune('A'+i))},
			ResSig:    []string{"resSig" + string(rune('A'+i))},
			Timestamp: uint64(baseTime - int64(i)), // Descending timestamp order
		}
		require.NoError(t, db.Create(record).Error)
	}

	// Test pagination - first page (3 items)
	messages, total, err := store.GetMessages(3, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, messages, 3)

	// Most recent should be first due to timestamp DESC order
	assert.Equal(t, uint64(1000), messages[0].ReqID)
	assert.Equal(t, uint64(1001), messages[1].ReqID)
	assert.Equal(t, uint64(1002), messages[2].ReqID)

	// Test pagination - second page (3 items)
	messages, total, err = store.GetMessages(3, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, messages, 3)
	assert.Equal(t, uint64(1003), messages[0].ReqID)
	assert.Equal(t, uint64(1004), messages[1].ReqID)
	assert.Equal(t, uint64(1005), messages[2].ReqID)

	// Test pagination - last page (partial)
	messages, total, err = store.GetMessages(3, 9)
	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, messages, 1)
	assert.Equal(t, uint64(1009), messages[0].ReqID)

	// Test pagination - beyond available records
	messages, total, err = store.GetMessages(3, 12)
	require.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, messages, 0)
}

// TestRPCStoreGetMessagesError tests error handling for GetMessages by corrupting the database schema
func TestRPCStoreGetMessagesError(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create store
	store := NewRPCStore(db)

	// Corrupt the schema by dropping the table
	err := db.Exec("DROP TABLE rpc_store").Error
	require.NoError(t, err)

	// Now try to get messages from the non-existent table
	messages, total, err := store.GetMessages(10, 0)
	assert.Error(t, err)
	assert.Empty(t, messages)
	assert.Equal(t, int64(0), total)
}

// TestRPCStoreGetMessageByID tests retrieving an RPC message by its request ID
func TestRPCStoreGetMessageByID(t *testing.T) {
	// Set up test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)

	// Create test data - several messages with different request IDs
	reqIDs := []uint64{1001, 1002, 1003}

	for i, reqID := range reqIDs {
		record := &RPCRecord{
			Sender:    "0xSender" + string(rune('A'+i)),
			ReqID:     reqID,
			Method:    "method" + string(rune('A'+i)),
			Params:    []byte(`{"test":"data"}`),
			Response:  []byte(`{"result":"ok"}`),
			ReqSig:    []string{"reqSig" + string(rune('A'+i))},
			ResSig:    []string{"resSig" + string(rune('A'+i))},
			Timestamp: uint64(time.Now().Unix() - int64(i)),
		}
		require.NoError(t, db.Create(record).Error)
	}

	// Test retrieval by existing ID
	message, err := store.GetMessageByID(1002)
	require.NoError(t, err)
	assert.NotNil(t, message)
	assert.Equal(t, uint64(1002), message.ReqID)
	assert.Equal(t, "0xSenderB", message.Sender)
	assert.Equal(t, "methodB", message.Method)

	// Test retrieval by non-existent ID
	message, err = store.GetMessageByID(9999)
	assert.Error(t, err)
	assert.Nil(t, message)
	assert.Contains(t, err.Error(), "record not found")
}

// TestRPCStoreStoreMessageError tests error handling for StoreMessage
func TestRPCStoreStoreMessageError(t *testing.T) {
	// Set up test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)

	// Create test data with invalid params that cannot be marshalled
	sender := "0xSender123"
	req := &RPCData{
		RequestID: 12345,
		Method:    "test_method",
		Params:    []any{make(chan int)}, // Channels cannot be marshalled to JSON
		Timestamp: uint64(time.Now().Unix()),
	}
	reqSig := []string{"sig1"}
	resBytes := []byte(`{"result": "ok"}`)
	resSig := []string{"resSig1"}

	// Attempt to store the message, should fail due to unmarshalable params
	err := store.StoreMessage(sender, req, reqSig, resBytes, resSig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}
