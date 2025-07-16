package nitrolite

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Sign hashes the provided data using Keccak256 and signs it with the given private key.
func Sign(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	dataHash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(dataHash.Bytes(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	return signature, nil
}

// Verify checks if the signature on the provided data was created by the given address.
func Verify(data []byte, sig []byte, address common.Address) (bool, error) {
	dataHash := crypto.Keccak256Hash(data)

	pubKeyRaw, err := crypto.Ecrecover(dataHash.Bytes(), sig)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr == address, nil
}
