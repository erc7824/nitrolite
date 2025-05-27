package nitrolite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GetChannelID returns the keccak256 hash of the ABI-encoded channel data.
// The encoding packs the two participants, the adjudicator, the challenge, and the nonce
// as static types (addresses padded to 32 bytes, and uint64 values in a 32-byte big-endian form).
func GetChannelID(ch Channel, chainID uint32) (common.Hash, error) {
	// ABI-encode the structure manually in the same order as Solidity
	participantsT, _ := abi.NewType("address[]", "", nil)
	adjudicatorT, _ := abi.NewType("address", "", nil)
	challengeT, _ := abi.NewType("uint64", "", nil)
	nonceT, _ := abi.NewType("uint64", "", nil)
	chainIdT, _ := abi.NewType("uint256", "", nil)
	arguments := abi.Arguments{
		{
			Type: participantsT,
		},
		{
			Type: adjudicatorT,
		},
		{
			Type: challengeT,
		},
		{
			Type: nonceT,
		},
		{
			Type: chainIdT,
		},
	}

	// Convert uint64 to *big.Int for ABI encoding
	chainIDCasted := new(big.Int).SetUint64(uint64(chainID))

	encoded, err := arguments.Pack(ch.Participants, ch.Adjudicator, ch.Challenge, ch.Nonce, chainIDCasted)
	if err != nil {
		return [32]byte{}, err
	}

	// Hash the encoded bytes with Keccak-256
	return crypto.Keccak256Hash(encoded), nil
}
