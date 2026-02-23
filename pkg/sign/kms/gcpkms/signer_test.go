package gcpkms

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseECPublicKeyPEM_Valid(t *testing.T) {
	// Use P-256 for the test since x509 natively supports it.
	// Real GCP KMS returns secp256k1, which x509 also handles.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	pub, err := parseECPublicKeyPEM(string(pemBlock))
	require.NoError(t, err)
	assert.Equal(t, key.PublicKey.X, pub.X)
	assert.Equal(t, key.PublicKey.Y, pub.Y)
}

func TestParseECPublicKeyPEM_InvalidPEM(t *testing.T) {
	_, err := parseECPublicKeyPEM("not a PEM block")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM")
}

func TestParseECPublicKeyPEM_NotECDSA(t *testing.T) {
	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("not a valid key"),
	})

	_, err := parseECPublicKeyPEM(string(pemBlock))
	assert.Error(t, err)
}
