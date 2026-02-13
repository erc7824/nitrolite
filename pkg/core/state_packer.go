package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type StatePackerV1 struct {
	assetStore AssetStore
}

func NewStatePackerV1(assetStore AssetStore) *StatePackerV1 {
	return &StatePackerV1{
		assetStore: assetStore,
	}
}

// PackState is a convenience function that creates a StatePackerV1 and packs the state.
// For production use, create a StatePackerV1 instance and reuse it.
func PackState(state State, assetStore AssetStore) ([]byte, error) {
	packer := NewStatePackerV1(assetStore)
	return packer.PackState(state)
}

// PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
// This matches the Solidity contract's two-step encoding:
//
//	signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
//	message = abi.encode(channelId, signingData)
//
// The signingData is encoded as dynamic bytes inside the outer abi.encode.
func (p *StatePackerV1) PackState(state State) ([]byte, error) {
	// Ensure HomeChannelID is present
	if state.HomeChannelID == nil {
		return nil, fmt.Errorf("state.HomeChannelID is required for packing")
	}

	// Convert HomeChannelID to bytes32
	channelID := common.HexToHash(*state.HomeChannelID)

	// Generate metadata from the state transition
	metadata, err := GetStateTransitionHash(state.Transition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state transitions hash: %w", err)
	}

	// Define the Ledger type to match Solidity
	ledgerType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "chainId", Type: "uint64"},
		{Name: "token", Type: "address"},
		{Name: "decimals", Type: "uint8"},
		{Name: "userAllocation", Type: "uint256"},
		{Name: "userNetFlow", Type: "int256"},
		{Name: "nodeAllocation", Type: "uint256"},
		{Name: "nodeNetFlow", Type: "int256"},
	})
	if err != nil {
		return nil, err
	}

	bytes32Type := abi.Type{T: abi.FixedBytesTy, Size: 32}

	// Step 1: Pack signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
	signingDataArgs := abi.Arguments{
		{Type: uint64Type},  // version
		{Type: uint8Type},   // intent
		{Type: bytes32Type}, // metadata
		{Type: ledgerType},  // homeState
		{Type: ledgerType},  // nonHomeState
	}

	// Define a private type to match Solidity's Ledger struct
	type contractLedger struct {
		ChainId        uint64
		Token          common.Address
		Decimals       uint8
		UserAllocation *big.Int
		UserNetFlow    *big.Int
		NodeAllocation *big.Int
		NodeNetFlow    *big.Int
	}

	homeDecimals, err := p.assetStore.GetTokenDecimals(state.HomeLedger.BlockchainID, state.HomeLedger.TokenAddress)
	if err != nil {
		return nil, err
	}

	// Convert decimal amounts to big.Int scaled to the token's smallest unit
	userBalanceBI, err := DecimalToBigInt(state.HomeLedger.UserBalance, homeDecimals)
	if err != nil {
		return nil, err
	}
	userNetFlowBI, err := DecimalToBigInt(state.HomeLedger.UserNetFlow, homeDecimals)
	if err != nil {
		return nil, err
	}
	nodeBalanceBI, err := DecimalToBigInt(state.HomeLedger.NodeBalance, homeDecimals)
	if err != nil {
		return nil, err
	}
	nodeNetFlowBI, err := DecimalToBigInt(state.HomeLedger.NodeNetFlow, homeDecimals)
	if err != nil {
		return nil, err
	}

	homeLedger := contractLedger{
		ChainId:        state.HomeLedger.BlockchainID,
		Token:          common.HexToAddress(state.HomeLedger.TokenAddress),
		Decimals:       homeDecimals,
		UserAllocation: userBalanceBI,
		UserNetFlow:    userNetFlowBI,
		NodeAllocation: nodeBalanceBI,
		NodeNetFlow:    nodeNetFlowBI,
	}

	// For nonHomeState, use escrow ledger if available, otherwise use zero values
	var nonHomeLedger contractLedger

	if state.EscrowLedger != nil {
		escrowDecimals, err := p.assetStore.GetTokenDecimals(state.EscrowLedger.BlockchainID, state.EscrowLedger.TokenAddress)
		if err != nil {
			return nil, err
		}

		escrowUserBalanceBI, err := DecimalToBigInt(state.EscrowLedger.UserBalance, escrowDecimals)
		if err != nil {
			return nil, err
		}
		escrowUserNetFlowBI, err := DecimalToBigInt(state.EscrowLedger.UserNetFlow, escrowDecimals)
		if err != nil {
			return nil, err
		}
		escrowNodeBalanceBI, err := DecimalToBigInt(state.EscrowLedger.NodeBalance, escrowDecimals)
		if err != nil {
			return nil, err
		}
		escrowNodeNetFlowBI, err := DecimalToBigInt(state.EscrowLedger.NodeNetFlow, escrowDecimals)
		if err != nil {
			return nil, err
		}

		nonHomeLedger = contractLedger{
			ChainId:        state.EscrowLedger.BlockchainID,
			Token:          common.HexToAddress(state.EscrowLedger.TokenAddress),
			Decimals:       escrowDecimals,
			UserAllocation: escrowUserBalanceBI,
			UserNetFlow:    escrowUserNetFlowBI,
			NodeAllocation: escrowNodeBalanceBI,
			NodeNetFlow:    escrowNodeNetFlowBI,
		}
	} else {
		nonHomeLedger = contractLedger{
			ChainId:        0,
			Token:          common.Address{},
			Decimals:       0,
			UserAllocation: big.NewInt(0),
			UserNetFlow:    big.NewInt(0),
			NodeAllocation: big.NewInt(0),
			NodeNetFlow:    big.NewInt(0),
		}
	}

	// Determine intent based on last transition
	intent := TransitionToIntent(state.Transition)

	signingData, err := signingDataArgs.Pack(
		state.Version,
		intent,
		metadata,
		homeLedger,
		nonHomeLedger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack signing data: %w", err)
	}

	// Step 2: Pack message = abi.encode(channelId, signingData)
	// This matches Solidity: Utils.pack(channelId, signingData) = abi.encode(channelId, signingData)
	// where signingData is dynamic bytes
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	outerArgs := abi.Arguments{
		{Type: bytes32Type}, // channelId
		{Type: bytesType},   // signingData (dynamic bytes)
	}

	packed, err := outerArgs.Pack(
		channelID,
		signingData,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to pack outer message: %w", err)
	}

	return packed, nil
}
