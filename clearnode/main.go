package main

import (
	"context"
	"embed"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed config/migrations/*/*.sql
var embedMigrations embed.FS

func main() {
	logger := NewLoggerIPFS("root")
	if len(os.Args) > 1 {
		// If a CLI command is provided, run it and exit
		runCli(logger, os.Args[1])
		return
	}

	config, err := LoadConfig(logger)
	if err != nil {
		logger.Fatal("failed to load configuration", "error", err)
	}

	db, err := ConnectToDB(config.dbConf)
	if err != nil {
		logger.Fatal("Failed to setup database", "error", err)
	}

	err = loadWalletCache(db)
	if err != nil {
		logger.Fatal("Failed to load wallet cache", "error", err)
	}

	signer, err := NewSigner(config.privateKeyHex)
	if err != nil {
		logger.Fatal("failed to initialise signer", "error", err)
	}
	logger.Info("broker signer initialized", "address", signer.GetAddress().Hex())

	rpcStore := NewRPCStore(db)

	// Initialize Prometheus metrics
	metrics := NewMetrics()
	// Map to store custody clients for later reference
	custodyClients := make(map[string]*Custody)

	authManager, err := NewAuthManager(signer.GetPrivateKey())
	if err != nil {
		logger.Fatal("failed to initialize auth manager", "error", err)
	}

	rpcNode := NewRPCNode(signer, logger)
	wsNotifier := NewWSNotifier(rpcNode.Notify, logger)
	appSessionService := NewAppSessionService(db, wsNotifier)
	channelService := NewChannelService(db, config.networks, signer)

	NewRPCRouter(rpcNode, config, signer, appSessionService, channelService, db, authManager, metrics, rpcStore, wsNotifier, logger)

	rpcListenAddr := ":8000"
	rpcListenEndpoint := "/ws"
	rpcMux := http.NewServeMux()
	rpcMux.HandleFunc(rpcListenEndpoint, rpcNode.HandleConnection)

	rpcServer := &http.Server{
		Addr:    rpcListenAddr,
		Handler: rpcMux,
	}

	for name, network := range config.networks {
		client, err := NewCustody(signer, db, wsNotifier, network.InfuraURL, network.CustodyAddress, network.AdjudicatorAddress, network.BalanceCHeckerAddress, network.ChainID, network.BlockStep, logger)
		if err != nil {
			logger.Warn("failed to initialize blockchain client", "network", name, "error", err)
			continue
		}
		custodyClients[name] = client
		go client.ListenEvents(context.Background())
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

	// Start metrcis monitoring
	go metrics.RecordMetricsPeriodically(db, custodyClients, logger)

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
		logger.Error("failed to shut down metrics server", "error", err)
	}

	logger.Info("shutdown complete")
}

func runCli(logger Logger, name string) {
	switch name {
	case "reconcile":
		runReconcileCli(logger)
	case "export-transactions":
		runExportTransactionsCli(logger)
	default:
		logger.Fatal("Unknown CLI command", "name", name)
	}
}
