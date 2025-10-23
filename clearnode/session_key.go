package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	AppNameClearnode = "clearnode"
)

// SessionKey represents a ledger layer session key
type SessionKey struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Address string `gorm:"column:address;uniqueIndex;not null"`

	WalletAddress string    `gorm:"column:wallet_address;index;not null"`
	Application   string    `gorm:"column:application;not null"`
	Allowance     *string   `gorm:"column:allowance;type:jsonb"` // JSON serialized allowances
	Scope         string    `gorm:"column:scope;not null;"`
	ExpiresAt     time.Time `gorm:"column:expires_at;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SessionKey) TableName() string {
	return "session_keys"
}

// sessionKeyCache maps session key addresses to wallet addresses
var sessionKeyCache sync.Map

// loadSessionKeyCache populates the cache with session keys
func loadSessionKeyCache(db *gorm.DB) error {
	var sessionKeys []SessionKey
	if err := db.Where("expires_at > ?", time.Now().UTC()).Find(&sessionKeys).Error; err != nil {
		return err
	}
	for _, sk := range sessionKeys {
		sessionKeyCache.Store(sk.Address, sk.WalletAddress)
	}
	return nil
}

// AddSessionKey stores a new session key with its metadata
// Only one session key per wallet+app combination is allowed - registering a new one invalidates existing ones
func AddSessionKey(db *gorm.DB, walletAddress, address, applicationName, scope string, allowances []Allowance, expirationTime time.Time) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		// Validate expiration time is in the future
		if expirationTime.IsZero() || expirationTime.Before(time.Now().UTC()) {
			return fmt.Errorf("expiration time must be set and in the future")
		}
		expirationTime = expirationTime.UTC()

		if scope == "" {
			scope = "all"
		}

		// Check for and remove existing session key for this wallet+app combination
		var existingKeys []SessionKey
		err := tx.Where("wallet_address = ? AND application = ?",
			walletAddress, applicationName).Find(&existingKeys).Error
		if err != nil {
			return fmt.Errorf("failed to check existing session keys: %w", err)
		}

		// Remove existing session keys for this app (invalidate them)
		for _, existingKey := range existingKeys {
			if err := tx.Delete(&existingKey).Error; err != nil {
				return fmt.Errorf("failed to remove existing session key: %w", err)
			}
			sessionKeyCache.Delete(existingKey.Address)
		}

		spendingCapJSON, err := json.Marshal(allowances)
		if err != nil {
			return fmt.Errorf("failed to serialize spending cap: %w", err)
		}

		spendingCapStr := string(spendingCapJSON)

		sessionKey := &SessionKey{
			Address:       address,
			WalletAddress: walletAddress,
			Application:   applicationName,
			Allowance:     &spendingCapStr,
			Scope:         scope,
			ExpiresAt:     expirationTime,
		}

		return tx.Create(sessionKey).Error
	})

	// Update cache only after transaction commits successfully
	if err == nil {
		sessionKeyCache.Store(address, walletAddress)
	}

	return err
}

// GetWalletBySessionKey retrieves the wallet address associated with a given signer
func GetWalletBySessionKey(sessionKeyAddress string) string {
	// Check session key cache first
	if w, ok := sessionKeyCache.Load(sessionKeyAddress); ok {
		return w.(string)
	}
	return ""
}

// GetSessionKeysByWallet retrieves all session keys for a given wallet address
func GetSessionKeysByWallet(db *gorm.DB, walletAddress string) ([]SessionKey, error) {
	var sessionKeys []SessionKey

	err := db.Where("wallet_address = ?", walletAddress).
		Order("created_at DESC").
		Find(&sessionKeys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session keys for wallet %s: %w", walletAddress, err)
	}

	return sessionKeys, nil
}

// GetActiveSessionKeysByWallet retrieves only active (non-expired) session keys for a wallet
func GetActiveSessionKeysByWallet(db *gorm.DB, walletAddress string, listOpts *ListOptions) ([]SessionKey, error) {
	var sessionKeys []SessionKey

	query := db.Where("wallet_address = ? AND expires_at > ?",
		walletAddress, time.Now().UTC()).
		Order("created_at DESC")

	if listOpts != nil {
		if listOpts.Limit > 0 {
			query = query.Limit(int(listOpts.Limit))
		}
		if listOpts.Offset > 0 {
			query = query.Offset(int(listOpts.Offset))
		}
	}

	err := query.Find(&sessionKeys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active session keys for wallet %s: %w", walletAddress, err)
	}

	return sessionKeys, nil
}

// GetSessionKey retrieves a specific session key
func GetSessionKey(db *gorm.DB, sessionKeyAddress string) (*SessionKey, error) {
	var sk SessionKey
	err := db.Where("address = ?", sessionKeyAddress).First(&sk).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session key %s: %w", sessionKeyAddress, err)
	}
	return &sk, nil
}

// CalculateSessionKeySpending calculates total amount spent by a session key for a specific asset
func CalculateSessionKeySpending(db *gorm.DB, sessionKeyAddress string, assetSymbol string) (decimal.Decimal, error) {
	type result struct {
		TotalSpent decimal.Decimal
	}

	var res result
	err := db.Model(&Entry{}).
		Where("session_key = ? AND asset_symbol = ?", sessionKeyAddress, assetSymbol).
		Select("COALESCE(SUM(debit), 0) AS total_spent").
		Scan(&res).Error

	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to calculate session key spending: %w", err)
	}

	return res.TotalSpent, nil
}

// ValidateSessionKeySpending checks if a session key can spend the requested amount without exceeding its allowance
func ValidateSessionKeySpending(db *gorm.DB, sessionKeyAddress string, assetSymbol string, requestedAmount decimal.Decimal) error {
	sessionKey, err := GetSessionKey(db, sessionKeyAddress)
	if err != nil {
		return fmt.Errorf("operation denied: failed to get session key: %w", err)
	}
	if sessionKey.Application == AppNameClearnode {
		return nil // Do not enforce limitations on clearnode session keys
	}

	// Check if session key has expired
	if time.Now().UTC().After(sessionKey.ExpiresAt) {
		return fmt.Errorf("operation denied: session key expired")
	}

	// If no spending cap is set, deny the transaction
	if sessionKey.Allowance == nil {
		return fmt.Errorf("operation denied: session key has no allowance configured")
	}

	var allowances []Allowance
	if err := json.Unmarshal([]byte(*sessionKey.Allowance), &allowances); err != nil {
		return fmt.Errorf("failed to parse spending cap: %w", err)
	}

	// Find the allowance for this asset
	var allowedAmount decimal.Decimal
	found := false
	for _, allowance := range allowances {
		if allowance.Asset == assetSymbol {
			var err error
			allowedAmount, err = decimal.NewFromString(allowance.Amount)
			if err != nil {
				return fmt.Errorf("operation denied: failed to parse allowed amount: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("operation denied: asset %s not allowed in session key spending cap", assetSymbol)
	}

	currentSpending, err := CalculateSessionKeySpending(db, sessionKeyAddress, assetSymbol)
	if err != nil {
		return err
	}

	// Check if new spending would exceed the cap
	newTotal := currentSpending.Add(requestedAmount)
	if newTotal.GreaterThan(allowedAmount) {
		return fmt.Errorf("operation denied: insufficient session key allowance: %s required, %s available",
			requestedAmount, allowedAmount.Sub(currentSpending))
	}

	return nil
}

func ValidateSessionKeyApplication(db *gorm.DB, sessionKeyAddress string, appApplication string) error {
	sessionKey, err := GetSessionKey(db, sessionKeyAddress)
	if err != nil {
		return fmt.Errorf("failed to get session key: %w", err)
	}

	if sessionKey.Application == AppNameClearnode {
		return nil
	}

	if sessionKey.Application != appApplication {
		return fmt.Errorf("session key application mismatch: session key is for '%s', but app session is for '%s'",
			sessionKey.Application, appApplication)
	}

	return nil
}
