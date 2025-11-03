package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v3"
)

const (
	checkChainIdCallTimeout = 5 * time.Second
	defaultBlockStep        = uint64(10000)
	blockchainsFileName     = "blockchains.yaml"
)

var (
	blockchainNameRegex  = regexp.MustCompile(`^[a-z][a-z_]+[a-z]$`)
	contractAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// BlockchainsConfig represents the root configuration structure for all blockchain settings.
// It contains default contract addresses that apply to all blockchains unless overridden,
// and a list of individual blockchain configurations.
type BlockchainsConfig struct {
	DefaultContractAddresses ContractAddressesConfig `yaml:"default_contract_addresses"`
	Blockchains              []BlockchainConfig      `yaml:"blockchains"`
}

// BlockchainConfig represents configuration for a single blockchain.
// It includes connection details, contract addresses, and scanning parameters.
type BlockchainConfig struct {
	// Name is the blockchain identifier (e.g., "polygon_amoy", "base_sepolia")
	// Must match pattern: lowercase letters and underscores only
	Name string `yaml:"name"`
	// ID is the chain ID used for RPC validation
	ID uint32 `yaml:"id"`
	// Disabled determines if this blockchain should be connected
	Disabled bool `yaml:"disabled"`
	// BlockchainRPC is populated from environment variable <NAME>_BLOCKCHAIN_RPC
	BlockchainRPC string
	// BlockStep defines the block range for scanning (default: 10000)
	BlockStep uint64 `yaml:"block_step"`
	// ContractAddresses can override the default addresses
	ContractAddresses ContractAddressesConfig `yaml:"contract_addresses"`
}

// ContractAddressesConfig holds Ethereum contract addresses for blockchain operations.
// All addresses must be valid Ethereum addresses (0x followed by 40 hex characters).
type ContractAddressesConfig struct {
	Custody        string `yaml:"custody"`
	Adjudicator    string `yaml:"adjudicator"`
	BalanceChecker string `yaml:"balance_checker"`
}

// LoadBlockchains loads and validates blockchain configurations from a YAML file.
// It reads from <configDirPath>/blockchains.yaml, validates all settings,
// verifies RPC connections, and returns a map of enabled blockchains indexed by chain ID.
//
// The function performs the following validations:
// - Contract addresses format (0x + 40 hex chars)
// - Blockchain names (lowercase with underscores)
// - RPC endpoint availability and chain ID matching
// - Required contract addresses (using defaults when not specified)
func LoadBlockchains(configDirPath string) (map[uint32]BlockchainConfig, error) {
	blockchainsPath := filepath.Join(configDirPath, blockchainsFileName)
	f, err := os.Open(blockchainsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg BlockchainsConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.verifyVariables(); err != nil {
		return nil, err
	}

	if err := cfg.verifyRPCs(); err != nil {
		return nil, err
	}

	enabledBlockchains := cfg.getEnabled()
	return enabledBlockchains, nil
}

// verifyVariables validates the configuration structure and applies defaults.
// It ensures all contract addresses are valid, applies default addresses where needed,
// and sets default block step values. This method modifies the config in place.
func (cfg *BlockchainsConfig) verifyVariables() error {
	defaults := cfg.DefaultContractAddresses
	if !contractAddressRegex.MatchString(defaults.Custody) && defaults.Custody != "" {
		return fmt.Errorf("invalid default custody contract address '%s'", defaults.Custody)
	}
	if !contractAddressRegex.MatchString(defaults.Adjudicator) && defaults.Adjudicator != "" {
		return fmt.Errorf("invalid default adjudicator contract address '%s'", defaults.Adjudicator)
	}
	if !contractAddressRegex.MatchString(defaults.BalanceChecker) && defaults.BalanceChecker != "" {
		return fmt.Errorf("invalid default balance checker contract address '%s'", defaults.BalanceChecker)
	}

	for i, bc := range cfg.Blockchains {
		if bc.Disabled {
			continue
		}

		if !blockchainNameRegex.MatchString(bc.Name) {
			return fmt.Errorf("invalid blockchain name '%s', should match snake_case format", bc.Name)
		}

		if bc.ContractAddresses.Custody == "" {
			if defaults.Custody == "" {
				return fmt.Errorf("missing default and blockchain-specific custody contract address for blockchain '%s'", bc.Name)
			} else {
				cfg.Blockchains[i].ContractAddresses.Custody = defaults.Custody
			}
		} else if !contractAddressRegex.MatchString(bc.ContractAddresses.Custody) {
			return fmt.Errorf("invalid custody contract address '%s' for blockchain '%s'", bc.ContractAddresses.Custody, bc.Name)
		}

		if bc.ContractAddresses.Adjudicator == "" {
			if defaults.Adjudicator == "" {
				return fmt.Errorf("missing default and blockchain-specific adjudicator contract address for blockchain '%s'", bc.Name)
			} else {
				cfg.Blockchains[i].ContractAddresses.Adjudicator = defaults.Adjudicator
			}
		} else if !contractAddressRegex.MatchString(bc.ContractAddresses.Adjudicator) {
			return fmt.Errorf("invalid adjudicator contract address '%s' for blockchain '%s'", bc.ContractAddresses.Adjudicator, bc.Name)
		}

		if bc.ContractAddresses.BalanceChecker == "" {
			if defaults.BalanceChecker == "" {
				return fmt.Errorf("missing default and blockchain-specific balance checker contract address for blockchain '%s'", bc.Name)
			} else {
				cfg.Blockchains[i].ContractAddresses.BalanceChecker = defaults.BalanceChecker
			}
		} else if !contractAddressRegex.MatchString(bc.ContractAddresses.BalanceChecker) {
			return fmt.Errorf("invalid balance checker contract address '%s' for blockchain '%s'", bc.ContractAddresses.BalanceChecker, bc.Name)
		}

		if bc.BlockStep == 0 {
			cfg.Blockchains[i].BlockStep = defaultBlockStep
		}
	}

	return nil
}

// verifyRPCs validates RPC endpoints for all enabled blockchains.
// It reads RPC URLs from environment variables following the pattern:
// <BLOCKCHAIN_NAME_UPPERCASE>_BLOCKCHAIN_RPC
// and verifies that each endpoint returns the expected chain ID.
func (cfg *BlockchainsConfig) verifyRPCs() error {
	for i, bc := range cfg.Blockchains {
		if bc.Disabled {
			continue
		}

		blockchainRPC := os.Getenv(fmt.Sprintf("%s_BLOCKCHAIN_RPC", strings.ToUpper(bc.Name)))
		if blockchainRPC == "" {
			return fmt.Errorf("missing blockchain RPC for blockchain '%s'", bc.Name)
		}

		if err := checkChainId(blockchainRPC, bc.ID); err != nil {
			return fmt.Errorf("blockchain '%s' ChainID check failed: %w", bc.Name, err)
		}

		cfg.Blockchains[i].BlockchainRPC = blockchainRPC
	}

	return nil
}

// getEnabled returns a map of enabled blockchains indexed by their chain ID.
// Only blockchains with enabled=true are included in the result.
func (cfg *BlockchainsConfig) getEnabled() map[uint32]BlockchainConfig {
	enabledBlockchains := make(map[uint32]BlockchainConfig)
	for _, bc := range cfg.Blockchains {
		if !bc.Disabled {
			enabledBlockchains[bc.ID] = bc
		}
	}
	return enabledBlockchains
}

// checkChainId connects to an RPC endpoint and verifies it returns the expected chain ID.
// This ensures the RPC URL points to the correct blockchain network.
// The function uses a 5-second timeout for the connection and chain ID query.
func checkChainId(blockchainRPC string, expectedChainID uint32) error {
	ctx, cancel := context.WithTimeout(context.Background(), checkChainIdCallTimeout)
	defer cancel()

	client, err := ethclient.DialContext(ctx, blockchainRPC)
	if err != nil {
		return fmt.Errorf("failed to connect to blockchain RPC: %w", err)
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID from blockchain RPC: %w", err)
	}

	if uint32(chainID.Uint64()) != expectedChainID {
		return fmt.Errorf("unexpected chain ID from blockchain RPC: got %d, want %d", chainID.Uint64(), expectedChainID)
	}

	return nil
}
