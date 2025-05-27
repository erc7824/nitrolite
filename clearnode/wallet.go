package main

import (
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

func GetWalletBySigner(db *gorm.DB, signerAddress string) (string, error) {
	w, ok := walletCache.Load(signerAddress)
	if ok {
		return w.(string), nil
	} else {
		return "", nil
	}
}

func AddSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	sw := &SignerWallet{
		Signer: signerAddress,
		Wallet: walletAddress,
	}

	if err := db.Create(sw).Error; err != nil {
		// Check if it's a primary key conflict (duplicate key error)
		// For Postgres, use error code "23505" for unique_violation
		//
		// If so - ignore it
		// var pgErr *pgconn.PgError
		// if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		// 	return nil
		// } else {
		// 	return err
		// }

		// FIXME
		return nil
	}

	walletCache.Store(signerAddress, walletAddress)
	return nil
}

func RemoveSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	sw := &SignerWallet{
		Signer: signerAddress,
		Wallet: walletAddress,
	}
	return db.Delete(&sw).Error
}
