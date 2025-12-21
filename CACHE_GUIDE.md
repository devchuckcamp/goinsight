# Query Caching Implementation Guide

## Overview

Phase 3 implements intelligent query result caching to reduce database load, improve response times, and minimize LLM API calls. The caching layer is designed to be performant, flexible, and easily extensible.

## Architecture

### Cache Components

```
internal/cache/
├── cache.go          # Cache interface and error handling
├── memory_cache.go   # In-memory implementation with TTL support
└── manager.go        # High-level cache management API
```

### Key Features

1. **In-Memory Caching**: Fast, thread-safe caching with O(1) lookups
2. **TTL Management**: Automatic expiration with configurable time-to-live
3. **LRU Eviction**: Removes least recently used entries when capacity reached
4. **Background Cleanup**: Periodic background task removes expired entries
5. **Query Hashing**: Consistent MD5 hashing for query deduplication
6. **Pattern Invalidation**: Bulk cache invalidation for related queries

## Cache Interface

```go
// Cache defines the caching contract
type Cache interface {
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Get(ctx context.Context, key string) (interface{}, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
    Clear(ctx context.Context) error
    GetStats(ctx context.Context) CacheStats
    Close() error
}
```

## Implementation Details

### MemoryCache Algorithm

**Data Structures**:
- `data map[string]*CacheEntry`: Main cache storage (hash -> entry)
- `expirations map[string]time.Time`: TTL tracking
- `accessTimes map[string]time.Time`: LRU tracking
- `queryHashes map[string]string`: Query -> hash mapping

**Algorithms**:

1. **Get Operation**: O(1) hash lookup + expiration check
   ```
   1. Look up entry by key
   2. Check if expired (compare with current time)
   3. Update access time for LRU tracking
   4. Return value or error
   ```

2. **Set Operation**: O(1) insertion with eviction
   ```
   1. Check if capacity exceeded
   2. If full: evict LRU entry (O(n) scan for oldest)
   3. Insert new entry with expiration time
   4. Update access time
   ```

3. **LRU Eviction**: O(n) scan of access times
   ```
   1. Find entry with minimum access time
   2. Delete entry from all maps
   3. Decrement size counter
   ```

4. **TTL Cleanup**: O(n) periodic sweep
   ```
   1. Background goroutine runs every 1 minute
   2. Scan expirations map
   3. Delete expired entries
   4. Update stats
   ```

### CacheManager API

High-level interface for service layer:

```go
// Cache query results with automatic hashing
CacheQueryResult(ctx, query, results, ttl)
GetCachedQueryResult(ctx, query) (results, found, error)

// Invalidation
InvalidateQuery(ctx, query)
InvalidatePattern(ctx, pattern)
ClearCache(ctx)

// Statistics
GetCacheStats(ctx) CacheStats
```

## Integration with FeedbackService

### Service Constructors

1. **NewFeedbackService()**: No caching
2. **NewFeedbackServiceWithProfiler()**: Profiling only
3. **NewFeedbackServiceWithCache()**: Caching only
4. **NewFeedbackServiceFull()**: All features (profiling + caching)

### Caching Strategy

The service implements two-level caching:

**Level 1: Question Cache**
- Key: User's question text
- Value: Complete AskResponse (summary, recommendations, actions)
- TTL: 5 minutes (configurable)
- Hit Rate: High for frequently asked questions
- Benefit: Avoids LLM insight generation on repeat questions

**Level 2: SQL Query Cache**
- Key: Generated SQL query hash
- Value: Query result rows
- TTL: 5 minutes (configurable)
- Hit Rate: Medium when different questions generate same SQL
- Benefit: Avoids database execution on repeated patterns

### Cache Flow

```
User Question
    ↓
[Question Cache Hit?] → Return cached response ✓
    ↓ No
[Generate SQL]
    ↓
[SQL Cache Hit?] → Cache SQL results, generate insights
    ↓ No
[Execute Query] → Cache results → Generate insights
    ↓
[Cache Response] → Return to user
```

## Configuration

### DefaultCacheConfig

```go
cache.Config{
    Enabled:    true,                    // Enable caching
    MaxSize:    1000,                   // Max 1000 entries
    DefaultTTL: 5 * time.Minute,        // 5 minute TTL
}
```

### Customization

```go
// Create custom cache manager
cacheManager := cache.NewCacheManager(
    true,                    // enabled
    500,                     // maxSize
    10 * time.Minute,        // defaultTTL
)

// Configure service
svc.SetCacheTTL(10 * time.Minute)
svc.CacheQueryResults(true)
```

## Cache Statistics

Track cache performance metrics:

```go
type CacheStats struct {
    Hits        int64  // Successful cache retrievals
    Misses      int64  // Cache lookups that failed
    Evictions   int64  // Entries removed (capacity or TTL)
    Size        int64  // Current number of entries
    MaxSize     int64  // Maximum capacity
    TTLCheckups int64  // Background cleanup runs
}
```

### Example: Get Hit Rate

```go
stats := svc.GetCacheStats(ctx)
hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100
fmt.Printf("Cache Hit Rate: %.2f%%\n", hitRate)
```

## Use Cases

### Case 1: Frequently Asked Questions
```
Question: "What are the top billing issues?"
Hit 1: Database query + LLM insight generation
Hit 2: Returned from question cache (immediate)
Hit 3: Returned from question cache (immediate)
Saving: 2 database queries, 2 LLM API calls
```

### Case 2: Similar Questions, Same SQL
```
Question A: "Show me billing problems"
Question B: "What billing issues are there?"
Both generate: SELECT * FROM feedback WHERE category = 'billing'
Hit A: Database query → cache SQL results
Hit B: Cache hit on SQL → skip database, generate insights
Saving: 1 database query
```

### Case 3: Data Updates
```
Admin updates feedback data
InvalidateCachePattern(ctx, "feedback")  // Clear related queries
Next question uses fresh data
Guarantee: Consistent with data source
```

## Performance Characteristics

| Operation | Time Complexity | Memory |
|-----------|-----------------|--------|
| Get       | O(1)            | -      |
| Set       | O(1) avg, O(n) evict | O(1) per entry |
| Delete    | O(1)            | -      |
| Cleanup   | O(n)            | -      |
| Stats     | O(n)            | -      |

## Example Usage

### Basic Setup

```go
// Initialize cache in main.go
cacheManager := cache.NewCacheManager(true, 1000, 5*time.Minute)
defer cacheManager.Close()

// Use with service
service := service.NewFeedbackServiceWithCache(
    repo, llmClient, jiraClient, cacheManager,
)
```

### Working with Cache

```go
// Get stats
stats := service.GetCacheStats(ctx)
fmt.Printf("Cache hits: %d, misses: %d\n", stats.Hits, stats.Misses)

// Clear specific query
service.InvalidateQueryCache(ctx, "SELECT * FROM feedback...")

// Clear pattern
service.InvalidateCachePattern(ctx, "feedback")

// Clear everything
service.ClearCache(ctx)

// Check if enabled
if service.IsCacheEnabled() {
    fmt.Println("Cache is active")
}
```

## Future Enhancements

1. **Redis Support**: Distributed caching across multiple instances
2. **Cache Warming**: Pre-populate cache with common queries
3. **Adaptive TTL**: Adjust TTL based on data change frequency
4. **Compression**: Reduce memory usage for large result sets
5. **Serialization**: JSON/Protobuf serialization for persistence
6. **Hit/Miss Tracking**: Per-query cache metrics

## Testing

```go
func TestCacheBasics(t *testing.T) {
    cache := cache.NewMemoryCache(100, 1*time.Minute)
    defer cache.Close()

    // Test Set/Get
    cache.Set(ctx, "key1", "value1", 1*time.Minute)
    result, err := cache.Get(ctx, "key1")
    assert.Equal(t, "value1", result)

    // Test Expiration
    cache.Set(ctx, "key2", "value2", 1*time.Millisecond)
    time.Sleep(10*time.Millisecond)
    _, err = cache.Get(ctx, "key2")
    assert.NotNil(t, err) // Should be expired

    // Test LRU Eviction
    smallCache := cache.NewMemoryCache(2, 1*time.Minute)
    smallCache.Set(ctx, "k1", "v1", 1*time.Minute)
    smallCache.Set(ctx, "k2", "v2", 1*time.Minute)
    smallCache.Set(ctx, "k3", "v3", 1*time.Minute) // Should evict k1
    _, err = smallCache.Get(ctx, "k1")
    assert.NotNil(t, err) // k1 should be evicted
}
```

## Monitoring

Monitor cache performance in production:

```go
// Log stats every minute
ticker := time.NewTicker(1 * time.Minute)
for range ticker.C {
    stats := service.GetCacheStats(ctx)
    log.Printf(
        "Cache: hits=%d, misses=%d, evictions=%d, size=%d/%d",
        stats.Hits, stats.Misses, stats.Evictions,
        stats.Size, stats.MaxSize,
    )
}
```

## Troubleshooting

### Cache Not Working
- Check `IsCacheEnabled()` returns true
- Verify TTL is > 0
- Check cache isn't cleared between requests

### High Eviction Rate
- Increase `MaxSize` in config
- Increase TTL to keep entries longer
- Monitor `GetCacheStats()` for insights

### Memory Usage Growing
- Reduce `MaxSize`
- Reduce `DefaultTTL`
- Clear cache periodically with `ClearCache()`
- Monitor individual entry size

---

**Last Updated**: December 20, 2025
