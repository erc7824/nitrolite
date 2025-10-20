package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clearCache() {
	custodySignerCache = sync.Map{}
	sessionKeyCache = sync.Map{}
}

func TestLoadWalletCache(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	require.NoError(t, db.Create(&SignerWallet{Signer: "alice", Wallet: "w1"}).Error)
	require.NoError(t, db.Create(&SignerWallet{Signer: "bob", Wallet: "w2"}).Error)

	require.NoError(t, loadCustodySignersCache(db))

	assert.Equal(t, "w1", GetWalletBySigner("alice"))
	assert.Equal(t, "w2", GetWalletBySigner("bob"))
	assert.Empty(t, GetWalletBySigner("carol"))
}

func TestAddSigner_New(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	// add a new mapping
	require.NoError(t, AddSigner(db, "w3", "charlie"))

	// should exist in DB
	var cnt int64
	require.NoError(t, db.Model(&SignerWallet{}).
		Where("signer = ?", "charlie").
		Count(&cnt).Error)
	assert.Equal(t, int64(1), cnt)

	// should be cached
	assert.Equal(t, "w3", GetWalletBySigner("charlie"))
}

func TestAddSigner_SamePair_NoError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	// pre-insert
	require.NoError(t, db.Create(&SignerWallet{Signer: "dan", Wallet: "w4"}).Error)

	// adding same pair again is a no-op
	require.NoError(t, AddSigner(db, "w4", "dan"))

	// still exactly one row
	var cnt int64
	require.NoError(t, db.Model(&SignerWallet{}).
		Where("signer = ?", "dan").
		Count(&cnt).Error)
	assert.Equal(t, int64(1), cnt)

	// cache should now have it
	assert.Equal(t, "w4", GetWalletBySigner("dan"))
}

func TestAddSigner_DifferentWallet_Error(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	// pre-insert
	require.NoError(t, db.Create(&SignerWallet{Signer: "eve", Wallet: "w5"}).Error)

	// attempt to re-bind to a different wallet
	err := AddSigner(db, "w5-different", "eve")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signer is already in use")

	// still one row
	var cnt int64
	require.NoError(t, db.Model(&SignerWallet{}).
		Where("signer = ?", "eve").
		Count(&cnt).Error)
	assert.Equal(t, int64(1), cnt)

	// cache remains empty
	assert.Empty(t, GetWalletBySigner("eve"))
}

func TestRemoveSigner(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	// insert via AddSigner so cache is populated
	require.NoError(t, AddSigner(db, "w6", "frank"))

	// remove it
	require.NoError(t, RemoveSigner(db, "w6", "frank"))

	// DB should no longer have it
	var cnt int64
	require.NoError(t, db.Model(&SignerWallet{}).
		Where("signer = ?", "frank").
		Count(&cnt).Error)
	assert.Equal(t, int64(0), cnt)

	assert.Equal(t, "", GetWalletBySigner("frank"))
}

func TestRemoveSigner_NonExistent_NoError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()
	clearCache()

	// removing a non-existent mapping should not error
	require.NoError(t, RemoveSigner(db, "any", "user"))

	// DB remains empty
	var cnt int64
	require.NoError(t, db.Model(&SignerWallet{}).Count(&cnt).Error)
	assert.Equal(t, int64(0), cnt)
}
