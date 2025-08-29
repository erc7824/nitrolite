// Package sign provides blockchain-agnostic cryptographic signing interfaces.
//
// This package defines core interfaces for digital signatures that can be
// implemented by various blockchain ecosystems while maintaining a consistent
// API for signing operations.
//
// The primary interfaces are:
//
//   - Signer: Core interface for cryptographic signing operations
//   - PublicKey: Interface for public key operations
//   - Address: Interface for blockchain addresses
//   - AddressRecoverer: Optional interface for signature-based address recovery
//
// # Security Design
//
// This package follows security best practices by:
//   - Never exposing private key material through interfaces
//   - Providing only necessary operations (signing and public key access)
//   - Supporting hardware security modules (HSM) and key management services (KMS)
//   - Preventing accidental private key leakage in logs or debugging
//
// Usage
//
//	// Create a new Ethereum signer from a hex-encoded private key
//	signer, err := sign.NewEthereumSigner(privateKeyHex)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Sign a message (provide hash, not raw message)
//	message := []byte("hello world")
//	hash := ethcrypto.Keccak256Hash(message)
//	signature, err := signer.Sign(hash.Bytes())
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get the address
//	address := signer.PublicKey().Address()
//	fmt.Println("Address:", address.String())
package sign
