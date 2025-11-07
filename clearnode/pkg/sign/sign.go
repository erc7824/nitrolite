package sign

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Signer is an interface for a blockchain-agnostic signer.
type Signer interface {
	PublicKey() PublicKey                // Public key associated with this signer.
	Sign(data []byte) (Signature, error) // Sign generates a signature for the given data.
	// TODO: add Address() Address               // Address derived from the signer's public key.
}

// AddressRecoverer is an interface for recovering addresses from signatures.
type AddressRecoverer interface {
	RecoverAddress(message []byte, signature Signature) (Address, error)
}

// PublicKey is an interface for a blockchain-agnostic public key.
type PublicKey interface {
	Address() Address
	Bytes() []byte
}

// Address is an interface for a blockchain-specific address.
type Address interface {
	fmt.Stringer // All addresses must have a string representation.

	// Equals returns true if this address equals the other address.
	Equals(other Address) bool
}

// Signature is a generic byte slice representing a cryptographic signature.
type Signature []byte

// Type represents the signature type/platform used for signatures.
type Type uint8

const (
	TypeEthereum Type = iota
	TypeUnknown       = 255
)

// String returns the string representation of the algorithm.
func (t Type) String() string {
	switch t {
	case TypeEthereum:
		return "Ethereum"
	default:
		return "Unknown"
	}
}

// Type returns the signature type for this signature based on its length and structure.
func (s Signature) Type() Type {
	if len(s) == 65 {
		// Standard Ethereum signature format (r: 32 bytes, s: 32 bytes, v: 1 byte)
		return TypeEthereum
	}
	return TypeUnknown
}

// MarshalJSON implements the json.Marshaler interface, encoding the signature as a hex string.
func (s Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (s *Signature) UnmarshalJSON(data []byte) error {
	var hexStr string
	if err := json.Unmarshal(data, &hexStr); err != nil {
		return err
	}
	decoded, err := hexutil.Decode(hexStr)
	if err != nil {
		return err
	}
	*s = decoded
	return nil
}

// String implements the fmt.Stringer interface
func (s Signature) String() string {
	return hexutil.Encode(s)
}

// NewAddressRecoverer creates an appropriate AddressRecoverer based on the signature algorithm.
func NewAddressRecoverer(sigType Type) (AddressRecoverer, error) {
	switch sigType {
	case TypeEthereum:
		return &EthereumAddressRecoverer{}, nil
	default:
		return nil, fmt.Errorf("unsupported signature type: %s", sigType.String())
	}
}

// NewAddressRecovererFromSignature creates an AddressRecoverer based on signature algorithm detection.
func NewAddressRecovererFromSignature(signature Signature) (AddressRecoverer, error) {
	return NewAddressRecoverer(signature.Type())
}
