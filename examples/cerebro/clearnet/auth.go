package clearnet

import (
	"context"
	"fmt"
	"math"

	"github.com/erc7824/nitrolite/clearnode/pkg/rpc"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
)

const (
	defaultExpirationPeriod = math.MaxUint64 // No expiration by default
)

type AuthChallengeParams struct {
	Address     string `json:"address"`
	SessionKey  string `json:"session_key"`
	Application string `json:"application"`
	Allowances  []any  `json:"allowances"`
	ExpiresAt   uint64 `json:"expires_at"`
	Scope       string `json:"scope"`
}

func (c *ClearnodeClient) Authenticate(wallet, signer sign.Signer) (rpc.AuthSigVerifyResponse, error) {
	if c.sessionKey != nil {
		return rpc.AuthSigVerifyResponse{}, nil // Already authenticated
	}

	params := rpc.AuthRequestRequest{
		Address:     wallet.PublicKey().Address().String(),
		SessionKey:  signer.PublicKey().Address().String(), // Using address as session key for simplicity
		Application: "clearnode",                           // Indicates that we create a session key with root permissions
		Allowances:  []rpc.Allowance{},                     // No allowances for now
		ExpiresAt:   defaultExpirationPeriod,
	}
	res, _, err := c.rpcClient.AuthWithSig(context.Background(), params, wallet)
	if err != nil {
		return rpc.AuthSigVerifyResponse{}, fmt.Errorf("authentication failed: %w", err)
	}

	c.sessionKey = signer
	return res, nil
}
