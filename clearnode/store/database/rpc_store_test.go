package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

// Helper function to convert signatures to strings for testing
func signaturesToStrings(sigs []sign.Signature) []string {
	result := make([]string, len(sigs))
	for i, sig := range sigs {
		result[i] = hexutil.Encode(sig)
	}
	return result
}

// TestRPCStoreNew tests the creation of a new RPCStore instance
func TestRPCStoreNew(t *testing.T) {
	// Set up test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)
	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

// // TestRPCStoreStoreMessage tests storing an RPC message in the database
// func TestRPCStoreStoreMessage(t *testing.T) {
// 	// Set up test database
// 	db, cleanup := SetupTestDB(t)
// 	defer cleanup()

// 	// Create a new RPCStore
// 	store := NewRPCStore(db)

// 	// Create test data
// 	sender := "0xSender123"
// 	timestamp := uint64(time.Now().Unix())
// 	reqID := uint64(12345)
// 	method := "test_method"
// 	paramsJSON := []byte(`{"key1":"value1","key2":42}`)
// 	reqSig := []sign.Signature{sign.Signature(hexutil.MustDecode("0x1234")), sign.Signature(hexutil.MustDecode("0x4567"))}
// 	resBytes := []byte(`{"result": "ok"}`)
// 	resSig := []sign.Signature{sign.Signature(hexutil.MustDecode("0x4321"))}

// 	// Create rpc.Params from JSON
// 	var params rpc.Params
// 	err := json.Unmarshal(paramsJSON, &params)
// 	require.NoError(t, err)

// 	// Create rpc.Payload
// 	req := &rpc.Payload{
// 		RequestID: reqID,
// 		Method:    method,
// 		Params:    params,
// 		Timestamp: timestamp,
// 	}

// 	// Store the message
// 	err = store.StoreMessage(sender, req, reqSig, resBytes, resSig)
// 	require.NoError(t, err)

// 	// Verify the message was stored correctly
// 	var count int64
// 	err = db.Model(&RPCRecord{}).Count(&count).Error
// 	require.NoError(t, err)
// 	assert.Equal(t, int64(1), count)

// 	var record RPCRecord
// 	err = db.First(&record).Error
// 	require.NoError(t, err)

// 	assert.Equal(t, sender, record.Sender)
// 	assert.Equal(t, reqID, record.ReqID)
// 	assert.Equal(t, method, record.Method)
// 	assert.Equal(t, timestamp, record.Timestamp)

// 	assert.ElementsMatch(t, signaturesToStrings(reqSig), record.ReqSig)
// 	assert.ElementsMatch(t, signaturesToStrings(resSig), record.ResSig)
// 	assert.Equal(t, resBytes, record.Response)

// 	// Verify params were stored correctly
// 	// The params are stored as a JSON object since req.Params is a map
// 	var storedParams map[string]interface{}
// 	err = json.Unmarshal(record.Params, &storedParams)
// 	require.NoError(t, err)

// 	// Verify the stored params match what we sent
// 	assert.Equal(t, "value1", storedParams["key1"])
// 	assert.Equal(t, float64(42), storedParams["key2"]) // JSON unmarshal numbers as float64
// }

// TestRPCStoreStoreMessageError tests error handling for StoreMessage
func TestRPCStoreStoreMessageError(t *testing.T) {
	// Set up test database
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	// Create a new RPCStore
	store := NewRPCStore(db)

	// Create test data with invalid params that cannot be marshalled
	sender := "0xSender123"

	// Create invalid params (with a channel that can't be marshalled)
	invalidParams := make(map[string]interface{})
	invalidParams["channel"] = make(chan int)
	invalidParamsJSON, _ := json.Marshal(invalidParams) // This will actually fail

	var params rpc.Params
	// Attempting to use invalid JSON will cause an error
	_ = json.Unmarshal([]byte(`{"key": "value"}`), &params)

	req := &rpc.Payload{
		RequestID: 12345,
		Method:    "test_method",
		Params:    params,
		Timestamp: uint64(time.Now().Unix()),
	}
	reqSig := []sign.Signature{sign.Signature([]byte("sig1"))}
	resBytes := invalidParamsJSON // Use the invalid JSON to trigger error
	resSig := []sign.Signature{sign.Signature([]byte("resSig1"))}

	// Attempt to store the message, should fail due to unmarshal-ale params
	err := store.StoreMessage(sender, req, reqSig, resBytes, resSig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}

// // TestRPCStoreGetRPCHistoryForUser tests getting RPC history for a specific user with pagination
// func TestRPCStoreGetRPCHistoryForUser(t *testing.T) {
// 	// Set up test database
// 	db, cleanup := SetupTestDB(t)
// 	defer cleanup()

// 	// Create a new RPCStore
// 	store := NewRPCStore(db)

// 	// Create test data
// 	user1 := "0xUser1"
// 	user2 := "0xUser2"
// 	baseTime := uint64(time.Now().Unix())

// 	// Create test records for user1
// 	user1Records := []RPCRecord{
// 		{Sender: user1, ReqID: 1, Method: "method1", Params: []byte(`[1]`), Timestamp: baseTime - 5, ReqSig: []string{"sig1"}, Response: []byte(`{"result":1}`), ResSig: []string{}},
// 		{Sender: user1, ReqID: 2, Method: "method2", Params: []byte(`[2]`), Timestamp: baseTime - 4, ReqSig: []string{"sig2"}, Response: []byte(`{"result":2}`), ResSig: []string{}},
// 		{Sender: user1, ReqID: 3, Method: "method3", Params: []byte(`[3]`), Timestamp: baseTime - 3, ReqSig: []string{"sig3"}, Response: []byte(`{"result":3}`), ResSig: []string{}},
// 		{Sender: user1, ReqID: 4, Method: "method4", Params: []byte(`[4]`), Timestamp: baseTime - 2, ReqSig: []string{"sig4"}, Response: []byte(`{"result":4}`), ResSig: []string{}},
// 		{Sender: user1, ReqID: 5, Method: "method5", Params: []byte(`[5]`), Timestamp: baseTime - 1, ReqSig: []string{"sig5"}, Response: []byte(`{"result":5}`), ResSig: []string{}},
// 	}

// 	// Create test records for user2
// 	user2Records := []RPCRecord{
// 		{Sender: user2, ReqID: 6, Method: "method6", Params: []byte(`[6]`), Timestamp: baseTime, ReqSig: []string{"sig6"}, Response: []byte(`{"result":6}`), ResSig: []string{}},
// 	}

// 	// Store all records
// 	for _, record := range append(user1Records, user2Records...) {
// 		err := db.Create(&record).Error
// 		require.NoError(t, err)
// 	}

// 	// Test cases
// 	testCases := []struct {
// 		name           string
// 		userID         string
// 		options        *ListOptions
// 		expectedReqIDs []uint64
// 		expectedCount  int
// 	}{
// 		{
// 			name:           "Default pagination for user1",
// 			userID:         user1,
// 			options:        &ListOptions{},
// 			expectedReqIDs: []uint64{5, 4, 3, 2, 1}, // Descending order
// 			expectedCount:  5,
// 		},
// 		{
// 			name:           "Limit only for user1",
// 			userID:         user1,
// 			options:        &ListOptions{Limit: 3},
// 			expectedReqIDs: []uint64{5, 4, 3}, // First 3 in descending order
// 			expectedCount:  3,
// 		},
// 		{
// 			name:           "Offset and limit for user1",
// 			userID:         user1,
// 			options:        &ListOptions{Offset: 2, Limit: 2},
// 			expectedReqIDs: []uint64{3, 2}, // Skip 2, take 2
// 			expectedCount:  2,
// 		},
// 		{
// 			name:           "Ascending sort for user1",
// 			userID:         user1,
// 			options:        func() *ListOptions { sortType := SortTypeAscending; return &ListOptions{Sort: &sortType} }(),
// 			expectedReqIDs: []uint64{1, 2, 3, 4, 5}, // Ascending order
// 			expectedCount:  5,
// 		},
// 		{
// 			name:           "User2 records only",
// 			userID:         user2,
// 			options:        &ListOptions{},
// 			expectedReqIDs: []uint64{6},
// 			expectedCount:  1,
// 		},
// 		{
// 			name:           "Non-existent user",
// 			userID:         "0xNonExistent",
// 			options:        &ListOptions{},
// 			expectedReqIDs: []uint64{},
// 			expectedCount:  0,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			records, err := store.GetRPCHistory(tc.userID, tc.options)
// 			require.NoError(t, err)
// 			assert.Len(t, records, tc.expectedCount)

// 			// Verify the records are in expected order
// 			for i, record := range records {
// 				if i < len(tc.expectedReqIDs) {
// 					assert.Equal(t, tc.expectedReqIDs[i], record.ReqID)
// 					assert.Equal(t, tc.userID, record.Sender)
// 				}
// 			}
// 		})
// 	}
// }
