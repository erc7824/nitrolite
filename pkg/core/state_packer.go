package core

import (
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
func (p *StatePackerV1) PackState(state State) ([]byte, error) {
	// Pack the state using the cross-chain state structure
	ledgerType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "chainId", Type: "uint64"},
		{Name: "token", Type: "address"},
		{Name: "decimals", Type: "uint8"},
		{Name: "participantBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "allocation", Type: "uint256"},
			{Name: "netFlow", Type: "int256"},
		}},
		{Name: "nodeBalance", Type: "tuple", Components: []abi.ArgumentMarshaling{
			{Name: "allocation", Type: "uint256"},
			{Name: "netFlow", Type: "int256"},
		}},
	})
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: uint256Type}, // version
		{Type: uint64Type},  // homeChainId
		{Type: uint8Type},   // intent
		{Type: ledgerType},  // homeState
		{Type: ledgerType},  // nonHomeState
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

	homeState := struct {
		ChainId            uint64
		Token              common.Address
		Decimals           uint8
		ParticipantBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
		NodeBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
	}{
		ChainId:  uint64(state.HomeLedger.BlockchainID),
		Token:    common.HexToAddress(state.HomeLedger.TokenAddress),
		Decimals: homeDecimals,
		ParticipantBalance: struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}{
			Allocation: userBalanceBI,
			NetFlow:    userNetFlowBI,
		},
		NodeBalance: struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}{
			Allocation: nodeBalanceBI,
			NetFlow:    nodeNetFlowBI,
		},
	}

	// For nonHomeState, use escrow ledger if available, otherwise use zero values
	var nonHomeState struct {
		ChainId            uint64
		Token              common.Address
		Decimals           uint8
		ParticipantBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
		NodeBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
	}

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

		nonHomeState = struct {
			ChainId            uint64
			Token              common.Address
			Decimals           uint8
			ParticipantBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
			NodeBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
		}{
			ChainId:  uint64(state.EscrowLedger.BlockchainID),
			Token:    common.HexToAddress(state.EscrowLedger.TokenAddress),
			Decimals: escrowDecimals,
			ParticipantBalance: struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}{
				Allocation: escrowUserBalanceBI,
				NetFlow:    escrowUserNetFlowBI,
			},
			NodeBalance: struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}{
				Allocation: escrowNodeBalanceBI,
				NetFlow:    escrowNodeNetFlowBI,
			},
		}
	} else {
		nonHomeState = struct {
			ChainId            uint64
			Token              common.Address
			Decimals           uint8
			ParticipantBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
			NodeBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
		}{
			ChainId:  0,
			Token:    common.Address{},
			Decimals: 0,
			ParticipantBalance: struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}{
				Allocation: big.NewInt(0),
				NetFlow:    big.NewInt(0),
			},
			NodeBalance: struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}{
				Allocation: big.NewInt(0),
				NetFlow:    big.NewInt(0),
			},
		}
	}

	// Determine intent based on last transition
	intent := uint8(0) // default intent
	if lastTransition := state.GetLastTransition(); lastTransition != nil {
		// Map transition type to intent
		// This is a simplified mapping - adjust based on actual requirements
		switch lastTransition.Type {
		case TransitionTypeTransferSend, TransitionTypeTransferReceive:
			intent = 1 // operate intent
		case TransitionTypeHomeDeposit, TransitionTypeEscrowDeposit:
			intent = 2 // deposit intent
		case TransitionTypeHomeWithdrawal, TransitionTypeEscrowWithdraw:
			intent = 3 // withdraw intent
		}
	}

	packed, err := args.Pack(
		big.NewInt(int64(state.Version)),
		uint64(state.HomeLedger.BlockchainID),
		intent,
		homeState,
		nonHomeState,
	)
	if err != nil {
		return nil, err
	}
	return packed, nil
}
