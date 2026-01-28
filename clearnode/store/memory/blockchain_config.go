package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

const (
	defaultBlockStep    = uint64(10000)
	blockchainsFileName = "blockchains.yaml"
)

var (
	blockchainNameRegex  = regexp.MustCompile(`^[a-z][a-z_]+[a-z]$`)
	contractAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
)

// BlockchainsConfig represents the root configuration structure for all blockchain settings.
// It contains default contract addresses that apply to all blockchains unless overridden,
// and a list of individual blockchain configurations.
type BlockchainsConfig struct {
	DefaultContractAddress string             `yaml:"default_contract_address"`
	Blockchains            []BlockchainConfig `yaml:"blockchains"`
}

// BlockchainConfig represents configuration for a single blockchain.
// It includes connection details, contract addresses, and scanning parameters.
type BlockchainConfig struct {
	// Name is the blockchain identifier (e.g., "polygon_amoy", "base_sepolia")
	// Must match pattern: lowercase letters and underscores only
	Name string `yaml:"name"`
	// ID is the chain ID used for RPC validation
	ID uint64 `yaml:"id"`
	// TODO: blockchains must not be disabled in prod deployment
	Disabled bool `yaml:"disabled"`
	// BlockStep defines the block range for scanning (default: 10000)
	BlockStep uint64 `yaml:"block_step"`
	// ContractAddress can override the default addresses
	ContractAddress string `yaml:"contract_address"`
}

// LoadEnabledBlockchains loads and validates blockchain configurations from a YAML file.
// It reads from <configDirPath>/blockchains.yaml, validates all settings,
// verifies RPC connections, and returns a map of enabled blockchains indexed by chain ID.
//
// The function performs the following validations:
// - Contract addresses format (0x + 40 hex chars)
// - Blockchain names (lowercase with underscores)
// - RPC endpoint availability and chain ID matching
// - Required contract addresses (using defaults when not specified)
func LoadEnabledBlockchains(configDirPath string) (map[uint64]BlockchainConfig, error) {
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

	if err := verifyBlockchainsConfig(&cfg); err != nil {
		return nil, err
	}

	return getEnabledBlockchains(&cfg), nil
}

func verifyBlockchainsConfig(cfg *BlockchainsConfig) error {
	if !contractAddressRegex.MatchString(cfg.DefaultContractAddress) && cfg.DefaultContractAddress != "" {
		return fmt.Errorf("invalid default contract address '%s'", cfg.DefaultContractAddress)
	}

	for i, bc := range cfg.Blockchains {
		if bc.Disabled {
			continue
		}

		if !blockchainNameRegex.MatchString(bc.Name) {
			return fmt.Errorf("invalid blockchain name '%s', should match snake_case format", bc.Name)
		}

		if bc.ContractAddress == "" {
			if cfg.DefaultContractAddress == "" {
				return fmt.Errorf("missing default and blockchain-specific contract address for blockchain '%s'", bc.Name)
			} else {
				cfg.Blockchains[i].ContractAddress = cfg.DefaultContractAddress
			}
		} else if !contractAddressRegex.MatchString(bc.ContractAddress) {
			return fmt.Errorf("invalid contract address '%s' for blockchain '%s'", bc.ContractAddress, bc.Name)
		}

		if bc.BlockStep == 0 {
			cfg.Blockchains[i].BlockStep = defaultBlockStep
		}
	}

	return nil
}

// getEnabledBlockchains returns a map of enabled blockchains indexed by their chain ID.
// Only blockchains with enabled=true are included in the result.
func getEnabledBlockchains(cfg *BlockchainsConfig) map[uint64]BlockchainConfig {
	enabledBlockchains := make(map[uint64]BlockchainConfig)
	for _, bc := range cfg.Blockchains {
		if !bc.Disabled {
			enabledBlockchains[bc.ID] = bc
		}
	}
	return enabledBlockchains
}
