package main

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewSigner(privateKeyHex string) (*Signer, error) {
	privKey, address, err := decodePrivateKey(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address from private key: %w", err)
	}

	return &Signer{
		privateKey: privKey,
		address:    address,
	}, nil
}

// Address returns the Ethereum address of the signer
func (s *Signer) Address() common.Address {
	return s.address
}

// Sign creates an ECDSA signature for the provided data
func (s *Signer) Sign(data []byte) ([]byte, error) {
	dataHash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(dataHash.Bytes(), s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	return signature, nil
}

func signNitroMessage(s *Signer, message []byte) ([]byte, error) {
	signature, err := s.Sign(message)
	if err != nil {
		return nil, err
	}

	if signature[64] < 27 {
		signature[64] += 27 // Adjust signature version
	}
	return signature, nil
}

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

func signChallenge(s *Signer, c AuthChallenge) ([]byte, error) {
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

	_, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	signature, err := s.Sign([]byte(rawData))
	if err != nil {
		return nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	return signature, nil
}

func decodePrivateKey(privateKeyHex string) (*ecdsa.PrivateKey, common.Address, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("failed to parse private key: %w", err)
	}

	address := crypto.PubkeyToAddress(privKey.PublicKey)
	return privKey, address, nil
}
