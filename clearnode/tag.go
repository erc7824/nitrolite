package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"gorm.io/gorm"
)

// UserTagModel represents the user tag model in the database.
type UserTagModel struct {
	Wallet string `gorm:"column:wallet;primaryKey"`
	Tag    string `gorm:"column:tag;uniqueIndex;not null"`
}

func (UserTagModel) TableName() string {
	return "user_tags"
}

// GenerateOrRetrieveUserTag checks if a user tag exists for the given wallet.
// If it does, it returns the existing tag. If not, it generates a new unique tag
// and stores it in the database, retrying up to 10 times if necessary.
func GenerateOrRetrieveUserTag(db *gorm.DB, wallet string) (*UserTagModel, error) {
	// Start transaction
	tx := db.Begin()
	defer tx.Rollback()

	// Check if the tag already exists
	var existingUserTag UserTagModel
	if err := tx.Where("wallet = ?", wallet).First(&existingUserTag).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check existing user tag: %v", err)
		}
	}

	// If it exists, return the existing tag
	if existingUserTag.Tag != "" {
		return &existingUserTag, nil
	}

	for i := 0; i < 10; i++ {
		generatedTag := GenerateRandomAlphanumericTag()
		model := &UserTagModel{
			Wallet: wallet,
			Tag:    generatedTag,
		}

		if err := tx.Create(model).Error; err != nil {
			// TODO: log an error and retry
			continue
		}

		// Commit the transaction
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %v", err)
		}

		return model, nil
	}

	return nil, fmt.Errorf("failed to generate a unique tag after multiple attempts")
}

// GetUserTagByWallet retrieves the user tag associated with a given wallet address.
func GetUserTagByWallet(db *gorm.DB, wallet string) (string, error) {
	if wallet == "" {
		return "", fmt.Errorf("wallet address cannot be empty")
	}

	var model UserTagModel
	if err := db.Where("wallet = ?", wallet).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", err
		}
		return "", fmt.Errorf("failed to retrieve record: %v", err)
	}
	return model.Tag, nil
}

// GetWalletByTag retrieves the wallet address associated with a given user tag.
func GetWalletByTag(db *gorm.DB, tag string) (UserTagModel, error) {
	if tag == "" {
		return UserTagModel{}, fmt.Errorf("tag cannot be empty")
	}

	tag = strings.ToUpper(tag)

	var model UserTagModel
	if err := db.Where("tag = ?", tag).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return UserTagModel{}, fmt.Errorf("no associated wallet for tag: %s", tag)
		}
		return UserTagModel{}, fmt.Errorf("failed to retrieve record: %v", err)
	}
	return model, nil
}

// GenerateRandomAlphanumericTag generates a random alphanumeric tag of length 6.
func GenerateRandomAlphanumericTag() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 6)
	charsetMaxIndex := big.NewInt(int64(len(charset) - 1))

	for i := range result {
		// Use crypto/rand.Int() for cryptographically secure random numbers
		randomIndex, err := rand.Int(rand.Reader, charsetMaxIndex)
		if err != nil {
			// If crypto/rand fails, this is a serious error that should cause a panic
			// since we can't generate secure random numbers
			panic(fmt.Sprintf("failed to generate secure random number: %v", err))
		}
		result[i] = charset[randomIndex.Int64()]
	}
	return string(result)
}
