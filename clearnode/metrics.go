package main

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"
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
	AuthRequests prometheus.Counter
	AuthSuccess  prometheus.Counter
	AuthFailure  prometheus.Counter

	// Channel metrics
	ChannelsTotal  prometheus.Gauge
	ChannelsOpen   prometheus.Gauge
	ChannelsClosed prometheus.Gauge

	// RPC method metrics
	RPCRequests *prometheus.CounterVec

	// Application metrics
	AppSessionsTotal prometheus.Gauge

	// Smart contract metrics
	BrokerBalanceAvailable *prometheus.GaugeVec
	BrokerChannelCount     *prometheus.GaugeVec
}

// NewMetrics initializes and registers Prometheus metrics
func NewMetrics() *Metrics {
	metrics := &Metrics{
		ConnectedClients: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_connected_clients",
			Help: "The current number of connected clients",
		}),
		ConnectionsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_connections_total",
			Help: "The total number of WebSocket connections made since server start",
		}),
		MessageReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_ws_messages_received_total",
			Help: "The total number of WebSocket messages received",
		}),
		MessageSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_ws_messages_sent_total",
			Help: "The total number of WebSocket messages sent",
		}),
		AuthRequests: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_auth_requests_total",
			Help: "The total number of authentication requests",
		}),
		AuthSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_auth_success_total",
			Help: "The total number of successful authentications",
		}),
		AuthFailure: promauto.NewCounter(prometheus.CounterOpts{
			Name: "clearnet_auth_failure_total",
			Help: "The total number of failed authentications",
		}),
		ChannelsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_channels_total",
			Help: "The total number of channels",
		}),
		ChannelsOpen: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_channels_open",
			Help: "The number of open channels",
		}),
		ChannelsClosed: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_channels_closed",
			Help: "The number of closed channels",
		}),
		RPCRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "clearnet_rpc_requests_total",
				Help: "The total number of RPC requests by method",
			},
			[]string{"method"},
		),
		AppSessionsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "clearnet_app_sessions_total",
			Help: "The total number of application sessions",
		}),
		BrokerBalanceAvailable: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clearnet_broker_balance_available",
				Help: "Available balance of the broker on the custody contract",
			},
			[]string{"network", "token"},
		),
		BrokerChannelCount: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clearnet_broker_channel_count",
				Help: "Number of channels for the broker on the custody contract",
			},
			[]string{"network", "token"},
		),
	}

	return metrics
}

func (m *Metrics) RecordMetricsPeriodically(db *gorm.DB, custodyClients map[string]*Custody) {
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
			// Refresh the list of tokens to monitor
			monitoredTokens := GetUniqueTokenAddresses(db)

			// Update metrics for each custody client
			for _, client := range custodyClients {
				client.UpdateBalanceMetrics(context.Background(), monitoredTokens, m)
			}
		}
	}
}

// UpdateChannelMetrics updates the channel metrics from the database
func (m *Metrics) UpdateChannelMetrics(db *gorm.DB) {
	var total, open, closed int64

	db.Model(&Channel{}).Count(&total)
	db.Model(&Channel{}).Where("status = ?", ChannelStatusOpen).Count(&open)
	db.Model(&Channel{}).Where("status = ?", ChannelStatusClosed).Count(&closed)

	m.ChannelsTotal.Set(float64(total))
	m.ChannelsOpen.Set(float64(open))
	m.ChannelsClosed.Set(float64(closed))
}

// UpdateAppSessionMetrics updates the application session metrics from the database
func (m *Metrics) UpdateAppSessionMetrics(db *gorm.DB) {
	var count int64
	db.Model(&AppSession{}).Count(&count)
	m.AppSessionsTotal.Set(float64(count))
}

// GetUniqueTokenAddresses returns a list of unique token addresses from the database
func GetUniqueTokenAddresses(db *gorm.DB) []common.Address {
	var tokens []string

	// Query unique token addresses from the channels table
	db.Model(&Channel{}).Distinct().Pluck("token", &tokens)

	// Convert to common.Address and remove empty strings
	addresses := make([]common.Address, 0, len(tokens))
	for _, tokenStr := range tokens {
		if tokenStr != "" {
			addresses = append(addresses, common.HexToAddress(tokenStr))
		}
	}

	return addresses
}
