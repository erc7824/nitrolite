package main

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Challenge represents an authentication challenge
type Challenge struct {
	Token     uuid.UUID // Random challenge token
	Address   string    // Address this challenge was created for
	CreatedAt time.Time // When the challenge was created
	ExpiresAt time.Time // When the challenge expires
	Completed bool      // Whether the challenge has been used
}

// AuthManager handles authentication challenges
type AuthManager struct {
	challenges     map[uuid.UUID]*Challenge // Challenge token -> Challenge
	challengesMu   sync.RWMutex
	challengeTTL   time.Duration
	maxChallenges  int
	cleanupTicker  *time.Ticker
	authSessions   map[string]time.Time // Address -> last active time
	authSessionsMu sync.RWMutex
	sessionTTL     time.Duration
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	am := &AuthManager{
		challenges:    make(map[uuid.UUID]*Challenge),
		challengeTTL:  5 * time.Minute,
		maxChallenges: 1000, // Prevent DoS
		cleanupTicker: time.NewTicker(10 * time.Minute),
		authSessions:  make(map[string]time.Time),
		sessionTTL:    24 * time.Hour,
	}

	// Start background cleanup
	go am.cleanupExpiredChallenges()
	return am
}

// GenerateChallenge creates a new challenge for a specific address
func (am *AuthManager) GenerateChallenge(address string) (uuid.UUID, error) {
	// Normalize address
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	// Create challenge with expiration
	now := time.Now()
	challenge := &Challenge{
		Token:     uuid.New(),
		Address:   address,
		CreatedAt: now,
		ExpiresAt: now.Add(am.challengeTTL),
		Completed: false,
	}

	// Store challenge
	am.challengesMu.Lock()
	defer am.challengesMu.Unlock()

	// Enforce max challenge limit (basic DoS protection)
	if len(am.challenges) >= am.maxChallenges {
		return uuid.UUID{}, errors.New("too many pending challenges")
	}

	am.challenges[challenge.Token] = challenge

	return challenge.Token, nil
}

// ValidateChallenge validates a challenge response
func (am *AuthManager) ValidateChallenge(challengeToken uuid.UUID, address string) error {
	// Normalize address
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	// Get the challenge
	am.challengesMu.Lock()
	defer am.challengesMu.Unlock()

	challenge, exists := am.challenges[challengeToken]
	if !exists {
		return errors.New("challenge not found")
	}

	// Verify the challenge was created for this address
	if challenge.Address != address {
		return errors.New("challenge was not created for this address")
	}

	// Check if challenge is expired
	if time.Now().After(challenge.ExpiresAt) {
		delete(am.challenges, challengeToken)
		return errors.New("challenge expired")
	}

	// Check if challenge is already used
	if challenge.Completed {
		delete(am.challenges, challengeToken)
		return errors.New("challenge already used")
	}

	// Mark challenge as completed
	challenge.Completed = true

	// Clean up
	challenge.ExpiresAt = time.Now().Add(30 * time.Second) // Keep briefly for reference

	// Register authenticated session
	am.registerAuthSession(address)

	return nil
}

// RegisterAuthSession registers an authenticated session
func (am *AuthManager) registerAuthSession(address string) {
	am.authSessionsMu.Lock()
	defer am.authSessionsMu.Unlock()
	am.authSessions[address] = time.Now()
}

// ValidateSession checks if a session is valid
func (am *AuthManager) ValidateSession(address string) bool {
	am.authSessionsMu.RLock()
	defer am.authSessionsMu.RUnlock()

	lastActive, exists := am.authSessions[address]
	if !exists {
		return false
	}

	// Check if session has expired
	if time.Now().After(lastActive.Add(am.sessionTTL)) {
		return false
	}

	return true
}

// UpdateSession updates the last active time for a session
func (am *AuthManager) UpdateSession(address string) bool {
	am.authSessionsMu.Lock()
	defer am.authSessionsMu.Unlock()

	_, exists := am.authSessions[address]
	if !exists {
		return false
	}

	am.authSessions[address] = time.Now()
	return true
}

// CleanupExpiredChallenges periodically removes expired challenges
func (am *AuthManager) cleanupExpiredChallenges() {
	for range am.cleanupTicker.C {
		now := time.Now()

		// Cleanup challenges
		am.challengesMu.Lock()
		for token, challenge := range am.challenges {
			if now.After(challenge.ExpiresAt) {
				delete(am.challenges, token)
			}
		}
		am.challengesMu.Unlock()

		// Cleanup sessions
		am.authSessionsMu.Lock()
		for addr, lastActive := range am.authSessions {
			if now.After(lastActive.Add(am.sessionTTL)) {
				delete(am.authSessions, addr)
			}
		}
		am.authSessionsMu.Unlock()
	}
}
