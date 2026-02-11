package sign

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKMSClient is a mock of KMSClientAPI.
type MockKMSClient struct {
	mock.Mock
}

func (m *MockKMSClient) GetPublicKey(ctx context.Context, params *kms.GetPublicKeyInput, optFns ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kms.GetPublicKeyOutput), args.Error(1)
}

func (m *MockKMSClient) Sign(ctx context.Context, params *kms.SignInput, optFns ...func(*kms.Options)) (*kms.SignOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kms.SignOutput), args.Error(1)
}

func TestKMSSigner(t *testing.T) {
	// Generate a real key pair for testing
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	assert.NoError(t, err)

	pubKeyBytes := crypto.FromECDSAPub(&privateKey.PublicKey)

	// Create SPKI structure for GetPublicKey response
	// We only care about SubjectPublicKey for this test as KMSSigner logic extracts it
	spki := struct {
		Algorithm        pkix.AlgorithmIdentifier
		SubjectPublicKey asn1.BitString
	}{
		Algorithm: pkix.AlgorithmIdentifier{
			Algorithm:  asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}, // id-ecPublicKey
		},
		SubjectPublicKey: asn1.BitString{
			Bytes:     pubKeyBytes,
			BitLength: len(pubKeyBytes) * 8,
		},
	}
	spkiBytes, err := asn1.Marshal(spki)
	assert.NoError(t, err)

	mockClient := new(MockKMSClient)
	keyID := "test-key-id"

	// Mock GetPublicKey
	keyIdPtr := &keyID
	mockClient.On("GetPublicKey", mock.Anything, mock.MatchedBy(func(input *kms.GetPublicKeyInput) bool {
		return *input.KeyId == keyID
	})).Return(&kms.GetPublicKeyOutput{
		KeyId:     keyIdPtr,
		PublicKey: spkiBytes,
		KeySpec:   types.KeySpecEccSecgP256k1,
	}, nil)

	signer, err := NewKMSSigner(mockClient, keyID)
	assert.NoError(t, err)
	assert.NotNil(t, signer)

	expectedAddr := NewEthereumAddress(crypto.PubkeyToAddress(privateKey.PublicKey))
	assert.True(t, expectedAddr.Equals(signer.PublicKey().Address()))

	// Test Sign
	hash := crypto.Keccak256([]byte("hello world"))

	// Generate signature using the private key
	sig, err := crypto.Sign(hash, privateKey)
	assert.NoError(t, err)

	// crypto.Sign returns [R || S || V] (65 bytes)
	// KMS Sign returns ASN.1 of R and S.
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])

	asn1Sig, err := asn1.Marshal(struct{ R, S *big.Int }{r, s})
	assert.NoError(t, err)

	mockClient.On("Sign", mock.Anything, mock.MatchedBy(func(input *kms.SignInput) bool {
		return *input.KeyId == keyID && string(input.Message) == string(hash)
	})).Return(&kms.SignOutput{
		Signature: asn1Sig,
		KeyId:     keyIdPtr,
	}, nil)

	signature, err := signer.Sign(hash)
	assert.NoError(t, err)
	assert.Len(t, signature, 65)

	// Verify signature
	recoveredAddr, err := RecoverAddressFromHash(hash, signature)
	assert.NoError(t, err)
	assert.True(t, signer.PublicKey().Address().Equals(recoveredAddr))
}
