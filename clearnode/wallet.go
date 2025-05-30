package main

import (
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SignerWallet struct {
	Signer string `gorm:"column:signer;primaryKey"`
	Wallet string `gorm:"column:wallet;index;not null"`
}

func (SignerWallet) TableName() string {
	return "signers"
}

var walletCache sync.Map

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

func GetWalletBySigner(signer string) string {
	if w, ok := walletCache.Load(signer); ok {
		return w.(string)
	}
	return ""
}

func AddSigner(db *gorm.DB, wallet, signer string) error {
	sw := &SignerWallet{Signer: signer, Wallet: wallet}

	if err := db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(sw).Error; err != nil {
		return err
	}

	walletCache.Store(signer, wallet)
	return nil
}

func RemoveSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	sw := &SignerWallet{
		Signer: signerAddress,
		Wallet: walletAddress,
	}
	return db.Delete(&sw).Error
}
