package clearnet

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"github.com/erc7824/nitrolite/examples/cerebro/unisig"
)

type AuthChallenge struct {
	AppName     string
	AppAddress  string
	Token       string
	Scope       string
	Wallet      string
	Participant string
	Expire      string
	Allowances  []any
}

func (c *ClearnodeClient) Authenticate(wallet, signer unisig.Signer) error {
	if c.signer != nil {
		return nil // Already authenticated
	}

	ch := AuthChallenge{
		Wallet:      wallet.Address().Hex(),
		Participant: signer.Address().Hex(), // Using address as session key for simplicity
		AppName:     "Yellow Bridge",
		Allowances:  []any{},                // No allowances for now
		Expire:      "",                     // No expiration for now
		Scope:       "",                     // No specific scope for now
		AppAddress:  wallet.Address().Hex(), // Using address as app address for simplicity
	}
	res, err := c.request("auth_request", nil,
		ch.Wallet,
		ch.Participant,
		ch.AppName,
		ch.Allowances,
		ch.Expire,
		ch.Scope,
		ch.AppAddress,
	)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}
	if res.Res.Method != "auth_challenge" || len(res.Res.Params) < 1 {
		return fmt.Errorf("unexpected response to auth_request: %v", res.Res)
	}

	challengeMap, ok := res.Res.Params[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid auth_challenge response format: %v", res.Res.Params[0])
	}
	challengeToken, ok := challengeMap["challenge_message"].(string)
	if !ok {
		return fmt.Errorf("challenge_message not found in auth_challenge response: %v", challengeMap)
	}

	ch.Token = challengeToken
	chSig, err := signChallenge(wallet, ch)
	if err != nil {
		return fmt.Errorf("failed to sign challenge: %w", err)
	}
	authVerifyChallenge := map[string]any{
		"challenge": challengeToken,
	}
	res, err = c.request("auth_verify", []string{hexutil.Encode(chSig)}, authVerifyChallenge)
	if err != nil {
		return fmt.Errorf("authentication verification failed: %w", err)
	}
	if res.Res.Method != "auth_verify" || len(res.Res.Params) < 1 {
		return fmt.Errorf("unexpected response to auth_verify: %v", res.Res)
	}

	verifyMap, ok := res.Res.Params[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid auth_verify response format: %v", res.Res.Params[0])
	}
	if authSuccess, _ := verifyMap["success"].(bool); !authSuccess {
		return fmt.Errorf("authentication failed: %v", verifyMap)
	}

	c.signer = signer
	return nil
}

func signChallenge(s unisig.Signer, c AuthChallenge) ([]byte, error) {
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
			"challenge":   c.Token,
			"scope":       c.Scope,
			"wallet":      c.Wallet,
			"application": c.AppAddress,
			"participant": c.Participant,
			"expire":      c.Expire,
			"allowances":  c.Allowances,
		},
	}

	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	signature, err := s.Sign(hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	return signature, nil
}

func signRPCData(signer unisig.Signer, rpcData RPCData) ([]byte, error) {
	dataBytes, err := json.Marshal(rpcData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC data: %w", err)
	}

	dataHash := crypto.Keccak256(dataBytes)
	signature, err := signer.Sign(dataHash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign RPC data: %w", err)
	}

	return signature, nil
}
