package sign_test

import (
	"fmt"
	"log"

	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign/ethereum"
)

// ExampleSigner demonstrates creating a signer and signing a message.
func ExampleSigner() {
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" // Example private key

	// Create a new signer. It returns the generic sign.Signer interface.
	signer, err := ethereum.NewEthereumSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	// You can now use the signer for generic operations.
	fmt.Println("Address:", signer.PrivateKey().PublicKey().Address())

	message := []byte("hello world")
	signature, err := signer.PrivateKey().Sign(message)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Signature length:", len(signature))
	// Output:
	// Address: 0x1Be31A94361a391bBaFB2a4CCd704F57dc04d4bb
	// Signature length: 65
}

// ExampleSignature_String demonstrates the String method of Signature.
func ExampleSignature_String() {
	sig := sign.Signature([]byte{0x01, 0x02, 0x03, 0x04})
	fmt.Println(sig.String())
	// Output:
	// 0x01020304
}

// ExampleSignature_MarshalJSON demonstrates JSON marshaling of signatures.
func ExampleSignature_MarshalJSON() {
	sig := sign.Signature([]byte{0x01, 0x02, 0x03, 0x04})
	jsonData, err := sig.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonData))
	// Output:
	// "0x01020304"
}

// ExampleSignature_UnmarshalJSON demonstrates JSON unmarshaling of signatures.
func ExampleSignature_UnmarshalJSON() {
	var sig sign.Signature
	jsonData := []byte(`"0x01020304"`)

	err := sig.UnmarshalJSON(jsonData)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%x\n", []byte(sig))
	// Output:
	// 01020304
}

// Example_blockchainSpecificFeatures demonstrates using blockchain-specific features.
// This shows how to call EIP-712 recovery for Ethereum directly from the implementation package.
func Example_blockchainSpecificFeatures() {
	// Example message for standard recovery
	message := []byte("hello world")

	// Create a signature using our signer
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	signer, err := ethereum.NewEthereumSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	signature, err := signer.PrivateKey().Sign(message)
	if err != nil {
		log.Fatal(err)
	}

	// Call the function directly from the `ethereum` package for message recovery
	recoveredAddr, err := ethereum.RecoverAddress(message, signature)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Verify it matches the signer's address
	signerAddr := signer.PrivateKey().PublicKey().Address().String()
	fmt.Printf("Addresses match: %t\n", recoveredAddr == signerAddr)
	// Output:
	// Addresses match: true
}

// Example_genericAddressRecovery demonstrates using the generic AddressRecoverer interface.
func Example_genericAddressRecovery() {
	message := []byte("hello world")

	// Create a signer
	pkHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	signer, err := ethereum.NewEthereumSigner(pkHex)
	if err != nil {
		log.Fatal(err)
	}

	// Sign the message
	signature, err := signer.PrivateKey().Sign(message)
	if err != nil {
		log.Fatal(err)
	}

	// Use the generic AddressRecoverer interface
	if recoverer, ok := signer.(sign.AddressRecoverer); ok {
		recoveredAddr, err := recoverer.RecoverAddress(message, signature)
		if err != nil {
			log.Fatal(err)
		}

		signerAddr := signer.PrivateKey().PublicKey().Address().String()
		fmt.Printf("Generic recovery works: %t\n", recoveredAddr == signerAddr)
	} else {
		fmt.Println("Signer does not support address recovery")
	}
	// Output:
	// Generic recovery works: true
}

// ExampleAddress_comparison demonstrates address comparison methods.
func ExampleAddress_comparison() {
	// Create two different signers
	pkHex1 := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	pkHex2 := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	signer1, err := ethereum.NewEthereumSigner(pkHex1)
	if err != nil {
		log.Fatal(err)
	}

	signer2, err := ethereum.NewEthereumSigner(pkHex2)
	if err != nil {
		log.Fatal(err)
	}

	// Get their addresses
	addr1 := signer1.PrivateKey().PublicKey().Address()
	addr2 := signer2.PrivateKey().PublicKey().Address()
	addr1Copy := signer1.PrivateKey().PublicKey().Address()

	// Test equality
	fmt.Printf("addr1 equals addr2: %t\n", addr1.Equals(addr2))
	fmt.Printf("addr1 equals addr1Copy: %t\n", addr1.Equals(addr1Copy))

	// Output:
	// addr1 equals addr2: false
	// addr1 equals addr1Copy: true
}
