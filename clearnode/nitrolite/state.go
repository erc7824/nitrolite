package nitrolite

import (
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

func PackState(channelID common.Hash, state State) ([]byte, error) {
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

	packed, err := args.Pack(channelID, state.Intent, state.Version, state.Data, state.Allocations)
	if err != nil {
		return nil, err
	}
	return packed, nil
}
