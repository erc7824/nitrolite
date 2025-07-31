package nitrolite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Intent uint8

const (
	IntentOPERATE    Intent = 0
	IntentINITIALIZE Intent = 1
	IntentRESIZE     Intent = 2
	IntentFINALIZE   Intent = 3
)

// EncodeState encodes channel state into a byte array using channelID, intent, version, state data, and allocations.
func EncodeState(channelID common.Hash, intent Intent, version *big.Int, stateData []byte, allocations []Allocation) ([]byte, error) {
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

	packed, err := args.Pack(channelID, uint8(intent), version, stateData, allocations)
	if err != nil {
		return nil, err
	}
	return packed, nil
}

// COMMON STATE DEFINITION

// Assumption: let's abstract from state channel framework, and try do define a common state without being limited by state-channel limitations.

type CommonState struct {
	State         UnsignedCommonState `json:"state"`
	OwnerSig      Signature           `json:"owner_sig"`
	ValidatorSigs []Signature         `json:"validator_sigs"`
}

// Adjudicator manages a registry of approved validator nodes. Each node has a signature weight.
// Adjudicator verifies that validatorSigs achieve a signature quorum threshold.
// Number of validators and quorum threshold can be adjusted in realtime by adjudicator contract.

type UnsignedCommonState struct {
	Version     *big.Int     `json:"version"`      // Common state version
	StateData   []byte       `json:"state_data"`   // Common state data
	ChainStates []ChainState `json:"chain_states"` // User allocation on each chain
}

type ChainState struct {
	ChainID     uint32            `json:"chain_id"`
	Allocations []TokenAllocation `json:"allocations"`
}

type TokenAllocation struct {
	TokenAddress common.Address `json:"token"`
	RawAmount    *big.Int       `json:"amount"`
}

// User deposits money on smart contract. Smart contract account is a big state channel with Yellow Network.
// Yellow network users are participants of this big state channel. State changes are validated by set of validators.
// Funds are in the same pool, delivering high liquidity and funds efficiency.

// User has 2 options to withdraw funds:
// 1. Withdraw with a set of validator signatures (cooperative withdraw).
// 	- Withdraw(CurrentCommonState, WithdrawReqValidatorSigs)
// 2. Call Withdraw without providing signatures (if network is down), which will initialize a withdraw. (like challenge or unlock period in yellow vault). Validator can use this window to submit a newer state.
// 	- Withdraw(CurrentCommonState)

// CLEARNODE COMMON STATE INTEGRATION

// To perform a transfer on Yellow Network, user must:

// Call GetTransferState([]AssetAmount) on a Clearnode
// Receive a new UnsignedCommonState,
// sign it and submit it to a Clearnode by calling ExecuteTransfer(Destination, UnsignedCommonState, Signature)
