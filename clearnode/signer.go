package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Signature = nitrolite.Signature

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
func (s *Signer) Sign(data []byte) (Signature, error) {
	return nitrolite.Sign(data, s.privateKey)
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
func RecoverAddress(message []byte, sig Signature) (string, error) {
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
	sig Signature) (string, error) {
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

	// 2. Fix V if needed (Ethereum uses 27/28, go-ethereum expects 0/1)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	// 3. Recover public key
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
