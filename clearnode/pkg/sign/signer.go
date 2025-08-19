package sign

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Signer handles signing operations using a private key.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    common.Address
}

// NewSigner creates a Signer from a hex-encoded private key.
func NewSigner(privateKeyHex string) (*Signer, error) {
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Derive and cache public key and address for efficiency.
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKey)

	return &Signer{
		privateKey: privateKey,
		publicKey:  publicKey,
		address:    address,
	}, nil
}

// Sign hashes data with Keccak256 and creates an ECDSA signature.
func (s *Signer) Sign(data []byte) (Signature, error) {
	if s.privateKey == nil {
		return nil, fmt.Errorf("signer has no private key")
	}

	dataHash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(dataHash.Bytes(), s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Adjust V from 0/1 to 27/28 for Ethereum compatibility.
	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

// GetPublicKey returns the cached public key.
func (s *Signer) GetPublicKey() *ecdsa.PublicKey {
	return s.publicKey
}

// GetPrivateKey returns the private key.
func (s *Signer) GetPrivateKey() *ecdsa.PrivateKey {
	return s.privateKey
}

// GetAddress returns the cached Ethereum address.
func (s *Signer) GetAddress() common.Address {
	return s.address
}

// RecoverAddress recovers an address from an EIP-191 signature.
func RecoverAddress(message []byte, sig Signature) (string, error) {
	msgHash := crypto.Keccak256Hash(message)
	addr, err := recoverAddressFromHash(msgHash.Bytes(), sig)
	if err != nil {
		return "", err
	}
	return addr.Hex(), nil
}

// RecoverAddressEip712 recovers an address from an EIP-712 signature.
func RecoverAddressEip712(typedData apitypes.TypedData, sig Signature) (string, error) {
	typedDataHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return "", fmt.Errorf("failed to generate EIP-712 hash: %w", err)
	}
	addr, err := recoverAddressFromHash(typedDataHash, sig)
	if err != nil {
		return "", err
	}
	return addr.Hex(), nil
}

// recoverAddressFromHash recovers an address from a hash and signature.
func recoverAddressFromHash(hash []byte, sig Signature) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length: 65 bytes expected, got %d", len(sig))
	}

	// Create a copy to not modify original signature.
	localSig := make(Signature, 65)
	copy(localSig, sig)

	// Normalize V from 27/28 to 0/1.
	if localSig[64] >= 27 {
		localSig[64] -= 27
	}

	pubKey, err := crypto.SigToPub(hash, localSig)
	if err != nil {
		return common.Address{}, fmt.Errorf("signature recovery failed: %w", err)
	}

	return crypto.PubkeyToAddress(*pubKey), nil
}
