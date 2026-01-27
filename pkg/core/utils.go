package core

import (
	"errors"
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

func TransitionToIntent(transition *Transition) (uint8, error) {
	if transition == nil {
		return 0, errors.New("at least one transition is expected")
	}

	switch transition.Type {
	case TransitionTypeTransferSend,
		TransitionTypeTransferReceive,
		TransitionTypeCommit,
		TransitionTypeRelease:
		return INTENT_OPERATE, nil
	case TransitionTypeFinalize:
		return INTENT_CLOSE, nil
	case TransitionTypeHomeDeposit:
		return INTENT_DEPOSIT, nil
	case TransitionTypeHomeWithdrawal:
		return INTENT_WITHDRAW, nil
	case TransitionTypeMutualLock:
		return INTENT_INITIATE_ESCROW_DEPOSIT, nil
	case TransitionTypeEscrowDeposit:
		return INTENT_FINALIZE_ESCROW_DEPOSIT, nil
	case TransitionTypeEscrowLock:
		return INTENT_INITIATE_ESCROW_WITHDRAWAL, nil
	case TransitionTypeEscrowWithdraw:
		return INTENT_FINALIZE_ESCROW_WITHDRAWAL, nil
	case TransitionTypeMigrate:
		return INTENT_INITIATE_MIGRATION, nil
	// TODO: Add:
	// FINALIZE_MIGRATION.
	default:
		return 0, errors.New("unexpected transition type: " + transition.Type.String())
	}
}

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

// GetHomeChannelID generates a unique identifier for a primary channel based on its definition.
// This matches the Solidity getChannelId function which computes keccak256(abi.encode(ChannelDefinition)).
// The metadata is derived from the asset: first 8 bytes of keccak256(asset) padded to 32 bytes.
func GetHomeChannelID(node, user, asset string, nonce uint64, challengeDuration uint32) (string, error) {
	// Generate metadata from asset
	userAddr := common.HexToAddress(user)
	nodeAddr := common.HexToAddress(node)
	metadata := GenerateChannelMetadata(asset)

	// Define the struct to match Solidity's ChannelDefinition
	type channelDefinition struct {
		ChallengeDuration uint32
		User              common.Address
		Node              common.Address
		Nonce             uint64
		Metadata          [32]byte
	}

	def := channelDefinition{
		ChallengeDuration: challengeDuration,
		User:              userAddr,
		Node:              nodeAddr,
		Nonce:             nonce,
		Metadata:          metadata,
	}

	// Define the struct type for ABI encoding
	channelDefType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "challengeDuration", Type: "uint32"},
		{Name: "user", Type: "address"},
		{Name: "node", Type: "address"},
		{Name: "nonce", Type: "uint64"},
		{Name: "metadata", Type: "bytes32"},
	})
	if err != nil {
		return "", err
	}

	args := abi.Arguments{
		{Type: channelDefType},
	}

	packed, err := args.Pack(def)
	if err != nil {
		return "", err
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetEscrowChannelID derives an escrow-specific channel ID based on a home channel and state version.
// This matches the Solidity getEscrowId function which computes keccak256(abi.encode(channelId, version)).
func GetEscrowChannelID(homeChannelID string, stateVersion uint64) (string, error) {
	rawHomeChannelID := common.HexToHash(homeChannelID)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelId
		{Type: uint64Type}, // version
	}

	packed, err := args.Pack(rawHomeChannelID, stateVersion)
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

func GetStateTransitionsHash(transitions []Transition) ([32]byte, error) {
	hash := [32]byte{}
	type contractTransition struct {
		Type      uint8
		TxId      string
		AccountId string
		Amount    string
	}
	transitionType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "type", Type: "uint8"},
		{Name: "txId", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "amount", Type: "string"},
	})
	if err != nil {
		return hash, fmt.Errorf("failed to create transition type: %w", err)
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.SliceTy, Elem: &transitionType}},
	}

	contractsTransitions := make([]contractTransition, len(transitions))

	for i, t := range transitions {
		contractsTransitions[i] = contractTransition{
			Type:      uint8(t.Type),
			TxId:      t.TxID,
			AccountId: t.AccountID,
			Amount:    t.Amount.String(),
		}
	}

	packed, err := args.Pack(
		contractsTransitions,
	)
	if err != nil {
		return hash, fmt.Errorf("failed to pack app state update: %w", err)
	}

	hash = crypto.Keccak256Hash(packed)
	return hash, nil
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

// GenerateChannelMetadata creates metadata from an asset by taking the first 8 bytes of keccak256(asset)
// and padding the rest with zeros to make a 32-byte array.
func GenerateChannelMetadata(asset string) [32]byte {
	assetHash := crypto.Keccak256Hash([]byte(asset))
	var metadata [32]byte
	copy(metadata[:8], assetHash[:8])
	return metadata
}
