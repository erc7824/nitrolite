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
	AccountID   string          `gorm:"column:account_id;not null;index:idx_account_asset_symbol;index:idx_account_participant"`
	AccountType AccountType     `gorm:"column:account_type;not null"`
	AssetSymbol string          `gorm:"column:asset_symbol;not null;index:idx_account_asset_symbol"`
	Participant string          `gorm:"column:participant;not null;index:idx_account_participant"`
	Credit      decimal.Decimal `gorm:"column:credit;type:decimal(38,18);not null"`
	Debit       decimal.Decimal `gorm:"column:debit;type:decimal(38,18);not null"`
	CreatedAt   time.Time
}

func (Entry) TableName() string {
	return "ledger"
}

type ParticipantLedger struct {
	participant string
	db          *gorm.DB
}

func GetParticipantLedger(db *gorm.DB, participant string) *ParticipantLedger {
	return &ParticipantLedger{participant: participant, db: db}
}

func (l *ParticipantLedger) Record(accountID string, assetSymbol string, amount decimal.Decimal) error {
	entry := &Entry{
		AccountID:   accountID,
		Participant: l.participant,
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

	fmt.Println("recording entry for: ", l.participant, " in account ", accountID, " ", assetSymbol, " ", amount)
	return l.db.Create(entry).Error
}

func (l *ParticipantLedger) Balance(accountID string, assetSymbol string) (decimal.Decimal, error) {
	type result struct {
		Balance decimal.Decimal `gorm:"column:balance"`
	}
	var res result
	if err := l.db.Model(&Entry{}).
		Where("account_id = ? AND asset_symbol = ? AND participant = ?", accountID, assetSymbol, l.participant).
		Select("COALESCE(SUM(credit),0) - COALESCE(SUM(debit),0) AS balance").
		Scan(&res).Error; err != nil {
		return decimal.Zero, err
	}
	return res.Balance, nil
}

func (l *ParticipantLedger) GetBalances(accountID string) ([]Balance, error) {
	type row struct {
		Asset   string          `gorm:"column:asset_symbol"`
		Balance decimal.Decimal `gorm:"column:balance"`
	}

	var rows []row
	if err := l.db.
		Model(&Entry{}).
		Where("account_id = ? AND participant = ?", accountID, l.participant).
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

func (l *ParticipantLedger) GetEntries(accountID, assetSymbol string) ([]Entry, error) {
	var entries []Entry
	q := l.db.Where("account_id = ? AND participant = ?", accountID, l.participant)
	if assetSymbol != "" {
		q = q.Where("asset_symbol = ?", assetSymbol)
	}

	if err := q.Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
