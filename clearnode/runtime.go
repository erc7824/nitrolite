package main

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/clearnode/store/memory"
	"github.com/erc7824/nitrolite/pkg/log"
	"github.com/erc7824/nitrolite/pkg/rpc"
	"github.com/erc7824/nitrolite/pkg/sign"
)

//go:embed config/migrations/*/*.sql
var embedMigrations embed.FS

var Version = "v1.0.0" // set at build time with -ldflags "-X main.Version=x.y.z"

type Backbone struct {
	NodeVersion                 string
	ChannelMinChallengeDuration uint32
	BlockchainRPCs              map[uint64]string

	DbStore     database.DatabaseStore
	MemoryStore memory.MemoryStore
	RpcNode     rpc.Node
	StateSigner sign.Signer
	TxSigner    sign.Signer
	Logger      log.Logger
}

type Config struct {
	Database                    database.DatabaseConfig
	ChannelMinChallengeDuration uint32 `yaml:"channel_min_challenge_duration" env:"CLEARNODE_CHANNEL_MIN_CHALLENGE_DURATION" env-default:"86400"` // 24 hours
	SignerKey                   string `yaml:"signer_key" env:"CLEARNODE_SIGNER_KEY,required"`
}

// InitBackbone initializes the backbone components of the application.
func InitBackbone() *Backbone {
	// ------------------------------------------------
	// Logger
	// ------------------------------------------------

	var loggerConf log.Config
	if err := cleanenv.ReadEnv(&loggerConf); err != nil {
		panic("failed to read logger config from env: " + err.Error())
	}
	logger := log.NewZapLogger(loggerConf)
	logger = logger.WithName("main")

	// ------------------------------------------------
	// (Preparation)
	// ------------------------------------------------

	configDirPath := os.Getenv("CLEARNODE_CONFIG_DIR_PATH")
	if configDirPath == "" {
		configDirPath = "."
	}

	configDotEnvPath := filepath.Join(configDirPath, ".env")
	logger.Info("loading .env file", "path", configDotEnvPath)
	if err := godotenv.Load(configDotEnvPath); err != nil {
		logger.Warn(".env file not found")
	}

	var conf Config
	if err := cleanenv.ReadEnv(&conf); err != nil {
		logger.Fatal("failed to read env", "err", err)
	}

	logger.Info("config loaded", "version", Version)

	// ------------------------------------------------
	// Database Store
	// ------------------------------------------------

	db, err := database.ConnectToDB(conf.Database, embedMigrations)
	if err != nil {
		logger.Fatal("failed to load database store", "error", err)
	}
	dbStore := database.NewDBStore(db)

	// ------------------------------------------------
	// Memory Store
	// ------------------------------------------------

	memoryStore, err := memory.NewMemoryStoreV1FromConfig(configDirPath)
	if err != nil {
		logger.Fatal("failed to load blockchains", "error", err)
	}

	// ------------------------------------------------
	// Signer
	// ------------------------------------------------

	stateSigner, err := sign.NewEthereumMsgSigner(conf.SignerKey)
	if err != nil {
		logger.Fatal("failed to initialise state signer", "error", err)
	}
	txSigner, err := sign.NewEthereumRawSigner(conf.SignerKey)
	if err != nil {
		logger.Fatal("failed to initialise tx signer", "error", err)
	}
	logger.Info("signer initialized", "address", stateSigner.PublicKey().Address())

	// ------------------------------------------------
	// RPC Node
	// ------------------------------------------------

	rpcNode, err := rpc.NewWebsocketNode(rpc.WebsocketNodeConfig{
		Logger: logger,
	})

	// ------------------------------------------------
	// Blockchain RPCs
	// ------------------------------------------------

	blockchains, err := memoryStore.GetBlockchains()
	if err != nil {
		logger.Fatal("failed to get blockchains", "error", err)
	}

	blockchainRPCs := make(map[uint64]string)
	for _, bc := range blockchains {
		envVarName := "CLEARNODE_BLOCKCHAIN_RPC_" + strings.ToUpper(bc.Name)
		rpcURL := os.Getenv(envVarName)
		if rpcURL == "" {
			logger.Fatal("blockchain RPC URL not set in env", "blockchainID", bc.ID, "env_var", envVarName)
		}

		// Test connection
		if err := checkChainId(rpcURL, bc.ID); err != nil {
			logger.Fatal("failed to verify blockchain RPC", "blockchainID", bc.ID, "error", err)
		}
		blockchainRPCs[bc.ID] = rpcURL
	}

	return &Backbone{
		NodeVersion:                 Version,
		ChannelMinChallengeDuration: conf.ChannelMinChallengeDuration,
		BlockchainRPCs:              blockchainRPCs,

		DbStore:     dbStore,
		MemoryStore: memoryStore,
		RpcNode:     rpcNode,
		StateSigner: stateSigner,
		TxSigner:    txSigner,
		Logger:      logger,
	}
}

// checkChainId connects to an RPC endpoint and verifies it returns the expected chain ID.
// This ensures the RPC URL points to the correct blockchain network.
// The function uses a 5-second timeout for the connection and chain ID query.
func checkChainId(blockchainRPC string, expectedChainID uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

	if chainID.Uint64() != expectedChainID {
		return fmt.Errorf("unexpected chain ID from blockchain RPC: got %d, want %d", chainID.Uint64(), expectedChainID)
	}

	return nil
}
