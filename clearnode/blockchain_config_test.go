package main

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
				DefaultContractAddresses: ContractAddressesConfig{
					Custody:        "0x0000000000000000000000000000000000000001",
					Adjudicator:    "0x0000000000000000000000000000000000000002",
					BalanceChecker: "0x0000000000000000000000000000000000000003",
				},
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody:     "0x1111111111111111111111111111111111111111",
							Adjudicator: "0x2222222222222222222222222222222222222222",
						},
						BlockStep: 10,
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
				assert.Equal(t, "0x1111111111111111111111111111111111111111", ethCfg.ContractAddresses.Custody)
				assert.Equal(t, "0x2222222222222222222222222222222222222222", ethCfg.ContractAddresses.Adjudicator)
				assert.Equal(t, "0x0000000000000000000000000000000000000003", ethCfg.ContractAddresses.BalanceChecker)
				assert.False(t, ethCfg.Disabled)
				assert.Equal(t, uint64(10), ethCfg.BlockStep)

				sepoliaCfg := blockchains[1]
				assert.Equal(t, "ethereum_sepolia", sepoliaCfg.Name)
				assert.Equal(t, uint32(11155111), sepoliaCfg.ID)
				assert.Equal(t, "0x0000000000000000000000000000000000000001", sepoliaCfg.ContractAddresses.Custody)
				assert.Equal(t, "0x0000000000000000000000000000000000000002", sepoliaCfg.ContractAddresses.Adjudicator)
				assert.Equal(t, "0x0000000000000000000000000000000000000003", sepoliaCfg.ContractAddresses.BalanceChecker)
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
				DefaultContractAddresses: ContractAddressesConfig{
					Custody:        "0x0000000000000000000000000000000000000001",
					Adjudicator:    "0x0000000000000000000000000000000000000002",
					BalanceChecker: "0x0000000000000000000000000000000000000003",
				},
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
				DefaultContractAddresses: ContractAddressesConfig{
					Custody:        "0x0000s00000000000000000000000000000000001",
					Adjudicator:    "0x0000s00000000000000000000000000000000002",
					BalanceChecker: "0x0000s00000000000000000000000000000000003",
				},
			},
			expectedErrorStr: "invalid default custody contract address '0x0000s00000000000000000000000000000000001'",
		},
		{
			name: "invalid default adjudicator address",
			cfg: BlockchainsConfig{
				DefaultContractAddresses: ContractAddressesConfig{
					Custody:        "0x0000000000000000000000000000000000000001",
					Adjudicator:    "0x0000s00000000000000000000000000000000002",
					BalanceChecker: "0x0000s00000000000000000000000000000000003",
				},
			},
			expectedErrorStr: "invalid default adjudicator contract address '0x0000s00000000000000000000000000000000002'",
		},
		{
			name: "invalid default balance checker address",
			cfg: BlockchainsConfig{
				DefaultContractAddresses: ContractAddressesConfig{
					Custody:        "",
					Adjudicator:    "",
					BalanceChecker: "0x0000s00000000000000000000000000000000003",
				},
			},
			expectedErrorStr: "invalid default balance checker contract address '0x0000s00000000000000000000000000000000003'",
		},
		{
			name: "missing custody address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:                1,
						Name:              "ethereum",
						ContractAddresses: ContractAddressesConfig{},
					},
				},
			},
			expectedErrorStr: "missing default and blockchain-specific custody contract address for blockchain 'ethereum'",
		},
		{
			name: "missing adjudicator address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody: "0x1111111111111111111111111111111111111111"},
					},
				},
			},
			expectedErrorStr: "missing default and blockchain-specific adjudicator contract address for blockchain 'ethereum'",
		},
		{
			name: "missing balance checker address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody:     "0x1111111111111111111111111111111111111111",
							Adjudicator: "0x2222222222222222222222222222222222222222",
						},
					},
				},
			},
			expectedErrorStr: "missing default and blockchain-specific balance checker contract address for blockchain 'ethereum'",
		},
		{
			name: "invalid custody address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody:        "0x0000s00000000000000000000000000000000001",
							Adjudicator:    "0x0000s00000000000000000000000000000000002",
							BalanceChecker: "0x0000s00000000000000000000000000000000003",
						},
					},
				},
			},
			expectedErrorStr: "invalid custody contract address '0x0000s00000000000000000000000000000000001' for blockchain 'ethereum'",
		},
		{
			name: "invalid adjudicator address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody:        "0x0000000000000000000000000000000000000001",
							Adjudicator:    "0x0000s00000000000000000000000000000000002",
							BalanceChecker: "0x0000s00000000000000000000000000000000003",
						},
					},
				},
			},
			expectedErrorStr: "invalid adjudicator contract address '0x0000s00000000000000000000000000000000002' for blockchain 'ethereum'",
		},
		{
			name: "invalid balance checker address",
			cfg: BlockchainsConfig{
				Blockchains: []BlockchainConfig{
					{
						ID:   1,
						Name: "ethereum",
						ContractAddresses: ContractAddressesConfig{
							Custody:        "0x0000000000000000000000000000000000000001",
							Adjudicator:    "0x0000000000000000000000000000000000000002",
							BalanceChecker: "0x0000s00000000000000000000000000000000003",
						},
					},
				},
			},
			expectedErrorStr: "invalid balance checker contract address '0x0000s00000000000000000000000000000000003' for blockchain 'ethereum'",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.verifyVariables()
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
