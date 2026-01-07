package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	uint8Type, _   = abi.NewType("uint8", "", nil)
	uint32Type, _  = abi.NewType("uint32", "", nil)
	uint256Type, _ = abi.NewType("uint256", "", nil)
)

// GetHomeChannelID generates a unique identifier for a primary channel between a node and a user.
func GetHomeChannelID(nodeAddress, userAddress, tokenAddress string, nonce, challenge uint64) (string, error) {
	nodeAddr := common.HexToAddress(nodeAddress)
	userAddr := common.HexToAddress(userAddress)

	tokenAddr := common.HexToAddress(tokenAddress)
	// TODO: decide token or asset

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}}, // node
		{Type: abi.Type{T: abi.AddressTy}}, // user
		{Type: abi.Type{T: abi.AddressTy}}, // asset
		{Type: uint32Type},                 // challenge
		{Type: uint256Type},                // nonce
	}

	packed, err := args.Pack(nodeAddr, userAddr, tokenAddr, challenge, nonce)
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

	packed, err := args.Pack(rawHomeChannelID, stateVersion)
	if err != nil {
		return "", err
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetStateID creates a unique hash representing a specific snapshot of a user's wallet and asset state.
func GetStateID(userWallet, asset string, epoch, version uint64) (string, error) {
	userAddr := common.HexToAddress(userWallet)

	args := abi.Arguments{
		{Type: abi.Type{T: abi.AddressTy}}, // userWallet
		{Type: abi.Type{T: abi.StringTy}},  // asset symbol/string
		{Type: uint256Type},                // epoch
		{Type: uint256Type},                // version
	}

	packed, err := args.Pack(
		userAddr,
		asset,
		new(big.Int).SetUint64(epoch),
		new(big.Int).SetUint64(version),
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack state ID: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// GetTransactionID calculates a unique transaction reference based on the participating account and the resulting state.
func GetTransactionID(toAccount, fromAccount string, senderNewStateID, receiverNewStateID *string) (string, error) {
	var packed []byte
	var err error

	// 1) User Initiated: Hash(ToAccount, SenderNewStateID)
	// 2) Node Initiated: Hash(FromAccount, ReceiverNewStateID)

	if senderNewStateID != nil {
		args := abi.Arguments{
			{Type: abi.Type{T: abi.StringTy}},               // ToAccount
			{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // SenderNewStateID
		}
		senderStateID := common.HexToHash(*senderNewStateID)
		packed, err = args.Pack(toAccount, senderStateID)
	} else if receiverNewStateID != nil {
		args := abi.Arguments{
			{Type: abi.Type{T: abi.StringTy}},               // FromAccount
			{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // ReceiverNewStateID
		}
		receiverStateID := common.HexToHash(*receiverNewStateID)
		packed, err = args.Pack(fromAccount, receiverStateID)
	} else {
		return "", fmt.Errorf("transaction must have either SenderNewStateID or ReceiverNewStateID")
	}

	if err != nil {
		return "", fmt.Errorf("failed to pack transaction ID arguments: %w", err)
	}

	return crypto.Keccak256Hash(packed).Hex(), nil
}

// PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
func PackState(channelID string, state State) ([]byte, error) {
	// TODO: refine with the current packing approach
	allocationType, err := abi.NewType("tuple[]", "", []abi.ArgumentMarshaling{
		{Name: "destination", Type: "address"},
		{Name: "token", Type: "address"},
		{Name: "amount", Type: "uint256"},
	})
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelID
		{Type: uint8Type},                // intent
		{Type: uint256Type},              // version
		{Type: abi.Type{T: abi.BytesTy}}, // stateData
		{Type: allocationType},           // allocations (tuple[])
	}

	packed, err := args.Pack(channelID, state.Version)
	if err != nil {
		return nil, err
	}
	return packed, nil
}

// UnpackState decodes ABI-packed bytes back into a State struct for off-chain processing.
func UnpackState(data []byte) (*State, error) {
	crossChainStateType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "version", Type: "uint256"},
		{Name: "homeChainId", Type: "uint64"},
		{Name: "intent", Type: "uint8"},
		{Name: "homeState", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "chainId", Type: "uint64"},
			{Name: "token", Type: "address"},
			{Name: "participantBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
				{Name: "allocation", Type: "uint256"},
				{Name: "netFlow", Type: "int256"},
			}},
			{Name: "nodeBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
				{Name: "allocation", Type: "uint256"},
				{Name: "netFlow", Type: "int256"},
			}},
		}},
		{Name: "nonHomeState", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "chainId", Type: "uint64"},
			{Name: "token", Type: "address"},
			{Name: "participantBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
				{Name: "allocation", Type: "uint256"},
				{Name: "netFlow", Type: "int256"},
			}},
			{Name: "nodeBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
				{Name: "allocation", Type: "uint256"},
				{Name: "netFlow", Type: "int256"},
			}},
		}},
		{Name: "participantSig", Type: "bytes"},
		{Name: "nodeSig", Type: "bytes"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cross chain state type: %w", err)
	}

	args := abi.Arguments{{Type: crossChainStateType}}
	unpacked, err := args.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack data: %w", err)
	}

	if len(unpacked) == 0 {
		return nil, fmt.Errorf("no data unpacked")
	}

	unpackedMap, ok := unpacked[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert unpacked data to map")
	}

	state := &State{}
	if version, ok := unpackedMap["version"].(*big.Int); ok {
		state.Version = version.Uint64()
	}
	if participantSig, ok := unpackedMap["participantSig"].([]byte); ok {
		sigStr := string(participantSig)
		state.UserSig = &sigStr
	}
	if nodeSig, ok := unpackedMap["nodeSig"].([]byte); ok {
		sigStr := string(nodeSig)
		state.NodeSig = &sigStr
	}

	return state, nil
}
