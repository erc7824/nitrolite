package main

import (
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type SignerWallet struct {
	Signer string `gorm:"column:signer;primaryKey"`
	Wallet string `gorm:"column:wallet;index;not null"`
}

func (SignerWallet) TableName() string {
	return "signers"
}

// walletCache is a thread-safe cache for signer-wallet mappings.
var walletCache sync.Map

// loadWalletCache populates the walletCache from the database.
func loadWalletCache(db *gorm.DB) error {
	var all []SignerWallet
	if err := db.Find(&all).Error; err != nil {
		return err
	}
	for _, sw := range all {
		walletCache.Store(sw.Signer, sw.Wallet)
	}
	return nil
}

// GetWalletBySigner retrieves the wallet address associated with a given signer from the cache.
func GetWalletBySigner(signer string) string {
	if w, ok := walletCache.Load(signer); ok {
		return w.(string)
	}
	return ""
}

// AddSigner adds a new signer-wallet mapping to the database.
func AddSigner(db *gorm.DB, wallet, signer string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var existingSigner SignerWallet

		err := tx.Where("signer = ?", signer).First(&existingSigner).Error
		switch {
		case err == nil:
			if existingSigner.Wallet == wallet {
				walletCache.Store(signer, wallet)
				return nil
			}
			return fmt.Errorf("signer is already in use for another waller")

		case errors.Is(err, gorm.ErrRecordNotFound):
			sw := &SignerWallet{Signer: signer, Wallet: wallet}
			if err := tx.Create(sw).Error; err != nil {
				return err
			}
			walletCache.Store(signer, wallet)
			return nil

		default:
			return err
		}
	})
}

// RemoveSigner deletes a signer-wallet mapping from the database and cache.
func RemoveSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	sw := &SignerWallet{
		Signer: signerAddress,
		Wallet: walletAddress,
	}
	if err := db.Delete(&sw).Error; err != nil {
		return err
	}

	walletCache.Delete(signerAddress)
	return nil
}
