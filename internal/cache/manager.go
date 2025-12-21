package cache

import (
	"context"
	"time"
)

// CacheManager provides high-level cache operations for query results
// Handles caching strategy, TTL management, and invalidation
type CacheManager struct {
	cache      Cache
	enabled    bool
	defaultTTL time.Duration
	maxSize    int64
}

// NewCacheManager creates a new cache manager instance
// Parameters:
//   - enabled: Whether caching is active
//   - maxSize: Maximum number of entries in cache
//   - defaultTTL: Default time-to-live for cached entries
func NewCacheManager(enabled bool, maxSize int64, defaultTTL time.Duration) *CacheManager {
	if defaultTTL == 0 {
		defaultTTL = 5 * time.Minute
	}

	var cache Cache
	if enabled {
		cache = NewMemoryCache(maxSize, defaultTTL)
	}

	return &CacheManager{
		cache:      cache,
		enabled:    enabled,
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}
}

// IsCacheEnabled checks if caching is enabled
func (cm *CacheManager) IsCacheEnabled() bool {
	return cm.enabled && cm.cache != nil
}

// CacheQueryResult stores query results in cache
func (cm *CacheManager) CacheQueryResult(ctx context.Context, query string, results interface{}, ttl time.Duration) error {
	if !cm.IsCacheEnabled() {
		return nil
	}

	if ttl == 0 {
		ttl = cm.defaultTTL
	}

	if memCache, ok := cm.cache.(*MemoryCache); ok {
		return memCache.CacheQuery(ctx, query, results, ttl)
	}

	return cm.cache.Set(ctx, query, results, ttl)
}

// GetCachedQueryResult retrieves cached query results
// Returns (results, found, error)
func (cm *CacheManager) GetCachedQueryResult(ctx context.Context, query string) (interface{}, bool, error) {
	if !cm.IsCacheEnabled() {
		return nil, false, nil
	}

	if memCache, ok := cm.cache.(*MemoryCache); ok {
		results, err := memCache.GetCachedQuery(ctx, query)
		if err != nil {
			// Key not found is not an error for this use case
			if cacheErr, ok := err.(*CacheError); ok && cacheErr.Code == ErrKeyNotFound {
				return nil, false, nil
			}
			return nil, false, err
		}
		return results, true, nil
	}

	results, err := cm.cache.Get(ctx, query)
	if err != nil {
		if cacheErr, ok := err.(*CacheError); ok && cacheErr.Code == ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	return results, true, nil
}

// InvalidateQuery removes a specific cached query
func (cm *CacheManager) InvalidateQuery(ctx context.Context, query string) error {
	if !cm.IsCacheEnabled() {
		return nil
	}

	if memCache, ok := cm.cache.(*MemoryCache); ok {
		return memCache.InvalidateQuery(ctx, query)
	}

	return cm.cache.Delete(ctx, query)
}

// InvalidatePattern removes cached queries matching a pattern
// Useful for: Invalidating all queries involving a table after UPDATE/DELETE
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	if !cm.IsCacheEnabled() {
		return nil
	}

	if memCache, ok := cm.cache.(*MemoryCache); ok {
		return memCache.InvalidatePattern(ctx, pattern)
	}

	return nil
}

// ClearCache removes all cached entries
func (cm *CacheManager) ClearCache(ctx context.Context) error {
	if !cm.IsCacheEnabled() {
		return nil
	}

	return cm.cache.Clear(ctx)
}

// GetCacheStats returns current cache statistics
func (cm *CacheManager) GetCacheStats(ctx context.Context) CacheStats {
	if !cm.IsCacheEnabled() {
		return CacheStats{}
	}

	return cm.cache.GetStats(ctx)
}

// Close closes the cache manager
func (cm *CacheManager) Close() error {
	if cm.cache != nil {
		return cm.cache.Close()
	}
	return nil
}

// CacheConfig holds configuration for caching
type CacheConfig struct {
	Enabled    bool
	MaxSize    int64
	DefaultTTL time.Duration
}

// DefaultCacheConfig returns sensible default cache configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		Enabled:    true,
		MaxSize:    1000,         // Max 1000 cached queries
		DefaultTTL: 5 * time.Minute, // 5 minute default TTL
	}
}
