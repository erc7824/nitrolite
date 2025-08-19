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
var _ sign.PrivateKey = (*PrivateKey)(nil)
var _ sign.PublicKey = (*PublicKey)(nil)
var _ sign.Address = (*Address)(nil)

// Address implements the sign.Address interface for Ethereum.
type Address struct{ common.Address }

func (a Address) String() string { return a.Address.Hex() }

// PublicKey implements the sign.PublicKey interface for Ethereum.
type PublicKey struct{ *ecdsa.PublicKey }

func (p PublicKey) Address() sign.Address {
	return Address{ethcrypto.PubkeyToAddress(*p.PublicKey)}
}
func (p PublicKey) Bytes() []byte { return ethcrypto.FromECDSAPub(p.PublicKey) }

// PrivateKey implements the sing.PrivateKey interface for Ethereum.
type PrivateKey struct{ *ecdsa.PrivateKey }

func (p PrivateKey) PublicKey() sign.PublicKey {
	return PublicKey{p.PrivateKey.Public().(*ecdsa.PublicKey)}
}
func (p PrivateKey) Bytes() []byte { return ethcrypto.FromECDSA(p.PrivateKey) }

// Sign first hashes with Keccak256, as is standard for Ethereum.
func (p PrivateKey) Sign(data []byte) (sign.Signature, error) {
	hash := ethcrypto.Keccak256Hash(data)
	sig, err := ethcrypto.Sign(hash.Bytes(), p.PrivateKey)
	if err != nil {
		return nil, err
	}
	// Adjust V from 0/1 to 27/28 for Ethereum compatibility.
	if sig[64] < 27 {
		sig[64] += 27
	}

	return sig, nil
}

// Signer is the Ethereum implementation of the sign.Signer interface.
type Signer struct {
	privateKey PrivateKey
}

func (s *Signer) Address() sign.Address       { return s.privateKey.PublicKey().Address() }
func (s *Signer) PublicKey() sign.PublicKey   { return s.privateKey.PublicKey() }
func (s *Signer) PrivateKey() sign.PrivateKey { return s.privateKey }

// NewSignerFromHex creates a new Ethereum signer from a hex-encoded private key.
func NewEthereumSigner(privateKeyHex string) (sign.Signer, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	key, err := ethcrypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("could not parse ethereum private key: %w", err)
	}
	return &Signer{privateKey: PrivateKey{key}}, nil
}

// RecoverAddressEIP712 is an Ethereum-specific function to recover an address from an EIP-712 signature.
func RecoverAddressEIP712(typedData apitypes.TypedData, sig sign.Signature) (string, error) {
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

// RecoverAddress is an Ethereum-specific function to recover an address from a standard message signature.
func RecoverAddress(message []byte, sig sign.Signature) (string, error) {
	msgHash := ethcrypto.Keccak256Hash(message)
	addr, err := recoverAddressFromHash(msgHash.Bytes(), sig)
	if err != nil {
		return "", err
	}
	return addr.Hex(), nil
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
