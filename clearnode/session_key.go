package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SessionKey represents both a session key and its signer relationship
type SessionKey struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	SignerAddress string `gorm:"column:signer_address;index;not null"`
	WalletAddress string `gorm:"column:wallet_address;index;not null"`

	SignerType string `gorm:"column:signer_type;not null;default:'session'"` // 'session', 'custody'

	// Session metadata fields (nullable for custody signers)
	ApplicationName *string    `gorm:"column:application_name"`
	SpendingCap     *string    `gorm:"column:spending_cap;type:text"`   // JSON serialized allowances
	UsedAllowance   *string    `gorm:"column:used_allowance;type:text"` // JSON serialized used amounts
	Scope           *string    `gorm:"column:scope"`
	ExpirationTime  *time.Time `gorm:"column:expiration_time"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (SessionKey) TableName() string {
	return "session_keys"
}

var sessionKeyCache sync.Map

// loadSessionKeyCache populates the sessionKeyCache from the database
func loadSessionKeyCache(db *gorm.DB) error {
	var allKeys []SessionKey
	if err := db.Find(&allKeys).Error; err != nil {
		return err
	}
	for _, sk := range allKeys {
		sessionKeyCache.Store(sk.SignerAddress, sk.WalletAddress)
	}
	return nil
}

// AddSessionKey stores a new session key with its metadata (replaces AddSessionKey + AddSigner)
// Only one session key per wallet+app combination is allowed - adding a new one invalidates existing ones
func AddSessionKey(db *gorm.DB, walletAddress, signerAddress, applicationName, scope string, allowances []Allowance, expirationTime time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Check for and remove existing session key for this wallet+app combination
		var existingKeys []SessionKey
		err := tx.Where("wallet_address = ? AND application_name = ? AND signer_type = ?",
			walletAddress, applicationName, "session").Find(&existingKeys).Error
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
			SignerAddress:   signerAddress,
			WalletAddress:   walletAddress,
			ApplicationName: &applicationName,
			SpendingCap:     &spendingCapStr,
			UsedAllowance:   &usedAllowanceStr,
			Scope:           &scope,
			ExpirationTime:  &expirationTime,
			SignerType:      "session",
		}

		err = tx.Create(sessionKey).Error
		if err != nil {
			return err
		}

		sessionKeyCache.Store(signerAddress, walletAddress)
		return nil
	})
}

// AddSigner creates a signer without session metadata (for custody, etc.) - replaces old AddSigner
func AddSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	// Check if signer already exists
	if w, ok := sessionKeyCache.Load(signerAddress); ok {
		if w.(string) == walletAddress {
			return nil // Already exists with same wallet
		}
		return fmt.Errorf("signer is already in use for another wallet")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Before adding a new signer, we need to ensure that the relationship is valid.
		// 1. A wallet address can't be used as a signer for another wallet.
		// 2. An address can't be used as a wallet if it is already a signer for another wallet.
		// 3. A signer can only be associated with one wallet.

		// A wallet address can't be used as a signer for another wallet.
		var count int64
		if err := tx.Model(&SessionKey{}).Where("wallet_address = ?", signerAddress).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("cannot use a wallet as a signer")
		}

		// Address can't be used as a wallet if it is already a signer for another wallet.
		if err := tx.Model(&SessionKey{}).Where("signer_address = ?", walletAddress).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("wallet is already in use as a signer")
		}

		// Signer cannot be used for another wallet if it already exists.
		var existingSigner SessionKey
		err := tx.Where("signer_address = ?", signerAddress).First(&existingSigner).Error
		switch {
		case err == nil:
			if existingSigner.WalletAddress == walletAddress {
				sessionKeyCache.Store(signerAddress, walletAddress)
				return nil
			}
			return fmt.Errorf("signer is already in use for another wallet")

		case err == gorm.ErrRecordNotFound:
			basicSigner := &SessionKey{
				SignerAddress: signerAddress,
				WalletAddress: walletAddress,
				SignerType:    "custody",
			}

			if err := tx.Create(basicSigner).Error; err != nil {
				return err
			}
			sessionKeyCache.Store(signerAddress, walletAddress)
			return nil

		default:
			return err
		}
	})
}

// GetWalletBySigner retrieves the wallet address associated with a given signer
func GetWalletBySigner(signerAddress string) string {
	if w, ok := sessionKeyCache.Load(signerAddress); ok {
		return w.(string)
	}
	return ""
}

// IsSessionKey checks if the given address is a session key
func IsSessionKey(db *gorm.DB, signerAddress string) bool {
	var count int64
	db.Model(&SessionKey{}).Where("signer_address = ? AND signer_type = ?", signerAddress, "session").Count(&count)
	return count > 0
}

// GetSessionKeysByWallet retrieves all session keys for a given wallet address
func GetSessionKeysByWallet(db *gorm.DB, walletAddress string) ([]SessionKey, error) {
	var sessionKeys []SessionKey

	err := db.Where("wallet_address = ? AND signer_type = ?", walletAddress, "session").
		Order("created_at DESC").
		Find(&sessionKeys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session keys for wallet %s: %w", walletAddress, err)
	}

	return sessionKeys, nil
}

// GetActiveSessionKeysByWallet retrieves only active (non-expired) session keys for a wallet
func GetActiveSessionKeysByWallet(db *gorm.DB, walletAddress string) ([]SessionKey, error) {
	var sessionKeys []SessionKey

	err := db.Where("wallet_address = ? AND signer_type = ? AND (expiration_time IS NULL OR expiration_time > ?)",
		walletAddress, "session", time.Now()).
		Order("created_at DESC").
		Find(&sessionKeys).Error

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active session keys for wallet %s: %w", walletAddress, err)
	}

	return sessionKeys, nil
}

// UpdateSessionKeyUsage recalculates and updates the used allowance for a session key based on ledger entries
func UpdateSessionKeyUsage(db *gorm.DB, signerAddress string) error {
	sessionKey, err := GetSessionKeyByKey(db, signerAddress)
	if err != nil {
		return fmt.Errorf("failed to get session key: %w", err)
	}

	if sessionKey.SpendingCap == nil {
		return fmt.Errorf("session key %s has no spending cap configured", signerAddress)
	}

	var allowances []Allowance
	if err := json.Unmarshal([]byte(*sessionKey.SpendingCap), &allowances); err != nil {
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
		Where("signer_address = ? AND signer_type = ?", signerAddress, "session").
		Update("used_allowance", usedAllowanceStr).Error
}

// GetSessionKeyByKey retrieves a specific session key by its signer address
func GetSessionKeyByKey(db *gorm.DB, signerAddress string) (*SessionKey, error) {
	var sk SessionKey
	err := db.Where("signer_address = ? AND signer_type = ?", signerAddress, "session").First(&sk).Error
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session key %s: %w", signerAddress, err)
	}
	return &sk, nil
}

// DeleteSessionKey removes a session key from the database
func DeleteSessionKey(db *gorm.DB, signerAddress string) error {
	err := db.Where("signer_address = ? AND signer_type = ?", signerAddress, "session").Delete(&SessionKey{}).Error
	if err != nil {
		return err
	}
	sessionKeyCache.Delete(signerAddress)
	return nil
}

// RemoveSigner deletes a signer from the database and cache
func RemoveSigner(db *gorm.DB, walletAddress, signerAddress string) error {
	err := db.Where("signer_address = ? AND wallet_address = ?", signerAddress, walletAddress).Delete(&SessionKey{}).Error
	if err != nil {
		return err
	}
	sessionKeyCache.Delete(signerAddress)
	return nil
}

// GetAllSignersForWallet gets all signers (any type) for a wallet
func GetAllSignersForWallet(db *gorm.DB, walletAddress string) ([]SessionKey, error) {
	var signers []SessionKey
	err := db.Where("wallet_address = ?", walletAddress).Find(&signers).Error
	return signers, err
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
	sessionKey, err := GetSessionKeyByKey(db, sessionKeyAddress)
	if err != nil {
		return fmt.Errorf("failed to get session key: %w", err)
	}

	// If no spending cap is set, deny the transaction
	if sessionKey.SpendingCap == nil {
		return fmt.Errorf("session key has no spending cap configured - transaction denied")
	}

	var allowances []Allowance
	if err := json.Unmarshal([]byte(*sessionKey.SpendingCap), &allowances); err != nil {
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
