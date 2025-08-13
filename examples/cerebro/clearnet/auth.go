package clearnet

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/erc7824/nitrolite/examples/cerebro/unisig"
)

type AuthChallengeParams struct {
	Address            string `json:"address"`
	SessionKey         string `json:"session_key"`
	AppName            string `json:"app_name"`
	Allowances         []any  `json:"allowances"`
	Expire             string `json:"expire"`
	Scope              string `json:"scope"`
	ApplicationAddress string `json:"application"`
}

func (c *ClearnodeClient) Authenticate(wallet, signer unisig.Signer) (string, error) {
	if c.signer != nil {
		return "", nil // Already authenticated
	}

	ch := AuthChallengeParams{
		Address:            wallet.Address().Hex(),
		SessionKey:         signer.Address().Hex(), // Using address as session key for simplicity
		AppName:            "Cerebro CLI",
		Allowances:         []any{},                // No allowances for now
		Expire:             "",                     // No expiration for now
		Scope:              "",                     // No specific scope for now
		ApplicationAddress: wallet.Address().Hex(), // Using address as app address for simplicity
	}
	res, err := c.request("auth_request", nil, ch)
	if err != nil {
		return "", fmt.Errorf("authentication request failed: %w", err)
	}
	if res.Res.Method != "auth_challenge" {
		return "", fmt.Errorf("unexpected response to auth_request: %v", res.Res)
	}

	challengeMap, ok := res.Res.Params.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid auth_challenge response format: %v", res.Res.Params)
	}
	challengeToken, ok := challengeMap["challenge_message"].(string)
	if !ok {
		return "", fmt.Errorf("challenge_message not found in auth_challenge response: %v", challengeMap)
	}

	chSig, err := signChallenge(wallet, ch, challengeToken)
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}
	authVerifyChallenge := map[string]any{
		"challenge": challengeToken,
	}
	res, err = c.request("auth_verify", []unisig.Signature{chSig}, authVerifyChallenge)
	if err != nil {
		return "", fmt.Errorf("authentication verification failed: %w", err)
	}
	if res.Res.Method != "auth_verify" {
		return "", fmt.Errorf("unexpected response to auth_verify: %v", res.Res)
	}

	verifyMap, ok := res.Res.Params.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid auth_verify response format: %v", res.Res.Params)
	}
	if authSuccess, _ := verifyMap["success"].(bool); !authSuccess {
		return "", fmt.Errorf("authentication failed: %v", verifyMap)
	}

	res, err = c.request("get_user_tag", nil, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get user tag: %w", err)
	}
	if res.Res.Method != "get_user_tag" {
		return "", fmt.Errorf("unexpected response to get_user_tag: %v", res.Res)
	}

	tagMap, ok := res.Res.Params.(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid auth_verify response format: %v", res.Res.Params)
	}
	userTag, _ := tagMap["tag"].(string)

	c.signer = signer
	return userTag, nil
}

func signChallenge(s unisig.Signer, c AuthChallengeParams, token string) (unisig.Signature, error) {
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": {
				{Name: "name", Type: "string"},
			},
			"Policy": {
				{Name: "challenge", Type: "string"},
				{Name: "scope", Type: "string"},
				{Name: "wallet", Type: "address"},
				{Name: "application", Type: "address"},
				{Name: "participant", Type: "address"},
				{Name: "expire", Type: "uint256"},
				{Name: "allowances", Type: "Allowance[]"},
			},
			"Allowance": {
				{Name: "asset", Type: "string"},
				{Name: "amount", Type: "uint256"},
			}},
		PrimaryType: "Policy",
		Domain: apitypes.TypedDataDomain{
			Name: c.AppName,
		},
		Message: map[string]interface{}{
			"challenge":   token,
			"scope":       c.Scope,
			"wallet":      c.Address,
			"application": c.ApplicationAddress,
			"participant": c.SessionKey,
			"expire":      c.Expire,
			"allowances":  c.Allowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return unisig.Signature{}, err
	}

	signature, err := s.Sign(hash)
	if err != nil {
		return unisig.Signature{}, fmt.Errorf("failed to sign challenge: %w", err)
	}

	return signature, nil
}

func signRPCData(signer unisig.Signer, rpcData RPCData) (unisig.Signature, error) {
	dataBytes, err := json.Marshal(rpcData)
	if err != nil {
		return unisig.Signature{}, fmt.Errorf("failed to marshal RPC data: %w", err)
	}

	dataHash := crypto.Keccak256(dataBytes)
	signature, err := signer.Sign(dataHash)
	if err != nil {
		return unisig.Signature{}, fmt.Errorf("failed to sign RPC data: %w", err)
	}

	return signature, nil
}
