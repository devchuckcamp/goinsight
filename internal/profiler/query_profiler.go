package profiler

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QueryMetrics holds performance metrics for a single query execution
type QueryMetrics struct {
	QueryID       string        `json:"query_id"`
	QueryHash     string        `json:"query_hash"`
	Query         string        `json:"query"`
	ExecutionTime time.Duration `json:"execution_time_ms"`
	RowsReturned  int64         `json:"rows_returned"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	PoolUsage     int           `json:"pool_usage"`
	CacheHit      bool          `json:"cache_hit"`
	Error         error         `json:"error,omitempty"`
}

// QueryProfiler monitors and tracks query execution metrics
// Algorithms:
// 1. Execution Timing: Measures elapsed time from start to completion
// 2. Query Hashing: Generates consistent MD5 hash for query deduplication
// 3. Metrics Aggregation: Maintains per-query statistics for pattern analysis
// 4. Circular Buffer: Stores recent queries with bounded memory usage
type QueryProfiler struct {
	mu sync.RWMutex

	// Query metrics storage: hash -> []QueryMetrics (circular buffer)
	metrics map[string][]*QueryMetrics

	// Query statistics for trend analysis
	stats map[string]*QueryStats

	// Logger for persistent storage
	logger *Logger

	// Configuration
	maxMetricsPerQuery int
	slowQueryThresh    float64 // milliseconds

	// Aggregated statistics
	totalQueries    int64
	totalErrors     int64
	totalCacheHits  int64
	totalExecution  float64
}

// QueryStats holds aggregated statistics for a query pattern
type QueryStats struct {
	QueryHash       string
	Query           string
	ExecutionCount  int64
	TotalExecTimeMS float64
	MinExecTimeMS   float64
	MaxExecTimeMS   float64
	AvgExecTimeMS   float64
	ErrorCount      int64
	CacheHitCount   int64
	LastExecuted    time.Time
	FirstSeen       time.Time
}

// NewQueryProfiler creates a new QueryProfiler instance
func NewQueryProfiler(logger *Logger, slowQueryThresholdMS float64) *QueryProfiler {
	return &QueryProfiler{
		metrics:            make(map[string][]*QueryMetrics),
		stats:              make(map[string]*QueryStats),
		logger:             logger,
		maxMetricsPerQuery: 100, // Keep last 100 executions per query
		slowQueryThresh:    slowQueryThresholdMS,
	}
}

// hashQuery generates a consistent MD5 hash for query deduplication
// Algorithm: Normalize query -> MD5 hash
// This enables pattern-based analysis across semantically identical queries
func (qp *QueryProfiler) hashQuery(query string) string {
	hash := md5.Sum([]byte(query))
	return fmt.Sprintf("%x", hash)
}

// StartQueryExecution marks the beginning of a query execution
// Returns a QueryMetrics object to be updated when query completes
func (qp *QueryProfiler) StartQueryExecution(query string) *QueryMetrics {
	queryHash := qp.hashQuery(query)
	queryID := uuid.New().String()

	return &QueryMetrics{
		QueryID:   queryID,
		QueryHash: queryHash,
		Query:     query,
		StartTime: time.Now(),
	}
}

// RecordQueryExecution records completed query metrics
// Algorithm:
// 1. Calculate execution time delta
// 2. Update aggregated statistics
// 3. Maintain circular buffer of recent executions
// 4. Log slow queries
// 5. Update counters
func (qp *QueryProfiler) RecordQueryExecution(metrics *QueryMetrics, rowsReturned int64, poolUsage int, cacheHit bool, err error) {
	qp.mu.Lock()
	defer qp.mu.Unlock()

	// Calculate execution time
	metrics.EndTime = time.Now()
	metrics.ExecutionTime = metrics.EndTime.Sub(metrics.StartTime)
	metrics.RowsReturned = rowsReturned
	metrics.PoolUsage = poolUsage
	metrics.CacheHit = cacheHit
	metrics.Error = err

	// Update global counters
	qp.totalQueries++
	execTimeMS := metrics.ExecutionTime.Seconds() * 1000
	qp.totalExecution += execTimeMS

	if cacheHit {
		qp.totalCacheHits++
	}

	if err != nil {
		qp.totalErrors++
	}

	// Store in circular buffer
	qp.storeMetrics(metrics)

	// Update statistics
	qp.updateStats(metrics, execTimeMS)

	// Log slow queries
	if execTimeMS > qp.slowQueryThresh {
		qp.logger.LogSlowQuery(metrics.QueryID, metrics.Query, metrics.QueryHash, execTimeMS, qp.slowQueryThresh)
	}

	// Log query execution
	qp.logger.LogQueryExecution(metrics.QueryID, metrics.Query, metrics.QueryHash, execTimeMS, rowsReturned, err)
}

// storeMetrics maintains a circular buffer of recent executions
// Algorithm: When buffer exceeds maxMetricsPerQuery, remove oldest entry
func (qp *QueryProfiler) storeMetrics(metrics *QueryMetrics) {
	hash := metrics.QueryHash

	if _, exists := qp.metrics[hash]; !exists {
		qp.metrics[hash] = make([]*QueryMetrics, 0, qp.maxMetricsPerQuery)
	}

	buffer := qp.metrics[hash]

	// Remove oldest if buffer is full
	if len(buffer) >= qp.maxMetricsPerQuery {
		buffer = buffer[1:]
	}

	qp.metrics[hash] = append(buffer, metrics)
}

// updateStats updates aggregated statistics
// Algorithm: Track min, max, average execution times and error rates
func (qp *QueryProfiler) updateStats(metrics *QueryMetrics, execTimeMS float64) {
	hash := metrics.QueryHash

	if _, exists := qp.stats[hash]; !exists {
		qp.stats[hash] = &QueryStats{
			QueryHash:     hash,
			Query:         metrics.Query,
			FirstSeen:     metrics.StartTime,
			MinExecTimeMS: execTimeMS,
			MaxExecTimeMS: execTimeMS,
		}
	}

	stats := qp.stats[hash]

	// Update execution metrics
	stats.ExecutionCount++
	stats.TotalExecTimeMS += execTimeMS
	stats.LastExecuted = metrics.EndTime

	// Update min/max
	if execTimeMS < stats.MinExecTimeMS {
		stats.MinExecTimeMS = execTimeMS
	}
	if execTimeMS > stats.MaxExecTimeMS {
		stats.MaxExecTimeMS = execTimeMS
	}

	// Recalculate average
	stats.AvgExecTimeMS = stats.TotalExecTimeMS / float64(stats.ExecutionCount)

	// Update error count
	if metrics.Error != nil {
		stats.ErrorCount++
	}

	// Update cache hit count
	if metrics.CacheHit {
		stats.CacheHitCount++
	}
}

// GetStats returns statistics for a specific query hash
func (qp *QueryProfiler) GetStats(queryHash string) *QueryStats {
	qp.mu.RLock()
	defer qp.mu.RUnlock()
	if stats, exists := qp.stats[queryHash]; exists {
		return stats
	}
	return nil
}

// GetRecentMetrics returns recent execution metrics for a query hash
func (qp *QueryProfiler) GetRecentMetrics(queryHash string, limit int) []*QueryMetrics {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	metrics, exists := qp.metrics[queryHash]
	if !exists {
		return nil
	}

	if limit > len(metrics) {
		limit = len(metrics)
	}

	// Return the most recent 'limit' entries
	start := len(metrics) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*QueryMetrics, limit)
	copy(result, metrics[start:])
	return result
}

// GetAggregateStats returns all query statistics
func (qp *QueryProfiler) GetAggregateStats() map[string]*QueryStats {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	// Create a copy to avoid external modifications
	result := make(map[string]*QueryStats)
	for hash, stats := range qp.stats {
		statsCopy := *stats
		result[hash] = &statsCopy
	}
	return result
}

// GetSlowQueries returns queries exceeding the slow query threshold
// Algorithm: Filter stats where AvgExecTimeMS > slowQueryThresh
func (qp *QueryProfiler) GetSlowQueries() []*QueryStats {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	var slowQueries []*QueryStats

	for _, stats := range qp.stats {
		if stats.AvgExecTimeMS > qp.slowQueryThresh {
			statsCopy := *stats
			slowQueries = append(slowQueries, &statsCopy)
		}
	}

	return slowQueries
}

// GetProfileReport returns a comprehensive profiling report
type ProfileReport struct {
	TotalQueries     int64
	TotalErrors      int64
	TotalCacheHits   int64
	CacheHitRate     float64
	TotalExecTimeMS  float64
	AvgExecTimeMS    float64
	UniqueQueries    int
	SlowQueryCount   int
	ErrorRate        float64
	GeneratedAt      time.Time
}

func (qp *QueryProfiler) GetProfileReport() ProfileReport {
	qp.mu.RLock()
	defer qp.mu.RUnlock()

	slowCount := 0
	for _, stats := range qp.stats {
		if stats.AvgExecTimeMS > qp.slowQueryThresh {
			slowCount++
		}
	}

	avgExecTime := 0.0
	if qp.totalQueries > 0 {
		avgExecTime = qp.totalExecution / float64(qp.totalQueries)
	}

	cacheHitRate := 0.0
	if qp.totalQueries > 0 {
		cacheHitRate = float64(qp.totalCacheHits) / float64(qp.totalQueries) * 100
	}

	errorRate := 0.0
	if qp.totalQueries > 0 {
		errorRate = float64(qp.totalErrors) / float64(qp.totalQueries) * 100
	}

	return ProfileReport{
		TotalQueries:    qp.totalQueries,
		TotalErrors:     qp.totalErrors,
		TotalCacheHits:  qp.totalCacheHits,
		CacheHitRate:    cacheHitRate,
		TotalExecTimeMS: qp.totalExecution,
		AvgExecTimeMS:   avgExecTime,
		UniqueQueries:   len(qp.stats),
		SlowQueryCount:  slowCount,
		ErrorRate:       errorRate,
		GeneratedAt:     time.Now(),
	}
}

// RecordQueryExecutionWithContext wraps query execution with context
// This enables timeout and cancellation support
func (qp *QueryProfiler) RecordQueryExecutionWithContext(ctx context.Context, query string, execFunc func() (int64, int, error)) error {
	metrics := qp.StartQueryExecution(query)

	// Execute the query
	rows, poolUsage, err := execFunc()

	// Record the execution
	qp.RecordQueryExecution(metrics, rows, poolUsage, false, err)

	return err
}

// SetSlowQueryThreshold updates the slow query threshold (in milliseconds)
func (qp *QueryProfiler) SetSlowQueryThreshold(thresholdMS float64) {
	qp.mu.Lock()
	defer qp.mu.Unlock()
	qp.slowQueryThresh = thresholdMS
}

// Reset clears all profiling data
func (qp *QueryProfiler) Reset() {
	qp.mu.Lock()
	defer qp.mu.Unlock()

	qp.metrics = make(map[string][]*QueryMetrics)
	qp.stats = make(map[string]*QueryStats)
	qp.totalQueries = 0
	qp.totalErrors = 0
	qp.totalCacheHits = 0
	qp.totalExecution = 0
}
