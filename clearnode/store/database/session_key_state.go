package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"gorm.io/gorm"
)

// SessionKeyStateV1 represents a session key state in the database.
// ID is Hash(user_address + session_key + version).
type SessionKeyStateV1 struct {
	ID             string                     `gorm:"column:id;primaryKey"`
	UserAddress    string                     `gorm:"column:user_address;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:1"`
	SessionKey     string                     `gorm:"column:session_key;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:2"`
	Version        uint64                     `gorm:"column:version;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:3"`
	ApplicationIDs []SessionKeyApplicationV1  `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	AppSessionIDs  []SessionKeyAppSessionIDV1 `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	ExpiresAt      time.Time                  `gorm:"column:expires_at;not null"`
	UserSig        string                     `gorm:"column:user_sig;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (SessionKeyStateV1) TableName() string {
	return "session_key_states_v1"
}

// SessionKeyApplicationV1 links a session key state to an application ID.
type SessionKeyApplicationV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	ApplicationID     string `gorm:"column:application_id;not null;primaryKey;priority:2;index"`
}

func (SessionKeyApplicationV1) TableName() string {
	return "session_key_applications_v1"
}

// SessionKeyAppSessionIDV1 links a session key state to an app session ID.
type SessionKeyAppSessionIDV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	AppSessionID      string `gorm:"column:app_session_id;not null;primaryKey;priority:2;index"`
}

func (SessionKeyAppSessionIDV1) TableName() string {
	return "session_key_app_sessions_v1"
}

// StoreSessionKeyState stores a new session key state version.
func (s *DBStore) StoreSessionKeyState(state app.AppSessionKeyStateV1) error {
	userAddress := strings.ToLower(state.UserAddress)
	sessionKey := strings.ToLower(state.SessionKey)

	id, err := app.GenerateSessionKeyStateIDV1(userAddress, sessionKey, state.Version)
	if err != nil {
		return fmt.Errorf("failed to generate session key state ID: %w", err)
	}

	dbState := SessionKeyStateV1{
		ID:          id,
		UserAddress: userAddress,
		SessionKey:  sessionKey,
		Version:     state.Version,
		ExpiresAt:   state.ExpiresAt.UTC(),
		UserSig:     state.UserSig,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.Create(&dbState).Error; err != nil {
		return fmt.Errorf("failed to store session key state: %w", err)
	}

	if len(state.ApplicationIDs) > 0 {
		applicationIDs := make([]SessionKeyApplicationV1, len(state.ApplicationIDs))
		for i, appID := range state.ApplicationIDs {
			applicationIDs[i] = SessionKeyApplicationV1{
				SessionKeyStateID: id,
				ApplicationID:     strings.ToLower(appID),
			}
		}
		if err := s.db.Create(&applicationIDs).Error; err != nil {
			return fmt.Errorf("failed to store application IDs: %w", err)
		}
	}

	if len(state.AppSessionIDs) > 0 {
		appSessionIDs := make([]SessionKeyAppSessionIDV1, len(state.AppSessionIDs))
		for i, sessID := range state.AppSessionIDs {
			appSessionIDs[i] = SessionKeyAppSessionIDV1{
				SessionKeyStateID: id,
				AppSessionID:      strings.ToLower(sessID),
			}
		}
		if err := s.db.Create(&appSessionIDs).Error; err != nil {
			return fmt.Errorf("failed to store app session IDs: %w", err)
		}
	}

	return nil
}

// GetLastKeyStates retrieves the latest session key states for a user with optional filtering.
// Returns only the highest-version row per session key that has not expired.
func (s *DBStore) GetLastKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)

	// Subquery to get the max version per session key for this user
	subQuery := s.db.Model(&SessionKeyStateV1{}).
		Select("user_address, session_key, MAX(version) as max_version").
		Where("user_address = ? AND expires_at > ?", wallet, time.Now().UTC()).
		Group("user_address, session_key")

	if sessionKey != nil && *sessionKey != "" {
		subQuery = subQuery.Where("session_key = ?", strings.ToLower(*sessionKey))
	}

	query := s.db.
		Joins("JOIN (?) AS latest ON session_key_states_v1.user_address = latest.user_address AND session_key_states_v1.session_key = latest.session_key AND session_key_states_v1.version = latest.max_version", subQuery).
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		Order("session_key_states_v1.updated_at DESC")

	var dbStates []SessionKeyStateV1
	if err := query.Find(&dbStates).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []app.AppSessionKeyStateV1{}, nil
		}
		return nil, fmt.Errorf("failed to get session key states: %w", err)
	}

	states := make([]app.AppSessionKeyStateV1, len(dbStates))
	for i, dbState := range dbStates {
		states[i] = dbSessionKeyStateToCore(&dbState)
	}

	return states, nil
}

// GetLatestSessionKeyState retrieves the latest version of a specific session key for a user.
// Returns nil if no state exists.
func (s *DBStore) GetLatestSessionKeyState(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var dbState SessionKeyStateV1
	err := s.db.
		Where("user_address = ? AND session_key = ? AND expires_at > ?", wallet, sessionKey, time.Now().UTC()).
		Order("version DESC").
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		First(&dbState).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest session key state: %w", err)
	}

	result := dbSessionKeyStateToCore(&dbState)
	return &result, nil
}

// GetAppSessionKeyOwner returns the user_address that owns the given session key
// authorized for the specified app session ID. Only non-expired, latest-version keys are considered.
func (s *DBStore) GetAppSessionKeyOwner(sessionKey, appSessionId string) (string, error) {
	sessionKey = strings.ToLower(sessionKey)
	appSessionId = strings.ToLower(appSessionId)

	// Subquery to get the application ID from the app session
	appSubQuery := s.db.Model(&AppSessionV1{}).Select("application").Where("id = ?", appSessionId)

	var dbState SessionKeyStateV1
	err := s.db.
		Joins("LEFT JOIN session_key_app_sessions_v1 ON session_key_app_sessions_v1.session_key_state_id = session_key_states_v1.id").
		Joins("LEFT JOIN session_key_applications_v1 ON session_key_applications_v1.session_key_state_id = session_key_states_v1.id").
		Where("session_key_states_v1.session_key = ? AND (session_key_app_sessions_v1.app_session_id = ? OR session_key_applications_v1.application_id = (?)) AND session_key_states_v1.expires_at > ?",
			sessionKey, appSessionId, appSubQuery, time.Now().UTC()).
		Order("session_key_states_v1.version DESC").
		First(&dbState).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("no active session key found for key %s and app session %s", sessionKey, appSessionId)
		}
		return "", fmt.Errorf("failed to get session key owner: %w", err)
	}

	return dbState.UserAddress, nil
}

func dbSessionKeyStateToCore(dbState *SessionKeyStateV1) app.AppSessionKeyStateV1 {
	applicationIDs := make([]string, len(dbState.ApplicationIDs))
	for i, a := range dbState.ApplicationIDs {
		applicationIDs[i] = a.ApplicationID
	}

	appSessionIDs := make([]string, len(dbState.AppSessionIDs))
	for i, a := range dbState.AppSessionIDs {
		appSessionIDs[i] = a.AppSessionID
	}

	return app.AppSessionKeyStateV1{
		UserAddress:    dbState.UserAddress,
		SessionKey:     dbState.SessionKey,
		Version:        dbState.Version,
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      dbState.ExpiresAt,
		UserSig:        dbState.UserSig,
	}
}
