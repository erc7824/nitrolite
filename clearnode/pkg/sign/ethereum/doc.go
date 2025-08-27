// Package ethereum provides Ethereum-specific implementation of the sign interfaces.
//
// This package implements the blockchain-agnostic signing interfaces defined in
// the parent sign package specifically for the Ethereum ecosystem.
//
// Features
//
//   - ECDSA signature generation using secp256k1 curve
//   - Keccak-256 message hashing (Ethereum standard)
//   - Address recovery from signatures
//   - Ethereum address format compatibility
//
// Usage
//
//	// Create a new Ethereum signer from a hex-encoded private key
//	signer, err := ethereum.NewEthereumSigner(privateKeyHex)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Sign a message
//	message := []byte("hello world")
//	signature, err := signer.Sign(message)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Get the address
//	address := signer.PublicKey().Address()
//	fmt.Println("Address:", address.String())
//
//	// Recover address from signature (if signer implements AddressRecoverer)
//	if recoverer, ok := signer.(sign.AddressRecoverer); ok {
//		recoveredAddr, err := recoverer.RecoverAddress(message, signature)
//		if err == nil {
//			fmt.Println("Recovered:", recoveredAddr.String())
//		}
//	}
//
// # Security
//
// Private keys are kept internal to the Signer struct and are never exposed
// through the public API. This design supports hardware wallets and key
// management services that cannot or should not expose private key material.
package ethereum
