package core

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPackState(t *testing.T) {
	t.Run("pack_state_with_transitions", func(t *testing.T) {
		// Setup mock asset store
		assetStore := newMockAssetStore()
		assetStore.AddToken(42, "0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320", 8)
		assetStore.AddToken(4242, "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", 14)

		channelID := "0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b"

		decimalFromString := func(s string) decimal.Decimal {
			d, err := decimal.NewFromString(s)
			if err != nil {
				t.Fatalf("failed to parse decimal from string %s: %v", s, err)
			}
			return d
		}
		// Create a state with transitions (metadata will be derived from transitions)
		state := State{
			Version:       24,
			Asset:         "test", // won't be used
			HomeChannelID: &channelID,
			Transition:    *NewTransition(TransitionTypeEscrowWithdraw, "tx2", "account2", decimal.NewFromInt(-50)),

			HomeLedger: Ledger{
				BlockchainID: 42,
				TokenAddress: "0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320",
				UserBalance:  decimalFromString("3"),
				UserNetFlow:  decimalFromString("2.00000001"),
				NodeBalance:  decimalFromString("0"),
				NodeNetFlow:  decimalFromString("-0.99999999"),
			},
			EscrowLedger: &Ledger{
				BlockchainID: 4242,
				TokenAddress: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
				UserBalance:  decimalFromString("3"),
				UserNetFlow:  decimalFromString("2.00000001"),
				NodeBalance:  decimalFromString("0"),
				NodeNetFlow:  decimalFromString("-0.99999999"),
			},
		}

		packer := NewStatePackerV1(assetStore)
		packed, err := packer.PackState(state)
		assert.NoError(t, err)
		assert.NotNil(t, packed)

		expectedPackedState := "0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000002200000000000000000000000000000000000000000000000000000000000000018000000000000000000000000000000000000000000000000000000000000000713d86a5d0df614471807009f2507fd08007c21de21cf2c8c1d3a12582e065c6a000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000001092000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000110d9316ec0000000000000000000000000000000000000000000000000000000b5e62103c2400000000000000000000000000000000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffa50cef950240"
		packedHex := hexutil.Encode(packed)
		assert.Equal(t, expectedPackedState, packedHex, "Packed state should match expected value")
	})

	t.Run("pack_state_without_escrow", func(t *testing.T) {
		// Setup mock asset store
		assetStore := newMockAssetStore()
		assetStore.AddToken(42, "0x90b7E285ab6cf4e3A2487669dba3E339dB8a3320", 8)
		assetStore.AddToken(4242, "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", 14)

		channelID := "0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b"

		decimalFromString := func(s string) decimal.Decimal {
			d, err := decimal.NewFromString(s)
			if err != nil {
				t.Fatalf("failed to parse decimal from string %s: %v", s, err)
			}
			return d
		}
		// Create a state with transitions (metadata will be derived from transitions)
		state := State{
			Version:       24,
			Asset:         "test",
			Transition:    *NewTransition(TransitionTypeHomeDeposit, "tx123", "account456", decimal.NewFromInt(1000)),
			HomeChannelID: &channelID,
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
		packed, err := packer.PackState(state)
		assert.NoError(t, err)
		assert.NotNil(t, packed)
		expectedPackedState := "0x3e9dd25a843e3a234c278c6f3fab3983949e2404b276cacb3c47ada06e00f74b0000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000022000000000000000000000000000000000000000000000000000000000000000180000000000000000000000000000000000000000000000000000000000000002756774c79382894c7bdc8e9fe73285e1650c10c820b05a8bba7d0aace4adb92a000000000000000000000000000000000000000000000000000000000000002a00000000000000000000000090b7e285ab6cf4e3a2487669dba3e339db8a332000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000011e1a300000000000000000000000000000000000000000000000000000000000bebc2010000000000000000000000000000000000000000000000000000000000000000fffffffffffffffffffffffffffffffffffffffffffffffffffffffffa0a1f010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
		packedHex := hexutil.Encode(packed)
		assert.Equal(t, expectedPackedState, packedHex, "Packed state should match expected value")
	})
}
