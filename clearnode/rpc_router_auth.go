package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

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
	if len(req.Params) < 7 {
		c.Fail("invalid parameters: expected 7 parameters")
		return
	}

	addr, ok := req.Params[0].(string)
	if !ok || addr == "" {
		c.Fail(fmt.Sprintf("invalid address: %v", req.Params[0]))
		return
	}

	sessionKey, ok := req.Params[1].(string)
	if !ok || sessionKey == "" {
		c.Fail(fmt.Sprintf("invalid session key: %v", req.Params[1]))
		return
	}

	appName, ok := req.Params[2].(string)
	if !ok || appName == "" {
		c.Fail(fmt.Sprintf("invalid application name: %v", req.Params[2]))
		return
	}

	rawAllowances := req.Params[3]
	allowances, err := parseAllowances(rawAllowances)
	if err != nil {
		c.Fail(fmt.Sprintf("invalid allowances: %s", err.Error()))
		return
	}

	expire, ok := req.Params[4].(string)
	if !ok {
		c.Fail(fmt.Sprintf("invalid expiration time: %v", req.Params[4]))
		return
	}

	scope, ok := req.Params[5].(string)
	if !ok {
		c.Fail(fmt.Sprintf("invalid scope: %v", req.Params[5]))
		return
	}

	applicationAddress, ok := req.Params[6].(string)
	if !ok {
		c.Fail(fmt.Sprintf("invalid application address: %v", req.Params[6]))
		return
	}

	logger.Debug("incoming auth request",
		"addr", addr,
		"sessionKey", sessionKey,
		"appName", appName,
		"rawAllowances", rawAllowances,
		"scope", scope,
		"expire", expire,
		"applicationAddress", applicationAddress)

	// Generate a challenge for this address
	token, err := r.AuthManager.GenerateChallenge(
		addr,
		sessionKey,
		appName,
		allowances,
		scope,
		expire,
		applicationAddress,
	)
	if err != nil {
		logger.Error("failed to generate challenge", "error", err)
		c.Fail("failed to generate challenge")
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

	if len(req.Params) < 1 {
		c.Fail("invalid parameters: expected at least 1 parameter")
		return
	}

	paramsJSON, err := json.Marshal(req.Params[0])
	if err != nil {
		c.Fail(fmt.Sprintf("invalid parameters format: %s", err.Error()))
		return
	}

	var authParams AuthVerifyParams
	if err := json.Unmarshal(paramsJSON, &authParams); err != nil {
		c.Fail(fmt.Sprintf("failed to parse auth parameters: %s", err.Error()))
		return
	}

	var authMethod string
	var policy *Policy
	var responseData any
	var rpcErrorMessage string
	if authParams.JWT != "" {
		authMethod = "jwt"
		policy, responseData, rpcErrorMessage = r.handleAuthJWTVerify(ctx, authParams)
	} else if len(c.Message.Sig) > 0 {
		authMethod = "signature"
		policy, responseData, rpcErrorMessage = r.handleAuthSigVerify(ctx, c.Message.Sig[0], authParams)
	} else {
		c.Fail("invalid authentication method: expected JWT or signature")
		return
	}

	r.Metrics.AuthAttemptsTotal.With(prometheus.Labels{
		"auth_method": authMethod,
	}).Inc()
	if rpcErrorMessage != "" {
		r.Metrics.AuthAttempsFail.With(prometheus.Labels{
			"auth_method": authMethod,
		}).Inc()
		c.Fail(rpcErrorMessage)
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
		c.Fail("authentication required")
		return
	}

	// Cast to Policy type
	p, ok := policy.(*Policy)
	if !ok {
		logger.Error("invalid policy type in storage", "type", fmt.Sprintf("%T", policy))
		c.Fail("invalid policy type in storage")
		return
	}

	// Check if session is still valid
	if !r.AuthManager.ValidateSession(p.Wallet) {
		// TODO: verify whether we should validate it by wallet instead of participant
		logger.Debug("session expired", "signerAddress", p.Wallet)
		c.Fail("session expired, please re-authenticate")
		return
	}

	// Update session activity timestamp
	r.AuthManager.UpdateSession(p.Wallet)

	if err := ValidateTimestamp(req.Timestamp, r.Config.msgExpiryTime); err != nil {
		logger.Debug("invalid message timestamp", "error", err)
		c.Fail("invalid message timestamp")
		return
	}

	c.Next()
}

// handleAuthJWTVerify verifies the JWT token and returns the policy, response data and rpc error message.
func (r *RPCRouter) handleAuthJWTVerify(ctx context.Context, authParams AuthVerifyParams) (*Policy, any, string) {
	logger := LoggerFromContext(ctx)

	claims, err := r.AuthManager.VerifyJWT(authParams.JWT)
	if err != nil {
		logger.Error("failed to verify JWT", "error", err)
		return nil, nil, "invalid JWT token"
	}

	return &claims.Policy, map[string]any{
		"address":     claims.Policy.Wallet,
		"session_key": claims.Policy.Participant,
		// "jwt_token":   newJwtToken, TODO: add refresh token
		"success": true,
	}, ""
}

// handleAuthJWTVerify verifies the challenge signature and returns the policy, response data and rpc error message.
func (r *RPCRouter) handleAuthSigVerify(ctx context.Context, sig string, authParams AuthVerifyParams) (*Policy, any, string) {
	logger := LoggerFromContext(ctx)

	challenge, err := r.AuthManager.GetChallenge(authParams.Challenge)
	if err != nil {
		logger.Error("failed to get challenge", "error", err)
		return nil, nil, "invalid challenge"
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
		return nil, nil, "invalid signature"
	}

	if err := r.AuthManager.ValidateChallenge(authParams.Challenge, recoveredAddress); err != nil {
		logger.Debug("challenge verification failed", "error", err)
		return nil, nil, "invalid challenge or signature"
	}

	// Store signer
	if err := AddSigner(r.DB, challenge.Address, challenge.SessionKey); err != nil {
		logger.Error("failed to create signer in db", "error", err)
		return nil, nil, "failed to create signer in db"
	}

	claims, jwtToken, err := r.AuthManager.GenerateJWT(challenge.Address, challenge.SessionKey, "", "", challenge.Allowances)
	if err != nil {
		logger.Error("failed to generate JWT token", "error", err)
		return nil, nil, "failed to generate JWT token"
	}

	return &claims.Policy, map[string]any{
		"address":     challenge.Address,
		"session_key": challenge.SessionKey,
		"jwt_token":   jwtToken,
		"success":     true,
	}, ""
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

func parseAllowances(rawAllowances any) ([]Allowance, error) {
	outerSlice, ok := rawAllowances.([]interface{})
	if !ok {
		return nil, fmt.Errorf("input is not a list of allowances")
	}

	result := make([]Allowance, len(outerSlice))

	for i, item := range outerSlice {
		innerSlice, ok := item.([]interface{})
		if !ok {
			return nil, fmt.Errorf("allowance at index %d is not a list", i)
		}
		if len(innerSlice) != 2 {
			return nil, fmt.Errorf("allowance at index %d must have exactly 2 elements (asset, amount)", i)
		}

		asset, ok1 := innerSlice[0].(string)
		amount, ok2 := innerSlice[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("allowance at index %d has non-string asset or amount", i)
		}

		result[i] = Allowance{
			Asset:  asset,
			Amount: amount,
		}
	}

	return result, nil
}
