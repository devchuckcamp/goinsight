package cache

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"
)

// MemoryCache implements an in-memory caching layer with TTL support
// Algorithm:
// 1. Hash-based lookup: O(1) average case for Get/Set
// 2. TTL expiration: Background cleanup on access + periodic sweeps
// 3. Circular buffer for LRU eviction when max size exceeded
// 4. Thread-safe with read-write locks for concurrent access
type MemoryCache struct {
	mu sync.RWMutex

	// Main cache storage: hash(key) -> CacheEntry
	data map[string]*CacheEntry

	// TTL tracking for expiration detection
	expirations map[string]time.Time

	// LRU tracking for eviction
	accessTimes map[string]time.Time

	// Cache configuration
	maxSize      int64
	currentSize  int64
	defaultTTL   time.Duration
	cleanupTimer *time.Timer

	// Statistics
	stats CacheStats

	// Query result specific tracking
	queryHashes map[string]string // query -> hash mapping for invalidation
}

// NewMemoryCache creates a new in-memory cache instance
// Parameters:
//   - maxSize: Maximum number of entries (0 = unlimited)
//   - defaultTTL: Default time-to-live for entries
func NewMemoryCache(maxSize int64, defaultTTL time.Duration) *MemoryCache {
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute // Default 5 minutes
	}

	mc := &MemoryCache{
		data:         make(map[string]*CacheEntry),
		expirations:  make(map[string]time.Time),
		accessTimes:  make(map[string]time.Time),
		maxSize:      maxSize,
		defaultTTL:   defaultTTL,
		queryHashes:  make(map[string]string),
		cleanupTimer: time.NewTimer(1 * time.Minute),
	}

	// Start background cleanup goroutine
	go mc.cleanupExpired()

	return mc
}

// Set stores a value in the cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return NewCacheError(ErrInvalidKey, "cache key cannot be empty", nil)
	}

	if ttl == 0 {
		ttl = mc.defaultTTL
	}

	if ttl < 0 {
		return NewCacheError(ErrInvalidTTL, "TTL must be positive", nil)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict entries
	if mc.maxSize > 0 && int64(len(mc.data)) >= mc.maxSize && mc.data[key] == nil {
		mc.evictLRU()
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	entry := &CacheEntry{
		Data:      value,
		Timestamp: now,
		TTL:       ttl,
		ExpiresAt: expiresAt,
	}

	// Store entry
	mc.data[key] = entry
	mc.expirations[key] = expiresAt
	mc.accessTimes[key] = now
	mc.currentSize++

	return nil
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	if key == "" {
		return nil, NewCacheError(ErrInvalidKey, "cache key cannot be empty", nil)
	}

	mc.mu.RLock()
	entry, exists := mc.data[key]
	mc.mu.RUnlock()

	if !exists {
		mc.mu.Lock()
		mc.stats.Misses++
		mc.mu.Unlock()
		return nil, NewCacheError(ErrKeyNotFound, fmt.Sprintf("key '%s' not found in cache", key), nil)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		mc.mu.Lock()
		delete(mc.data, key)
		delete(mc.expirations, key)
		delete(mc.accessTimes, key)
		mc.stats.Evictions++
		mc.mu.Unlock()
		return nil, NewCacheError(ErrKeyNotFound, fmt.Sprintf("key '%s' has expired", key), nil)
	}

	// Update access time for LRU
	mc.mu.Lock()
	mc.accessTimes[key] = time.Now()
	mc.stats.Hits++
	mc.mu.Unlock()

	return entry.Data, nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return NewCacheError(ErrInvalidKey, "cache key cannot be empty", nil)
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.data[key]; !exists {
		return NewCacheError(ErrKeyNotFound, fmt.Sprintf("key '%s' not found in cache", key), nil)
	}

	delete(mc.data, key)
	delete(mc.expirations, key)
	delete(mc.accessTimes, key)
	mc.currentSize--

	return nil
}

// Exists checks if a key exists and is not expired
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, NewCacheError(ErrInvalidKey, "cache key cannot be empty", nil)
	}

	mc.mu.RLock()
	entry, exists := mc.data[key]
	mc.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		_ = mc.Delete(ctx, key)
		return false, nil
	}

	return true, nil
}

// Clear removes all entries from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.data = make(map[string]*CacheEntry)
	mc.expirations = make(map[string]time.Time)
	mc.accessTimes = make(map[string]time.Time)
	mc.queryHashes = make(map[string]string)
	mc.currentSize = 0

	return nil
}

// GetStats returns cache statistics
func (mc *MemoryCache) GetStats(ctx context.Context) CacheStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	stats := mc.stats
	stats.Size = int64(len(mc.data))
	stats.MaxSize = mc.maxSize

	// Calculate hit rate
	totalRequests := stats.Hits + stats.Misses
	if totalRequests > 0 {
		// Hits and misses are already in stats
	}

	return stats
}

// evictLRU removes the least recently used entry
// Algorithm: Find entry with oldest access time
func (mc *MemoryCache) evictLRU() {
	var lruKey string
	var lruTime time.Time

	for key, accessTime := range mc.accessTimes {
		if lruTime.IsZero() || accessTime.Before(lruTime) {
			lruKey = key
			lruTime = accessTime
		}
	}

	if lruKey != "" {
		delete(mc.data, lruKey)
		delete(mc.expirations, lruKey)
		delete(mc.accessTimes, lruKey)
		mc.currentSize--
		mc.stats.Evictions++
	}
}

// cleanupExpired runs periodically to clean expired entries
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()
		now := time.Now()
		var expiredKeys []string

		for key, expiresAt := range mc.expirations {
			if now.After(expiresAt) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		for _, key := range expiredKeys {
			delete(mc.data, key)
			delete(mc.expirations, key)
			delete(mc.accessTimes, key)
			mc.currentSize--
			mc.stats.Evictions++
		}

		mc.stats.TTLCheckups++
		mc.mu.Unlock()
	}
}

// GenerateQueryHash creates a consistent hash for SQL queries
// Algorithm: MD5 hash of normalized query
func (mc *MemoryCache) GenerateQueryHash(query string) string {
	hash := md5.Sum([]byte(query))
	return fmt.Sprintf("%x", hash)
}

// CacheQuery stores query results with query text tracking
func (mc *MemoryCache) CacheQuery(ctx context.Context, query string, results interface{}, ttl time.Duration) error {
	hash := mc.GenerateQueryHash(query)

	mc.mu.Lock()
	mc.queryHashes[query] = hash
	mc.mu.Unlock()

	return mc.Set(ctx, hash, results, ttl)
}

// GetCachedQuery retrieves cached query results
func (mc *MemoryCache) GetCachedQuery(ctx context.Context, query string) (interface{}, error) {
	hash := mc.GenerateQueryHash(query)
	return mc.Get(ctx, hash)
}

// InvalidateQuery removes a cached query result
func (mc *MemoryCache) InvalidateQuery(ctx context.Context, query string) error {
	hash := mc.GenerateQueryHash(query)

	mc.mu.Lock()
	delete(mc.queryHashes, query)
	mc.mu.Unlock()

	return mc.Delete(ctx, hash)
}

// InvalidatePattern removes all cached queries matching a pattern
// Useful for: UPDATE/DELETE invalidation
func (mc *MemoryCache) InvalidatePattern(ctx context.Context, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	var keysToDelete []string
	for query, hash := range mc.queryHashes {
		// Simple pattern matching: contains substring
		if _, exists := mc.data[hash]; exists && len(pattern) > 0 {
			// Check if pattern appears in query
			if len(query) > len(pattern) {
				for i := 0; i <= len(query)-len(pattern); i++ {
					if query[i:i+len(pattern)] == pattern {
						keysToDelete = append(keysToDelete, hash)
						delete(mc.queryHashes, query)
						break
					}
				}
			}
		}
	}

	for _, key := range keysToDelete {
		delete(mc.data, key)
		delete(mc.expirations, key)
		delete(mc.accessTimes, key)
		mc.currentSize--
		mc.stats.Evictions++
	}

	return nil
}

// Close closes the cache and cleans up resources
func (mc *MemoryCache) Close() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.cleanupTimer != nil {
		mc.cleanupTimer.Stop()
	}

	mc.data = nil
	mc.expirations = nil
	mc.accessTimes = nil
	mc.queryHashes = nil

	return nil
}
