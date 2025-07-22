package nitrolite

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signature []byte

func (s Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(hexutil.Encode(s))
}

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

func (s Signature) String() string {
	return hexutil.Encode(s)
}

func SignaturesToStrings(signatures []Signature) []string {
	strs := make([]string, len(signatures))
	for i, sig := range signatures {
		strs[i] = sig.String()
	}
	return strs
}

func SignaturesFromStrings(strs []string) ([]Signature, error) {
	signatures := make([]Signature, len(strs))
	for i, str := range strs {
		sig, err := hexutil.Decode(str)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signature %d (%s): %w", i, str, err)
		}
		signatures[i] = sig
	}
	return signatures, nil
}

// Sign hashes the provided data using Keccak256 and signs it with the given private key.
func Sign(data []byte, privateKey *ecdsa.PrivateKey) (Signature, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}

	dataHash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(dataHash.Bytes(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: got %d, want 65", len(signature))
	}

	// This step is necessary to remain compatible with the ecrecover precompile
	if signature[64] < 27 {
		signature[64] += 27
	}

	return signature, nil
}

// Verify checks if the signature on the provided data was created by the given address.
func Verify(data []byte, sig Signature, address common.Address) (bool, error) {
	dataHash := crypto.Keccak256Hash(data)

	// Create a copy of the signature to avoid modifying the original
	sigToVerify := make(Signature, len(sig))
	copy(sigToVerify, sig)

	// Ensure the signature is in the correct format
	if sigToVerify[64] >= 27 {
		sigToVerify[64] -= 27
	}

	pubKeyRaw, err := crypto.Ecrecover(dataHash.Bytes(), sigToVerify)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	pubKey, err := crypto.UnmarshalPubkey(pubKeyRaw)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddr == address, nil
}
