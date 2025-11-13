package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Mode string

const (
	ModeProduction Mode = "production"
	ModeTest       Mode = "test"
)

const (
	configDirPathEnv     = "CLEARNODE_CONFIG_DIR_PATH"
	defaultConfigDirPath = "."
	defaultMessageExpiry = 60 // in seconds
)

// Config represents the overall application configuration
type Config struct {
	mode          Mode
	blockchains   map[uint32]BlockchainConfig
	assets        AssetsConfig
	privateKeyHex string
	dbConf        DatabaseConfig
	msgExpiryTime int // Time in seconds for message timestamp validation
}

// LoadConfig builds configuration from environment variables
func LoadConfig(logger Logger) (*Config, error) {
	logger = logger.NewSystem("config")

	configDirPath := os.Getenv(configDirPathEnv)
	if configDirPath == "" {
		configDirPath = defaultConfigDirPath
	}

	// Load .env files
	configDotEnvPath := filepath.Join(configDirPath, ".env")
	logger.Info("loading .env file", "path", configDotEnvPath)
	if err := godotenv.Load(configDotEnvPath); err != nil {
		logger.Warn(".env file not found")
	}

	mode := Mode(os.Getenv("CLEARNODE_MODE"))
	if mode == "" {
		mode = ModeProduction
	} else if mode != ModeProduction && mode != ModeTest {
		logger.Fatal("invalid CLEARNODE_MODE value", "value", mode)
	}
	logger.Info("set mode", "value", mode)

	// Get database URL from environment variables
	var dbConf DatabaseConfig
	dbURL := os.Getenv("CLEARNODE_DATABASE_URL")

	// If DATABASE_URL is not empty, parse the connection string
	// Otherwise, read the envs in usual way
	if dbURL != "" {
		var err error
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

	messageTimestampExpiry := defaultMessageExpiry
	if messageExpiry := os.Getenv("MSG_EXPIRY_TIME"); messageExpiry != "" {
		if parsed, err := strconv.Atoi(messageExpiry); err == nil && parsed > 0 {
			messageTimestampExpiry = parsed
		} else {
			logger.Warn("Invalid MSG_EXPIRY_TIME", "messageExpiry", messageExpiry)
		}
	}
	logger.Info("set message expiry time", "value", messageTimestampExpiry)

	blockchains, err := LoadBlockchains(configDirPath)
	if err != nil {
		logger.Fatal("failed to load blockchains", "error", err)
	}

	assets, err := LoadAssets(configDirPath)
	if err != nil {
		logger.Fatal("failed to load assets", "error", err)
	}

	config := Config{
		mode:          mode,
		blockchains:   blockchains,
		assets:        assets,
		privateKeyHex: privateKeyHex,
		dbConf:        dbConf,
		msgExpiryTime: messageTimestampExpiry,
	}

	return &config, nil
}
