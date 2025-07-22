package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Allowance represents allowances for connection
type Allowance struct {
	Asset  string `json:"asset"`
	Amount string `json:"amount"`
}

// Signer handles signing operations using a private key
type Signer struct {
	privateKey *ecdsa.PrivateKey
}

// NewSigner creates a new signer from a hex-encoded private key
func NewSigner(privateKeyHex string) (*Signer, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	return &Signer{privateKey: privateKey}, nil
}

// Sign creates an ECDSA signature for the provided data
func (s *Signer) Sign(data []byte) ([]byte, error) {
	sig, err := nitrolite.Sign(data, s.privateKey)
	if err != nil {
		return nil, err
	}

	signature := make([]byte, 65)
	copy(signature[0:32], sig.R[:])
	copy(signature[32:64], sig.S[:])

	if sig.V >= 27 {
		signature[64] = sig.V - 27
	}
	return signature, nil
}

// NitroSign creates a signature for the provided state in nitrolite.Signature format
func (s *Signer) NitroSign(encodedState []byte) (nitrolite.Signature, error) {
	sig, err := nitrolite.Sign(encodedState, s.privateKey)
	if err != nil {
		return nitrolite.Signature{}, fmt.Errorf("failed to sign encoded state: %w", err)
	}
	return nitrolite.Signature{
		V: sig.V,
		R: sig.R,
		S: sig.S,
	}, nil
}

// GetPublicKey returns the public key associated with the signer
func (s *Signer) GetPublicKey() *ecdsa.PublicKey {
	return s.privateKey.Public().(*ecdsa.PublicKey)
}

// GetPrivateKey returns the private key used by the signer
func (s *Signer) GetPrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}

// GetAddress returns the address derived from the signer's public key
func (s *Signer) GetAddress() common.Address {
	return crypto.PubkeyToAddress(*s.GetPublicKey())
}

// RecoverAddress takes the original message and its hex-encoded signature, and returns the address
func RecoverAddress(message []byte, signatureHex string) (string, error) {
	sig, err := hexutil.Decode(signatureHex)
	if err != nil {
		return "", fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sig) != 65 {
		return "", fmt.Errorf("invalid signature length: got %d, want 65", len(sig))
	}

	if sig[64] >= 27 {
		sig[64] -= 27
	}

	msgHash := crypto.Keccak256Hash(message)

	pubkey, err := crypto.SigToPub(msgHash.Bytes(), sig)
	if err != nil {
		return "", fmt.Errorf("signature recovery failed: %w", err)
	}

	addr := crypto.PubkeyToAddress(*pubkey)
	return addr.Hex(), nil
}

func RecoverAddressFromEip712Signature(
	addrHex string,
	challengeToken string,
	sessionKey string,
	appName string,
	allowances []Allowance,
	scope string,
	application string,
	expire string,
	signatureHex string) (string, error) {
	convertedAllowances := convertAllowances(allowances)

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
			Name: appName,
		},
		Message: map[string]interface{}{
			"challenge":   challengeToken,
			"scope":       scope,
			"wallet":      addrHex,
			"application": application,
			"participant": sessionKey,
			"expire":      expire,
			"allowances":  convertedAllowances,
		},
	}

	// 1. Hash the typed data (domain separator + message struct hash)
	typedDataHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return "", err
	}

	// 2. Example signature
	sig, err := hexutil.Decode(signatureHex)
	if err != nil {
		return "", err
	}

	// 3. Fix V if needed (Ethereum uses 27/28, go-ethereum expects 0/1)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	// 4. Recover public key
	pubKey, err := crypto.SigToPub(typedDataHash, sig)
	if err != nil {
		return "", err
	}

	signerAddress := crypto.PubkeyToAddress(*pubKey)

	return signerAddress.Hex(), nil
}

func convertAllowances(input []Allowance) []map[string]interface{} {
	out := make([]map[string]interface{}, len(input))
	for i, a := range input {
		amountInt, ok := new(big.Int).SetString(a.Amount, 10)
		if !ok {
			log.Printf("Invalid amount in allowance: %s", a.Amount)
			continue
		}
		out[i] = map[string]interface{}{
			"asset":  a.Asset,
			"amount": amountInt,
		}
	}
	return out
}
