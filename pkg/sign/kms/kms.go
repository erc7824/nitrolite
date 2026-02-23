// Package kms provides shared utilities for KMS-based Ethereum signers.
//
// KMS services (GCP, AWS, etc.) return signatures in DER/ASN.1 format,
// while Ethereum expects a 65-byte R || S || V format. This package
// handles the conversion between these formats.
package kms

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// secp256k1 curve order N and half-N for S-value normalization.
var (
	secp256k1N     = crypto.S256().Params().N
	secp256k1HalfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
)

// ParseDERSignature extracts R and S values from a DER-encoded ASN.1 signature.
//
// The expected DER format is:
//
//	0x30 <total-len> 0x02 <r-len> <r-bytes> 0x02 <s-len> <s-bytes>
func ParseDERSignature(derSig []byte) (r, s *big.Int, err error) {
	if len(derSig) < 8 {
		return nil, nil, errors.New("DER signature too short")
	}
	if derSig[0] != 0x30 {
		return nil, nil, fmt.Errorf("expected DER SEQUENCE tag 0x30, got 0x%02x", derSig[0])
	}

	// Total length of the contents
	seqLen := int(derSig[1])
	if seqLen+2 != len(derSig) {
		return nil, nil, fmt.Errorf("DER sequence length mismatch: header says %d, have %d bytes", seqLen+2, len(derSig))
	}

	rest := derSig[2:]

	// Parse R
	r, rest, err = parseDERInteger(rest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse R: %w", err)
	}

	// Parse S
	s, rest, err = parseDERInteger(rest)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse S: %w", err)
	}

	if len(rest) != 0 {
		return nil, nil, fmt.Errorf("trailing %d bytes after DER signature", len(rest))
	}

	return r, s, nil
}

// parseDERInteger parses a DER-encoded INTEGER and returns the remaining bytes.
func parseDERInteger(data []byte) (*big.Int, []byte, error) {
	if len(data) < 2 {
		return nil, nil, errors.New("DER integer too short")
	}
	if data[0] != 0x02 {
		return nil, nil, fmt.Errorf("expected DER INTEGER tag 0x02, got 0x%02x", data[0])
	}

	intLen := int(data[1])
	if intLen+2 > len(data) {
		return nil, nil, fmt.Errorf("DER integer length %d exceeds available %d bytes", intLen, len(data)-2)
	}

	val := new(big.Int).SetBytes(data[2 : 2+intLen])
	return val, data[2+intLen:], nil
}

// NormalizeS ensures S is in the lower half of the curve order (canonical form).
// Ethereum requires S <= N/2 to prevent signature malleability.
func NormalizeS(s *big.Int) *big.Int {
	if s.Cmp(secp256k1HalfN) > 0 {
		return new(big.Int).Sub(secp256k1N, s)
	}
	return s
}

// RecoverV determines the recovery ID (V value) for an Ethereum signature
// by trying both possible values (0 and 1) and checking which one recovers
// to the expected public key.
func RecoverV(hash []byte, r, s *big.Int, expectedPub *ecdsa.PublicKey) (uint8, error) {
	sig := make([]byte, 65)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	for v := uint8(0); v < 2; v++ {
		sig[64] = v
		recovered, err := crypto.SigToPub(hash, sig)
		if err != nil {
			continue
		}
		if recovered.X.Cmp(expectedPub.X) == 0 && recovered.Y.Cmp(expectedPub.Y) == 0 {
			return v, nil
		}
	}

	return 0, errors.New("failed to recover V: neither V=0 nor V=1 matched the expected public key")
}

// DERToEthereumSignature converts a DER-encoded signature from a KMS service
// into a 65-byte Ethereum signature (R[32] || S[32] || V[1]).
//
// It parses the DER encoding, normalizes S, recovers V, and adjusts V to 27/28.
func DERToEthereumSignature(hash, derSig []byte, expectedPub *ecdsa.PublicKey) ([]byte, error) {
	r, s, err := ParseDERSignature(derSig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER signature: %w", err)
	}

	s = NormalizeS(s)

	v, err := RecoverV(hash, r, s, expectedPub)
	if err != nil {
		return nil, fmt.Errorf("failed to recover V: %w", err)
	}

	ethSig := make([]byte, 65)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(ethSig[32-len(rBytes):32], rBytes)
	copy(ethSig[64-len(sBytes):64], sBytes)
	ethSig[64] = v + 27

	return ethSig, nil
}

// ValidateSecp256k1PublicKey checks that the given public key is on the secp256k1 curve.
func ValidateSecp256k1PublicKey(pub *ecdsa.PublicKey) error {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return errors.New("public key is nil")
	}
	if !secp256k1.S256().IsOnCurve(pub.X, pub.Y) {
		return errors.New("public key is not on the secp256k1 curve")
	}
	return nil
}
