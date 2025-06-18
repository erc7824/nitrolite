package main

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Entry represents a ledger entry in the database
type Entry struct {
	ID          uint            `gorm:"primaryKey"`
	AccountID   string          `gorm:"column:account_id;not null;index:idx_account_asset_symbol;index:idx_account_wallet"`
	AccountType AccountType     `gorm:"column:account_type;not null"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null;index:idx_account_asset_symbol"`
	Wallet      string          `gorm:"column:wallet;not null;index:idx_account_wallet"`
	Credit      decimal.Decimal `gorm:"column:credit;type:decimal(38,18);not null"`
	Debit       decimal.Decimal `gorm:"column:debit;type:decimal(38,18);not null"`
	CreatedAt   time.Time
}

func (Entry) TableName() string {
	return "ledger"
}

type WalletLedger struct {
	wallet string
	db     *gorm.DB
}

func GetWalletLedger(db *gorm.DB, wallet string) *WalletLedger {
	return &WalletLedger{wallet: wallet, db: db}
}

func (l *WalletLedger) Record(accountID string, assetSymbol string, amount decimal.Decimal) error {
	entry := &Entry{
		AccountID:   accountID,
		Wallet:      l.wallet,
		AssetSymbol: assetSymbol,
		Credit:      decimal.Zero,
		Debit:       decimal.Zero,
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
	return l.db.Create(entry).Error
}

func (l *WalletLedger) Transfer(fromAccountID, toAccountID, assetSymbol string, amount decimal.Decimal) error {
	balance, err := l.Balance(l.wallet, assetSymbol)
	if err != nil {
		return fmt.Errorf("failed to check participant balance: %w", err)
	}

	if amount.GreaterThan(balance) {
		return fmt.Errorf("insufficient funds: %s for asset %s", l.wallet, assetSymbol)
	}

	if amount.IsNegative() || balance.IsNegative() {
		return fmt.Errorf("invalid transfer")
	}

	if err = l.Record(fromAccountID, assetSymbol, amount.Neg()); err != nil {
		return fmt.Errorf("failed to debit source account: %w", err)
	}
	if err = l.Record(toAccountID, assetSymbol, amount); err != nil {
		return fmt.Errorf("failed to credit destination account: %w", err)
	}

	return nil
}

func (l *WalletLedger) Balance(accountID string, assetSymbol string) (decimal.Decimal, error) {
	type result struct {
		Balance decimal.Decimal `gorm:"column:balance"`
	}
	var res result
	if err := l.db.Model(&Entry{}).
		Where("account_id = ? AND asset_symbol = ? AND wallet = ?", accountID, assetSymbol, l.wallet).
		Select("COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Scan(&res).Error; err != nil {
		return decimal.Zero, err
	}
	return res.Balance, nil
}

func (l *WalletLedger) GetBalances(accountID string) ([]Balance, error) {
	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	if err := l.db.
		Model(&Entry{}).
		Where("account_id = ? AND wallet = ?", accountID, l.wallet).
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

func (l *WalletLedger) GetEntries(accountID, assetSymbol string) ([]Entry, error) {
	var entries []Entry
	q := l.db.Model(&Entry{})

	if accountID != "" {
		q = q.Where("account_id = ?", accountID)
	}

	if l.wallet != "" {
		q = q.Where("wallet = ?", l.wallet)
	}

	if assetSymbol != "" {
		q = q.Where("asset_symbol = ?", assetSymbol)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func getAppSessionBalances(tx *gorm.DB, appSessionID string) (map[string]decimal.Decimal, error) {
	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	if err := tx.
		Model(&Entry{}).
		Where("account_id = ?", appSessionID).
		Select("asset_symbol", "COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Group("asset_symbol").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch balances for account %s: %w", appSessionID, err)
	}

	result := make(map[string]decimal.Decimal, len(rows))
	for _, r := range rows {
		result[r.Asset] = r.Balance
	}
	return result, nil
}
