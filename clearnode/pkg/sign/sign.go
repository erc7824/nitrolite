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
}

// AddressRecoverer is an optional interface that signers can implement
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
