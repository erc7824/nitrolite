package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionType int

const (
	TransactionTypeTransfer      TransactionType = 100
	TransactionTypeDeposit       TransactionType = 201
	TransactionTypeWithdrawal    TransactionType = 202
	TransactionTypeAppDeposit    TransactionType = 301
	TransactionTypeAppWithdrawal TransactionType = 302
)

var (
	ErrInvalidLedgerTransactionType = errors.New("invalid ledger transaction type")
)

type LedgerTransaction struct {
	ID          uint            `gorm:"primaryKey"`
	Hash        string          `gorm:"column:hash;not null;uniqueIndex"`
	Type        TransactionType `gorm:"column:tx_type;not null;index:idx_type;index:idx_from_to_account"`
	FromAccount string          `gorm:"column:from_account;not null;index:idx_from_account;index:idx_from_to_account"`
	ToAccount   string          `gorm:"column:to_account;not null;index:idx_to_account;index:idx_from_to_account"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(38,18);not null"`
	CreatedAt   time.Time
}

func (tx *LedgerTransaction) JSON() TransactionResponse {
	return TransactionResponse{
		TxHash:      tx.Hash,
		TxType:      tx.Type.String(),
		FromAccount: tx.FromAccount,
		ToAccount:   tx.ToAccount,
		Asset:       tx.AssetSymbol,
		Amount:      tx.Amount,
		CreatedAt:   tx.CreatedAt,
	}
}

func (LedgerTransaction) TableName() string {
	return "ledger_transactions"
}

func (t *LedgerTransaction) BeforeCreate(tx *gorm.DB) (err error) {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	// Set a temporary hash. This will be overwritten in AfterCreate.
	t.Hash = fmt.Sprintf("pending-%p", t)
	return nil
}

func (t *LedgerTransaction) AfterCreate(tx *gorm.DB) (err error) {
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

func RecordLedgerTransaction(tx *gorm.DB, txType TransactionType, fromAccount, toAccount, assetSymbol string, amount decimal.Decimal) (*LedgerTransaction, error) {
	transaction := &LedgerTransaction{
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

func GetLedgerTransactions(tx *gorm.DB, accountID, assetSymbol string, txType *TransactionType) ([]LedgerTransaction, error) {
	var transactions []LedgerTransaction
	q := tx.Model(&LedgerTransaction{})

	if accountID != "" {
		q = q.Where("from_account = ? OR to_account = ?", accountID, accountID)

	}
	if assetSymbol != "" {
		q = q.Where("asset_symbol = ?", assetSymbol)
	}
	if txType != nil {
		q = q.Where("tx_type = ?", txType)
	}

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

// TransactionTypeToString converts integer transaction type to string
func (t TransactionType) String() string {
	switch t {
	case TransactionTypeTransfer:
		return "transfer"
	case TransactionTypeDeposit:
		return "deposit"
	case TransactionTypeWithdrawal:
		return "withdrawal"
	case TransactionTypeAppDeposit:
		return "app_deposit"
	case TransactionTypeAppWithdrawal:
		return "app_withdrawal"
	default:
		return ""
	}
}

// parseLedgerTransactionType converts string transaction type to integer
func parseLedgerTransactionType(s string) (TransactionType, error) {
	switch s {
	case "transfer":
		return TransactionTypeTransfer, nil
	case "deposit":
		return TransactionTypeDeposit, nil
	case "withdrawal":
		return TransactionTypeWithdrawal, nil
	case "app_deposit":
		return TransactionTypeAppDeposit, nil
	case "app_withdrawal":
		return TransactionTypeAppWithdrawal, nil
	default:
		return 0, ErrInvalidLedgerTransactionType
	}
}
