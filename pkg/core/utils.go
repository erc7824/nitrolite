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
	uint64Type, _  = abi.NewType("uint64", "", nil)
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

	// Convert challenge to uint32 and nonce to big.Int for ABI packing
	challenge32 := uint32(challenge)
	nonceBI := new(big.Int).SetUint64(nonce)

	packed, err := args.Pack(nodeAddr, userAddr, tokenAddr, challenge32, nonceBI)
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

// PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
func PackState(state State) ([]byte, error) {
	// Pack the state using the cross-chain state structure
	ledgerType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
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

	// Convert user balance and netflow to big.Int
	userBalanceBI := state.HomeLedger.UserBalance.BigInt()
	userNetFlowBI := state.HomeLedger.UserNetFlow.BigInt()
	nodeBalanceBI := state.HomeLedger.NodeBalance.BigInt()
	nodeNetFlowBI := state.HomeLedger.NodeNetFlow.BigInt()

	homeState := struct {
		ChainId            uint64
		Token              common.Address
		ParticipantBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
		NodeBalance struct {
			Allocation *big.Int
			NetFlow    *big.Int
		}
	}{
		ChainId: uint64(state.HomeLedger.BlockchainID),
		Token:   common.HexToAddress(state.HomeLedger.TokenAddress),
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
		escrowUserBalanceBI := state.EscrowLedger.UserBalance.BigInt()
		escrowUserNetFlowBI := state.EscrowLedger.UserNetFlow.BigInt()
		escrowNodeBalanceBI := state.EscrowLedger.NodeBalance.BigInt()
		escrowNodeNetFlowBI := state.EscrowLedger.NodeNetFlow.BigInt()

		nonHomeState = struct {
			ChainId            uint64
			Token              common.Address
			ParticipantBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
			NodeBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
		}{
			ChainId: uint64(state.EscrowLedger.BlockchainID),
			Token:   common.HexToAddress(state.EscrowLedger.TokenAddress),
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
			ParticipantBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
			NodeBalance struct {
				Allocation *big.Int
				NetFlow    *big.Int
			}
		}{
			ChainId: 0,
			Token:   common.Address{},
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
