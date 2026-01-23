package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

var (
	uint8Type, _   = abi.NewType("uint8", "", nil)
	uint32Type, _  = abi.NewType("uint32", "", nil)
	uint64Type, _  = abi.NewType("uint64", "", nil)
	uint256Type, _ = abi.NewType("uint256", "", nil)
)

// ValidateDecimalPrecision validates that an amount doesn't exceed the maximum allowed decimal places.
func ValidateDecimalPrecision(amount decimal.Decimal, maxDecimals uint8) error {
	if amount.Exponent() < -int32(maxDecimals) {
		return fmt.Errorf("amount exceeds maximum decimal precision: max %d decimals allowed, got %d", maxDecimals, -amount.Exponent())
	}
	return nil
}

// DecimalToBigInt converts a decimal.Decimal amount to *big.Int scaled to the token's smallest unit.
// For example, 1.23 USDC (6 decimals) becomes 1230000.
// This is used when preparing amounts for smart contract calls.
func DecimalToBigInt(amount decimal.Decimal, decimals uint8) (*big.Int, error) {
	// Multiply by 10^decimals to convert to smallest unit
	multiplier := decimal.New(1, int32(decimals))
	scaled := amount.Mul(multiplier)

	err := ValidateDecimalPrecision(scaled, 0)
	if err != nil {
		return nil, err
	}

	// Convert to big.Int, truncating any remaining fractional part
	return scaled.BigInt(), nil
}

// GetHomeChannelID generates a unique identifier for a primary channel between a node and a user.
func GetHomeChannelID(nodeAddress, userAddress, asset string, nonce uint64, challenge uint32) (string, error) {
	nodeAddr := common.HexToAddress(nodeAddress)
	userAddr := common.HexToAddress(userAddress)

	assetHash := crypto.Keccak256Hash([]byte(asset))
	assetID := assetHash[:8] // Use first 8 bytes as asset identifier

	metadata := make([]byte, 32)
	copy(metadata[:8], assetID)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}},              // node
		{Type: abi.Type{T: abi.AddressTy}},              // user
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // metadata
		{Type: uint32Type},                              // challenge
		{Type: uint64Type},                              // nonce
	}

	packed, err := args.Pack(nodeAddr, userAddr, metadata, challenge, nonce)
	if err != nil {
		return "", err
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetEscrowChannelID derives an escrow-specific channel ID based on a home channel and state version.
func GetEscrowChannelID(homeChannelID string, stateVersion uint64) (string, error) {
	rawHomeChannelID := common.HexToHash(homeChannelID)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // homeChannelID
		{Type: uint256Type},                             // stateVersion
	}

	stateVersionBI := new(big.Int).SetUint64(stateVersion)

	packed, err := args.Pack(rawHomeChannelID, stateVersionBI)
	if err != nil {
		return "", err
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetStateID creates a unique hash representing a specific snapshot of a user's wallet and asset state.
func GetStateID(userWallet, asset string, epoch, version uint64) string {
	userAddr := common.HexToAddress(userWallet)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}}, // userWallet
		{Type: abi.Type{T: abi.StringTy}},  // asset symbol/string
		{Type: uint256Type},                // epoch
		{Type: uint256Type},                // version
	}

	packed, _ := args.Pack(
		userAddr,
		asset,
		new(big.Int).SetUint64(epoch),
		new(big.Int).SetUint64(version),
	)

	return crypto.Keccak256Hash(packed).Hex()
}

// GetSenderTransactionID calculates and returns a unique transaction ID reference for actions initiated by user.
func GetSenderTransactionID(toAccount string, senderNewStateID string) (string, error) {
	return getTransactionID(toAccount, senderNewStateID)
}

// GetReceiverTransactionID calculates and returns a unique transaction ID reference for actions initiated by node.
func GetReceiverTransactionID(fromAccount, receiverNewStateID string) (string, error) {
	return getTransactionID(fromAccount, receiverNewStateID)
}

func getTransactionID(account, newStateID string) (string, error) {
	args := abi.Arguments{
		{Type: abi.Type{T: abi.StringTy}},
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}},
	}

	receiverStateID := common.HexToHash(newStateID)
	packed, err := args.Pack(account, receiverStateID)
	if err != nil {
		return "", fmt.Errorf("failed to pack transaction ID arguments: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}
