package main

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthManager(t *testing.T) {
	signingKey, _ := crypto.GenerateKey()
	authManager, err := NewAuthManager(signingKey)
	require.NoError(t, err)
	require.NotNil(t, authManager)

	// Generate a challenge
	challenge, err := authManager.GenerateChallenge("addr", "session_key", "application", []Allowance{}, "", 0)
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

func TestAuthManagerJwtManagement(t *testing.T) {
	signingKey, _ := crypto.GenerateKey()
	authManager, err := NewAuthManager(signingKey)
	require.NoError(t, err)
	require.NotNil(t, authManager)

	wallet := "0x1234567890123456789012345678901234567890"
	sessionKey := "0x6966978ce78df3228993aa46984eab6d68bbe195"
	scope := "test_scope"
	application := "test_application"

	// Before JWT generation, session should not be valid
	valid := authManager.ValidateSession(wallet)
	assert.False(t, valid, "Session should not be valid before JWT verification")

	_, token, err := authManager.GenerateJWT(wallet, sessionKey, scope, application, []Allowance{
		{
			Asset:  "usdc",
			Amount: "100000",
		},
	}, uint64(time.Now().Add(1*time.Hour).Unix()))
	require.NoError(t, err)

	// After JWT generation but before verification, session should still not be valid
	valid = authManager.ValidateSession(wallet)
	assert.False(t, valid, "Session should not be valid after JWT generation but before verification")

	claims, err := authManager.VerifyJWT(token)
	require.NoError(t, err)

	// Basic JWT verification
	assert.Equal(t, wallet, claims.Policy.Wallet)
	assert.Equal(t, sessionKey, claims.Policy.SessionKey)
	assert.Equal(t, scope, claims.Policy.Scope)
	assert.Equal(t, application, claims.Policy.Application)

	// After JWT verification, session should be valid
	valid = authManager.ValidateSession(wallet)
	assert.True(t, valid, "Session should be valid after JWT verification")
}

func TestAuthManagerJwtSessionRegistration(t *testing.T) {
	signingKey, _ := crypto.GenerateKey()
	authManager, err := NewAuthManager(signingKey)
	require.NoError(t, err)
	require.NotNil(t, authManager)

	wallet := "0x1234567890123456789012345678901234567890"
	sessionKey := "0x6966978ce78df3228993aa46984eab6d68bbe195"

	// Generate JWT
	_, token, err := authManager.GenerateJWT(wallet, sessionKey, "", "", []Allowance{}, uint64(time.Now().Add(1*time.Hour).Unix()))
	require.NoError(t, err)

	// Before verification, session should not be valid
	valid := authManager.ValidateSession(wallet)
	assert.False(t, valid, "Session should not be valid before JWT verification")

	// Verify JWT
	_, err = authManager.VerifyJWT(token)
	require.NoError(t, err)

	// After verification, session should be valid
	valid = authManager.ValidateSession(wallet)
	assert.True(t, valid, "Session should be valid after JWT verification")

	// Update session should work
	updated := authManager.UpdateSession(wallet)
	assert.True(t, updated, "Should be able to update session after JWT verification")
}

func TestAuthManagerJwtExpiration(t *testing.T) {
	signingKey, _ := crypto.GenerateKey()

	// We're testing session expiration, not JWT expiration,
	// so keep the JWT valid for longer than the session
	am := &AuthManager{
		challenges:     make(map[uuid.UUID]*Challenge),
		challengeTTL:   5 * time.Minute,
		authSessions:   make(map[string]time.Time),
		sessionTTL:     250 * time.Millisecond, // Short TTL for testing
		cleanupTicker:  time.NewTicker(10 * time.Minute),
		maxChallenges:  1000,
		authSigningKey: signingKey,
	}

	wallet := "0x1234567890123456789012345678901234567890"
	sessionKey := "0x6966978ce78df3228993aa46984eab6d68bbe195"

	// Create a JWT with custom claims for longer expiration
	policy := Policy{
		Wallet:      wallet,
		SessionKey:  sessionKey,
		Scope:       "",
		Application: "",
		Allowances:  []Allowance{},
		ExpiresAt:   time.Now().Add(5 * time.Minute), // Longer expiration for JWT
	}

	claims := JWTClaims{
		Policy: policy,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)), // Longer expiration
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "clearnode",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(am.authSigningKey)
	require.NoError(t, err)

	// Verify JWT should register a session
	_, err = am.VerifyJWT(tokenString)
	require.NoError(t, err)

	// Session should be valid immediately
	valid := am.ValidateSession(wallet)
	assert.True(t, valid, "Session should be valid after JWT verification")

	// Wait for session to expire
	time.Sleep(300 * time.Millisecond)

	// Session should be invalid after expiration
	valid = am.ValidateSession(wallet)
	assert.False(t, valid, "Session should be invalid after expiration")
}

func TestUpdateExpiredSession(t *testing.T) {
	signingKey, _ := crypto.GenerateKey()
	am := &AuthManager{
		challenges:     make(map[uuid.UUID]*Challenge),
		challengeTTL:   5 * time.Minute,
		authSessions:   make(map[string]time.Time),
		sessionTTL:     250 * time.Millisecond, // Short TTL for testing
		cleanupTicker:  time.NewTicker(10 * time.Minute),
		maxChallenges:  1000,
		authSigningKey: signingKey,
	}

	wallet := "0x1234567890123456789012345678901234567890"

	// Register the session
	am.registerAuthSession(wallet)

	// Verify session is valid
	valid := am.ValidateSession(wallet)
	assert.True(t, valid, "Session should be valid immediately after registration")

	// Wait for session to expire
	time.Sleep(300 * time.Millisecond)

	// Session should be invalid after expiration
	valid = am.ValidateSession(wallet)
	assert.False(t, valid, "Session should be invalid after expiration")

	// Attempt to update the expired session
	updated := am.UpdateSession(wallet)

	// According to current implementation, UpdateSession returns false for non-existent sessions
	// but does not check expiration - it just checks if the session exists in the map
	assert.True(t, updated, "UpdateSession returns true if session exists in map, even if expired")

	// Verify if the session is now valid after update
	valid = am.ValidateSession(wallet)
	assert.True(t, valid, "Session should be valid after update")
}
