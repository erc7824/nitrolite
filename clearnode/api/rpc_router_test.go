package api

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestRPCRouter(t *testing.T) (*RPCRouter, *gorm.DB, func()) {
	db, dbCleanup := SetupTestDB(t)

	// Use a test private key
	privateKeyHex := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	signer, err := NewSigner(privateKeyHex)
	require.NoError(t, err)

	logger := NewLoggerIPFS("root.test")

	node := NewRPCNode(signer, logger)
	wsNotifier := NewWSNotifier(node.Notify, logger)

	blockchains := map[uint32]BlockchainConfig{
		137: {
			Name:          "polygon",
			ID:            137,
			BlockchainRPC: "https://polygon-mainnet.infura.io/v3/test",
			ContractAddresses: ContractAddressesConfig{
				Custody:     "0xCustodyAddress",
				Adjudicator: "0xAdjudicatorAddress",
			},
		},
		42220: {
			Name:          "celo",
			ID:            42220,
			BlockchainRPC: "https://celo-mainnet.infura.io/v3/test",
			ContractAddresses: ContractAddressesConfig{
				Custody:     "0xCustodyAddress2",
				Adjudicator: "0xAdjudicatorAddress2",
			},
		},
	}

	config := &Config{blockchains: blockchains, assets: AssetsConfig{}, msgExpiryTime: 60}
	channelService := NewChannelService(db, blockchains, &config.assets, signer)

	// Create an instance of RPCRouter
	router := &RPCRouter{
		Node:              node,
		Config:            config,
		Signer:            signer,
		AppSessionService: NewAppSessionService(db, wsNotifier),
		ChannelService:    channelService,
		DB:                db,
		wsNotifier:        wsNotifier,
		MessageCache:      NewMessageCache(60 * time.Second),
		lg:                logger.WithName("rpc-router"),
		Metrics:           NewMetricsWithRegistry(prometheus.NewRegistry()),
	}

	return router, router.DB, func() {
		dbCleanup()
	}
}
