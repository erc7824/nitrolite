package sign

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

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

// Address is an interface for a blockchain-specific address.
type Address interface {
	fmt.Stringer // All addresses must have a string representation.
}

// PublicKey is an interface for a blockchain-agnostic public key.
type PublicKey interface {
	Address() Address
	Bytes() []byte
}

// PrivateKey is an interface for a blockchain-agnostic private key.
type PrivateKey interface {
	PublicKey() PublicKey
	// Sign generates a signature for the given data.
	// Note: Any hashing must be done by the specific implementation.
	Sign(data []byte) (Signature, error)
	Bytes() []byte
}

// Signer is an interface for a blockchain-agnostic signer.
// It provides methods to access the address, public key, and private key.
type Signer interface {
	Address() Address
	PublicKey() PublicKey
	PrivateKey() PrivateKey
}
