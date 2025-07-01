package main

import (
	"fmt"
	"math/rand"

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
			return "", fmt.Errorf("user tag does not exist for wallet: %s", wallet)
		}
		return "", fmt.Errorf("failed to retrieve record: %v", err)
	}
	return model.Tag, nil
}

// GetWalletByTag retrieves the wallet address associated with a given user tag.
func GetWalletByTag(db *gorm.DB, tag string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}

	var model UserTagModel
	if err := db.Where("tag = ?", tag).First(&model).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve record: %v", err)
	}
	return model.Wallet, nil
}

// GenerateRandomAlphanumericTag generates a random alphanumeric tag of length 8.
func GenerateRandomAlphanumericTag() string {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, 8)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
