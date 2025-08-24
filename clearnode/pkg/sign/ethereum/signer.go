package ethereum

import (
	"crypto/ecdsa"
	"fmt"
	"strings"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Ensure our types implement the interfaces at compile time.
var _ sign.Signer = (*Signer)(nil)
var _ sign.AddressRecoverer = (*Signer)(nil)
var _ sign.PublicKey = (*PublicKey)(nil)
var _ sign.Address = (*Address)(nil)

// Address implements the sign.Address interface for Ethereum.
type Address struct{ common.Address }

func (a Address) String() string { return a.Address.Hex() }

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

// Signer is the Ethereum implementation of the sign.Signer interface.
type Signer struct {
	privateKey *ecdsa.PrivateKey
	publicKey  PublicKey
}

func (s *Signer) PublicKey() sign.PublicKey { return s.publicKey }

// Sign first hashes with Keccak256, as is standard for Ethereum.
func (s *Signer) Sign(data []byte) (sign.Signature, error) {
	hash := ethcrypto.Keccak256Hash(data)
	sig, err := ethcrypto.Sign(hash.Bytes(), s.privateKey)
	if err != nil {
		return nil, err
	}
	// Adjust V from 0/1 to 27/28 for Ethereum compatibility.
	if sig[64] < 27 {
		sig[64] += 27
	}
	return sig, nil
}

// RecoverAddress implements the AddressRecoverer interface.
func (s *Signer) RecoverAddress(message []byte, signature sign.Signature) (sign.Address, error) {
	return RecoverAddress(message, signature)
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

// RecoverAddressEIP712 is an Ethereum-specific function to recover an address from an EIP-712 signature.
func RecoverAddressEIP712(typedData apitypes.TypedData, sig sign.Signature) (sign.Address, error) {
	typedDataHash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate EIP-712 hash: %w", err)
	}
	addr, err := recoverAddressFromHash(typedDataHash, sig)
	if err != nil {
		return nil, err
	}
	return Address{addr}, nil
}

// RecoverAddress is an Ethereum-specific function to recover an address from a standard message signature.
func RecoverAddress(message []byte, sig sign.Signature) (sign.Address, error) {
	msgHash := ethcrypto.Keccak256Hash(message)
	addr, err := recoverAddressFromHash(msgHash.Bytes(), sig)
	if err != nil {
		return nil, err
	}
	return Address{addr}, nil
}

// recoverAddressFromHash is an internal helper for signature recovery.
func recoverAddressFromHash(hash []byte, sig sign.Signature) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("invalid signature length")
	}
	localSig := make(sign.Signature, 65)
	copy(localSig, sig)
	if localSig[64] >= 27 {
		localSig[64] -= 27
	}
	pubKey, err := ethcrypto.SigToPub(hash, localSig)
	if err != nil {
		return common.Address{}, fmt.Errorf("signature recovery failed: %w", err)
	}
	return ethcrypto.PubkeyToAddress(*pubKey), nil
}
