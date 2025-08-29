package sign

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		tests := []struct {
			sigType  Type
			expected string
		}{
			{TypeEthereum, "Ethereum"},
			{TypeUnknown, "Unknown"},
			{Type(99), "Unknown"},
		}

		for _, test := range tests {
			assert.Equal(t, test.expected, test.sigType.String())
		}
	})
}

func TestSignature(t *testing.T) {
	t.Run("Type detection", func(t *testing.T) {
		tests := []struct {
			name     string
			sig      Signature
			expected Type
		}{
			{
				name:     "Ethereum signature (65 bytes)",
				sig:      make(Signature, 65),
				expected: TypeEthereum,
			},
			{
				name:     "Short signature",
				sig:      make(Signature, 32),
				expected: TypeUnknown,
			},
			{
				name:     "Long signature",
				sig:      make(Signature, 128),
				expected: TypeUnknown,
			},
			{
				name:     "Empty signature",
				sig:      Signature{},
				expected: TypeUnknown,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				assert.Equal(t, test.expected, test.sig.Type())
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

func TestAddressRecovererFactory(t *testing.T) {
	t.Run("NewAddressRecoverer with supported algorithm", func(t *testing.T) {
		recoverer, err := NewAddressRecoverer(TypeEthereum)
		require.NoError(t, err)
		assert.NotNil(t, recoverer)

		_, ok := recoverer.(*EthereumAddressRecoverer)
		assert.True(t, ok)
	})

	t.Run("NewAddressRecoverer with unsupported algorithm", func(t *testing.T) {
		recoverer, err := NewAddressRecoverer(Type(99))
		assert.Error(t, err)
		assert.Nil(t, recoverer)
		assert.Contains(t, err.Error(), "unsupported signature type: Unknown")
	})

	t.Run("NewAddressRecovererFromSignature", func(t *testing.T) {
		sig := make(Signature, 65)
		recoverer, err := NewAddressRecovererFromSignature(sig)
		require.NoError(t, err)
		assert.NotNil(t, recoverer)

		shortSig := make(Signature, 32)
		recoverer, err = NewAddressRecovererFromSignature(shortSig)
		assert.Error(t, err)
		assert.Nil(t, recoverer)
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
		result := sig.Type()
		assert.Equal(t, uint8(255), uint8(result))
		assert.Equal(t, "Unknown", result.String())
		assert.Equal(t, "0x", sig.String())
	})
}
