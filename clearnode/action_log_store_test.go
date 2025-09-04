package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActionLogStore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewActionLogStore(db)
	assert.NotNil(t, store)
	assert.NotNil(t, store.db)
}

func TestStore(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewActionLogStore(db)

	userID := "0xUser123"
	label := ActionLabel("misbehavior_spam")
	metadata := map[string]interface{}{"severity": "high"}
	metadataBytes, err := json.Marshal(metadata)
	require.NoError(t, err)

	err = store.Store(context.Background(), userID, label, metadataBytes)
	require.NoError(t, err)

	var record UserActionLog
	err = db.First(&record).Error
	require.NoError(t, err)

	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, label, record.Label)
	assert.Equal(t, metadataBytes, record.Metadata)
	assert.False(t, record.CreatedAt.IsZero())
}

func TestStoreWithNilMetadata(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewActionLogStore(db)

	userID := "0xUser456"
	label := ActionLabel("misbehavior_timeout")

	err := store.Store(context.Background(), userID, label, nil)
	require.NoError(t, err)

	var record UserActionLog
	err = db.First(&record).Error
	require.NoError(t, err)

	assert.Equal(t, userID, record.UserID)
	assert.Equal(t, label, record.Label)
	assert.Nil(t, record.Metadata)
}

func TestList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewActionLogStore(db)

	user1 := "user-1"
	user2 := "user-2"
	labelSpam := ActionLabel("spam")
	labelLogin := ActionLabel("login")
	labelTimeout := ActionLabel("timeout")

	recordsToCreate := []UserActionLog{
		{UserID: user1, Label: labelLogin}, // Oldest
		{UserID: user2, Label: labelSpam},
		{UserID: user1, Label: labelSpam},
		{UserID: user1, Label: labelTimeout},
		{UserID: user2, Label: labelLogin},
		{UserID: user1, Label: labelSpam}, // Newest
	}

	for _, record := range recordsToCreate {
		err := db.Create(&record).Error
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond)
	}

	type expectedResult struct {
		UserID string
		Label  ActionLabel
	}

	testCases := []struct {
		name            string
		userID          string
		label           ActionLabel
		options         *ListOptions
		expectedResults []expectedResult
	}{
		{
			name:   "Filter by user ID only",
			userID: user1,
			label:  "",
			// Expected order for user1 is newest to oldest: spam, timeout, spam, login
			expectedResults: []expectedResult{
				{user1, labelSpam},
				{user1, labelTimeout},
				{user1, labelSpam},
				{user1, labelLogin},
			},
		},
		{
			name:    "Filter by user ID with limit",
			userID:  user1,
			label:   "",
			options: &ListOptions{Limit: 2},
			// Expected order is the first 2 newest for user1: spam, timeout
			expectedResults: []expectedResult{
				{user1, labelSpam},
				{user1, labelTimeout},
			},
		},
		{
			name:   "Filter by label only",
			userID: "",
			label:  labelSpam,
			// Expected order for "spam" is newest to oldest: user1, user1, user2
			expectedResults: []expectedResult{
				{user1, labelSpam},
				{user1, labelSpam},
				{user2, labelSpam},
			},
		},
		{
			name:   "Filter by both user ID and label",
			userID: user1,
			label:  labelSpam,
			expectedResults: []expectedResult{
				{UserID: user1, Label: labelSpam},
				{UserID: user1, Label: labelSpam},
			},
		},
		{
			name:            "No results",
			userID:          user2,
			label:           labelTimeout,
			expectedResults: []expectedResult{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var userIDPtr *string
			if tc.userID != "" {
				userIDPtr = &tc.userID
			}
			var labelPtr *ActionLabel
			if tc.label != "" {
				labelPtr = &tc.label
			}
			if tc.options == nil {
				tc.options = &ListOptions{}
			}

			records, err := store.List(context.Background(), userIDPtr, labelPtr, tc.options)
			require.NoError(t, err)
			assert.Len(t, records, len(tc.expectedResults))

			for i, record := range records {
				assert.Equal(t, tc.expectedResults[i].UserID, record.UserID)
				assert.Equal(t, tc.expectedResults[i].Label, record.Label)
			}
		})
	}
}

func TestCount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewActionLogStore(db)

	user1 := "0xUser1"
	user2 := "0xUser2"
	labelSpam := ActionLabel("spam")
	labelTimeout := ActionLabel("timeout")

	records := []UserActionLog{
		{UserID: user1, Label: labelSpam},
		{UserID: user1, Label: labelTimeout},
		{UserID: user1, Label: labelSpam},
		{UserID: user2, Label: labelSpam},
	}
	require.NoError(t, db.Create(&records).Error)

	testCases := []struct {
		name          string
		userID        string
		label         ActionLabel
		expectedCount int64
	}{
		{"Count by user and label", user1, labelSpam, 2},
		{"Count by user only", user1, "", 3},
		{"Count by label only", "", labelSpam, 3},
		{"Count non-existent label", user1, "non_existent", 0},
		{"Count non-existent user", "non_existent", labelSpam, 0},
		{"Count all records", "", "", 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var userIDPtr *string
			if tc.userID != "" {
				userIDPtr = &tc.userID
			}

			var labelPtr *ActionLabel
			if tc.label != "" {
				labelPtr = &tc.label
			}

			count, err := store.Count(context.Background(), userIDPtr, labelPtr)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCount, count)
		})
	}
}

func TestTableName(t *testing.T) {
	record := UserActionLog{}
	assert.Equal(t, "user_action_logs", record.TableName())
}
