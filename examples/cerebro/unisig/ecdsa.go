package unisig

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EcdsaSigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewEcdsaSigner(privateKeyHex string) (*EcdsaSigner, error) {
	privKey, address, err := DecodeEcdsaPrivateKey(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address from private key: %w", err)
	}

	return &EcdsaSigner{
		privateKey: privKey,
		address:    address,
	}, nil
}

// Address returns the Ethereum address of the signer
func (s *EcdsaSigner) Address() common.Address {
	return s.address
}

// Sign creates an ECDSA signature for the provided message
func (s *EcdsaSigner) Sign(msg []byte) ([]byte, error) {
	signature, err := crypto.Sign(msg, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	return signature, nil
}

func DecodeEcdsaPrivateKey(privateKeyHex string) (*ecdsa.PrivateKey, common.Address, error) {
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
