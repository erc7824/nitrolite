package main

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TODO: add a file with db migrations
// TODO: update docs
// TODO: desice on format of application identifier field
// TODO: test on a real deployment

// Each wallet-signer can have only one active policy per app per scope.
type DBPolicy struct {
	ID          uint          `gorm:"primaryKey;column:id"`
	Wallet      string        `gorm:"column:wallet;not null"`
	Participant string        `gorm:"column:participant;not null"`
	Scope       string        `gorm:"column:scope;not null"`
	Application string        `gorm:"column:application;not null"`
	Used        bool          `gorm:"column:used;type:boolean;not null;default:false"`
	Allowances  []DBAllowance `gorm:"column:allowance;type:json;serializer:json;not null"`
	ExpiresAt   time.Time     `gorm:"column:expires_at;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DBAllowance struct {
	Asset  string          `json:"asset"`
	Amount decimal.Decimal `json:"amount"`
}

func (DBPolicy) TableName() string {
	return "policies"
}

// AddPolicy inserts a new policy, replacing any existing unused policy
func AddPolicy(db *gorm.DB, policy *Policy) error {
	return db.Transaction(func(tx *gorm.DB) error {
		allowances := make([]DBAllowance, 0, len(policy.Allowances))
		for _, allowance := range policy.Allowances {
			amount, err := decimal.NewFromString(allowance.Amount)
			if err != nil {
				return err
			}
			allowances = append(allowances, DBAllowance{
				Asset:  allowance.Asset,
				Amount: amount,
			})
		}

		// Delete any existing unused policies for this key
		if err := tx.
			Where("application = ? AND wallet = ? AND participant = ? AND scope = ? AND used = FALSE",
				policy.Application, policy.Wallet, policy.Participant, policy.Scope).
			Delete(&DBPolicy{}).Error; err != nil {
			return err
		}
		// Create the new policy
		if err := tx.Create(
			&DBPolicy{
				Wallet:      policy.Wallet,
				Participant: policy.Participant,
				Scope:       policy.Scope,
				Application: policy.Application,
				Allowances:  allowances,
				ExpiresAt:   policy.ExpiresAt,
				Used:        false,
			},
		).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetPolicy fetches a non-used, non-expired policy
func GetPolicy(db *gorm.DB, application, wallet, participant, scope string) (*DBPolicy, error) {
	var p DBPolicy
	now := time.Now()
	err := db.
		Where("application = ? AND wallet = ? AND participant = ? AND scope = ? AND used = FALSE AND expires_at > ?",
			application, wallet, participant, scope, now).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// UsePolicy marks a policy as used and saves it
func UsePolicy(db *gorm.DB, policy *DBPolicy) error {
	policy.Used = true
	return db.Save(policy).Error
}
