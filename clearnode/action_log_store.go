package main

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ActionLabel defines a custom type for user action labels.
type ActionLabel string

const (
// Example: LabelMaliciousChallenge ActionLabel = "malicious_challenge"
)

// UserActionLog is the data and database model for storing user action logs.
type UserActionLog struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    string      `gorm:"column:user_id;type:varchar(255);not null;index" json:"user_id"`
	Label     ActionLabel `gorm:"column:label;type:varchar(255);not null" json:"label"`
	Metadata  []byte      `gorm:"column:metadata;type:text" json:"metadata,omitempty"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"created_at"`
}

// TableName specifies the table name for the UserActionLog model.
func (UserActionLog) TableName() string {
	return "user_action_logs"
}

// Store is an interface that defines the contract for storing and retrieving user action logs.
type Store interface {
	Store(ctx context.Context, userID string, label ActionLabel, metadata []byte) error
	List(ctx context.Context, userID *string, label *ActionLabel, options *ListOptions) ([]UserActionLog, error)
	Count(ctx context.Context, userID *string, label *ActionLabel) (int64, error)
}

type ActionLogStore struct {
	db *gorm.DB
}

// NewActionLogStore creates a new ActionLogStore instance.
func NewActionLogStore(db *gorm.DB) *ActionLogStore {
	return &ActionLogStore{db: db}
}

// Store saves a new user action log record in the database.
func (s *ActionLogStore) Store(ctx context.Context, userID string, label ActionLabel, metadata []byte) error {
	record := &UserActionLog{
		UserID:   userID,
		Label:    label,
		Metadata: metadata,
	}
	return s.db.WithContext(ctx).Create(record).Error
}

// List retrieves user action logs with optional filtering and pagination.
func (s *ActionLogStore) List(ctx context.Context, userID *string, label *ActionLabel, options *ListOptions) ([]UserActionLog, error) {
	query := applyListOptions(s.db.WithContext(ctx), "created_at", SortTypeDescending, options)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if label != nil {
		query = query.Where("label = ?", *label)
	}

	var logs []UserActionLog
	err := query.Find(&logs).Error
	return logs, err
}

// Count returns the count of user action records, with optional filtering.
func (s *ActionLogStore) Count(ctx context.Context, userID *string, label *ActionLabel) (int64, error) {
	query := s.db.WithContext(ctx).Model(&UserActionLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if label != nil {
		query = query.Where("label = ?", *label)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}
