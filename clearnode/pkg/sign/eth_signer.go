package sign

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Ensure our types implement the interfaces at compile time.
var _ Signer = (*EthereumSigner)(nil)
var _ AddressRecoverer = (*EthereumAddressRecoverer)(nil)
var _ PublicKey = (*EthereumPublicKey)(nil)
var _ Address = (*EthereumAddress)(nil)

// EthereumAddress implements the Address interface for Ethereum.
type EthereumAddress struct{ common.Address }

func (a EthereumAddress) String() string { return a.Address.Hex() }

// NewEthereumAddress creates a new Ethereum address from a common.Address.
func NewEthereumAddress(addr common.Address) EthereumAddress {
	return EthereumAddress{addr}
}

// NewEthereumAddressFromHex creates a new Ethereum address from a hex string.
func NewEthereumAddressFromHex(hexAddr string) EthereumAddress {
	return EthereumAddress{common.HexToAddress(hexAddr)}
}

// Equals returns true if this address equals the other address.
func (a EthereumAddress) Equals(other Address) bool {
	if otherAddr, ok := other.(EthereumAddress); ok {
		return a.Address == otherAddr.Address
	}
	// Fallback to string comparison for cross-blockchain compatibility
	return a.String() == other.String()
}

// EthereumPublicKey implements the PublicKey interface for Ethereum.
type EthereumPublicKey struct{ *ecdsa.PublicKey }

func (p EthereumPublicKey) Address() Address {
	return EthereumAddress{ethcrypto.PubkeyToAddress(*p.PublicKey)}
}
func (p EthereumPublicKey) Bytes() []byte { return ethcrypto.FromECDSAPub(p.PublicKey) }

// NewEthereumPublicKey creates a new Ethereum public key from an ECDSA public key.
func NewEthereumPublicKey(pub *ecdsa.PublicKey) EthereumPublicKey {
	return EthereumPublicKey{pub}
}

// NewEthereumPublicKeyFromBytes creates a new Ethereum public key from raw bytes.
func NewEthereumPublicKeyFromBytes(pubBytes []byte) (EthereumPublicKey, error) {
	pub, err := ethcrypto.UnmarshalPubkey(pubBytes)
	if err != nil {
		return EthereumPublicKey{}, fmt.Errorf("failed to unmarshal public key: %w", err)
	}
	return EthereumPublicKey{pub}, nil
}

// EthereumSigner is the Ethereum implementation of the Signer interface.
type EthereumSigner struct {
	privateKey *ecdsa.PrivateKey
	publicKey  EthereumPublicKey
}

func (s *EthereumSigner) PublicKey() PublicKey { return s.publicKey }

// Sign expects the input data to be a hash (e.g., Keccak256 hash).
func (s *EthereumSigner) Sign(hash []byte) (Signature, error) {
	sig, err := ethcrypto.Sign(hash, s.privateKey)
	if err != nil {
		return nil, err
	}
	// Adjust V from 0/1 to 27/28 for Ethereum compatibility.
	if sig[64] < 27 {
		sig[64] += 27
	}
	return Signature(sig), nil
}

// NewEthereumSigner creates a new Ethereum signer from a hex-encoded private key.
func NewEthereumSigner(privateKeyHex string) (Signer, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	key, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("could not parse ethereum private key: %w", err)
	}
	return &EthereumSigner{
		privateKey: key,
		publicKey:  EthereumPublicKey{key.Public().(*ecdsa.PublicKey)},
	}, nil
}

// EthereumAddressRecoverer implements the AddressRecoverer interface for Ethereum.
type EthereumAddressRecoverer struct{}

// RecoverAddress implements the AddressRecoverer interface.
func (r *EthereumAddressRecoverer) RecoverAddress(message []byte, signature Signature) (Address, error) {
	hash := ethcrypto.Keccak256Hash(message)
	return RecoverAddressFromHash(hash.Bytes(), signature)
}

// RecoverAddressFromHash recovers an address from a signature using a pre-computed hash.
func RecoverAddressFromHash(hash []byte, sig Signature) (Address, error) {
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
	return EthereumAddress{ethcrypto.PubkeyToAddress(*pubKey)}, nil
}
