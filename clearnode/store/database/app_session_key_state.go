package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AppSessionKeyStateV1 represents a session key state in the database.
// ID is Hash(user_address + session_key + version).
type AppSessionKeyStateV1 struct {
	ID             string                        `gorm:"column:id;primaryKey"`
	UserAddress    string                        `gorm:"column:user_address;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:1"`
	SessionKey     string                        `gorm:"column:session_key;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:2"`
	Version        uint64                        `gorm:"column:version;not null;uniqueIndex:idx_session_key_states_v1_user_key_ver,priority:3"`
	ApplicationIDs []AppSessionKeyApplicationV1  `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	AppSessionIDs  []AppSessionKeyAppSessionIDV1 `gorm:"foreignKey:SessionKeyStateID;references:ID"`
	ExpiresAt      time.Time                     `gorm:"column:expires_at;not null"`
	UserSig        string                        `gorm:"column:user_sig;not null"`
	CreatedAt      time.Time
}

func (AppSessionKeyStateV1) TableName() string {
	return "app_session_key_states_v1"
}

// AppSessionKeyHeadV1 represents the current head (latest version) for a (user_address, session_key) pair
// This table provides O(1) reads and proper row-level locking for session key updates
type AppSessionKeyHeadV1 struct {
	UserAddress string    `gorm:"column:user_address;primaryKey"`
	SessionKey  string    `gorm:"column:session_key;primaryKey"`
	Version     uint64    `gorm:"column:version;not null"`
	ExpiresAt   time.Time `gorm:"column:expires_at;not null"`
	UserSig     string    `gorm:"column:user_sig;not null"`

	// Reference to history
	HistoryID *string `gorm:"column:history_id"` // References current state in app_session_key_states_v1

	UpdatedAt time.Time
}

func (AppSessionKeyHeadV1) TableName() string {
	return "app_session_key_heads_v1"
}

// SessionKeyApplicationV1 links a session key state to an application ID.
type AppSessionKeyApplicationV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	ApplicationID     string `gorm:"column:application_id;not null;primaryKey;priority:2;index"`
}

func (AppSessionKeyApplicationV1) TableName() string {
	return "app_session_key_applications_v1"
}

// AppSessionKeyAppSessionIDV1 links a session key state to an app session ID.
type AppSessionKeyAppSessionIDV1 struct {
	SessionKeyStateID string `gorm:"column:session_key_state_id;not null;primaryKey;priority:1"`
	AppSessionID      string `gorm:"column:app_session_id;not null;primaryKey;priority:2;index"`
}

func (AppSessionKeyAppSessionIDV1) TableName() string {
	return "app_session_key_app_sessions_v1"
}

// StoreAppSessionKeyState stores a new session key state version.
// IMPORTANT: This method MUST be called within a transaction (ExecuteInTransaction).
// It assumes the head row was already locked via GetLastAppSessionKeyState.
// It inserts the new state into history and updates/creates the head row.
func (s *DBStore) StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error {
	// Safety guard: ensure we're in a transaction
	if !s.inTx {
		return fmt.Errorf("StoreAppSessionKeyState must be called within a transaction (use ExecuteInTransaction)")
	}

	userAddress := strings.ToLower(state.UserAddress)
	sessionKey := strings.ToLower(state.SessionKey)

	id, err := app.GenerateSessionKeyStateIDV1(userAddress, sessionKey, state.Version)
	if err != nil {
		return fmt.Errorf("failed to generate session key state ID: %w", err)
	}

	// Step 1: Insert into history (app_session_key_states_v1)
	dbState := AppSessionKeyStateV1{
		ID:          id,
		UserAddress: userAddress,
		SessionKey:  sessionKey,
		Version:     state.Version,
		ExpiresAt:   state.ExpiresAt.UTC(),
		UserSig:     state.UserSig,
	}

	if err := s.db.Create(&dbState).Error; err != nil {
		return fmt.Errorf("failed to store session key state: %w", err)
	}

	// Store related ApplicationIDs
	if len(state.ApplicationIDs) > 0 {
		applicationIDs := make([]AppSessionKeyApplicationV1, len(state.ApplicationIDs))
		for i, appID := range state.ApplicationIDs {
			applicationIDs[i] = AppSessionKeyApplicationV1{
				SessionKeyStateID: id,
				ApplicationID:     strings.ToLower(appID),
			}
		}
		if err := s.db.Create(&applicationIDs).Error; err != nil {
			return fmt.Errorf("failed to store application IDs: %w", err)
		}
	}

	// Store related AppSessionIDs
	if len(state.AppSessionIDs) > 0 {
		appSessionIDs := make([]AppSessionKeyAppSessionIDV1, len(state.AppSessionIDs))
		for i, sessID := range state.AppSessionIDs {
			appSessionIDs[i] = AppSessionKeyAppSessionIDV1{
				SessionKeyStateID: id,
				AppSessionID:      strings.ToLower(sessID),
			}
		}
		if err := s.db.Create(&appSessionIDs).Error; err != nil {
			return fmt.Errorf("failed to store app session IDs: %w", err)
		}
	}

	// Step 2: Upsert head (INSERT ... ON CONFLICT UPDATE)
	// Use raw SQL for proper ON CONFLICT behavior
	now := time.Now().UTC()
	err = s.db.Exec(`
		INSERT INTO app_session_key_heads_v1
		(user_address, session_key, version, expires_at, user_sig, history_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_address, session_key)
		DO UPDATE SET
			version = EXCLUDED.version,
			expires_at = EXCLUDED.expires_at,
			user_sig = EXCLUDED.user_sig,
			history_id = EXCLUDED.history_id,
			updated_at = EXCLUDED.updated_at
		WHERE app_session_key_heads_v1.version < EXCLUDED.version
	`, userAddress, sessionKey, state.Version, state.ExpiresAt.UTC(), state.UserSig, id, now).Error

	if err != nil {
		return fmt.Errorf("failed to upsert app session key head: %w", err)
	}

	return nil
}

// GetLastAppSessionKeyStates retrieves the latest session key states for a user with optional filtering.
// Returns only non-expired session keys. Optimized to use 3 queries total regardless of N.
func (s *DBStore) GetLastAppSessionKeyStates(wallet string, sessionKey *string) ([]app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)

	// Query 1: Fetch all heads
	query := s.db.Model(&AppSessionKeyHeadV1{}).
		Where("user_address = ? AND expires_at > ?", wallet, time.Now().UTC()).
		Order("updated_at DESC")

	if sessionKey != nil && *sessionKey != "" {
		query = query.Where("session_key = ?", strings.ToLower(*sessionKey))
	}

	var heads []AppSessionKeyHeadV1
	if err := query.Find(&heads).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return []app.AppSessionKeyStateV1{}, nil
		}
		return nil, fmt.Errorf("failed to get session key heads: %w", err)
	}

	if len(heads) == 0 {
		return []app.AppSessionKeyStateV1{}, nil
	}

	// Collect all history IDs for batch fetching
	historyIDs := make([]string, 0, len(heads))
	headByHistoryID := make(map[string]*AppSessionKeyHeadV1)
	for i := range heads {
		if heads[i].HistoryID != nil {
			historyIDs = append(historyIDs, *heads[i].HistoryID)
			headByHistoryID[*heads[i].HistoryID] = &heads[i]
		}
	}

	// Query 2: Batch fetch all ApplicationIDs
	var appIDLinks []AppSessionKeyApplicationV1
	if len(historyIDs) > 0 {
		err := s.db.Model(&AppSessionKeyApplicationV1{}).
			Where("session_key_state_id IN ?", historyIDs).
			Find(&appIDLinks).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to fetch application IDs: %w", err)
		}
	}

	// Query 3: Batch fetch all AppSessionIDs
	var appSessionIDLinks []AppSessionKeyAppSessionIDV1
	if len(historyIDs) > 0 {
		err := s.db.Model(&AppSessionKeyAppSessionIDV1{}).
			Where("session_key_state_id IN ?", historyIDs).
			Find(&appSessionIDLinks).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to fetch app session IDs: %w", err)
		}
	}

	// Build maps for fast lookup
	appIDsByHistoryID := make(map[string][]string)
	for _, link := range appIDLinks {
		appIDsByHistoryID[link.SessionKeyStateID] = append(appIDsByHistoryID[link.SessionKeyStateID], link.ApplicationID)
	}

	appSessionIDsByHistoryID := make(map[string][]string)
	for _, link := range appSessionIDLinks {
		appSessionIDsByHistoryID[link.SessionKeyStateID] = append(appSessionIDsByHistoryID[link.SessionKeyStateID], link.AppSessionID)
	}

	// Assemble results
	states := make([]app.AppSessionKeyStateV1, 0, len(heads))
	for _, head := range heads {
		state := app.AppSessionKeyStateV1{
			UserAddress: head.UserAddress,
			SessionKey:  head.SessionKey,
			Version:     head.Version,
			ExpiresAt:   head.ExpiresAt,
			UserSig:     head.UserSig,
		}

		// Add related IDs if history exists
		if head.HistoryID != nil {
			state.ApplicationIDs = appIDsByHistoryID[*head.HistoryID]
			state.AppSessionIDs = appSessionIDsByHistoryID[*head.HistoryID]
		}

		// Ensure non-nil slices
		if state.ApplicationIDs == nil {
			state.ApplicationIDs = []string{}
		}
		if state.AppSessionIDs == nil {
			state.AppSessionIDs = []string{}
		}

		states = append(states, state)
	}

	return states, nil
}

// GetLastAppSessionKeyVersion returns the latest version of a non-expired session key state for a user.
// Returns 0 if no state exists. Uses head table for O(1) lookup.
func (s *DBStore) GetLastAppSessionKeyVersion(wallet, sessionKey string) (uint64, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	var result struct {
		Version uint64
	}
	err := s.db.Model(&AppSessionKeyHeadV1{}).
		Select("version").
		Where("user_address = ? AND session_key = ? AND expires_at > ?", wallet, sessionKey, time.Now().UTC()).
		Take(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to check session key state: %w", err)
	}

	return result.Version, nil
}

// GetLastAppSessionKeyState retrieves the latest version of a specific session key for a user.
// When called within a transaction (s.inTx == true), this method acquires a FOR UPDATE lock
// on the head row. Returns nil if no state exists or if expired.
func (s *DBStore) GetLastAppSessionKeyState(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error) {
	wallet = strings.ToLower(wallet)
	sessionKey = strings.ToLower(sessionKey)

	if s.inTx {
		return s.getLastAppSessionKeyStateWithLock(wallet, sessionKey)
	}

	return s.getLastAppSessionKeyStateNoLock(wallet, sessionKey)
}

// getLastAppSessionKeyStateWithLock reads the head with FOR UPDATE lock (transaction mode)
func (s *DBStore) getLastAppSessionKeyStateWithLock(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error) {
	var head AppSessionKeyHeadV1

	// Try to lock the head row with FOR UPDATE
	query := s.db.Model(&AppSessionKeyHeadV1{}).
		Where("user_address = ? AND session_key = ? AND expires_at > ?", wallet, sessionKey, time.Now().UTC()).
		Clauses(clause.Locking{Strength: "UPDATE"})

	err := query.First(&head).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No head found or expired
			return nil, nil
		}
		return nil, fmt.Errorf("failed to lock app session key head: %w", err)
	}

	// Need to fetch ApplicationIDs and AppSessionIDs from history
	if head.HistoryID == nil {
		// No history reference, return basic state
		return &app.AppSessionKeyStateV1{
			UserAddress: head.UserAddress,
			SessionKey:  head.SessionKey,
			Version:     head.Version,
			ExpiresAt:   head.ExpiresAt,
			UserSig:     head.UserSig,
		}, nil
	}

	// Fetch full state from history to get related IDs
	var dbState AppSessionKeyStateV1
	err = s.db.
		Where("id = ?", *head.HistoryID).
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// History row not found, return head data only
			return &app.AppSessionKeyStateV1{
				UserAddress: head.UserAddress,
				SessionKey:  head.SessionKey,
				Version:     head.Version,
				ExpiresAt:   head.ExpiresAt,
				UserSig:     head.UserSig,
			}, nil
		}
		return nil, fmt.Errorf("failed to get app session key history: %w", err)
	}

	result := dbSessionKeyStateToCore(&dbState)
	return &result, nil
}

// getLastAppSessionKeyStateNoLock reads the head without locking (non-transaction mode)
func (s *DBStore) getLastAppSessionKeyStateNoLock(wallet, sessionKey string) (*app.AppSessionKeyStateV1, error) {
	var head AppSessionKeyHeadV1

	err := s.db.Model(&AppSessionKeyHeadV1{}).
		Where("user_address = ? AND session_key = ? AND expires_at > ?", wallet, sessionKey, time.Now().UTC()).
		First(&head).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get app session key head: %w", err)
	}

	// Fetch full state from history to get related IDs
	if head.HistoryID == nil {
		return &app.AppSessionKeyStateV1{
			UserAddress: head.UserAddress,
			SessionKey:  head.SessionKey,
			Version:     head.Version,
			ExpiresAt:   head.ExpiresAt,
			UserSig:     head.UserSig,
		}, nil
	}

	var dbState AppSessionKeyStateV1
	err = s.db.
		Where("id = ?", *head.HistoryID).
		Preload("ApplicationIDs").
		Preload("AppSessionIDs").
		First(&dbState).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &app.AppSessionKeyStateV1{
				UserAddress: head.UserAddress,
				SessionKey:  head.SessionKey,
				Version:     head.Version,
				ExpiresAt:   head.ExpiresAt,
				UserSig:     head.UserSig,
			}, nil
		}
		return nil, fmt.Errorf("failed to get app session key history: %w", err)
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

	var dbState AppSessionKeyStateV1
	err := s.db.
		Joins("LEFT JOIN app_session_key_app_sessions_v1 ON app_session_key_app_sessions_v1.session_key_state_id = app_session_key_states_v1.id").
		Joins("LEFT JOIN app_session_key_applications_v1 ON app_session_key_applications_v1.session_key_state_id = app_session_key_states_v1.id").
		Where("app_session_key_states_v1.session_key = ? AND (app_session_key_app_sessions_v1.app_session_id = ? OR app_session_key_applications_v1.application_id = (?)) AND app_session_key_states_v1.expires_at > ?",
			sessionKey, appSessionId, appSubQuery, time.Now().UTC()).
		Order("app_session_key_states_v1.version DESC").
		First(&dbState).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("no active session key found for key %s and app session %s", sessionKey, appSessionId)
		}
		return "", fmt.Errorf("failed to get session key owner: %w", err)
	}

	return dbState.UserAddress, nil
}

func dbSessionKeyStateToCore(dbState *AppSessionKeyStateV1) app.AppSessionKeyStateV1 {
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
