package main

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const (
	ErrGetAccountBalance = "failed to get account balance"
	ErrRecordLedgerEntry = "failed to record a ledger entry"
)

// Entry represents a ledger entry in the database
type Entry struct {
	ID          uint            `gorm:"primaryKey"`
	AccountID   string          `gorm:"column:account_id;not null;index:idx_account_asset_symbol;index:idx_account_wallet"`
	AccountType AccountType     `gorm:"column:account_type;not null"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null;index:idx_account_asset_symbol"`
	Wallet      string          `gorm:"column:wallet;not null;index:idx_account_wallet"`
	Credit      decimal.Decimal `gorm:"column:credit;type:varchar(78);not null"`
	Debit       decimal.Decimal `gorm:"column:debit;type:varchar(78);not null"`
	SessionKey  *string         `gorm:"column:session_key;index:idx_session_key"`
	CreatedAt   time.Time
}

func (Entry) TableName() string {
	return "ledger"
}

type WalletLedger struct {
	wallet common.Address
	db     *gorm.DB
}

// AccountID represents a unique identifier for an account, which can be a wallet or an application session.
// Main reason for creating this type is to ensure the address format is consistent ( e.g., no downcase conversion).
type AccountID string

func NewAccountID(accountID string) AccountID {
	if !common.IsHexAddress(accountID) {
		return AccountID(accountID)
	}

	return AccountID(common.HexToAddress(accountID).Hex())
}

func (a AccountID) String() string {
	return string(a)
}

func GetWalletLedger(db *gorm.DB, wallet common.Address) *WalletLedger {
	return &WalletLedger{wallet: wallet, db: db}
}

func (l *WalletLedger) Record(accountID AccountID, assetSymbol string, amount decimal.Decimal, sessionKey *string) error {
	entry := &Entry{
		AccountID:   accountID.String(),
		Wallet:      l.wallet.Hex(),
		AssetSymbol: assetSymbol,
		Credit:      decimal.Zero,
		Debit:       decimal.Zero,
		SessionKey:  sessionKey,
		CreatedAt:   time.Now(),
	}

	if amount.IsPositive() {
		entry.Credit = amount
	} else if amount.IsNegative() {
		entry.Debit = amount.Abs()
	} else {
		return nil
	}

	fmt.Println("recording entry for: ", l.wallet, " in account ", accountID, " ", assetSymbol, " ", amount)

	err := l.db.Create(entry).Error
	if err != nil {
		return RPCErrorf(ErrRecordLedgerEntry+" : %w", err)
	}
	return nil
}

func (l *WalletLedger) Balance(accountID AccountID, assetSymbol string) (decimal.Decimal, error) {
	switch l.db.Dialector.Name() {
	case "postgres":
		var result struct {
			Balance decimal.Decimal
		}
		err := l.db.Model(&Entry{}).
			Where("account_id = ? AND asset_symbol = ? AND wallet = ?", accountID.String(), assetSymbol, l.wallet.Hex()).
			Select("COALESCE(SUM(credit), 0) - COALESCE(SUM(debit), 0) AS balance").
			Scan(&result).Error

		if err != nil {
			return decimal.Zero, err
		}
		return result.Balance, nil

	case "sqlite":
		// Fetch all records and sum in Go to avoid SQLite's floating-point conversion for big numbers.
		var entries []Entry
		err := l.db.Model(&Entry{}).
			Where("account_id = ? AND asset_symbol = ? AND wallet = ?", accountID.String(), assetSymbol, l.wallet.Hex()).
			Find(&entries).Error

		if err != nil {
			return decimal.Zero, err
		}

		balance := decimal.Zero
		for _, entry := range entries {
			balance = balance.Add(entry.Credit).Sub(entry.Debit)
		}
		return balance, nil

	default:
		return decimal.Zero, fmt.Errorf("unsupported database driver: %s", l.db.Dialector.Name())
	}
}

func (l *WalletLedger) GetBalances(accountID AccountID) ([]Balance, error) {
	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	if err := l.db.
		Model(&Entry{}).
		Where("account_id = ? AND wallet = ?", accountID.String(), l.wallet.Hex()).
		Select("asset_symbol", "COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Group("asset_symbol").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	balances := make([]Balance, len(rows))
	for i, r := range rows {
		balances[i] = Balance{
			Asset:  r.Asset,
			Amount: r.Balance,
		}
	}
	return balances, nil
}

func (l *WalletLedger) GetEntries(accountID *AccountID, assetSymbol string) ([]Entry, error) {
	var entries []Entry
	q := l.db.Model(&Entry{})

	if accountID != nil && accountID.String() != "" {
		q = q.Where("account_id = ?", accountID.String())
	}

	// TODO: design a better way to handle the case when wallet is not set
	if l.wallet.Hex() != common.HexToAddress("").Hex() {
		q = q.Where("wallet = ?", l.wallet.Hex())
	}

	if assetSymbol != "" {
		q = q.Where("asset_symbol = ?", assetSymbol)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func getAppSessionBalances(tx *gorm.DB, appSessionID AccountID) (map[string]decimal.Decimal, error) {
	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	if err := tx.
		Model(&Entry{}).
		Where("account_id = ?", appSessionID.String()).
		Select("asset_symbol", "COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Group("asset_symbol").
		Scan(&rows).Error; err != nil {
		return nil, RPCErrorf("failed to fetch balances for account %s: %w", appSessionID, err)
	}

	result := make(map[string]decimal.Decimal, len(rows))
	for _, r := range rows {
		result[r.Asset] = r.Balance
	}
	return result, nil
}
