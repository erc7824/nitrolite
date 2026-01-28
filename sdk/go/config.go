// Package sdk provides a simple Go SDK for interacting with Clearnode instances.
package sdk

import (
	"log"
	"os"
	"time"
)

// Config holds the configuration options for the Clearnode client.
type Config struct {
	// URL is the WebSocket URL of the clearnode server
	URL string

	// HandshakeTimeout is the maximum time to wait for initial connection
	HandshakeTimeout time.Duration

	// PingInterval is the interval between keep-alive pings
	PingInterval time.Duration

	// ErrorHandler is called when connection errors occur
	ErrorHandler func(error)

	// BlockchainRPCs maps blockchain IDs to their RPC endpoints
	// Used by SmartClient for on-chain operations
	BlockchainRPCs map[uint64]string
}

// Option is a functional option for configuring the Client.
type Option func(*Config)

// DefaultConfig returns the default configuration with sensible defaults.
var DefaultConfig = Config{
	HandshakeTimeout: 5 * time.Second,
	PingInterval:     5 * time.Second,
	ErrorHandler:     defaultErrorHandler,
}

// defaultErrorHandler logs errors to stderr.
func defaultErrorHandler(err error) {
	if err != nil {
		log.New(os.Stderr, "[clearnode] ", log.LstdFlags).Printf("connection error: %v", err)
	}
}

// WithHandshakeTimeout sets the maximum time to wait for initial connection.
func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *Config) {
		c.HandshakeTimeout = d
	}
}

// WithPingInterval sets the interval between keep-alive pings.
func WithPingInterval(d time.Duration) Option {
	return func(c *Config) {
		c.PingInterval = d
	}
}

// WithErrorHandler sets a custom error handler for connection errors.
// The handler is called when the connection encounters an error or is closed.
func WithErrorHandler(fn func(error)) Option {
	return func(c *Config) {
		c.ErrorHandler = fn
	}
}
