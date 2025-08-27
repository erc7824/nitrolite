package sign

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlgorithm(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		tests := []struct {
			alg      Algorithm
			expected string
		}{
			{AlgorithmKeccak256, "Keccak256"},
			{AlgorithmECDSA, "ECDSA"},
			{AlgorithmUnknown, "Unknown"},
			{Algorithm(99), "Unknown"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, test.alg.String())
		}
	})
}

func TestSignature(t *testing.T) {
	t.Run("Algorithm detection", func(t *testing.T) {
		tests := []struct {
			name     string
			sig      Signature
			expected Algorithm
		}{
			{
				name:     "Ethereum signature (65 bytes)",
				sig:      make(Signature, 65),
				expected: AlgorithmKeccak256,
			},
			{
				name:     "Short signature",
				sig:      make(Signature, 32),
				expected: AlgorithmUnknown,
			},
			{
				name:     "Long signature",
				sig:      make(Signature, 128),
				expected: AlgorithmUnknown,
			},
			{
				name:     "Empty signature",
				sig:      Signature{},
				expected: AlgorithmUnknown,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, test.expected, test.sig.Alg())
			})
		}
	})

	t.Run("JSON marshaling", func(t *testing.T) {
		sig := Signature{0x01, 0x02, 0x03}
		
		// Marshal to JSON
		jsonData, err := json.Marshal(sig)
		require.NoError(t, err)
		
		// Should be hex encoded
		expected := `"0x010203"`
		assert.Equal(t, expected, string(jsonData))
		
		// Unmarshal back
		var unmarshaled Signature
		err = json.Unmarshal(jsonData, &unmarshaled)
		require.NoError(t, err)
		
		assert.Equal(t, sig, unmarshaled)
	})

	t.Run("JSON unmarshaling errors", func(t *testing.T) {
		tests := []struct {
			name     string
			jsonData string
		}{
			{"Invalid JSON", `{invalid}`},
			{"Invalid hex", `"0xinvalidhex"`},
			{"Non-string", `123`},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				var sig Signature
				err := json.Unmarshal([]byte(test.jsonData), &sig)
				assert.Error(t, err)
			})
		}
	})

	t.Run("String representation", func(t *testing.T) {
		sig := Signature{0x01, 0x23, 0x45}
		expected := "0x012345"
		assert.Equal(t, expected, sig.String())
	})
}

func TestAddressRecoverer(t *testing.T) {
	t.Run("NewAddressRecoverer with supported algorithm", func(t *testing.T) {
		recoverer, err := NewAddressRecoverer(AlgorithmKeccak256)
		require.NoError(t, err)
		assert.NotNil(t, recoverer)
		
		// Should be a Keccak256Recoverer
		_, ok := recoverer.(*Keccak256Recoverer)
		assert.True(t, ok)
	})

	t.Run("NewAddressRecoverer with unsupported algorithm", func(t *testing.T) {
		recoverer, err := NewAddressRecoverer(AlgorithmECDSA)
		assert.Error(t, err)
		assert.Nil(t, recoverer)
		assert.Contains(t, err.Error(), "unsupported algorithm: ECDSA")
	})

	t.Run("NewAddressRecovererFromSignature", func(t *testing.T) {
		// Ethereum-sized signature
		sig := make(Signature, 65)
		recoverer, err := NewAddressRecovererFromSignature(sig)
		require.NoError(t, err)
		assert.NotNil(t, recoverer)
		
		// Unknown algorithm signature
		shortSig := make(Signature, 32)
		recoverer, err = NewAddressRecovererFromSignature(shortSig)
		assert.Error(t, err)
		assert.Nil(t, recoverer)
	})
}

func TestKeccak256Recoverer(t *testing.T) {
	t.Run("RecoverAddress returns placeholder error", func(t *testing.T) {
		recoverer := &Keccak256Recoverer{}
		message := []byte("test message")
		signature := make(Signature, 65)
		
		addr, err := recoverer.RecoverAddress(message, signature)
		assert.Error(t, err)
		assert.Nil(t, addr)
		assert.Contains(t, err.Error(), "Keccak256 recovery requires blockchain-specific implementation")
	})
}

func TestSignatureEdgeCases(t *testing.T) {
	t.Run("Empty signature JSON marshaling", func(t *testing.T) {
		sig := Signature{}
		jsonData, err := json.Marshal(sig)
		require.NoError(t, err)
		assert.Equal(t, `"0x"`, string(jsonData))
	})

	t.Run("Nil signature handling", func(t *testing.T) {
		var sig Signature
		// Empty signature should return Unknown algorithm (255)
		result := sig.Alg()
		assert.Equal(t, uint8(255), uint8(result))
		assert.Equal(t, "Unknown", result.String())
		assert.Equal(t, "0x", sig.String())
	})
}