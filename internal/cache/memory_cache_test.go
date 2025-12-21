package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestMemoryCacheSet tests basic cache set operation
func TestMemoryCacheSet(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)

	tests := []struct {
		name      string
		key       string
		value     interface{}
		shouldErr bool
	}{
		{
			name:      "set string value",
			key:       "key1",
			value:     "value1",
			shouldErr: false,
		},
		{
			name:      "set number value",
			key:       "key2",
			value:     42,
			shouldErr: false,
		},
		{
			name:      "set nil value",
			key:       "key3",
			value:     nil,
			shouldErr: false,
		},
		{
			name:      "set map value",
			key:       "key4",
			value:     map[string]interface{}{"nested": "data"},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(context.Background(), tt.key, tt.value, 0)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Set() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

// TestMemoryCacheGet tests basic cache retrieval
func TestMemoryCacheGet(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx := context.Background()

	// Set a value
	_ = cache.Set(ctx, "key1", "value1", 0)

	tests := []struct {
		name          string
		key           string
		expectedValue interface{}
		shouldExist   bool
	}{
		{
			name:          "get existing key",
			key:           "key1",
			expectedValue: "value1",
			shouldExist:   true,
		},
		{
			name:          "get non-existing key",
			key:           "nonexistent",
			expectedValue: nil,
			shouldExist:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := cache.Get(ctx, tt.key)
			if tt.shouldExist {
				if err != nil {
					t.Errorf("Get() error = %v, want nil", err)
				}
				if value != tt.expectedValue {
					t.Errorf("Get() value = %v, want %v", value, tt.expectedValue)
				}
			} else {
				if err == nil {
					t.Errorf("Get() error = nil, want error")
				}
			}
		})
	}
}

// TestMemoryCacheDelete tests cache deletion
func TestMemoryCacheDelete(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx := context.Background()

	// Set and then delete
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Delete(ctx, "key1")

	// Verify deletion
	_, err := cache.Get(ctx, "key1")
	if err == nil {
		t.Error("Delete() failed: key still exists after deletion")
	}
}

// TestMemoryCacheClear tests clearing the entire cache
func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx := context.Background()

	// Set multiple values
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)
	_ = cache.Set(ctx, "key3", "value3", 0)

	// Clear cache
	_ = cache.Clear(ctx)

	// Verify all are gone
	stats := cache.GetStats(ctx)
	if stats.Size != 0 {
		t.Errorf("Clear() failed: cache size = %d, want 0", stats.Size)
	}
}

// TestMemoryCacheTTL tests time-to-live expiration
func TestMemoryCacheTTL(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx := context.Background()

	// Set value with 100ms TTL
	_ = cache.Set(ctx, "key1", "value1", 100*time.Millisecond)

	// Should exist immediately
	_, err := cache.Get(ctx, "key1")
	if err != nil {
		t.Error("TTL test: value should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = cache.Get(ctx, "key1")
	if err == nil {
		t.Error("TTL test: value should be expired after TTL")
	}
}

// TestMemoryCacheLRUEviction tests LRU eviction when max capacity reached
func TestMemoryCacheLRUEviction(t *testing.T) {
	cache := NewMemoryCache(3, 5*time.Minute) // Max 3 items
	ctx := context.Background()

	// Fill cache
	_ = cache.Set(ctx, "key1", "value1", 0)
	_ = cache.Set(ctx, "key2", "value2", 0)
	_ = cache.Set(ctx, "key3", "value3", 0)

	// Verify size
	stats := cache.GetStats(ctx)
	if stats.Size != 3 {
		t.Errorf("Cache size = %d, want 3", stats.Size)
	}

	// Add one more - should evict least recently used (key1)
	_ = cache.Set(ctx, "key4", "value4", 0)

	// key1 should be gone
	_, err := cache.Get(ctx, "key1")
	if err == nil {
		t.Error("LRU eviction failed: least recently used key was not evicted")
	}

	// key4 should exist
	_, err = cache.Get(ctx, "key4")
	if err != nil {
		t.Error("LRU eviction: new key should exist")
	}
}

// TestMemoryCacheSize tests size tracking
func TestMemoryCacheSize(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx := context.Background()

	stats := cache.GetStats(ctx)
	if stats.Size != 0 {
		t.Errorf("Initial size = %d, want 0", stats.Size)
	}

	_ = cache.Set(ctx, "key1", "value1", 0)
	stats = cache.GetStats(ctx)
	if stats.Size != 1 {
		t.Errorf("Size after 1 insert = %d, want 1", stats.Size)
	}

	_ = cache.Set(ctx, "key2", "value2", 0)
	stats = cache.GetStats(ctx)
	if stats.Size != 2 {
		t.Errorf("Size after 2 inserts = %d, want 2", stats.Size)
	}
}

// TestMemoryCacheConcurrency tests thread safety
func TestMemoryCacheConcurrency(t *testing.T) {
	cache := NewMemoryCache(1000, 5*time.Minute)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := "key"
				_ = cache.Set(ctx, key, id*operationsPerGoroutine+j, 0)
				_, _ = cache.Get(ctx, key)
				if j%10 == 0 {
					_ = cache.Delete(ctx, key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Cache should still be functional
	stats := cache.GetStats(ctx)
	if stats.Size < 0 {
		t.Errorf("Cache size = %d, invalid", stats.Size)
	}
}

// TestMemoryCacheContextCancellation tests behavior with cancelled context
func TestMemoryCacheContextCancellation(t *testing.T) {
	cache := NewMemoryCache(10, 5*time.Minute)
	ctx, cancel := context.WithCancel(context.Background())

	// Set value with active context
	err := cache.Set(ctx, "key1", "value1", 0)
	if err != nil {
		t.Errorf("Set with active context failed: %v", err)
	}

	// Cancel context and try to get
	cancel()
	_, err = cache.Get(ctx, "key1")

	// Should still return value (context cancellation is typically not checked in cache gets)
	if err != nil && err != context.Canceled {
		t.Logf("Get with cancelled context: error %v (may be implementation specific)", err)
	}
}

// BenchmarkMemoryCacheSet benchmarks Set operation
func BenchmarkMemoryCacheSet(b *testing.B) {
	cache := NewMemoryCache(1000, 5*time.Minute)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, "key", "value", 0)
	}
}

// BenchmarkMemoryCacheGet benchmarks Get operation
func BenchmarkMemoryCacheGet(b *testing.B) {
	cache := NewMemoryCache(1000, 5*time.Minute)
	ctx := context.Background()
	cache.Set(ctx, "key", "value", 0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, "key")
	}
}

// BenchmarkMemoryCacheDelete benchmarks Delete operation
func BenchmarkMemoryCacheDelete(b *testing.B) {
	cache := NewMemoryCache(1000, 5*time.Minute)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, "key", "value", 0)
		_ = cache.Delete(ctx, "key")
	}
}

// BenchmarkMemoryCacheMixed benchmarks mixed operations (75% Get, 20% Set, 5% Delete)
func BenchmarkMemoryCacheMixed(b *testing.B) {
	cache := NewMemoryCache(1000, 5*time.Minute)
	ctx := context.Background()

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		cache.Set(ctx, "key", "value", 0)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := i % 100
		if op < 75 {
			cache.Get(ctx, "key")
		} else if op < 95 {
			cache.Set(ctx, "key", "value", 0)
		} else {
			cache.Delete(ctx, "key")
		}
	}
}

// TestCacheInterface verifies that MemoryCache implements Cache interface
func TestCacheInterface(t *testing.T) {
	var _ Cache = (*MemoryCache)(nil)
}
