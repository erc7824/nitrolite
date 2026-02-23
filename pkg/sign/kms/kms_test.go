package kms

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test private key (same one used across the sign package tests).
const testPrivateKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

func setupTestKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.HexToECDSA(testPrivateKeyHex)
	require.NoError(t, err)
	return key
}

// buildDER constructs a DER-encoded ECDSA signature from R and S values.
func buildDER(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Add leading zero if high bit set (ASN.1 integer is signed)
	if len(rBytes) > 0 && rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0x00}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// 0x02 <r-len> <r> 0x02 <s-len> <s>
	seq := []byte{0x02, byte(len(rBytes))}
	seq = append(seq, rBytes...)
	seq = append(seq, 0x02, byte(len(sBytes)))
	seq = append(seq, sBytes...)

	// 0x30 <total-len> <seq>
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

func TestParseDERSignature(t *testing.T) {
	expectedR, _ := new(big.Int).SetString("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", 16)
	expectedS, _ := new(big.Int).SetString("483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8", 16)

	der := buildDER(expectedR, expectedS)

	r, s, err := ParseDERSignature(der)
	require.NoError(t, err)
	assert.Equal(t, expectedR, r)
	assert.Equal(t, expectedS, s)
}

func TestParseDERSignature_WithLeadingZeros(t *testing.T) {
	// R with high bit set — DER adds 0x00 prefix
	rVal, _ := new(big.Int).SetString("ff"+
		"be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", 16)
	sVal := big.NewInt(42)

	der := buildDER(rVal, sVal)

	r, s, err := ParseDERSignature(der)
	require.NoError(t, err)
	assert.Equal(t, rVal, r)
	assert.Equal(t, sVal, s)
}

func TestParseDERSignature_Errors(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"too short", []byte{0x30, 0x01}},
		{"wrong tag", []byte{0x31, 0x06, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01}},
		{"length mismatch", []byte{0x30, 0x10, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01}},
		{"missing S", []byte{0x30, 0x03, 0x02, 0x01, 0x01}},
		{"trailing bytes", []byte{0x30, 0x07, 0x02, 0x01, 0x01, 0x02, 0x01, 0x01, 0xff}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseDERSignature(tc.data)
			assert.Error(t, err)
		})
	}
}

func TestNormalizeS(t *testing.T) {
	// S in lower half — should be unchanged
	smallS := big.NewInt(42)
	assert.Equal(t, smallS, NormalizeS(smallS))

	// S exactly at half N — should be unchanged
	assert.Equal(t, secp256k1HalfN, NormalizeS(new(big.Int).Set(secp256k1HalfN)))

	// S above half N — should be flipped to N - S
	highS := new(big.Int).Add(secp256k1HalfN, big.NewInt(1))
	normalized := NormalizeS(highS)
	expected := new(big.Int).Sub(secp256k1N, highS)
	assert.Equal(t, expected, normalized)

	// S = N - 1 (maximum valid S) — should flip to 1
	maxS := new(big.Int).Sub(secp256k1N, big.NewInt(1))
	assert.Equal(t, big.NewInt(1), NormalizeS(maxS))
}

func TestRecoverV(t *testing.T) {
	key := setupTestKey(t)
	hash := crypto.Keccak256([]byte("test message"))

	// Sign with go-ethereum to get a known-good signature
	sig, err := crypto.Sign(hash, key)
	require.NoError(t, err)

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	expectedV := sig[64] // 0 or 1

	v, err := RecoverV(hash, r, s, &key.PublicKey)
	require.NoError(t, err)
	assert.Equal(t, expectedV, v)
}

func TestRecoverV_WrongPubKey(t *testing.T) {
	key := setupTestKey(t)
	hash := crypto.Keccak256([]byte("test message"))

	sig, err := crypto.Sign(hash, key)
	require.NoError(t, err)

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])

	// Use a different key — V recovery should fail
	otherKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	_, err = RecoverV(hash, r, s, &otherKey.PublicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to recover V")
}

func TestDERToEthereumSignature(t *testing.T) {
	key := setupTestKey(t)
	hash := crypto.Keccak256([]byte("test message for DER conversion"))

	// Sign with go-ethereum
	rawSig, err := crypto.Sign(hash, key)
	require.NoError(t, err)

	r := new(big.Int).SetBytes(rawSig[:32])
	s := new(big.Int).SetBytes(rawSig[32:64])

	// Build a DER signature from R and S
	der := buildDER(r, s)

	// Convert back via our function
	ethSig, err := DERToEthereumSignature(hash, der, &key.PublicKey)
	require.NoError(t, err)
	require.Len(t, ethSig, 65)

	// V should be 27 or 28
	assert.True(t, ethSig[64] == 27 || ethSig[64] == 28)

	// Verify the signature recovers to the correct address
	recoverSig := make([]byte, 65)
	copy(recoverSig, ethSig)
	recoverSig[64] -= 27
	recovered, err := crypto.SigToPub(hash, recoverSig)
	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.X, recovered.X)
	assert.Equal(t, key.PublicKey.Y, recovered.Y)
}

func TestDERToEthereumSignature_HighS(t *testing.T) {
	key := setupTestKey(t)
	hash := crypto.Keccak256([]byte("test high S normalization"))

	rawSig, err := crypto.Sign(hash, key)
	require.NoError(t, err)

	r := new(big.Int).SetBytes(rawSig[:32])
	s := new(big.Int).SetBytes(rawSig[32:64])

	// Force S to high value (above N/2)
	highS := new(big.Int).Sub(secp256k1N, s)
	if highS.Cmp(secp256k1HalfN) <= 0 {
		// Original S was already high, use it directly
		highS = s
		s = new(big.Int).Sub(secp256k1N, s)
	}

	der := buildDER(r, highS)

	ethSig, err := DERToEthereumSignature(hash, der, &key.PublicKey)
	require.NoError(t, err)

	// The S in the output should be normalized (low)
	outS := new(big.Int).SetBytes(ethSig[32:64])
	assert.True(t, outS.Cmp(secp256k1HalfN) <= 0, "S should be in lower half after normalization")
}

func TestValidateSecp256k1PublicKey(t *testing.T) {
	key := setupTestKey(t)
	assert.NoError(t, ValidateSecp256k1PublicKey(&key.PublicKey))

	// Nil key
	assert.Error(t, ValidateSecp256k1PublicKey(nil))

	// Key with nil X
	assert.Error(t, ValidateSecp256k1PublicKey(&ecdsa.PublicKey{}))
}
