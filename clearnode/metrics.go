package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

// Metrics contains all Prometheus metrics for the application
type Metrics struct {
	// WebSocket connection metrics
	ConnectedClients prometheus.Gauge
	ConnectionsTotal prometheus.Counter
	MessageReceived  prometheus.Counter
	MessageSent      prometheus.Counter

	// Authentication metrics
	AuthRequests       prometheus.Counter
	AuthAttemptsTotal  *prometheus.CounterVec
	AuthAttempsSuccess *prometheus.CounterVec
	AuthAttempsFail    *prometheus.CounterVec

	// Transfer metrics
	TransferAttemptsTotal   prometheus.Counter
	TransferAttemptsSuccess prometheus.Counter
	TransferAttemptsFail    prometheus.Counter

	// Channel & app sessions metrics
	Channels    *prometheus.GaugeVec
	AppSessions *prometheus.GaugeVec

	// RPC method metrics
	RPCRequests *prometheus.CounterVec

	// Smart contract metrics
	BrokerBalanceAvailable *prometheus.GaugeVec
	BrokerChannelCount     *prometheus.GaugeVec

	// Broker wallet metrics
	BrokerWalletBalance *prometheus.GaugeVec
}

// NewMetrics initializes and registers Prometheus metrics
func NewMetrics() *Metrics {
	return NewMetricsWithRegistry(nil)
}

// NewMetricsWithRegistry initializes and registers Prometheus metrics with a custom registry
func NewMetricsWithRegistry(registry prometheus.Registerer) *Metrics {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}
	factory := promauto.With(registry)

	metrics := &Metrics{
		ConnectedClients: factory.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_connected_clients",
			Help: "The current number of connected clients",
		}),
		ConnectionsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_connections_total",
			Help: "The total number of WebSocket connections made since server start",
		}),
		MessageReceived: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_ws_messages_received_total",
			Help: "The total number of WebSocket messages received",
		}),
		MessageSent: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_ws_messages_sent_total",
			Help: "The total number of WebSocket messages sent",
		}),
		AuthRequests: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_auth_requests_total",
			Help: "The total number of auth_requests (get challenge code)",
		}),
		AuthAttemptsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "clearnet_auth_attempts_total",
				Help: "The total number of authentication attempts",
			},
			[]string{"auth_method"},
		),
		AuthAttempsSuccess: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "clearnet_auth_attempts_success",
				Help: "The total number of successfull authentication attempts",
			},
			[]string{"auth_method"},
		),
		AuthAttempsFail: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "clearnet_auth_attempts_fail",
				Help: "The total number of failed authentication attempts",
			},
			[]string{"auth_method"},
		),
		TransferAttemptsTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_transfer_attempts_total",
			Help: "The total number of transfer attempts",
		}),
		TransferAttemptsSuccess: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_transfer_attempts_success",
			Help: "The total number of successful transfer attempts",
		}),
		TransferAttemptsFail: factory.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_transfer_attempts_fail",
			Help: "The total number of failed transfer attempts",
		}),
		Channels: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "clearnet_channels",
			Help: "The number of channels",
		},
			[]string{"status"},
		),
		AppSessions: factory.NewGaugeVec(prometheus.GaugeOpts{
			Name: "clearnet_app_sessions",
			Help: "The number of application sessions",
		},
			[]string{"status"},
		),
		RPCRequests: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "clearnet_rpc_requests_total",
				Help: "The total number of RPC requests by method",
			},
			[]string{"method", "status"},
		),
		BrokerBalanceAvailable: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clearnet_broker_balance_available",
				Help: "Available balance of the broker on the custody contract",
			},
			[]string{"blockchainID", "token", "asset"},
		),
		BrokerChannelCount: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clearnet_broker_channel_count",
				Help: "Number of channels for the broker on the custody contract",
			},
			[]string{"blockchainID"},
		),
		BrokerWalletBalance: factory.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clearnet_broker_wallet_balance",
				Help: "Broker wallet balance",
			},
			[]string{"blockchainID", "token", "asset"},
		),
	}

	return metrics
}

func (m *Metrics) RecordMetricsPeriodically(db *gorm.DB, custodyClients map[uint32]*Custody, logger Logger) {
	logger = logger.NewSystem("metrics")
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
			ctx = SetContextLogger(ctx, logger)

			// Update metrics for each custody client
			for _, custodyClient := range custodyClients {
				custodyClient.UpdateBalanceMetrics(ctx, m)
			}
		}
	}
}

func (m *Metrics) UpdateChannelMetrics(db *gorm.DB) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount

	err := db.Model(&Channel{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error
	if err != nil {
		// handle error (e.g., log it)
		return
	}

	// Reset the gauge vector before setting new values
	m.Channels.Reset()

	for _, row := range results {
		m.Channels.WithLabelValues(row.Status).Set(float64(row.Count))
	}
}

// UpdateAppSessionMetrics updates the application session metrics from the database
func (m *Metrics) UpdateAppSessionMetrics(db *gorm.DB) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount

	err := db.Model(&AppSession{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error
	if err != nil {
		// Log or handle error
		return
	}

	// Stage values to avoid partial update issues
	tmp := make(map[string]float64)
	for _, row := range results {
		tmp[row.Status] = float64(row.Count)
	}

	// Now safely update the GaugeVec
	m.AppSessions.Reset()
	for status, count := range tmp {
		m.AppSessions.WithLabelValues(status).Set(count)
	}
}
