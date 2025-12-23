package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type TransactionType string

const (
	TransactionTypeTransfer       TransactionType = "transfer"
	TransactionTypeCommit         TransactionType = "commit"
	TransactionTypeRelease        TransactionType = "release"
	TransactionTypeHomeDeposit    TransactionType = "home_deposit"
	TransactionTypeHomeWithdrawal TransactionType = "home_withdrawal"
	TransactionTypeMutualLock     TransactionType = "mutual_lock"
	TransactionTypeEscrowDeposit  TransactionType = "escrow_deposit"
	TransactionTypeEscrowLock     TransactionType = "escrow_lock"
	TransactionTypeEscrowWithdraw TransactionType = "escrow_withdraw"
	TransactionTypeMigrate        TransactionType = "migrate"

	// Aliases for backward compatibility
	TransactionTypeAppDeposit    TransactionType = TransactionTypeCommit  // Deprecated: use TransactionTypeCommit
	TransactionTypeAppWithdrawal TransactionType = TransactionTypeRelease // Deprecated: use TransactionTypeRelease
	TransactionTypeDeposit       TransactionType = "deposit"              // Legacy deposit type
	TransactionTypeWithdrawal    TransactionType = "withdrawal"           // Legacy withdrawal type
	TransactionTypeEscrowUnlock  TransactionType = "escrow_unlock"        // Legacy escrow unlock type
)

var (
	ErrInvalidLedgerTransactionType = "invalid ledger transaction type"
	ErrRecordTransaction            = "failed to record transaction"
)

// LedgerTransaction represents an immutable transaction in the system
// ID is deterministic based on transaction initiation:
// 1) Initiated by User: Hash(ToAccount, SenderNewStateID)
// 2) Initiated by Node: Hash(FromAccount, ReceiverNewStateID)
type LedgerTransaction struct {
	// ID is a 64-character deterministic hash
	ID                 string          `gorm:"column:id;primaryKey;size:64"`
	Type               TransactionType `gorm:"column:tx_type;not null;index:idx_type;index:idx_from_to_account"`
	AssetSymbol        string          `gorm:"column:asset_symbol;not null"`
	FromAccount        string          `gorm:"column:from_account;not null;index:idx_from_account;index:idx_from_to_account"`
	ToAccount          string          `gorm:"column:to_account;not null;index:idx_to_account;index:idx_from_to_account"`
	SenderNewStateID   *string         `gorm:"column:sender_new_state_id;size:64"`
	ReceiverNewStateID *string         `gorm:"column:receiver_new_state_id;size:64"`
	Amount             decimal.Decimal `gorm:"column:amount;type:decimal(38,18);not null"`
	CreatedAt          time.Time
}

func (LedgerTransaction) TableName() string {
	return "ledger_transactions"
}

// generateTransactionID generates a deterministic transaction ID
// For now, we use a simple hash of transaction components
// TODO: Implement proper deterministic ID generation based on specification:
// 1) Initiated by User: Hash(ToAccount, SenderNewStateID)
// 2) Initiated by Node: Hash(FromAccount, ReceiverNewStateID)
func generateTransactionID(txType TransactionType, fromAccount, toAccount AccountID, assetSymbol string, amount decimal.Decimal, timestamp time.Time) string {
	data := fmt.Sprintf("%s:%s:%s:%s:%s:%d", txType, fromAccount.String(), toAccount.String(), assetSymbol, amount.String(), timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// RecordLedgerTransactionWithID records a new ledger transaction in the database with a specific ID.
func RecordLedgerTransactionWithID(tx *gorm.DB, id string, txType TransactionType, fromAccount, toAccount AccountID, assetSymbol string, amount decimal.Decimal, senderNewStateID, receiverNewStateID *string) (*LedgerTransaction, error) {
	transaction := &LedgerTransaction{
		ID:                 id,
		Type:               txType,
		AssetSymbol:        assetSymbol,
		FromAccount:        fromAccount.String(),
		ToAccount:          toAccount.String(),
		SenderNewStateID:   senderNewStateID,
		ReceiverNewStateID: receiverNewStateID,
		Amount:             amount.Abs(),
	}

	err := tx.Create(transaction).Error
	if err != nil {
		return nil, fmt.Errorf(ErrRecordTransaction+" : %w", err)
	}
	return transaction, nil
}

// RecordLedgerTransaction records a new ledger transaction in the database with an auto-generated ID.
// This is a convenience wrapper that maintains backward compatibility.
func RecordLedgerTransaction(tx *gorm.DB, txType TransactionType, fromAccount, toAccount AccountID, assetSymbol string, amount decimal.Decimal) (*LedgerTransaction, error) {
	id := generateTransactionID(txType, fromAccount, toAccount, assetSymbol, amount, time.Now())
	return RecordLedgerTransactionWithID(tx, id, txType, fromAccount, toAccount, assetSymbol, amount, nil, nil)
}

// GetLedgerTransactions retrieves ledger transactions based on the provided filters.
func GetLedgerTransactions(db *gorm.DB, accountID AccountID, assetSymbol string, txType *TransactionType) ([]LedgerTransaction, error) {
	var transactions []LedgerTransaction

	q := db.Model(&LedgerTransaction{})

	if accountID.String() != "" {
		q = q.Where("from_account = ? OR to_account = ?", accountID.String(), accountID.String())
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

// String returns the string representation of the transaction type
func (t TransactionType) String() string {
	return string(t)
}

// ParseLedgerTransactionType converts string transaction type to TransactionType
func ParseLedgerTransactionType(s string) (TransactionType, error) {
	switch TransactionType(s) {
	case TransactionTypeTransfer,
		TransactionTypeCommit,
		TransactionTypeRelease,
		TransactionTypeHomeDeposit,
		TransactionTypeHomeWithdrawal,
		TransactionTypeMutualLock,
		TransactionTypeEscrowDeposit,
		TransactionTypeEscrowLock,
		TransactionTypeEscrowWithdraw,
		TransactionTypeMigrate,
		TransactionTypeDeposit,
		TransactionTypeWithdrawal,
		TransactionTypeEscrowUnlock:
		return TransactionType(s), nil
	case "app_deposit":
		return TransactionTypeCommit, nil
	case "app_withdrawal":
		return TransactionTypeRelease, nil
	default:
		return "", errors.New(ErrInvalidLedgerTransactionType)
	}
}
