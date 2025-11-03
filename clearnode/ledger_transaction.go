package main

import (
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
	TransactionTypeEscrowLock    TransactionType = 401
	TransactionTypeEscrowUnlock  TransactionType = 402
)

var (
	ErrInvalidLedgerTransactionType = RPCErrorf("invalid ledger transaction type")
	ErrRecordTransaction            = "failed to record transaction"
)

type LedgerTransaction struct {
	ID          uint            `gorm:"primaryKey"`
	Type        TransactionType `gorm:"column:tx_type;not null;index:idx_type;index:idx_from_to_account"`
	FromAccount string          `gorm:"column:from_account;not null;index:idx_from_account;index:idx_from_to_account"`
	ToAccount   string          `gorm:"column:to_account;not null;index:idx_to_account;index:idx_from_to_account"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(38,18);not null"`
	CreatedAt   time.Time
}

func (LedgerTransaction) TableName() string {
	return "ledger_transactions"
}

// RecordLedgerTransaction records a new ledger transaction in the database.
func RecordLedgerTransaction(tx *gorm.DB, txType TransactionType, fromAccount, toAccount AccountID, assetSymbol string, amount decimal.Decimal) (*LedgerTransaction, error) {
	transaction := &LedgerTransaction{
		Type:        txType,
		FromAccount: fromAccount.String(),
		ToAccount:   toAccount.String(),
		AssetSymbol: assetSymbol,
		Amount:      amount.Abs(),
	}

	err := tx.Create(transaction).Error
	if err != nil {
		return nil, RPCErrorf(ErrRecordTransaction+" : %w", err)
	}
	return transaction, nil
}

type TransactionWithTags struct {
	LedgerTransaction
	FromAccountTag string `gorm:"column:from_tag"`
	ToAccountTag   string `gorm:"column:to_tag"`
}

// GetLedgerTransactions retrieves ledger transactions based on the provided filters.
func GetLedgerTransactionsWithTags(db *gorm.DB, accountID AccountID, assetSymbol string, txType *TransactionType) ([]TransactionWithTags, error) {
	var transactions []TransactionWithTags

	q := db.Model(&LedgerTransaction{}).
		Joins("LEFT JOIN user_tags AS from_tags ON from_tags.wallet = ledger_transactions.from_account").
		Joins("LEFT JOIN user_tags AS to_tags ON to_tags.wallet = ledger_transactions.to_account").
		Select("ledger_transactions.*, from_tags.tag as from_tag, to_tags.tag as to_tag")

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
	case TransactionTypeEscrowLock:
		return "escrow_lock"
	case TransactionTypeEscrowUnlock:
		return "escrow_unlock"
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
	case "escrow_lock":
		return TransactionTypeEscrowLock, nil
	case "escrow_unlock":
		return TransactionTypeEscrowUnlock, nil
	default:
		return 0, ErrInvalidLedgerTransactionType
	}
}

// FormatTransactions formats multiple transactions with user tags.
func FormatTransactions(db *gorm.DB, transactions []TransactionWithTags) ([]TransactionResponse, error) {
	if len(transactions) == 0 {
		return []TransactionResponse{}, nil
	}

	responses := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = TransactionResponse{
			Id:             tx.ID,
			TxType:         tx.Type.String(),
			FromAccount:    tx.FromAccount,
			FromAccountTag: tx.FromAccountTag, // Will be empty string if not found
			ToAccount:      tx.ToAccount,
			ToAccountTag:   tx.ToAccountTag, // Will be empty string if not found
			Asset:          tx.AssetSymbol,
			Amount:         tx.Amount,
			CreatedAt:      tx.CreatedAt,
		}
	}

	return responses, nil
}
