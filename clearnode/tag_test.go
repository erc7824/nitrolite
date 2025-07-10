package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func Test_GetUserTagByWallet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	wallet := "0x1234567890abcdef1234567890abcdef12345678"
	tag, err := GetUserTagByWallet(db, wallet)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	require.Empty(t, tag, "Tag should be nil for non-existing wallet")

	// Try to resolve the wallet by unassociated tag
	walletRetrieved, err := GetWalletByTag(db, "non-existing-tag")
	require.Contains(t, err.Error(), "no associated wallet for tag")
	require.Empty(t, walletRetrieved, "Wallet should be empty for non-existing tag")

	// Try to find tag for an empty wallet
	tag, err = GetUserTagByWallet(db, "")
	require.Contains(t, err.Error(), "wallet address cannot be empty")
	require.Empty(t, tag, "Tag should be nil for empty wallet")

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
	walletRetrieved, err = GetWalletByTag(db, model.Tag)
	require.NoError(t, err)
	require.Equal(t, wallet, walletRetrieved.Wallet, "Retrieved wallet should match the original wallet")
}

func Test_GenerateRandomAlphaNumericTag(t *testing.T) {
	tag1 := GenerateRandomAlphanumericTag()
	require.Equal(t, len(tag1), 6)

	tag2 := GenerateRandomAlphanumericTag()
	require.Equal(t, len(tag2), 6)

	require.NotEqual(t, tag1, tag2, "Tags should be different")
}
