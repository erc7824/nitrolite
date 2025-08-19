package sign

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
)

// Signer defines the interface for a cryptographic signer.
type CryptoSigner interface {
	// Sign creates a signature for the given data.
	Sign(data []byte) (Signature, error)

	// Address returns the signer's Ethereum address.
	Address() common.Address

	// PublicKey returns the signer's public key.
	PublicKey() *ecdsa.PublicKey
}
