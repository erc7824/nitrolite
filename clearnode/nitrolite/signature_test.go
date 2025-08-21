package nitrolite

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

// TestSignAndVerify ensures that a signature created with Sign can be verified correctly.
func TestSignAndVerify(t *testing.T) {
	// Generate a new ECDSA key pair.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Original data to sign.
	data := []byte("Hello, Ethereum!")

	// Sign the data.
	sig, err := Sign(data, privateKey)
	if err != nil {
		t.Fatalf("failed to sign data: %v", err)
	}

	// Verify the signature with the correct data and address.
	valid, err := Verify(data, sig, address)
	if err != nil {
		t.Fatalf("failed to verify signature: %v", err)
	}
	if !valid {
		t.Fatal("expected signature to be valid")
	}

	// Verify with modified data (should fail).
	modifiedData := []byte("Hello, modified!")
	valid, err = Verify(modifiedData, sig, address)
	if err != nil {
		t.Fatalf("error verifying signature with modified data: %v", err)
	}
	if valid {
		t.Fatal("expected signature to be invalid for modified data")
	}

	// Verify with a different address (should fail).
	otherKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate alternative private key: %v", err)
	}
	wrongAddress := crypto.PubkeyToAddress(otherKey.PublicKey)
	valid, err = Verify(data, sig, wrongAddress)
	if err != nil {
		t.Fatalf("error verifying signature with wrong address: %v", err)
	}
	if valid {
		t.Fatal("expected signature to be invalid for the wrong address")
	}
}

// TestVerifyInvalidSignature demonstrates that a tampered signature fails verification.
func TestVerifyInvalidSignature(t *testing.T) {
	// Generate key and sign data.
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	data := []byte("Test data")
	sig, err := Sign(data, privateKey)
	if err != nil {
		t.Fatalf("failed to sign data: %v", err)
	}

	// Tamper with the signature (flip some bit).
	sig[0] ^= 0xff

	// Use the original public address.
	publicAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Attempt to verify the tampered signature.
	valid, err := Verify(data, sig, publicAddress)
	if err == nil && valid {
		t.Fatal("expected tampered signature to be invalid")
	}
}

// TestSignInvalidKey simulates error handling by passing a nil key to Sign.
func TestSignInvalidKey(t *testing.T) {
	data := []byte("Data with nil key")
	_, err := Sign(data, nil)
	if err == nil {
		t.Fatal("expected an error when signing with a nil key")
	}
}
