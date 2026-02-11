package sign

import (
	"context"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Ensure KMSSigner implements the Signer interface.
var _ Signer = (*KMSSigner)(nil)

// KMSClientAPI defines the interface for the AWS KMS client operations we use.
type KMSClientAPI interface {
	GetPublicKey(ctx context.Context, params *kms.GetPublicKeyInput, optFns ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(ctx context.Context, params *kms.SignInput, optFns ...func(*kms.Options)) (*kms.SignOutput, error)
}

// KMSSigner is a signer that uses AWS KMS to sign data.
type KMSSigner struct {
	client    KMSClientAPI
	keyID     string
	publicKey PublicKey
}

// NewKMSSigner creates a new KMSSigner.
func NewKMSSigner(client KMSClientAPI, keyID string) (*KMSSigner, error) {
	// Retrieve the public key from KMS to verify access and initialize the signer.
	pubKeyOutput, err := client.GetPublicKey(context.Background(), &kms.GetPublicKeyInput{
		KeyId: aws.String(keyID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from KMS: %w", err)
	}

	if pubKeyOutput.KeySpec != types.KeySpecEccSecgP256k1 {
		return nil, fmt.Errorf("unsupported key spec: %s, expected %s", pubKeyOutput.KeySpec, types.KeySpecEccSecgP256k1)
	}

	// Parse the SubjectPublicKeyInfo (SPKI) structure
	var spki struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal(pubKeyOutput.PublicKey, &spki); err != nil {
		return nil, fmt.Errorf("failed to parse public key ASN.1: %w", err)
	}

	// The SubjectPublicKey bytes are the uncompressed point.
	pubKeyBytes := spki.SubjectPublicKey.Bytes
	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	return &KMSSigner{
		client:    client,
		keyID:     keyID,
		publicKey: NewEthereumPublicKey(pubKey),
	}, nil
}

func (s *KMSSigner) PublicKey() PublicKey {
	return s.publicKey
}

// Sign expects the input data to be a hash (e.g., Keccak256 hash).
func (s *KMSSigner) Sign(hash []byte) (Signature, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("invalid hash length: %d", len(hash))
	}

	signInput := &kms.SignInput{
		KeyId:            aws.String(s.keyID),
		Message:          hash,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
	}

	signOutput, err := s.client.Sign(context.Background(), signInput)
	if err != nil {
		return nil, fmt.Errorf("failed to sign with KMS: %w", err)
	}

	// Parse the DER-encoded signature
	var rs struct {
		R, S *big.Int
	}
	if _, err := asn1.Unmarshal(signOutput.Signature, &rs); err != nil {
		return nil, fmt.Errorf("failed to parse signature ASN.1: %w", err)
	}

	// Canonicalize S to be in the lower half of the curve order (EIP-2)
	secp256k1N := crypto.S256().Params().N
	halfN := new(big.Int).Div(secp256k1N, big.NewInt(2))
	if rs.S.Cmp(halfN) > 0 {
		rs.S.Sub(secp256k1N, rs.S)
	}

	// Recover V
	// Try both 0 and 1
	var signature []byte
	found := false
	for _, v := range []byte{0, 1} {
		// Construct signature: R || S || V
		sig := make([]byte, 65)
		copy(sig[0:32], padBytes(rs.R.Bytes(), 32))
		copy(sig[32:64], padBytes(rs.S.Bytes(), 32))
		sig[64] = v

		// Recover public key
		recoveredPub, err := crypto.SigToPub(hash, sig)
		if err == nil {
			recoveredAddr := crypto.PubkeyToAddress(*recoveredPub)
			knownAddr := common.HexToAddress(s.publicKey.Address().String())
			if recoveredAddr == knownAddr {
				signature = sig
				signature[64] += 27 // Adjust V for Ethereum (27 or 28)
				found = true
				break
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("failed to recover valid signature")
	}

	return Signature(signature), nil
}

// padBytes left-pads the byte slice with zeros to the given length.
func padBytes(b []byte, length int) []byte {
	if len(b) >= length {
		return b
	}
	padded := make([]byte, length)
	copy(padded[length-len(b):], b)
	return padded
}
