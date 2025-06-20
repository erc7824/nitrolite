package main

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/erc7824/nitrolite/clearnode/nitrolite"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestMarshalEventEquivalence(t *testing.T) {
	event := nitrolite.CustodyResized{
		ChannelId:        [32]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		DeltaAllocations: []*big.Int{big.NewInt(1), big.NewInt(2)},
		Raw: types.Log{
			Address: common.HexToAddress("0x1122334455667788990011223344556677889900"),
			Topics:  []common.Hash{common.HexToHash("0xaabbcc")},
			Data:    []byte("some log data"),
		},
	}
	eventCopy := event
	eventCopy.Raw = types.Log{}
	expectedBytes, err := json.Marshal(eventCopy)
	require.NoError(t, err)

	// Test generic function
	actualBytes, err := MarshalEvent(event)
	require.NoError(t, err)
	require.True(t, bytes.Equal(expectedBytes, actualBytes), "The marshaled output from the generic function does not match the original.")
}
