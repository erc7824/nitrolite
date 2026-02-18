package core

import (
	"errors"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackChallengeState(t *testing.T) {
	t.Parallel()
	// Setup mock asset store
	assetStore := newMockAssetStore()
	assetStore.AddToken(42, "0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320", 8)

	channelID := "0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b"

	decimalFromString := func(s string) decimal.Decimal {
		d, err := decimal.NewFromString(s)
		if err != nil {
			t.Fatalf("failed to parse decimal from string %s: %v", s, err)
		}
		return d
	}

	state := State{
		Version:       24,
		Asset:         "test",
		HomeChannelID: &channelID,
		Transition:    *NewTransition(TransitionTypeHomeDeposit, "tx123", "account456", decimal.NewFromInt(1000)),
		HomeLedger: Ledger{
			BlockchainID: 42,
			TokenAddress: "0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320",
			UserBalance:  decimalFromString("3"),
			UserNetFlow:  decimalFromString("2.00000001"),
			NodeBalance:  decimalFromString("0"),
			NodeNetFlow:  decimalFromString("-0.99999999"),
		},
		EscrowLedger: nil,
	}

	packer := NewStatePackerV1(assetStore)
	packed, err := packer.PackChallengeState(state)
	require.NoError(t, err)
	assert.NotNil(t, packed)
	packedState, err := packer.PackState(state)
	require.NoError(t, err)

	assert.NotEqual(t, packedState, packed, "PackChallengeState should output different bytes than PackState")
	t.Logf("Packed Challenge State: %s", hexutil.Encode(packed))
}

func TestPackState_Errors(t *testing.T) {
	t.Parallel()
	t.Run("missing_home_channel_id", func(t *testing.T) {
		t.Parallel()
		assetStore := newMockAssetStore()
		state := State{
			Version:       1,
			Asset:         "test",
			HomeChannelID: nil, // Intentionally nil
		}

		packer := NewStatePackerV1(assetStore)
		_, err := packer.PackState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "state.HomeChannelID is required")
	})

	t.Run("asset_store_error", func(t *testing.T) {
		t.Parallel()
		// Let's create a local failing mock
		failingStore := &failingAssetStore{}

		channelID := "0x123"
		state := State{
			Version:       1,
			HomeChannelID: &channelID,
			Transition:    *NewTransition(TransitionTypeHomeDeposit, "tx", "acc", decimal.Zero),
			HomeLedger: Ledger{
				BlockchainID: 1,
				TokenAddress: "0xToken",
			},
		}

		packer := NewStatePackerV1(failingStore)
		_, err := packer.PackState(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock error")
	})
}

type failingAssetStore struct{}

func (f *failingAssetStore) GetAssetDecimals(_ string) (uint8, error) {
	return 0, errors.New("mock error")
}

func (f *failingAssetStore) GetTokenDecimals(_ uint64, _ string) (uint8, error) {
	return 0, errors.New("mock error")
}

func (f *failingAssetStore) GetTokenAddress(_ string, _ uint64) (string, error) {
	return "", errors.New("mock error")
}

func (f *failingAssetStore) GetSuggestedBlockchainID(_ string) (uint64, error) {
	return 0, errors.New("mock error")
}

func (f *failingAssetStore) AssetExistsOnBlockchain(_ uint64, _ string) (bool, error) {
	return false, errors.New("mock error")
}
