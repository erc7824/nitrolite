package eth

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Ensure our types implement the interfaces at compile time.
var _ sign.Signer = (*Signer)(nil)
var _ sign.AddressRecoverer = (*AddressRecoverer)(nil)
var _ sign.PublicKey = (*PublicKey)(nil)
var _ sign.Address = (*Address)(nil)

// Address implements the sign.Address interface for Ethereum.
type Address struct{ common.Address }

func (a Address) String() string { return a.Address.Hex() }

// NewAddress creates a new Ethereum address from a common.Address.
func NewAddress(addr common.Address) Address {
	return Address{addr}
}

// NewAddressFromHex creates a new Ethereum address from a hex string.
func NewAddressFromHex(hexAddr string) Address {
	return Address{common.HexToAddress(hexAddr)}
}

// Equals returns true if this address equals the other address.
func (a Address) Equals(other sign.Address) bool {
	if otherAddr, ok := other.(Address); ok {
		return a.Address == otherAddr.Address
	}
	// Fallback to string comparison for cross-blockchain compatibility
	return a.String() == other.String()
}

// PublicKey implements the sign.PublicKey interface for Ethereum.
type PublicKey struct{ *ecdsa.PublicKey }

func (p PublicKey) Address() sign.Address {
	return Address{ethcrypto.PubkeyToAddress(*p.PublicKey)}
}
func (p PublicKey) Bytes() []byte { return ethcrypto.FromECDSAPub(p.PublicKey) }

// NewPublicKey creates a new Ethereum public key from an ECDSA public key.
func NewPublicKey(pub *ecdsa.PublicKey) PublicKey {
	return PublicKey{pub}
}

// NewPublicKeyFromBytes creates a new Ethereum public key from raw bytes.
func NewPublicKeyFromBytes(pubBytes []byte) (PublicKey, error) {
	pub, err := ethcrypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return PublicKey{}, fmt.Errorf("failed to unmarshal public key: %w", err)
	}
	return PublicKey{pub}, nil
}

// Signer is the Ethereum implementation of the sign.Signer interface.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	publicKey  PublicKey
}

func (s *Signer) PublicKey() sign.PublicKey { return s.publicKey }

// AddressRecoverer implements the sign.AddressRecoverer interface for Ethereum.
type AddressRecoverer struct{}

// RecoverAddress implements the AddressRecoverer interface.
// It expects the message to be the original unhashed message and will hash it internally.
func (r *AddressRecoverer) RecoverAddress(message []byte, signature sign.Signature) (sign.Address, error) {
	hash := ethcrypto.Keccak256Hash(message)
	return RecoverAddress(hash.Bytes(), signature)
}

// Sign expects the input data to be a hash (e.g., Keccak256 hash).
func (s *Signer) Sign(hash []byte) (sign.Signature, error) {
	sig, err := ethcrypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, err
	}
	// Adjust V from 0/1 to 27/28 for Ethereum compatibility.
	if sig[64] < 27 {
		sig[64] += 27
	}
	return sign.Signature(sig), nil
}

// NewSignerFromHex creates a new Ethereum signer from a hex-encoded private key.
func NewEthereumSigner(privateKeyHex string) (sign.Signer, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	key, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("could not parse ethereum private key: %w", err)
	}
	return &Signer{
		privateKey: key,
		publicKey:  PublicKey{key.Public().(*ecdsa.PublicKey)},
	}, nil
}

// RecoverAddress recovers an address from a signature using a pre-computed hash.
func RecoverAddress(hash []byte, sig sign.Signature) (sign.Address, error) {
	if len(sig) != 65 {
		return nil, fmt.Errorf("invalid signature length")
	}
	localSig := make([]byte, 65)
	copy(localSig, sig)
	if localSig[64] >= 27 {
		localSig[64] -= 27
	}
	pubKey, err := ethcrypto.SigToPub(hash, localSig)
	if err != nil {
		return nil, fmt.Errorf("signature recovery failed: %w", err)
	}
	return Address{ethcrypto.PubkeyToAddress(*pubKey)}, nil
}
