package main

import (
	"time"

	"gorm.io/gorm"
)

// UserActionLog represents a user action log record in the database
type UserActionLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    string    `gorm:"column:user_id;type:varchar(255);not null;index"`
	Label     string    `gorm:"column:label;type:varchar(255);not null"`
	Timestamp uint64    `gorm:"column:timestamp;not null;index"`
	Metadata  []byte    `gorm:"column:metadata;type:text"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName specifies the table name for the UserActionLog model
func (UserActionLog) TableName() string {
	return "user_action_log"
}

// UserActionLogStore handles user action log storage and retrieval
type UserActionLogStore struct {
	db *gorm.DB
}

// NewUserActionLogStore creates a new UserActionLogStore instance
func NewUserActionLogStore(db *gorm.DB) *UserActionLogStore {
	return &UserActionLogStore{db: db}
}

// StoreUserAction stores a user action log record in the database
func (s *UserActionLogStore) StoreUserAction(userID, label string, metadata []byte) error {
	record := &UserActionLog{
		UserID:    userID,
		Label:     label,
		Timestamp: uint64(time.Now().Unix()),
		Metadata:  metadata,
		CreatedAt: time.Now(),
	}

	return s.db.Create(record).Error
}

// ListUserActions retrieves user action logs with optional filtering and pagination.
func (s *UserActionLogStore) ListUserActions(userID *string, label *string, options *ListOptions) ([]UserActionLog, error) {
	query := applyListOptions(s.db, "timestamp", SortTypeDescending, options)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if label != nil {
		query = query.Where("label = ?", *label)
	}

	var actions []UserActionLog
	err := query.Find(&actions).Error
	return actions, err
}

// CountUserActionsByUserAndLabel returns the count of user action records for a specific user and label
func (s *UserActionLogStore) CountUserActionsByUserAndLabel(userID, label string) (int64, error) {
	var count int64
	err := s.db.Model(&UserActionLog{}).Where("user_id = ? AND label = ?", userID, label).Count(&count).Error
	return count, err
}
