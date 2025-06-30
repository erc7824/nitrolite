package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionType string

var (
	TransactionTypeTransfer      TransactionType = "transfer"
	TransactionTypeDeposit       TransactionType = "deposit"
	TransactionTypeWithdrawal    TransactionType = "withdrawal"
	TransactionTypeAppDeposit    TransactionType = "app_deposit"
	TransactionTypeAppWithdrawal TransactionType = "app_withdrawal"
)

type Transaction struct {
	ID          uint            `gorm:"primaryKey"`
	Hash        string          `gorm:"column:hash;not null;uniqueIndex"`
	Type        TransactionType `gorm:"column:tx_type;not null;index:idx_type;index:idx_from_to_account"`
	FromAccount string          `gorm:"column:from_account;not null;index:idx_from_account;index:idx_from_to_account"`
	ToAccount   string          `gorm:"column:to_account;not null;index:idx_to_account;index:idx_from_to_account"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(38,18);not null"`
	CreatedAt   time.Time
}

func (Transaction) TableName() string {
	return "transactions"
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	// Set a temporary hash. This will be overwritten in AfterCreate.
	t.Hash = "pending"
	return nil
}

func (t *Transaction) AfterCreate(tx *gorm.DB) (err error) {
	txHash := getTransactionHash(
		t.ID,
		t.Type,
		t.FromAccount,
		t.ToAccount,
		t.AssetSymbol,
		t.Amount,
		t.CreatedAt.UnixMilli(),
	)

	return tx.Model(t).UpdateColumn("hash", txHash).Error
}

func RecordTransaction(tx *gorm.DB, txType TransactionType, fromAccount, toAccount, assetSymbol string, amount decimal.Decimal) (*Transaction, error) {
	transaction := &Transaction{
		Type:        txType,
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		AssetSymbol: assetSymbol,
		Amount:      amount,
	}

	err := tx.Create(transaction).Error

	fmt.Println("recording transaction from: ", fromAccount, " to ", toAccount, " ", assetSymbol, " ", amount)

	return transaction, err
}

func GetTransactions(tx *gorm.DB, accountID, assetSymbol string) ([]Transaction, error) {
	var transactions []Transaction
	q := tx.Model(&Transaction{})

	if accountID != "" {
		q = q.Where("from_account = ? OR to_account = ?", accountID, accountID)

	}
	if assetSymbol != "" {
		q = q.Where("asset_symbol = ?", assetSymbol)
	}

	q = q.Order("created_at DESC")

	if err := q.Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

// getTransactionHash generates a deterministic hash for a transaction based on its attributes.
func getTransactionHash(id uint, txType TransactionType, fromAccount, toAccount, assetSymbol string, amount decimal.Decimal, timestamp int64) string {
	dataString := fmt.Sprintf(
		"%d:%s:%s:%s:%s:%s:%s",
		id,
		txType,
		fromAccount,
		toAccount,
		assetSymbol,
		amount.String(),
		strconv.FormatInt(timestamp, 10), // Use FormatInt for int64
	)

	hash := sha256.New()
	hash.Write([]byte(dataString))
	return hex.EncodeToString(hash.Sum(nil))
}
