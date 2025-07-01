package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetUserTagByWallet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	wallet := "0x1234567890abcdef1234567890abcdef12345678"
	tag, err := GetUserTagByWallet(db, wallet)
	assert.Contains(t, err.Error(), "user tag does not exist for wallet")
	assert.Empty(t, tag, "Tag should be nil for non-existing wallet")

	// Create a user tag
	model, err := GenerateOrRetrieveUserTag(db, wallet)
	require.NoError(t, err)
	require.NotNil(t, model)

	// Make sure the tag is not regenerated
	model2, err := GenerateOrRetrieveUserTag(db, wallet)
	require.NoError(t, err)
	require.NotNil(t, model2)
	require.Equal(t, model.Tag, model2.Tag, "Tags should match for the same wallet")

	// Retrieve the tag by wallet
	retrievedTag, err := GetUserTagByWallet(db, wallet)
	require.NoError(t, err)
	require.Equal(t, model.Tag, retrievedTag)

	// Retrieve wallet by tag
	walletRetrieved, err := GetWalletByTag(db, model.Tag)
	require.NoError(t, err)
	require.Equal(t, wallet, walletRetrieved, "Retrieved wallet should match the original wallet")
}

func Test_GenerateRandomAlphaNumericTag(t *testing.T) {
	tag1 := GenerateRandomAlphanumericTag()
	require.Equal(t, len(tag1), 8)

	tag2 := GenerateRandomAlphanumericTag()
	require.Equal(t, len(tag2), 8)

	require.NotEqual(t, tag1, tag2, "Tags should be different")
}
