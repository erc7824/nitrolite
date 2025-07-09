package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
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

	return transaction, err
}

// GetLedgerTransactions retrieves ledger transactions based on the provided filters.
func GetLedgerTransactions(tx *gorm.DB, accountID AccountID, assetSymbol string, txType *TransactionType) ([]LedgerTransaction, error) {
	var transactions []LedgerTransaction
	q := tx.Model(&LedgerTransaction{})

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

// FormatTransactionsWithTags formats multiple transactions with user tags.
func FormatTransactionsWithTags(db *gorm.DB, transactions []LedgerTransaction) ([]TransactionResponse, error) {
	if len(transactions) == 0 {
		return []TransactionResponse{}, nil
	}

	// Collect all unique wallet addresses from transactions
	walletTags := make(map[string]string)
	for _, tx := range transactions {
		if common.IsHexAddress(tx.FromAccount) {
			walletTags[tx.FromAccount] = ""
		}
		if common.IsHexAddress(tx.ToAccount) {
			walletTags[tx.ToAccount] = ""
		}
	}

	uniqueAccountAddresses := make([]string, 0, len(walletTags))
	for addr := range walletTags {
		uniqueAccountAddresses = append(uniqueAccountAddresses, addr)
	}

	var userTags []UserTagModel
	if len(uniqueAccountAddresses) > 0 {
		if err := db.Where("wallet IN ?", uniqueAccountAddresses).Find(&userTags).Error; err != nil {
			return nil, fmt.Errorf("failed to fetch user tags: %w", err)
		}
	}

	// Assign tags to wallet addresses where available
	for _, tag := range userTags {
		walletTags[tag.Wallet] = tag.Tag
	}

	responses := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = TransactionResponse{
			Id:             tx.ID,
			TxType:         tx.Type.String(),
			FromAccount:    tx.FromAccount,
			FromAccountTag: walletTags[tx.FromAccount], // Will be empty string if not found
			ToAccount:      tx.ToAccount,
			ToAccountTag:   walletTags[tx.ToAccount], // Will be empty string if not found
			Asset:          tx.AssetSymbol,
			Amount:         tx.Amount,
			CreatedAt:      tx.CreatedAt,
		}
	}

	return responses, nil
}
