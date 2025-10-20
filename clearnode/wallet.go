package main

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// SignerWallet represents a custody signer for a wallet
// Possibly will be deprecated with a new version of smart contracts
type SignerWallet struct {
	Signer string `gorm:"column:signer;primaryKey"`
	Wallet string `gorm:"column:wallet;index;not null"`
}

func (SignerWallet) TableName() string {
	return "signers"
}

// custodySignerCache maps custody signer addresses to wallet addresses
var custodySignerCache sync.Map

// sessionKeyCache maps session key addresses to wallet addresses
var sessionKeyCache sync.Map

// loadCustodySignersCache populates the cache with custody signers from the signers table
func loadCustodySignersCache(db *gorm.DB) error {
	var signers []SignerWallet
	if err := db.Find(&signers).Error; err != nil {
		return err
	}
	for _, s := range signers {
		custodySignerCache.Store(s.Signer, s.Wallet)
	}
	return nil
}

// AddSigner adds a new custody signer to the database
func AddSigner(db *gorm.DB, wallet, signer string) error {
	// Check if the signer already exists for this wallet in custody cache
	if w, ok := custodySignerCache.Load(signer); ok {
		if w.(string) == wallet {
			return nil // Already exists with the same wallet
		}
		return fmt.Errorf("signer %s is already in use for another wallet", signer)
	}

	// Check if it exists as a session key
	if _, ok := sessionKeyCache.Load(signer); ok {
		return fmt.Errorf("address is already in use as a session key: %s", signer)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// Before adding a new signer, we need to ensure that the relationship is valid.
		// 1. A wallet address can't be used as a signer for another wallet.
		// 2. An address can't be used as a wallet if it is already a signer for another wallet.
		// 3. A signer can only be associated with one wallet.

		// Check if signer is used as a wallet in signers table (can't use a wallet as signer)
		var count int64
		if err := tx.Model(&SignerWallet{}).Where("wallet = ?", signer).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("cannot use a wallet address as a signer")
		}

		// Check if signer is used as a wallet in session_keys (can't use a wallet as signer)
		var countSessionKeys int64
		if err := tx.Model(&SessionKey{}).Where("wallet_address = ?", signer).Count(&countSessionKeys).Error; err != nil {
			return err
		}
		if countSessionKeys > 0 {
			return fmt.Errorf("cannot use a wallet address as a signer")
		}

		// Check if address is already a signer for another wallet in signers table
		var existingSigner SignerWallet
		err := tx.Where("signer = ?", signer).First(&existingSigner).Error
		switch {
		case err == nil:
			if existingSigner.Wallet == wallet {
				return nil // Already exists with same wallet
			}
			return fmt.Errorf("signer is already in use for another wallet")

		case err == gorm.ErrRecordNotFound:
			// Check session_keys table too
			var existingSessionKey SessionKey
			if err := tx.Where("signer_address = ?", signer).First(&existingSessionKey).Error; err == nil {
				return fmt.Errorf("signer is already in use as a session key")
			} else if err != gorm.ErrRecordNotFound {
				return err
			}

			// Create new signer
			newSigner := &SignerWallet{
				Signer: signer,
				Wallet: wallet,
			}

			return tx.Create(newSigner).Error

		default:
			return err
		}
	})

	// Update cache only after transaction commits successfully
	if err == nil {
		custodySignerCache.Store(signer, wallet)
	}

	return err
}

// RemoveSigner deletes a custody signer from the database and cache
func RemoveSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	sw := &SignerWallet{
		Signer: signerAddress,
		Wallet: walletAddress,
	}
	if err := db.Delete(&sw).Error; err != nil {
		return err
	}

	custodySignerCache.Delete(signerAddress)
	return nil
}
