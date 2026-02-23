// Package gcpkms implements [sign.Signer] using Google Cloud KMS.
//
// It wraps a GCP Cloud KMS asymmetric signing key (secp256k1) and converts
// the KMS-produced DER signatures into Ethereum-compatible 65-byte format.
package gcpkms

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"

	"github.com/erc7824/nitrolite/pkg/sign"
	kmssign "github.com/erc7824/nitrolite/pkg/sign/kms"
)

// GCPKMSSigner implements [sign.Signer] using a GCP Cloud KMS secp256k1 key.
//
// The key must be created with:
//   - Algorithm: EC_SIGN_SECP256K1_SHA256
//   - Purpose: ASYMMETRIC_SIGN
//   - Protection level: HSM (recommended)
type GCPKMSSigner struct {
	client      *kms.KeyManagementClient
	keyName     string // full resource name including version
	publicKey   sign.EthereumPublicKey
	ecPublicKey *ecdsa.PublicKey
}

// NewSigner creates a new GCP KMS signer.
//
// keyResourceName must be the full key version resource name:
// projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}/cryptoKeyVersions/{version}
func NewSigner(ctx context.Context, keyResourceName string) (*GCPKMSSigner, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	return newSignerWithClient(ctx, client, keyResourceName)
}

// newSignerWithClient creates a signer with an injected KMS client (for testing).
func newSignerWithClient(ctx context.Context, client *kms.KeyManagementClient, keyResourceName string) (*GCPKMSSigner, error) {
	// Fetch the public key from KMS
	resp, err := client.GetPublicKey(ctx, &kmspb.GetPublicKeyRequest{
		Name: keyResourceName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from KMS: %w", err)
	}

	// Parse PEM-encoded public key
	ecPub, err := parseECPublicKeyPEM(resp.Pem)
	if err != nil {
		return nil, fmt.Errorf("failed to parse KMS public key: %w", err)
	}

	// Validate it's on secp256k1
	if err := kmssign.ValidateSecp256k1PublicKey(ecPub); err != nil {
		return nil, fmt.Errorf("KMS key validation failed: %w", err)
	}

	return &GCPKMSSigner{
		client:      client,
		keyName:     keyResourceName,
		publicKey:   sign.NewEthereumPublicKey(ecPub),
		ecPublicKey: ecPub,
	}, nil
}

// PublicKey returns the cached Ethereum public key derived from the KMS key.
func (s *GCPKMSSigner) PublicKey() sign.PublicKey {
	return s.publicKey
}

// Sign signs the given hash using GCP KMS and returns an Ethereum-compatible signature.
//
// The hash should be a 32-byte digest (e.g., Keccak256). GCP KMS will sign this
// hash directly using the secp256k1 key. The returned DER signature is then
// converted to Ethereum's 65-byte R || S || V format.
func (s *GCPKMSSigner) Sign(hash []byte) (sign.Signature, error) {
	resp, err := s.client.AsymmetricSign(context.Background(), &kmspb.AsymmetricSignRequest{
		Name: s.keyName,
		Digest: &kmspb.Digest{
			Digest: &kmspb.Digest_Sha256{
				Sha256: hash,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("KMS AsymmetricSign failed: %w", err)
	}

	ethSig, err := kmssign.DERToEthereumSignature(hash, resp.Signature, s.ecPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert KMS signature to Ethereum format: %w", err)
	}

	return sign.Signature(ethSig), nil
}

// Close closes the underlying KMS client connection.
func (s *GCPKMSSigner) Close() error {
	return s.client.Close()
}

// parseECPublicKeyPEM parses a PEM-encoded EC public key.
func parseECPublicKeyPEM(pemStr string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not ECDSA, got %T", pub)
	}

	return ecPub, nil
}
