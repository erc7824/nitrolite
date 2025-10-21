package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SessionKey represents a ledger layer session key
type SessionKey struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	SignerAddress string `gorm:"column:signer_address;uniqueIndex;not null"`
	WalletAddress string `gorm:"column:wallet_address;index;not null"`

	AppName       string    `gorm:"column:app_name;not null"`
	AppAddress    string    `gorm:"column:app_address;not null;default:''"`
	Allowance     *string   `gorm:"column:allowance;type:text"`      // JSON serialized allowances
	UsedAllowance *string   `gorm:"column:used_allowance;type:text"` // JSON serialized used amounts
	Scope         string    `gorm:"column:scope;not null;default:'all'"`
	ExpiresAt     time.Time `gorm:"column:expires_at;not null"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SessionKey) TableName() string {
	return "session_keys"
}

// loadSessionKeyCache populates the cache with session keys
func loadSessionKeyCache(db *gorm.DB) error {
	var sessionKeys []SessionKey
	if err := db.Where("expires_at > ?", time.Now().UTC()).Find(&sessionKeys).Error; err != nil {
		return err
	}
	for _, sk := range sessionKeys {
		sessionKeyCache.Store(sk.SignerAddress, sk.WalletAddress)
	}
	return nil
}

// AddSessionKey stores a new session key with its metadata
// Only one session key per wallet+app combination is allowed - adding a new one invalidates existing ones
func AddSessionKey(db *gorm.DB, walletAddress, signerAddress, applicationName, applicationAddress, scope string, allowances []Allowance, expirationTime time.Time) error {
	if applicationName == "clearnode" && len(allowances) == 0 {
		return AddSigner(db, walletAddress, signerAddress)
	}

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
		err := tx.Where("wallet_address = ? AND app_name = ?",
			walletAddress, applicationName).Find(&existingKeys).Error
		if err != nil {
			return fmt.Errorf("failed to check existing session keys: %w", err)
		}

		// Remove existing session keys for this app (invalidate them)
		for _, existingKey := range existingKeys {
			if err := tx.Delete(&existingKey).Error; err != nil {
				return fmt.Errorf("failed to remove existing session key: %w", err)
			}
			sessionKeyCache.Delete(existingKey.SignerAddress)
		}

		spendingCapJSON, err := json.Marshal(allowances)
		if err != nil {
			return fmt.Errorf("failed to serialize spending cap: %w", err)
		}

		usedAllowanceJSON, err := json.Marshal([]Allowance{})
		if err != nil {
			return fmt.Errorf("failed to serialize used allowance: %w", err)
		}

		spendingCapStr := string(spendingCapJSON)
		usedAllowanceStr := string(usedAllowanceJSON)

		sessionKey := &SessionKey{
			SignerAddress: signerAddress,
			WalletAddress: walletAddress,
			AppName:       applicationName,
			AppAddress:    applicationAddress,
			Allowance:     &spendingCapStr,
			UsedAllowance: &usedAllowanceStr,
			Scope:         scope,
			ExpiresAt:     expirationTime,
		}

		return tx.Create(sessionKey).Error
	})

	// Update cache only after transaction commits successfully
	if err == nil {
		sessionKeyCache.Store(signerAddress, walletAddress)
	}

	return err
}

// GetWalletBySigner retrieves the wallet address associated with a given signer
func GetWalletBySigner(signer string) string {
	// Check session key cache first
	if w, ok := sessionKeyCache.Load(signer); ok {
		return w.(string)
	}
	// Check custody signer cache
	if w, ok := custodySignerCache.Load(signer); ok {
		return w.(string)
	}
	return ""
}

// IsSessionKey checks if the given address is a session key (not a custody signer)
func IsSessionKey(signerAddress string) bool {
	_, ok := sessionKeyCache.Load(signerAddress)
	return ok
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

// UpdateSessionKeyUsage recalculates and updates the used allowance for a session key based on ledger entries
func UpdateSessionKeyUsage(db *gorm.DB, signerAddress string) error {
	sessionKey, err := GetSessionKeyBySigner(db, signerAddress)
	if err != nil {
		return fmt.Errorf("failed to get session key: %w", err)
	}

	if sessionKey.Allowance == nil {
		return fmt.Errorf("session key %s has no spending cap configured", signerAddress)
	}

	var allowances []Allowance
	if err := json.Unmarshal([]byte(*sessionKey.Allowance), &allowances); err != nil {
		return fmt.Errorf("failed to parse spending cap: %w", err)
	}

	// Calculate used amounts for each asset
	var usedAllowances []Allowance
	for _, allowance := range allowances {
		usedAmount, err := GetSessionKeySpending(db, signerAddress, allowance.Asset)
		if err != nil {
			return fmt.Errorf("failed to get spending for asset %s: %w", allowance.Asset, err)
		}

		usedAllowances = append(usedAllowances, Allowance{
			Asset:  allowance.Asset,
			Amount: usedAmount.String(),
		})
	}

	usedAllowanceJSON, err := json.Marshal(usedAllowances)
	if err != nil {
		return fmt.Errorf("failed to serialize used allowance: %w", err)
	}

	usedAllowanceStr := string(usedAllowanceJSON)
	return db.Model(&SessionKey{}).
		Where("signer_address = ?", signerAddress).
		Update("used_allowance", usedAllowanceStr).Error
}

// GetSessionKeyBySigner retrieves a specific session key by its signer address
func GetSessionKeyBySigner(db *gorm.DB, signerAddress string) (*SessionKey, error) {
	var sk SessionKey
	err := db.Where("signer_address = ?", signerAddress).First(&sk).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session key %s: %w", signerAddress, err)
	}
	return &sk, nil
}

// GetSessionKeySpending calculates total amount spent by a session key for a specific asset
func GetSessionKeySpending(db *gorm.DB, sessionKeyAddress string, assetSymbol string) (decimal.Decimal, error) {
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
	if _, ok := custodySignerCache.Load(sessionKeyAddress); ok {
		return nil // Custody signers don't have spending limitations
	}

	sessionKey, err := GetSessionKeyBySigner(db, sessionKeyAddress)
	if err != nil {
		return fmt.Errorf("failed to get session key: %w", err)
	}

	// Check if session key has expired
	if time.Now().UTC().After(sessionKey.ExpiresAt) {
		return fmt.Errorf("session key expired")
	}

	// If no spending cap is set, deny the transaction
	if sessionKey.Allowance == nil {
		return fmt.Errorf("session key has no allowance")
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
				return fmt.Errorf("failed to parse allowed amount: %w", err)
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("asset %s not allowed in session key spending cap", assetSymbol)
	}

	currentSpending, err := GetSessionKeySpending(db, sessionKeyAddress, assetSymbol)
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
