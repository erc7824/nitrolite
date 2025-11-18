package main

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Challenge represents an authentication challenge
type Challenge struct {
	Token               uuid.UUID   // Random challenge token
	Address             string      // Address this challenge was created for
	SessionKey          string      // SessionKey that is going to be used for this session
	Application         string      // Name of the application which opened the connection
	Allowances          []Allowance // Allowances for this connection
	Scope               string      // Policy scope
	SessionKeyExpiresAt uint64      // Session key expiration Unix timestamp (in seconds)
	CreatedAt           time.Time   // When the challenge was created
	ChallengeExpiresAt  time.Time   // When the challenge expires
	Completed           bool        // Whether the challenge has been used
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
	authSigningKey *ecdsa.PrivateKey // Private key used to sign the jwts
}

type JWTClaims struct {
	Policy Policy `json:"policy"` // Application policy details
	jwt.RegisteredClaims
}

type Policy struct {
	Wallet      string      `json:"wallet"`      // Main wallet address authorizing the session
	SessionKey  string      `json:"session_key"` // Delegated session key address
	Scope       string      `json:"scope"`       // Permission scope (e.g., "app.create", "ledger.readonly")
	Application string      `json:"application"` // Application public address
	Allowances  []Allowance `json:"allowance"`   // Array of asset allowances
	ExpiresAt   time.Time   `json:"expiration"`  // Expiration Unix timestamp (in seconds)
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(signingKey *ecdsa.PrivateKey) (*AuthManager, error) {
	am := &AuthManager{
		challenges:     make(map[uuid.UUID]*Challenge),
		challengeTTL:   5 * time.Minute,
		maxChallenges:  1000, // Prevent DoS
		cleanupTicker:  time.NewTicker(10 * time.Minute),
		authSessions:   make(map[string]time.Time),
		sessionTTL:     24 * time.Hour,
		authSigningKey: signingKey,
	}

	// Start background cleanup
	go am.cleanupExpiredChallenges()
	return am, nil
}

// GenerateChallenge creates a new challenge for a specific address
func (am *AuthManager) GenerateChallenge(
	address string,
	sessionKey string,
	application string,
	allowances []Allowance,
	scope string,
	expiresAt uint64,
) (uuid.UUID, error) {
	// Normalize address
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	// Create challenge with expiration
	now := time.Now()
	challenge := &Challenge{
		Token:               uuid.New(),
		Address:             address,
		SessionKey:          sessionKey,
		Application:         application,
		Allowances:          allowances,
		Scope:               scope,
		SessionKeyExpiresAt: expiresAt,
		CreatedAt:           now,
		ChallengeExpiresAt:  now.Add(am.challengeTTL),
		Completed:           false,
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

func (am *AuthManager) GetChallenge(challengeToken uuid.UUID) (*Challenge, error) {
	// Get the challenge
	am.challengesMu.Lock()
	defer am.challengesMu.Unlock()

	challenge, exists := am.challenges[challengeToken]
	if !exists {
		return nil, errors.New("challenge not found")
	}

	return challenge, nil
}

// ValidateChallenge validates a challenge response
func (am *AuthManager) ValidateChallenge(challengeToken uuid.UUID, recoveredSigner string) error {
	// Normalize address
	if !strings.HasPrefix(recoveredSigner, "0x") {
		recoveredSigner = "0x" + recoveredSigner
	}

	// Get the challenge
	am.challengesMu.Lock()
	defer am.challengesMu.Unlock()

	challenge, exists := am.challenges[challengeToken]
	if !exists {
		return errors.New("challenge not found")
	}

	// Verify the challenge was created for this address
	if challenge.Address != recoveredSigner {
		return fmt.Errorf("challenge address mismatch, expected %s, got %s", challenge.Address, recoveredSigner)
	}

	// Check if challenge is expired
	if time.Now().After(challenge.ChallengeExpiresAt) {
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
	challenge.ChallengeExpiresAt = time.Now().Add(30 * time.Second) // Keep briefly for reference

	// Register authenticated session
	am.registerAuthSession(recoveredSigner)

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

func (am *AuthManager) GenerateJWT(address, sessionKey, scope, application string, allowances []Allowance, sessionKeyExpiresAt uint64) (*JWTClaims, string, error) {
	policy := Policy{
		Wallet:      address,
		SessionKey:  sessionKey,
		Scope:       scope,
		Application: application,
		Allowances:  allowances,
		ExpiresAt:   time.Unix(int64(sessionKeyExpiresAt), 0),
	}
	claims := JWTClaims{
		Policy: policy,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.sessionTTL)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "clearnode", // TODO: make configurable
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(am.authSigningKey)
	if err != nil {
		return nil, "", err
	}

	return &claims, tokenString, nil
}

func (am *AuthManager) VerifyJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, nil
		}

		return &am.authSigningKey.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid JWT token claims")
	}

	if err := am.validateClaims(claims); err != nil {
		return nil, err
	}

	// Register authenticated session
	am.registerAuthSession(claims.Policy.Wallet)

	return claims, nil
}

func (am *AuthManager) validateClaims(claims *JWTClaims) error {
	issuer, err := claims.GetIssuer()
	if err != nil {
		return errors.New("failed to get issuer from JWT token claims")
	}
	expiration, err := claims.GetExpirationTime()
	if err != nil {
		return errors.New("failed to get expiration from JWT token claims")
	}

	if issuer != "clearnode" {
		return errors.New("invalid JWT token claims")
	}
	if expiration.Before(time.Now()) {
		return errors.New("expired JWT token")
	}

	return nil
}

// CleanupExpiredChallenges periodically removes expired challenges
func (am *AuthManager) cleanupExpiredChallenges() {
	for range am.cleanupTicker.C {
		now := time.Now()

		// Cleanup challenges
		am.challengesMu.Lock()
		for token, challenge := range am.challenges {
			if now.After(challenge.ChallengeExpiresAt) {
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
