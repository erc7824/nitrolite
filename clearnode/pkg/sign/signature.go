package sign

import (
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Signature represents a 65-byte ECDSA signature [R || S || V].
type Signature []byte

// MarshalJSON implements the json.Marshaler interface.
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

// String implements the fmt.Stringer interface, returning the hex string representation of the signature.
func (s Signature) String() string {
	return hexutil.Encode(s)
}

// SignaturesToStrings converts a slice of Signatures to a slice of strings.
func SignaturesToStrings(signatures []Signature) []string {
	strs := make([]string, len(signatures))
	for i, sig := range signatures {
		strs[i] = sig.String()
	}
	return strs
}

// SignaturesFromStrings converts a slice of hex strings to a slice of Signatures.
func SignaturesFromStrings(strs []string) ([]Signature, error) {
	signatures := make([]Signature, len(strs))
	for i, str := range strs {
		sig, err := hexutil.Decode(str)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature %d ('%s'): %w", i, str, err)
		}
		signatures[i] = sig
	}
	return signatures, nil
}
