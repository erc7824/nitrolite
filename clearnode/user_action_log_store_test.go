package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserActionLogStoreNew(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserActionLogStore(db)
	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

func TestUserActionLogStoreUserAction(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserActionLogStore(db)

	userID := "0xUser123"
	label := "misbehavior_spam"
	metadata := map[string]interface{}{
		"severity": "high",
		"count":    5,
		"details":  "User sent too many requests",
	}
	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)

	err = store.StoreUserAction(userID, label, metadataBytes)
	require.NoError(t, err)

	var count int64
	err = db.Model(&UserActionLog{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	var record UserActionLog
	err = db.First(&record).Error
	require.NoError(t, err)

	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, label, record.Label)
	assert.Greater(t, record.Timestamp, uint64(0))
	assert.Equal(t, metadataBytes, record.Metadata)
	assert.False(t, record.CreatedAt.IsZero())

	var storedMetadata map[string]interface{}
	err = json.Unmarshal(record.Metadata, &storedMetadata)
	require.NoError(t, err)
	assert.Equal(t, "high", storedMetadata["severity"])
	assert.Equal(t, float64(5), storedMetadata["count"])
	assert.Equal(t, "User sent too many requests", storedMetadata["details"])
}

func TestUserActionLogStoreUserActionWithNilMetadata(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserActionLogStore(db)

	userID := "0xUser123"
	label := "misbehavior_timeout"

	err := store.StoreUserAction(userID, label, nil)
	require.NoError(t, err)

	var record UserActionLog
	err = db.First(&record).Error
	require.NoError(t, err)

	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, label, record.Label)
	assert.Nil(t, record.Metadata)
}

func TestListUserActions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserActionLogStore(db)

	user1 := "user-1"
	user2 := "user-2"
	baseTime := uint64(time.Now().Unix())

	recordsToCreate := []UserActionLog{
		{UserID: user1, Label: "login", Timestamp: baseTime - 10},
		{UserID: user2, Label: "spam", Timestamp: baseTime - 9},
		{UserID: user1, Label: "spam", Timestamp: baseTime - 8},
		{UserID: user1, Label: "timeout", Timestamp: baseTime - 7},
		{UserID: user2, Label: "login", Timestamp: baseTime - 6},
		{UserID: user1, Label: "spam", Timestamp: baseTime - 5},
	}

	for _, record := range recordsToCreate {
		record.CreatedAt = time.Now()
		err := db.Create(&record).Error
		require.NoError(t, err)
	}

	stringPtr := func(s string) *string { return &s }
	sortAsc := SortTypeAscending

	type expectedResult struct {
		UserID string
		Label  string
	}

	testCases := []struct {
		name            string
		userID          *string
		label           *string
		options         *ListOptions
		expectedResults []expectedResult
	}{
		{
			name:    "Filter by user ID only",
			userID:  stringPtr(user1),
			label:   nil,
			options: &ListOptions{},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "timeout"},
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "login"},
			},
		},
		{
			name:    "Filter by user ID with limit",
			userID:  stringPtr(user1),
			label:   nil,
			options: &ListOptions{Limit: 2},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "timeout"},
			},
		},
		{
			name:    "Filter by user ID with offset and limit",
			userID:  stringPtr(user1),
			label:   nil,
			options: &ListOptions{Offset: 1, Limit: 2},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "timeout"},
				{UserID: user1, Label: "spam"},
			},
		},
		{
			name:    "Filter by user ID with ascending sort",
			userID:  stringPtr(user1),
			label:   nil,
			options: &ListOptions{Sort: &sortAsc},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "login"},
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "timeout"},
				{UserID: user1, Label: "spam"},
			},
		},
		{
			name:    "Filter by label only",
			userID:  nil,
			label:   stringPtr("spam"),
			options: &ListOptions{},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "spam"},
				{UserID: user2, Label: "spam"},
			},
		},
		{
			name:    "Filter by label with limit",
			userID:  nil,
			label:   stringPtr("spam"),
			options: &ListOptions{Limit: 2},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "spam"},
				{UserID: user1, Label: "spam"},
			},
		},
		{
			name:            "Filter by both user ID and label",
			userID:          stringPtr(user1),
			label:           stringPtr("spam"),
			options:         &ListOptions{},
			expectedResults: []expectedResult{{UserID: user1, Label: "spam"}, {UserID: user1, Label: "spam"}},
		},
		{
			name:            "Filter by both user ID and label with no results",
			userID:          stringPtr(user2),
			label:           stringPtr("timeout"),
			options:         &ListOptions{},
			expectedResults: []expectedResult{},
		},
		{
			name:    "Filter by neither user ID nor label",
			userID:  nil,
			label:   nil,
			options: &ListOptions{},
			expectedResults: []expectedResult{
				{UserID: user1, Label: "spam"},
				{UserID: user2, Label: "login"},
				{UserID: user1, Label: "timeout"},
				{UserID: user1, Label: "spam"},
				{UserID: user2, Label: "spam"},
				{UserID: user1, Label: "login"},
			},
		},
		{
			name:            "Filter by non-existent user",
			userID:          stringPtr("non-existent-user"),
			label:           nil,
			options:         &ListOptions{},
			expectedResults: []expectedResult{},
		},
		{
			name:            "Filter by non-existent label",
			userID:          nil,
			label:           stringPtr("non-existent-label"),
			options:         &ListOptions{},
			expectedResults: []expectedResult{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			records, err := store.ListUserActions(tc.userID, tc.label, tc.options)
			require.NoError(t, err)
			assert.Len(t, records, len(tc.expectedResults))

			for i, record := range records {
				assert.Equal(t, tc.expectedResults[i].UserID, record.UserID)
				assert.Equal(t, tc.expectedResults[i].Label, record.Label)
			}
		})
	}
}

func TestUserActionLogCountUserActionsByUserAndLabel(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserActionLogStore(db)

	user1 := "0xUser1"
	user2 := "0xUser2"
	baseTime := uint64(time.Now().Unix())

	records := []UserActionLog{
		{UserID: user1, Label: "spam", Timestamp: baseTime - 5},
		{UserID: user1, Label: "timeout", Timestamp: baseTime - 4},
		{UserID: user1, Label: "spam", Timestamp: baseTime - 3},
		{UserID: user1, Label: "spam", Timestamp: baseTime - 2},
		{UserID: user2, Label: "spam", Timestamp: baseTime - 1},
		{UserID: user2, Label: "timeout", Timestamp: baseTime},
	}

	for _, record := range records {
		record.CreatedAt = time.Now()
		err := db.Create(&record).Error
		require.NoError(t, err)
	}

	count, err := store.CountUserActionsByUserAndLabel(user1, "spam")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	count, err = store.CountUserActionsByUserAndLabel(user1, "timeout")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = store.CountUserActionsByUserAndLabel(user2, "spam")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = store.CountUserActionsByUserAndLabel(user2, "timeout")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = store.CountUserActionsByUserAndLabel(user1, "non_existent")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	count, err = store.CountUserActionsByUserAndLabel("0xNonExistent", "spam")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestUserActionLogTableName(t *testing.T) {
	record := UserActionLog{}
	assert.Equal(t, "user_action_log", record.TableName())
}
