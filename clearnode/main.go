package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erc7824/nitrolite/clearnode/api"
	"github.com/erc7824/nitrolite/clearnode/custody"
	"github.com/erc7824/nitrolite/clearnode/metrics/prometheus"
	"github.com/erc7824/nitrolite/clearnode/pkg/log"
	"github.com/erc7824/nitrolite/clearnode/pkg/sign"
	"github.com/erc7824/nitrolite/clearnode/store/db"
	"gorm.io/gorm"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed config/migrations/*/*.sql
var embedMigrations embed.FS

func main() {
	cfg := log.Config{
		Format: "json",
		Level:  log.LevelDebug,
	}

	logger := log.NewZapLogger(cfg)
	if len(os.Args) > 1 {
		// If a CLI command is provided, run it and exit
		runCli(logger, os.Args[1])
		return
	}

	config, err := LoadConfig(logger)
	if err != nil {
		logger.Fatal("failed to load configuration", "error", err)
	}

	database, err := db.ConnectToDB(config.dbConf, embedMigrations)
	if err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}

	err = db.LoadSessionKeyCache(database)
	if err != nil {
		logger.Fatal("Failed to load session key cache", "error", err)
	}

	signer, err := sign.NewEthereumSigner(config.privateKeyHex)
	if err != nil {
		logger.Fatal("failed to initialise signer", "error", err)
	}
	logger.Info("broker signer initialized", "address", signer.PublicKey().Address())

	rpcStore := db.NewRPCStore(database)

	// Initialize Prometheus metrics
	metrics := prometheus.NewMetrics()
	// Map to store custody clients for later reference
	custodyClients := make(map[uint32]*custody.Custody)

	rpcNode := api.NewRPCNode(signer, logger)
	wsNotifier := api.NewWSNotifier(rpcNode.Notify, logger)

	api.NewRPCRouter(rpcNode, config, signer, database, metrics, rpcStore, wsNotifier, logger)

	rpcListenAddr := ":8000"
	rpcListenEndpoint := "/ws"
	rpcMux := http.NewServeMux()
	rpcMux.HandleFunc(rpcListenEndpoint, rpcNode.HandleConnection)

	rpcServer := &http.Server{
		Addr:    rpcListenAddr,
		Handler: rpcMux,
	}

	for chainID, blockchain := range config.blockchains {
		client, err := custody.NewCustody(signer, database, wsNotifier, blockchain, &config.assets, logger)
		if err != nil {
			logger.Fatal("failed to initialize blockchain client", "chainID", chainID, "error", err)
			continue
		}
		custodyClients[chainID] = client
		go client.ListenEvents(context.Background())
	}

	// Start blockchain action worker for all custody clients
	// TODO: This can be moved to a separate worker process in the future for better scalability
	if len(custodyClients) > 0 {
		custodyClients := make(map[uint32]custody.CustodyInterface, len(custodyClients))
		for chainID, client := range custodyClients {
			custodyClients[chainID] = client
		}
		worker := NewBlockchainWorker(database, custodyClients, logger)
		go worker.Start(context.Background())
	}

	metricsListenAddr := ":4242"
	metricsEndpoint := "/metrics"
	// Set up a separate mux for metrics
	metricsMux := http.NewServeMux()
	metricsMux.Handle(metricsEndpoint, promhttp.Handler())

	// Start metrics server on a separate port
	metricsServer := &http.Server{
		Addr:    metricsListenAddr,
		Handler: metricsMux,
	}

	// Start metrics monitoring
	go RecordMetricsPeriodically(database, custodyClients, metrics, logger)

	go func() {
		logger.Info("Prometheus metrics available", "listenAddr", metricsListenAddr, "endpoint", metricsEndpoint)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failure", "error", err)
		}
	}()

	// Start the main HTTP server.
	go func() {
		logger.Info("RPC server available", "listenAddr", rpcListenAddr, "endpoint", rpcListenEndpoint)
		if err := rpcServer.ListenAndServe(); err != nil {
			logger.Fatal("RPC server failure", "error", err)
		}
	}()

	// Wait for shutdown signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down")

	// Shutdown metrics server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shut down metrics server", "error", err)
	}

	// Shutdown RPC server
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rpcServer.Shutdown(ctx); err != nil {
		logger.Error("failed to shut down RPC server", "error", err)
	}

	logger.Info("shutdown complete")
}

func runCli(logger log.Logger, name string) {
	switch name {
	case "reconcile":
		runReconcileCli(logger)
	case "export-transactions":
		runExportTransactionsCli(logger)
	default:
		logger.Fatal("Unknown CLI command", "name", name)
	}
}

func RecordMetricsPeriodically(db *gorm.DB, custodyClients map[uint32]*custody.Custody, m *prometheus.Metrics, logger log.Logger) {
	logger = logger.WithName("metrics")
	dbTicker := time.NewTicker(15 * time.Second)
	defer dbTicker.Stop()

	balanceTicker := time.NewTicker(30 * time.Second)
	defer balanceTicker.Stop()
	for {
		select {
		case <-dbTicker.C:
			m.UpdateChannelMetrics(db)
			m.UpdateAppSessionMetrics(db)
		case <-balanceTicker.C:
			ctx := context.Background()
			ctx = log.SetContextLogger(ctx, logger)

			// Update metrics for each custody client
			for _, custodyClient := range custodyClients {
				custodyClient.UpdateBalanceMetrics(ctx, m)
			}
		}
	}
}
