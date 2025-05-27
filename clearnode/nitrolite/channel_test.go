package nitrolite

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestGetChannelID(t *testing.T) {
	ch := Channel{
		Participants: []common.Address{common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"), common.HexToAddress("0x2B5AD5c4795c026514f8317c7a215E218DcCD6cF")},
		Adjudicator:  common.HexToAddress("0x2e234DAe75C793f67A35089C9d99245E1C58470b"),
		Challenge:    3600,
		Nonce:        1,
	}

	channelID, err := GetChannelID(ch, 31337)
	require.NoError(t, err)
	require.Equal(t, "0x55100d2206d2f305bf5c19cc2b4347f6d962308bb5211b7a1e88d674aa866ff4", common.Hash(channelID).Hex())
}
