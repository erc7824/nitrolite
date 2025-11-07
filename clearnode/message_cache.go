package main

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	cleanupTargetFraction = 10   // Target: cleanup when ~1/10th of cache size new entries added
	minCleanupInterval    = 10   // Minimum cleanup interval in operations
	maxCleanupInterval    = 1000 // Maximum cleanup interval in operations
)

// MessageCache provides a thread-safe cache for tracking recent RPC messages
// to prevent duplicate message processing within the message expiry window.
// Important implications:
// - If message addition stops, expired entries will remain until the next Add()
// - This is intentional - it avoids cleanup overhead during read-heavy periods
// - The cache will not grow unbounded because:
//   - Expired entries are treated as non-existent by Exists()
//   - Next Add() will eventually trigger cleanup
type MessageCache struct {
	entries        map[string]int64 // hash -> expiry timestamp (Unix ms)
	mu             sync.RWMutex
	ttl            time.Duration
	cleanupCounter int
	cleanupEvery   int // Dynamically calculated based on cache size
}

// NewMessageCache creates a new MessageCache instance with the specified TTL.
func NewMessageCache(ttl time.Duration) *MessageCache {
	return &MessageCache{
		entries:      make(map[string]int64),
		ttl:          ttl,
		cleanupEvery: minCleanupInterval, // Initial value, will be recalculated
	}
}

// Add adds a message hash to the cache with an expiry time of TTL from now.
// It also performs periodic cleanup of expired entries.
func (mc *MessageCache) Add(hash string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.entries[hash] = time.Now().Add(mc.ttl).UnixMilli()

	// TODO: improve this logic to avoid cleanup on every Add when under high load
	// Lazy cleanup every N operations
	mc.cleanupCounter++
	if mc.cleanupCounter >= mc.cleanupEvery {
		mc.cleanupExpiredLocked()
		mc.recalculateCleanupInterval()
		mc.cleanupCounter = 0
	}
}

// Exists checks if a message hash exists in the cache and has not expired.
// Returns true if the message is still valid (cached and not expired).
func (mc *MessageCache) Exists(hash string) bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	expiryTime, exists := mc.entries[hash]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().UnixMilli() > expiryTime {
		return false
	}

	return true
}

// Remove explicitly removes a message hash from the cache.
// This can be used if a message processing fails and should be allowed to retry immediately.
func (mc *MessageCache) Remove(hash string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.entries, hash)
}

// cleanupExpiredLocked removes all expired entries from the cache.
// The caller is responsible for acquiring mc.mu.Lock() before calling this function.
// This design avoids double-locking and allows the Add() method to hold the lock
// for both adding an entry and performing cleanup in a single critical section.
func (mc *MessageCache) cleanupExpiredLocked() {
	now := time.Now().UnixMilli()
	for hash, expiryTime := range mc.entries {
		if now > expiryTime {
			delete(mc.entries, hash)
		}
	}
}

// recalculateCleanupInterval adjusts the cleanup frequency based on cache size.
// The adaptive approach scales cleanup frequency with cache pressure, which is more efficient than constant cleanup.
// Target: cleanup when ~cleanupTargetFraction% new entries added since last cleanup.
// Bounded between min and max operations.
func (mc *MessageCache) recalculateCleanupInterval() {
	size := len(mc.entries)

	interval := size / cleanupTargetFraction

	// Apply bounds
	if interval < minCleanupInterval {
		mc.cleanupEvery = minCleanupInterval
	} else if interval > maxCleanupInterval {
		mc.cleanupEvery = maxCleanupInterval
	} else {
		mc.cleanupEvery = interval
	}
}

// HashMessage generates a unique hash for an RPC message using Keccak256.
// It hashes the raw JSON bytes of the request, which includes the method, params, and timestamp.
// Keccak256 is used because it's:
//   - Faster than SHA-256 (~300 MB/s vs ~5 MB/s)
//   - Already available in go-ethereum/crypto
//   - Provides sufficient collision resistance for cache deduplication
func HashMessage(msg *RPCMessage) string {
	if msg == nil || msg.Req == nil {
		return ""
	}

	hash := crypto.Keccak256(msg.Req.rawBytes)
	return hex.EncodeToString(hash)
}
