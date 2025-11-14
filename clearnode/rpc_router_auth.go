package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

type AuthRequestParams struct {
	Address     string      `json:"address"`     // The wallet address requesting authentication
	SessionKey  string      `json:"session_key"` // The session key for the authentication
	Application string      `json:"application"` // The name of the application requesting authentication
	Allowances  []Allowance `json:"allowances"`  // Allowances for the application
	ExpiresAt   uint64      `json:"expires_at"`  // Expiration time for the authentication
	Scope       string      `json:"scope"`       // Scope of the authentication
}

// AuthResponse represents the server's challenge response
type AuthResponse struct {
	ChallengeMessage uuid.UUID `json:"challenge_message"` // The message to sign
}

// AuthVerifyParams represents parameters for completing authentication
type AuthVerifyParams struct {
	Challenge uuid.UUID `json:"challenge"` // The challenge token
	JWT       string    `json:"jwt"`       // Optional JWT to use for logging in
}

func (r *RPCRouter) HandleAuthRequest(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	// Track auth request metrics
	r.Metrics.AuthRequests.Inc()

	// Parse the parameters
	var authParams AuthRequestParams
	if err := parseParams(req.Params, &authParams); err != nil {
		c.Fail(err, "failed to parse auth parameters")
		return
	}

	logger.Debug("incoming auth request",
		"addr", authParams.Address,
		"sessionKey", authParams.SessionKey,
		"application", authParams.Application,
		"rawAllowances", authParams.Allowances,
		"scope", authParams.Scope,
		"expires_at", authParams.ExpiresAt)

	// Generate a challenge for this address
	token, err := r.AuthManager.GenerateChallenge(
		authParams.Address,
		authParams.SessionKey,
		authParams.Application,
		authParams.Allowances,
		authParams.Scope,
		authParams.ExpiresAt,
	)
	if err != nil {
		logger.Error("failed to generate challenge", "error", err)
		c.Fail(err, "failed to generate challenge")
		return
	}

	// Create challenge response
	challengeRes := AuthResponse{
		ChallengeMessage: token,
	}

	c.Succeed("auth_challenge", challengeRes)
}

func (r *RPCRouter) HandleAuthVerify(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	var authParams AuthVerifyParams
	if err := parseParams(req.Params, &authParams); err != nil {
		c.Fail(err, "failed to parse auth parameters")
		return
	}

	var authMethod string
	var policy *Policy
	var responseData any
	var err error
	if authParams.JWT != "" {
		authMethod = "jwt"
		policy, responseData, err = r.handleAuthJWTVerify(ctx, authParams)
	} else if len(c.Message.Sig) > 0 {
		authMethod = "signature"
		policy, responseData, err = r.handleAuthSigVerify(ctx, c.Message.Sig[0], authParams)
	} else {
		c.Fail(nil, "invalid authentication method: expected JWT or signature")
		return
	}

	r.Metrics.AuthAttemptsTotal.With(prometheus.Labels{
		"auth_method": authMethod,
	}).Inc()
	if err != nil {
		r.Metrics.AuthAttempsFail.With(prometheus.Labels{
			"auth_method": authMethod,
		}).Inc()
		c.Fail(err, "authentication failed")
		return
	}

	r.Metrics.AuthAttempsSuccess.With(prometheus.Labels{
		"auth_method": authMethod,
	}).Inc()

	c.UserID = policy.Wallet
	c.Storage.Set(ConnectionStoragePolicyKey, policy)
	c.Succeed(req.Method, responseData)
	logger.Info("authentication successful",
		"authMethod", authMethod,
		"userID", c.UserID)
}

func (r *RPCRouter) AuthMiddleware(c *RPCContext) {
	ctx := c.Context
	logger := LoggerFromContext(ctx)
	req := c.Message.Req

	// Get policy from storage
	policy, ok := c.Storage.Get(ConnectionStoragePolicyKey)
	if !ok || policy == nil || c.UserID == "" {
		c.Fail(nil, "authentication required")
		return
	}

	// Cast to Policy type
	p, ok := policy.(*Policy)
	if !ok {
		logger.Error("invalid policy type in storage", "type", fmt.Sprintf("%T", policy))
		c.Fail(nil, "invalid policy type in storage")
		return
	}

	// Check if session is still valid
	if !r.AuthManager.ValidateSession(p.Wallet) {
		// TODO: verify whether we should validate it by wallet instead of participant
		logger.Debug("session expired", "signerAddress", p.Wallet)
		c.Fail(nil, "session expired, please re-authenticate")
		return
	}

	// Update session activity timestamp
	r.AuthManager.UpdateSession(p.Wallet)

	if err := ValidateTimestamp(req.Timestamp, r.Config.msgExpiryTime); err != nil {
		logger.Debug("invalid message timestamp", "error", err)
		c.Fail(nil, "invalid message timestamp")
		return
	}

	c.Next()
}

// handleAuthJWTVerify verifies the JWT token and returns the policy, response data and error.
func (r *RPCRouter) handleAuthJWTVerify(ctx context.Context, authParams AuthVerifyParams) (*Policy, any, error) {
	logger := LoggerFromContext(ctx)

	claims, err := r.AuthManager.VerifyJWT(authParams.JWT)
	if err != nil {
		logger.Error("failed to verify JWT", "error", err)
		return nil, nil, RPCErrorf("invalid JWT token")
	}

	return &claims.Policy, map[string]any{
		"address":     claims.Policy.Wallet,
		"session_key": claims.Policy.SessionKey,
		// "jwt_token":   newJwtToken, TODO: add refresh token
		"success": true,
	}, nil
}

// handleAuthJWTVerify verifies the challenge signature and returns the policy, response data and error.
func (r *RPCRouter) handleAuthSigVerify(ctx context.Context, sig Signature, authParams AuthVerifyParams) (*Policy, any, error) {
	logger := LoggerFromContext(ctx)

	challenge, err := r.AuthManager.GetChallenge(authParams.Challenge)
	if err != nil {
		logger.Error("failed to get challenge", "error", err)
		return nil, nil, RPCErrorf("invalid challenge")
	}
	recoveredAddress, err := RecoverAddressFromEip712Signature(
		challenge.Address,
		challenge.Token.String(),
		challenge.SessionKey,
		challenge.Application,
		challenge.Allowances,
		challenge.Scope,
		challenge.SessionKeyExpiresAt,
		sig)
	if err != nil {
		logger.Error("failed to recover address from signature", "error", err)
		return nil, nil, RPCErrorf("invalid signature")
	}

	if err := r.AuthManager.ValidateChallenge(authParams.Challenge, recoveredAddress); err != nil {
		logger.Debug("challenge verification failed", "error", err)
		return nil, nil, RPCErrorf("invalid challenge or signature")
	}

	// Generate the User tag
	if _, err = GenerateOrRetrieveUserTag(r.DB, challenge.Address); err != nil {
		logger.Error("failed to store user tag in db", "error", err)
		return nil, nil, fmt.Errorf("failed to store user tag in db")
	}

	// TODO: to use expiration specified in the Policy, instead of just setting 1 hour
	claims, jwtToken, err := r.AuthManager.GenerateJWT(challenge.Address, challenge.SessionKey, challenge.Scope, challenge.Application, challenge.Allowances, challenge.SessionKeyExpiresAt)
	if err != nil {
		logger.Error("failed to generate JWT token", "error", err)
		return nil, nil, RPCErrorf("failed to generate JWT token")
	}

	// Validate allowances against supported assets before storing session key
	if err := validateAllowances(&r.Config.assets, challenge.Allowances); err != nil {
		logger.Error("unsupported asset in allowances", "error", err, "allowances", challenge.Allowances)
		return nil, nil, RPCErrorf("unsupported token: %w", err)
	}

	exists, err := CheckSessionKeyExists(r.DB, challenge.Address, challenge.SessionKey)
	if err != nil {
		logger.Error("failed to check existing session key", "error", err, "sessionKey", challenge.SessionKey)
		return nil, nil, err
	}

	if !exists {
		if err := AddSessionKey(r.DB, challenge.Address, challenge.SessionKey, challenge.Application, challenge.Scope, challenge.Allowances, claims.Policy.ExpiresAt); err != nil {
			logger.Error("failed to store session key", "error", err, "sessionKey", challenge.SessionKey)
			return nil, nil, err
		}
	}

	return &claims.Policy, map[string]any{
		"address":     challenge.Address,
		"session_key": challenge.SessionKey,
		"jwt_token":   jwtToken,
		"success":     true,
	}, nil
}

func ValidateTimestamp(ts uint64, expirySeconds int) error {
	if ts < 1_000_000_000_000 || ts > 9_999_999_999_999 {
		return fmt.Errorf("invalid timestamp %d: must be 13-digit Unix ms", ts)
	}
	t := time.UnixMilli(int64(ts)).UTC()
	if time.Since(t) > time.Duration(expirySeconds)*time.Second {
		return fmt.Errorf("timestamp expired: %s older than %d s", t.Format(time.RFC3339Nano), expirySeconds)
	}
	return nil
}

// validateAllowances validates that all assets in allowances are valid abd supported by the system
func validateAllowances(assetsCfg *AssetsConfig, allowances []Allowance) error {
	if len(allowances) == 0 {
		return nil
	}

	supportedSymbols := make(map[string]bool)
	for _, asset := range assetsCfg.Assets {
		if !asset.Disabled {
			supportedSymbols[asset.Symbol] = true
		}
	}

	for _, allowance := range allowances {
		if !supportedSymbols[allowance.Asset] {
			return fmt.Errorf("asset '%s' is not supported", allowance.Asset)
		}

		amount, err := decimal.NewFromString(allowance.Amount)
		if err != nil {
			return fmt.Errorf("invalid amount '%s' for asset '%s': %w", allowance.Amount, allowance.Asset, err)
		}

		if amount.LessThan(decimal.Zero) {
			return fmt.Errorf("allowance amount cannot be negative for asset '%s', got '%s'", allowance.Asset, allowance.Amount)
		}
	}

	return nil
}
