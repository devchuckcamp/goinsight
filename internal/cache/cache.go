package cache

import (
	"context"
	"time"
)

// CacheEntry represents a cached value with metadata
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Cache defines the interface for caching implementations
type Cache interface {
	// Set stores a value in the cache with a TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all entries from the cache
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) CacheStats

	// Close closes the cache and cleans up resources
	Close() error
}

// CacheStats holds cache performance metrics
type CacheStats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int64
	MaxSize     int64
	TTLCheckups int64
}

// CacheError represents cache-related errors
type CacheError struct {
	Code    string
	Message string
	Err     error
}

const (
	// Error codes
	ErrKeyNotFound = "key_not_found"
	ErrCacheFull   = "cache_full"
	ErrInvalidTTL  = "invalid_ttl"
	ErrInvalidKey  = "invalid_key"
)

// NewCacheError creates a new CacheError
func NewCacheError(code, message string, err error) *CacheError {
	return &CacheError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Error implements the error interface
func (e *CacheError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
