package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ERC-1271 magic value returned by isValidSignature when the signature is valid.
// bytes4(keccak256("isValidSignature(bytes32,bytes)"))
var erc1271MagicValue = [4]byte{0x16, 0x26, 0xba, 0x7e}

// isValidSignature(bytes32,bytes) selector
var isValidSignatureSelector = crypto.Keccak256([]byte("isValidSignature(bytes32,bytes)"))[:4]

// Ethereum interface for ERC-1271 verification (matches existing Ethereum interface in the codebase).
type ERC1271Verifier interface {
	CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error)
	CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error)
}

// IsContract checks if the given address is a smart contract (has code deployed).
func IsContract(ctx context.Context, client ERC1271Verifier, addr common.Address) (bool, error) {
	code, err := client.CodeAt(ctx, addr, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check contract code: %w", err)
	}
	return len(code) > 0, nil
}

// VerifyERC1271Signature calls isValidSignature(bytes32,bytes) on a smart contract
// wallet and returns true if the contract considers the signature valid.
func VerifyERC1271Signature(ctx context.Context, client ERC1271Verifier, contractAddr common.Address, hash []byte, signature []byte) (bool, error) {
	// ABI-encode the call: isValidSignature(bytes32 hash, bytes signature)
	// Selector (4 bytes) + hash (32 bytes padded) + offset to bytes (32 bytes) + length (32 bytes) + signature data (padded to 32)
	callData := make([]byte, 0, 4+32+32+32+len(signature)+32)

	// Function selector
	callData = append(callData, isValidSignatureSelector...)

	// bytes32 hash (already 32 bytes)
	if len(hash) != 32 {
		return false, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	callData = append(callData, hash...)

	// Offset to bytes parameter (always 64 = 0x40 for two fixed params)
	offset := make([]byte, 32)
	offset[31] = 64
	callData = append(callData, offset...)

	// Length of signature bytes
	sigLen := make([]byte, 32)
	bigLen := big.NewInt(int64(len(signature)))
	bigLen.FillBytes(sigLen)
	callData = append(callData, sigLen...)

	// Signature data (padded to 32-byte boundary)
	callData = append(callData, signature...)
	if pad := len(signature) % 32; pad != 0 {
		callData = append(callData, make([]byte, 32-pad)...)
	}

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &contractAddr,
		Data: callData,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("isValidSignature call failed: %w", err)
	}

	if len(result) < 32 {
		return false, fmt.Errorf("isValidSignature returned %d bytes, expected 32", len(result))
	}

	// The magic value is in the first 4 bytes of the 32-byte return value
	return result[0] == erc1271MagicValue[0] &&
		result[1] == erc1271MagicValue[1] &&
		result[2] == erc1271MagicValue[2] &&
		result[3] == erc1271MagicValue[3], nil
}
