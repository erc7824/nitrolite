package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthManager(t *testing.T) {
	authManager := NewAuthManager()
	require.NotNil(t, authManager)

	// Generate a challenge
	challenge, err := authManager.GenerateChallenge("addr")
	require.NoError(t, err)
	require.NotEmpty(t, challenge)

	// Verify challenge exists
	authManager.challengesMu.RLock()
	savedChallenge, exists := authManager.challenges[challenge]
	authManager.challengesMu.RUnlock()
	require.True(t, exists)
	assert.False(t, savedChallenge.Completed)
}

func TestAuthManagerSessionManagement(t *testing.T) {
	am := &AuthManager{
		challenges:    make(map[uuid.UUID]*Challenge),
		challengeTTL:  250 * time.Millisecond,
		authSessions:  make(map[string]time.Time),
		sessionTTL:    500 * time.Millisecond,
		cleanupTicker: time.NewTicker(10 * time.Minute),
		maxChallenges: 1000,
	}

	// Add a test session
	testAddr := "0x1234567890123456789012345678901234567890"
	am.registerAuthSession(testAddr)

	// Verify session is valid
	valid := am.ValidateSession(testAddr)
	assert.True(t, valid)

	// Update session
	time.Sleep(125 * time.Millisecond)
	updated := am.UpdateSession(testAddr)
	assert.True(t, updated)

	// Verify still valid
	valid = am.ValidateSession(testAddr)
	assert.True(t, valid)

	// Wait for session to expire
	time.Sleep(500 * time.Millisecond)
	valid = am.ValidateSession(testAddr)
	assert.False(t, valid)
}
