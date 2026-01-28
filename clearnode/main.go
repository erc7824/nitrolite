package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/erc7824/nitrolite/clearnode/api"
	"github.com/erc7824/nitrolite/clearnode/event_handlers"
	"github.com/erc7824/nitrolite/clearnode/store/database"
	"github.com/erc7824/nitrolite/pkg/blockchain/evm"
)

func main() {
	bb := InitBackbone()
	logger := bb.Logger
	ctx := context.Background()

	api.NewRPCRouter(bb.NodeVersion, bb.ChannelMinChallengeDuration,
		bb.RpcNode, bb.Signer, bb.DbStore, bb.MemoryStore, bb.Logger)

	rpcListenAddr := ":7824"
	rpcListenEndpoint := "/ws"
	rpcMux := http.NewServeMux()
	rpcMux.HandleFunc(rpcListenEndpoint, bb.RpcNode.ServeHTTP)

	rpcServer := &http.Server{
		Addr:    rpcListenAddr,
		Handler: rpcMux,
	}

	blockchains, err := bb.MemoryStore.GetBlockchains()
	if err != nil {
		logger.Fatal("failed to get blockchains from memory store", "error", err)
	}

	wrapInTx := func(handler func(database.DatabaseStore) error) error {
		return bb.DbStore.ExecuteInTransaction(handler)
	}
	useEHV1StoreInTx := func(h event_handlers.StoreTxHandler) error {
		return wrapInTx(func(s database.DatabaseStore) error { return h(s) })
	}

	eventHandlerService := event_handlers.NewEventHandlerService(useEHV1StoreInTx, logger)

	for _, b := range blockchains {
		rpcURL, ok := bb.BlockchainRPCs[b.ID]
		if !ok {
			logger.Fatal("no RPC URL configured for blockchain", "blockchainID", b.ID)
		}

		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			logger.Fatal("failed to connect to EVM Node")
		}
		reactor := evm.NewReactor(b.ID, eventHandlerService, bb.DbStore.StoreContractEvent)
		l := evm.NewListener(common.HexToAddress(b.ContractAddress), client, b.ID, b.BlockStep, logger, reactor.HandleEvent, bb.DbStore.GetLatestEvent)
		if err := l.Listen(ctx); err != nil {
			logger.Fatal("failed to start EVM listener")
		}

		blockchainClient, err := evm.NewClient(common.HexToAddress(b.ContractAddress), client, bb.Signer, b.ID, bb.MemoryStore)
		if err != nil {
			logger.Fatal("failed to create EVM client")
		}

		worker := NewBlockchainWorker(blockchainClient, bb.DbStore, logger)
		go worker.Start(ctx)
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

	go func() {
		logger.Info("prometheus metrics available", "listenAddr", metricsListenAddr, "endpoint", metricsEndpoint)
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
