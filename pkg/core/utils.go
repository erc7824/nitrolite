package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// TODO: GetHomeChannelID(), GetEscrowChannelID(), GetStateID(), GetTransactionID()

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

	intentType, err := abi.NewType("uint8", "", nil)
	if err != nil {
		return nil, err
	}
	versionType, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: abi.Type{T: abi.FixedBytesTy, Size: 32}}, // channelID
		{Type: intentType},               // intent
		{Type: versionType},              // version
		{Type: abi.Type{T: abi.BytesTy}}, // stateData
		{Type: allocationType},           // allocations (tuple[])
	}

	packed, err := args.Pack(channelID, state.Version)
	if err != nil {
		return nil, err
	}
	return packed, nil
}

// UnpackState decodes ABI-packed CrossChainState bytes into a State struct, extracting version and signatures.
func UnpackState(data []byte) (*State, error) {
	// Define the CrossChainState tuple type
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

	args := abi.Arguments{
		{Type: crossChainStateType},
	}

	unpacked, err := args.Unpack(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack data: %w", err)
	}

	if len(unpacked) == 0 {
		return nil, fmt.Errorf("no data unpacked")
	}

	// Convert unpacked data to State struct
	unpackedMap, ok := unpacked[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert unpacked data to map")
	}

	state := &State{}

	// Extract version - skip if not present
	if version, ok := unpackedMap["version"].(*big.Int); ok {
		state.Version = version.Uint64()
	}

	// Extract signatures - skip if not present
	if participantSig, ok := unpackedMap["participantSig"].([]byte); ok {
		sigStr := string(participantSig)
		state.UserSig = &sigStr
	}

	if nodeSig, ok := unpackedMap["nodeSig"].([]byte); ok {
		sigStr := string(nodeSig)
		state.NodeSig = &sigStr
	}

	// Note: Other fields like homeState, nonHomeState, intent, homeChainId
	// are not mapped to the State struct from types.go as they don't exist there
	// They would need to be handled separately if needed

	return state, nil
}
