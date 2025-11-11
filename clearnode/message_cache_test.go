package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMessageCache(t *testing.T) {
	t.Run("NewMessageCache", func(t *testing.T) {
		ttl := 60 * time.Second
		cache := NewMessageCache(ttl)

		require.NotNil(t, cache)
		require.Equal(t, ttl, cache.ttl)
		require.Equal(t, minCleanupInterval, cache.cleanupEvery)
		require.NotNil(t, cache.entries)
		require.Equal(t, 0, len(cache.entries))
	})

	t.Run("Add_And_Exists", func(t *testing.T) {
		cache := NewMessageCache(60 * time.Second)
		hash := "test-hash-123"

		// Initially should not exist
		require.False(t, cache.Exists(hash))

		// Add hash
		cache.Add(hash)

		// Should now exist
		require.True(t, cache.Exists(hash))

		// Check that multiple adds don't break anything
		cache.Add(hash)
		require.True(t, cache.Exists(hash))
	})

	t.Run("Expiry", func(t *testing.T) {
		ttl := 100 * time.Millisecond
		cache := NewMessageCache(ttl)
		hash := "expiring-hash"

		// Add hash
		cache.Add(hash)

		// Should exist immediately
		require.True(t, cache.Exists(hash))

		// Wait for expiry
		time.Sleep(150 * time.Millisecond)

		// Should not exist after expiry
		require.False(t, cache.Exists(hash))
	})

	t.Run("Remove", func(t *testing.T) {
		cache := NewMessageCache(60 * time.Second)
		hash := "removable-hash"

		// Add hash
		cache.Add(hash)

		require.True(t, cache.Exists(hash))

		// Remove hash
		cache.Remove(hash)

		// Should not exist after removal
		require.False(t, cache.Exists(hash))

		// Removing non-existent hash should not panic
		cache.Remove("non-existent-hash")
	})

	t.Run("CleanupAdaptive", func(t *testing.T) {
		ttl := 5 * time.Millisecond
		cache := NewMessageCache(ttl)

		// Add entries and let them expire
		for i := 0; i < 100; i++ {
			cache.Add(string(rune(i)))
		}

		// Wait for expiry
		time.Sleep(10 * time.Millisecond)

		// Get initial size (should have 100 expired entries)
		cache.mu.RLock()
		initialSize := len(cache.entries)
		cache.mu.RUnlock()

		require.Equal(t, 100, initialSize, "expected 100 entries before cleanup")

		// Trigger cleanup by adding new entries
		// With minCleanupInterval=10, cleanup should trigger after 10 adds
		for i := 0; i < minCleanupInterval+1; i++ {
			cache.Add("new-" + string(rune(i)))
		}

		// Check that cleanup happened
		cache.mu.RLock()
		finalSize := len(cache.entries)
		cache.mu.RUnlock()

		// Should have only the new entries (old 100 expired and cleaned up)
		require.LessOrEqual(t, finalSize, minCleanupInterval+1, "expected cleanup to reduce size")
	})

	t.Run("RecalculateCleanupInterval", func(t *testing.T) {
		cache := NewMessageCache(60 * time.Second)

		tests := []struct {
			cacheSize        int
			expectedInterval int
		}{
			{0, minCleanupInterval},     // Empty cache
			{50, minCleanupInterval},    // Small cache
			{100, minCleanupInterval},   // At minimum threshold
			{500, 50},                   // Medium cache: 500/10 = 50
			{1000, 100},                 // Large cache: 1000/10 = 100
			{5000, 500},                 // Larger cache: 5000/10 = 500
			{10000, maxCleanupInterval}, // At maximum threshold
			{20000, maxCleanupInterval}, // Beyond maximum
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("size_%d", tt.cacheSize), func(t *testing.T) {
				// Simulate cache size by adding entries
				cache.entries = make(map[string]int64, tt.cacheSize)
				for i := 0; i < tt.cacheSize; i++ {
					cache.entries[string(rune(i))] = time.Now().UnixMilli()
				}

				// Recalculate interval
				cache.recalculateCleanupInterval()

				require.Equal(t, tt.expectedInterval, cache.cleanupEvery,
					"for cache size %d", tt.cacheSize)
			})
		}
	})

	t.Run("Concurrency", func(t *testing.T) {
		cache := NewMessageCache(1 * time.Second)
		numGoroutines := 100
		numOpsPerGoroutine := 100

		var wg sync.WaitGroup
		wg.Add(numGoroutines * 3) // 3 types of operations

		// Concurrent adds
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOpsPerGoroutine; j++ {
					hash := string(rune(id*1000 + j))
					cache.Add(hash)
				}
			}(i)
		}

		// Concurrent exists checks
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOpsPerGoroutine; j++ {
					hash := string(rune(id*1000 + j))
					cache.Exists(hash)
				}
			}(i)
		}

		// Concurrent removes
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < numOpsPerGoroutine; j++ {
					hash := string(rune(id*1000 + j))
					cache.Remove(hash)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// If we got here without panicking or deadlocking, the test passes
	})

	t.Run("ExpiryDuringExists", func(t *testing.T) {
		ttl := 5 * time.Millisecond
		cache := NewMessageCache(ttl)
		hash := "test-hash"

		cache.Add(hash)

		// Hash exists immediately
		require.True(t, cache.Exists(hash))

		// Wait for expiry
		time.Sleep(10 * time.Millisecond)

		// Should return false even though entry is still in map
		require.False(t, cache.Exists(hash))

		// Verify the expired entry is still in the map (lazy deletion)
		cache.mu.RLock()
		_, stillInMap := cache.entries[hash]
		cache.mu.RUnlock()

		require.True(t, stillInMap, "expired entry should still be in map before cleanup")
	})

	t.Run("MultipleEntriesExpiry", func(t *testing.T) {
		ttl := 100 * time.Millisecond
		cache := NewMessageCache(ttl)

		// Add multiple entries at different times
		cache.Add("hash1")
		time.Sleep(30 * time.Millisecond)
		cache.Add("hash2")
		time.Sleep(30 * time.Millisecond)
		cache.Add("hash3")

		// All should exist
		require.True(t, cache.Exists("hash1"))
		require.True(t, cache.Exists("hash2"))
		require.True(t, cache.Exists("hash3"))

		// Wait for first to expire
		time.Sleep(60 * time.Millisecond)

		// hash1 should be expired, others still valid
		require.False(t, cache.Exists("hash1"))
		require.True(t, cache.Exists("hash2"))
		require.True(t, cache.Exists("hash3"))

		// Wait for all to expire
		time.Sleep(100 * time.Millisecond)

		// All should be expired
		require.False(t, cache.Exists("hash1"))
		require.False(t, cache.Exists("hash2"))
		require.False(t, cache.Exists("hash3"))
	})
}

func TestHashMessage(t *testing.T) {
	t.Run("Well formatted message", func(t *testing.T) {
		// Create test RPC message
		reqData := &RPCData{
			RequestID: 123,
			Method:    "transfer",
			Params:    []any{"param1", "param2"},
			Timestamp: 1234567890,
		}

		// Marshal to get rawBytes
		rawBytes, err := json.Marshal([]any{reqData.RequestID, reqData.Method, reqData.Params, reqData.Timestamp})
		require.NoError(t, err)
		reqData.rawBytes = rawBytes

		msg := &RPCMessage{
			Req: reqData,
		}

		// Generate hash
		hash1 := HashMessage(msg)

		// Check hash is not empty
		require.NotEmpty(t, hash1)

		// Check hash is hexadecimal string (64 chars for Keccak256)
		require.Equal(t, 64, len(hash1))

		// Same message should produce same hash
		hash2 := HashMessage(msg)
		require.Equal(t, hash1, hash2)

		// Different message should produce different hash
		reqData2 := &RPCData{
			RequestID: 456,
			Method:    "transfer",
			Params:    []any{"param1", "param2"},
			Timestamp: 1234567890,
		}
		rawBytes2, err := json.Marshal([]any{reqData2.RequestID, reqData2.Method, reqData2.Params, reqData2.Timestamp})
		require.NoError(t, err)
		reqData2.rawBytes = rawBytes2

		msg2 := &RPCMessage{
			Req: reqData2,
		}

		hash3 := HashMessage(msg2)
		require.NotEqual(t, hash1, hash3)
	})

	t.Run("Nil message", func(t *testing.T) {
		var msg *RPCMessage = nil
		hash := HashMessage(msg)
		require.Equal(t, "", hash)
	})

	t.Run("Nil RPCData", func(t *testing.T) {
		msg := &RPCMessage{
			Req: nil,
		}
		hash := HashMessage(msg)
		require.Equal(t, "", hash)
	})
}
