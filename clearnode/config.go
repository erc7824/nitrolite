package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// knownNetworks maps network name prefixes to their respective chain IDs.
// Each prefix is used to find corresponding environment variables:
// - {PREFIX}_INFURA_URL: The Infura endpoint URL for the network
// - {PREFIX}_CUSTODY_CONTRACT_ADDRESS: The custody contract address
var knownNetworks = map[string]uint32{
	"POLYGON":     137,
	"ETH_SEPOLIA": 11155111,
	"CELO":        42220,
	"BASE":        8453,
	"WORLD_CHAIN": 480,
	"ROOTSTOCK":   30,
	"FLOW":        747,
	"LOCALNET":    1337,
	"ETH_MAINNET": 1,
	"ANVIL":       31337,
}

// NetworkConfig represents configuration for a blockchain network
type NetworkConfig struct {
	Name                  string
	ChainID               uint32
	InfuraURL             string
	CustodyAddress        string
	AdjudicatorAddress    string
	BalanceCHeckerAddress string // TODO: add balance checker method into our smart contract
	BlockStep             uint64
}

// Config represents the overall application configuration
type Config struct {
	networks      map[uint32]*NetworkConfig
	privateKeyHex string
	dbConf        DatabaseConfig
	msgExpiryTime int // Time in seconds for message timestamp validation
}

// LoadConfig builds configuration from environment variables
func LoadConfig(logger Logger) (*Config, error) {
	logger = logger.NewSystem("config")

	var err error
	// Load environment variables
	if err = godotenv.Load(); err != nil {
		logger.Warn(".env file not found")
	}

	// Get database URL from environment variables
	var dbConf DatabaseConfig
	dbURL := os.Getenv("CLEARNODE_DATABASE_URL")

	// If DATABASE_URL is not empty, parse the connection string
	// Otherwise, read the envs in usual way
	if dbURL != "" {
		dbConf, err = ParseConnectionString(dbURL)
		if err != nil {
			logger.Error("failed to parse connection string", "err", err)
			return nil, err
		}
	} else {
		// Read db config
		if err := cleanenv.ReadEnv(&dbConf); err != nil {
			logger.Error("failed to read env", "err", err)
			return nil, err
		}
	}

	// Retrieve the private key.
	privateKeyHex := os.Getenv("BROKER_PRIVATE_KEY")
	if privateKeyHex == "" {
		logger.Fatal("BROKER_PRIVATE_KEY environment variable is required")
	}

	messageTimestampExpiry := 60
	if messageExpiry := os.Getenv("MSG_EXPIRY_TIME"); messageExpiry != "" {
		if parsed, err := strconv.Atoi(messageExpiry); err == nil && parsed > 0 {
			messageTimestampExpiry = parsed
		} else {
			logger.Warn("Invalid MSG_EXPIRY_TIME", "messageExpiry", messageExpiry)
		}
	}
	logger.Info("set message expiry time", "value", messageTimestampExpiry)

	config := Config{
		networks:      make(map[uint32]*NetworkConfig),
		privateKeyHex: privateKeyHex,
		dbConf:        dbConf,
		msgExpiryTime: messageTimestampExpiry,
	}

	// Process each network
	envs := os.Environ()
	for network, chainID := range knownNetworks {
		infuraURL := ""
		custodyAddress := ""
		adjudicatorAddress := ""
		balanceCheckerAddress := ""
		blockStep := uint64(10000) // Default block step for reconcile

		// Look for matching environment variables
		for _, env := range envs {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := parts[0]
			value := parts[1]

			if strings.HasPrefix(key, network+"_INFURA_URL") {
				infuraURL = value
			} else if strings.HasPrefix(key, network+"_CUSTODY_CONTRACT_ADDRESS") {
				custodyAddress = value
			} else if strings.HasPrefix(key, network+"_ADJUDICATOR_ADDRESS") {
				adjudicatorAddress = value
			} else if strings.HasPrefix(key, network+"_BALANCE_CHECKER_ADDRESS") {
				balanceCheckerAddress = value
			} else if strings.HasPrefix(key, network+"_BLOCK_STEP") {
				if step, err := strconv.ParseUint(value, 10, 64); err == nil && step > 0 {
					blockStep = step
				} else {
					logger.Warn("Invalid BLOCK_STEP value", "network", network, "value", value)
				}
			}
		}

		// Only add network if both required variables are present
		if infuraURL != "" && custodyAddress != "" && adjudicatorAddress != "" && balanceCheckerAddress != "" {
			networkLower := strings.ToLower(network)
			config.networks[chainID] = &NetworkConfig{
				Name:                  networkLower,
				ChainID:               chainID,
				InfuraURL:             infuraURL,
				CustodyAddress:        custodyAddress,
				AdjudicatorAddress:    adjudicatorAddress,
				BalanceCHeckerAddress: balanceCheckerAddress,
				BlockStep:             blockStep,
			}
		}
	}

	return &config, nil
}
