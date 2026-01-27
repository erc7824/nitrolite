package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockchainConfig_verifyVariables(t *testing.T) {
	tcs := []struct {
		name             string
		cfg              BlockchainsConfig
		expectedErrorStr string
		assertFunc       func(t *testing.T, blockchains []BlockchainConfig)
	}{
		{
			name: "valid config",
			cfg: BlockchainsConfig{
				DefaultContractAddress: "0x0000000000000000000000000000000000000001",
				Blockchains: []BlockchainConfig{
					{
						ID:              1,
						Name:            "ethereum",
						ContractAddress: "0x1111111111111111111111111111111111111111",
						BlockStep:       10,
					},
					{
						ID:   11155111,
						Name: "ethereum_sepolia",
					},
				},
			},
			expectedErrorStr: "",
			assertFunc: func(t *testing.T, blockchains []BlockchainConfig) {
				require.Len(t, blockchains, 2)

				ethCfg := blockchains[0]
				assert.Equal(t, "ethereum", ethCfg.Name)
				assert.Equal(t, uint32(1), ethCfg.ID)
				assert.Equal(t, "0x1111111111111111111111111111111111111111", ethCfg.ContractAddress)
				assert.False(t, ethCfg.Disabled)
				assert.Equal(t, uint64(10), ethCfg.BlockStep)

				sepoliaCfg := blockchains[1]
				assert.Equal(t, "ethereum_sepolia", sepoliaCfg.Name)
				assert.Equal(t, uint32(11155111), sepoliaCfg.ID)
				assert.Equal(t, "0x0000000000000000000000000000000000000001", sepoliaCfg.ContractAddress)
				assert.False(t, sepoliaCfg.Disabled)
				assert.Equal(t, defaultBlockStep, sepoliaCfg.BlockStep)
			},
		},
		{
			name: "invalid name 1",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						Name: "Invalid Name!",
						ID:   1,
					},
				},
			},
			expectedErrorStr: "invalid blockchain name 'Invalid Name!', should match snake_case format",
		},
		{
			name: "invalid name 2",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						Name: "_foo_",
						ID:   1,
					},
				},
			},
			expectedErrorStr: "invalid blockchain name '_foo_', should match snake_case format",
		},
		{
			name: "disabled blockchain",
			cfg: BlockchainsConfig{
				DefaultContractAddress: "0x0000000000000000000000000000000000000001",
				Blockchains: []BlockchainConfig{
					{
						ID:       1,
						Name:     "ethereum",
						Disabled: false,
					},
					{
						ID:       11155111,
						Name:     "_ethereum_sepolia_",
						Disabled: true,
					},
				},
			},
			expectedErrorStr: "",
			assertFunc: func(t *testing.T, blockchains []BlockchainConfig) {
				require.Len(t, blockchains, 2)

				ethCfg := blockchains[0]
				assert.Equal(t, "ethereum", ethCfg.Name)
				assert.Equal(t, uint32(1), ethCfg.ID)

				sepoliaCfg := blockchains[1]
				assert.Equal(t, "_ethereum_sepolia_", sepoliaCfg.Name)
				assert.Equal(t, uint32(11155111), sepoliaCfg.ID)
			},
		},
		{
			name: "invalid default custody address",
			cfg: BlockchainsConfig{
				DefaultContractAddress: "0x0000s00000000000000000000000000000000001",
			},
			expectedErrorStr: "invalid default contract address '0x0000s00000000000000000000000000000000001'",
		},
		{
			name: "missing custody address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:              1,
						Name:            "ethereum",
						ContractAddress: "",
					},
				},
			},
			expectedErrorStr: "missing default and blockchain-specific custody contract address for blockchain 'ethereum'",
		},
		{
			name: "invalid custody address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:              1,
						Name:            "ethereum",
						ContractAddress: "0x0000s00000000000000000000000000000000001",
					},
				},
			},
			expectedErrorStr: "invalid custody contract address '0x0000s00000000000000000000000000000000001' for blockchain 'ethereum'",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyBlockchainsConfig(&tc.cfg)
			if tc.expectedErrorStr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErrorStr, err.Error())
				return
			}

			require.NoError(t, err)
			if tc.assertFunc != nil {
				tc.assertFunc(t, tc.cfg.Blockchains)
			}
		})
	}
}
