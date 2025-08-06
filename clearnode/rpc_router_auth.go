package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type AuthRequestParams struct {
	Address            string      `json:"address"`     // The wallet address requesting authentication
	SessionKey         string      `json:"session_key"` // The session key for the authentication
	AppName            string      `json:"app_name"`    // The name of the application requesting authentication
	Allowances         []Allowance `json:"allowances"`  // Allowances for the application
	Expire             string      `json:"expire"`      // Expiration time for the authentication
	Scope              string      `json:"scope"`       // Scope of the authentication
	ApplicationAddress string      `json:"application"` // The address of the application requesting authentication
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
		"appName", authParams.AppName,
		"rawAllowances", authParams.Allowances,
		"scope", authParams.Scope,
		"expire", authParams.Expire,
		"applicationAddress", authParams.ApplicationAddress)

	// Generate a challenge for this address
	token, err := r.AuthManager.GenerateChallenge(
		authParams.Address,
		authParams.SessionKey,
		authParams.AppName,
		authParams.Allowances,
		authParams.Scope,
		authParams.Expire,
		authParams.ApplicationAddress,
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
		"session_key": claims.Policy.Participant,
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
		challenge.AppName,
		challenge.Allowances,
		challenge.Scope,
		challenge.ApplicationAddress,
		challenge.Expire,
		sig)
	if err != nil {
		logger.Error("failed to recover address from signature", "error", err)
		return nil, nil, RPCErrorf("invalid signature")
	}

	if err := r.AuthManager.ValidateChallenge(authParams.Challenge, recoveredAddress); err != nil {
		logger.Debug("challenge verification failed", "error", err)
		return nil, nil, RPCErrorf("invalid challenge or signature")
	}

	// Store signer
	if err := AddSigner(r.DB, challenge.Address, challenge.SessionKey); err != nil {
		logger.Error("failed to create signer in db", "error", err)
		return nil, nil, err
	}

	// Generate the User tag
	if _, err = GenerateOrRetrieveUserTag(r.DB, challenge.Address); err != nil {
		logger.Error("failed to store user tag in db", "error", err)
		return nil, nil, fmt.Errorf("failed to store user tag in db")
	}

	// TODO: to use expiration specified in the Policy, instead of just setting 1 hour
	claims, jwtToken, err := r.AuthManager.GenerateJWT(challenge.Address, challenge.SessionKey, challenge.Scope, challenge.AppName, challenge.Allowances)
	if err != nil {
		logger.Error("failed to generate JWT token", "error", err)
		return nil, nil, RPCErrorf("failed to generate JWT token")
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
