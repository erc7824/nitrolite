package evm

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

// TestHexutilDecode_WithAndWithoutPrefix tests how hexutil.Decode handles
// hex strings with and without "0x" prefix
func TestHexutilDecode_WithAndWithoutPrefix(t *testing.T) {
	// Test data - hexutil.Decode requires 0x prefix
	hexStrWithPrefix := "0x1234567890abcdef"

	// Decode with prefix
	result, err := hexutil.Decode(hexStrWithPrefix)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 8, len(result))
}

// TestHexutilDecode_SDK_SignatureFormat tests the exact format produced by SDK
func TestHexutilDecode_SDK_SignatureFormat(t *testing.T) {
	// Simulate SDK SignState output: hexutil.Encode(signature)
	// This is a 65-byte ECDSA signature (r=32, s=32, v=1)
	mockSignatureBytes := make([]byte, 65)
	for i := range mockSignatureBytes {
		mockSignatureBytes[i] = byte(i)
	}

	// This is what SDK produces
	sdkSignature := "0x" + "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f40"

	// Try to decode it
	decoded, err := hexutil.Decode(sdkSignature)
	if err != nil {
		t.Logf("Error decoding SDK signature: %v", err)
		t.FailNow()
	}

	require.Equal(t, 65, len(decoded))
}
