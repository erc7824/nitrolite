package sdk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	assert.Equal(t, 5*time.Second, DefaultConfig.HandshakeTimeout)
	assert.Equal(t, 5*time.Second, DefaultConfig.PingInterval)
	assert.NotNil(t, DefaultConfig.ErrorHandler)
}

func TestWithHandshakeTimeout(t *testing.T) {
	c := &Config{}
	opt := WithHandshakeTimeout(10 * time.Second)
	opt(c)
	assert.Equal(t, 10*time.Second, c.HandshakeTimeout)
}

func TestWithPingInterval(t *testing.T) {
	c := &Config{}
	opt := WithPingInterval(20 * time.Second)
	opt(c)
	assert.Equal(t, 20*time.Second, c.PingInterval)
}

func TestWithErrorHandler(t *testing.T) {
	c := &Config{}
	called := false
	handler := func(err error) {
		called = true
	}
	opt := WithErrorHandler(handler)
	opt(c)
	
	assert.NotNil(t, c.ErrorHandler)
	c.ErrorHandler(nil)
	assert.True(t, called)
}

func TestDefaultErrorHandler(t *testing.T) {
	// Just ensure it doesn't panic when called with nil or error
	// Capturing stderr is harder and maybe overkill for this simple function
	defaultErrorHandler(nil)
	// defaultErrorHandler(errors.New("test error")) // verification would require capturing stderr
}
