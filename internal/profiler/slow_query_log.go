package profiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// SlowQueryEntry represents a detected slow query
type SlowQueryEntry struct {
	DetectedAt    time.Time `json:"detected_at"`
	QueryID       string    `json:"query_id"`
	QueryHash     string    `json:"query_hash"`
	Query         string    `json:"query"`
	ExecutionMS   float64   `json:"execution_ms"`
	ThresholdMS   float64   `json:"threshold_ms"`
	ExceededByMS  float64   `json:"exceeded_by_ms"`
	RowsReturned  int64     `json:"rows_returned"`
	Occurrences   int64     `json:"occurrences"`
	LastOccurred  time.Time `json:"last_occurred"`
}

// SlowQueryAnalysis holds analysis of slow query patterns
type SlowQueryAnalysis struct {
	TotalSlowQueries      int64
	UniqueSlowQueries     int
	MostFrequentQuery     *SlowQueryEntry
	SlowestQuery          *SlowQueryEntry
	AverageSlowestTimeMS  float64
	TotalSlowExecTimeMS   float64
	SlowQueryPercentage   float64
	AnalysisGeneratedAt   time.Time
}

// SlowQueryLogger tracks and analyzes slow query patterns
// Algorithms:
// 1. Aggregation: Groups identical queries and tracks patterns
// 2. Deviation Detection: Identifies when performance degrades
// 3. Trend Analysis: Tracks slow query frequency over time
// 4. Recommendation Engine: Suggests index creation and query optimization
type SlowQueryLogger struct {
	mu sync.RWMutex

	// Slow query tracking: hash -> SlowQueryEntry
	slowQueries map[string]*SlowQueryEntry

	// Execution history for trend analysis
	history []*SlowQueryEntry

	// File for persistent logging
	logFilePath string
	logFile     *os.File

	// Configuration
	maxHistorySize      int
	slowQueryThreshold  float64
	degredationFactor   float64 // Alert when execution time increases by this factor
	warningThreshold    float64 // Warning level (slightly higher than threshold)

	// Metrics
	totalSlowQueryCount int64
}

// NewSlowQueryLogger creates a new SlowQueryLogger instance
func NewSlowQueryLogger(logDir string, slowQueryThresholdMS float64) (*SlowQueryLogger, error) {
	if logDir == "" {
		logDir = "./logs"
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "slow_queries.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open slow query log file: %w", err)
	}

	logger := &SlowQueryLogger{
		slowQueries:        make(map[string]*SlowQueryEntry),
		history:            make([]*SlowQueryEntry, 0, 1000),
		logFilePath:        logPath,
		logFile:            file,
		maxHistorySize:     1000,
		slowQueryThreshold: slowQueryThresholdMS,
		degredationFactor:  1.2, // Alert if 20% slower than average
		warningThreshold:   slowQueryThresholdMS * 1.5,
	}

	return logger, nil
}

// RecordSlowQuery records a slow query detection
// Algorithm:
// 1. Check if query already tracked
// 2. Update aggregated entry (count, last occurrence)
// 3. Append to history for trend analysis
// 4. Persist to log file
// 5. Check for performance degradation
func (sql *SlowQueryLogger) RecordSlowQuery(queryID, query, queryHash string, executionMS, thresholdMS float64, rowsReturned int64) error {
	sql.mu.Lock()
	defer sql.mu.Unlock()

	exceededByMS := executionMS - thresholdMS

	entry := &SlowQueryEntry{
		DetectedAt:   time.Now(),
		QueryID:      queryID,
		QueryHash:    queryHash,
		Query:        query,
		ExecutionMS:  executionMS,
		ThresholdMS:  thresholdMS,
		ExceededByMS: exceededByMS,
		RowsReturned: rowsReturned,
		LastOccurred: time.Now(),
	}

	// Update or create aggregated entry
	if existing, exists := sql.slowQueries[queryHash]; exists {
		existing.Occurrences++
		existing.LastOccurred = time.Now()
		entry.Occurrences = existing.Occurrences
	} else {
		entry.Occurrences = 1
	}

	sql.slowQueries[queryHash] = entry

	// Add to history (circular buffer)
	sql.appendToHistory(entry)

	// Increment total slow query count
	sql.totalSlowQueryCount++

	// Persist to log file
	if err := sql.persistEntry(entry); err != nil {
		return fmt.Errorf("failed to persist slow query: %w", err)
	}

	// Check for performance degradation
	sql.checkPerformanceDegradation(entry)

	return nil
}

// appendToHistory maintains a circular buffer of slow query history
func (sql *SlowQueryLogger) appendToHistory(entry *SlowQueryEntry) {
	if len(sql.history) >= sql.maxHistorySize {
		// Remove oldest entry
		sql.history = sql.history[1:]
	}
	sql.history = append(sql.history, entry)
}

// persistEntry writes the slow query entry to the log file
func (sql *SlowQueryLogger) persistEntry(entry *SlowQueryEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if sql.logFile != nil {
		if _, err := sql.logFile.Write(append(data, '\n')); err != nil {
			return err
		}
	}

	return nil
}

// checkPerformanceDegradation detects when query performance degrades
// Algorithm: Compare current execution time with historical average
func (sql *SlowQueryLogger) checkPerformanceDegradation(entry *SlowQueryEntry) {
	hash := entry.QueryHash

	// Find historical entries for this query
	var historicalTimes []float64
	for _, hist := range sql.history {
		if hist.QueryHash == hash {
			historicalTimes = append(historicalTimes, hist.ExecutionMS)
		}
	}

	if len(historicalTimes) < 3 {
		return // Need at least 3 data points for meaningful analysis
	}

	// Calculate average
	sum := 0.0
	for _, t := range historicalTimes {
		sum += t
	}
	avgTime := sum / float64(len(historicalTimes))

	// Check if current execution is significantly slower
	if entry.ExecutionMS > avgTime*sql.degredationFactor {
		// Performance has degraded - this would trigger an alert
		// In production, this could send to alerting service
	}
}

// GetSlowQueryEntry returns details for a specific slow query
func (sql *SlowQueryLogger) GetSlowQueryEntry(queryHash string) *SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	if entry, exists := sql.slowQueries[queryHash]; exists {
		entryCopy := *entry
		return &entryCopy
	}
	return nil
}

// GetAllSlowQueries returns all tracked slow queries
func (sql *SlowQueryLogger) GetAllSlowQueries() []*SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	entries := make([]*SlowQueryEntry, 0, len(sql.slowQueries))
	for _, entry := range sql.slowQueries {
		entryCopy := *entry
		entries = append(entries, &entryCopy)
	}

	return entries
}

// GetMostFrequentSlowQueries returns slow queries sorted by occurrence count
func (sql *SlowQueryLogger) GetMostFrequentSlowQueries(limit int) []*SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	entries := make([]*SlowQueryEntry, 0, len(sql.slowQueries))
	for _, entry := range sql.slowQueries {
		entryCopy := *entry
		entries = append(entries, &entryCopy)
	}

	// Sort by occurrences (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Occurrences > entries[j].Occurrences
	})

	if limit > len(entries) {
		limit = len(entries)
	}

	return entries[:limit]
}

// GetSlowestQueries returns slow queries sorted by execution time
func (sql *SlowQueryLogger) GetSlowestQueries(limit int) []*SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	entries := make([]*SlowQueryEntry, 0, len(sql.slowQueries))
	for _, entry := range sql.slowQueries {
		entryCopy := *entry
		entries = append(entries, &entryCopy)
	}

	// Sort by execution time (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ExecutionMS > entries[j].ExecutionMS
	})

	if limit > len(entries) {
		limit = len(entries)
	}

	return entries[:limit]
}

// GetAnalysis returns comprehensive analysis of slow query patterns
// Algorithm: Aggregates all slow query data and identifies patterns
func (sql *SlowQueryLogger) GetAnalysis(totalQueries int64, totalExecTimeMS float64) SlowQueryAnalysis {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	analysis := SlowQueryAnalysis{
		TotalSlowQueries:   sql.totalSlowQueryCount,
		UniqueSlowQueries:  len(sql.slowQueries),
		AnalysisGeneratedAt: time.Now(),
	}

	if len(sql.slowQueries) == 0 {
		return analysis
	}

	// Find most frequent
	maxOccurrences := int64(0)
	for _, entry := range sql.slowQueries {
		if entry.Occurrences > maxOccurrences {
			maxOccurrences = entry.Occurrences
			entryCopy := *entry
			analysis.MostFrequentQuery = &entryCopy
		}
	}

	// Find slowest
	maxExecTime := 0.0
	for _, entry := range sql.slowQueries {
		if entry.ExecutionMS > maxExecTime {
			maxExecTime = entry.ExecutionMS
			entryCopy := *entry
			analysis.SlowestQuery = &entryCopy
		}
	}

	// Calculate averages
	totalSlowTime := 0.0
	for _, entry := range sql.slowQueries {
		totalSlowTime += entry.ExecutionMS * float64(entry.Occurrences)
	}

	analysis.TotalSlowExecTimeMS = totalSlowTime
	if analysis.TotalSlowQueries > 0 {
		analysis.AverageSlowestTimeMS = totalSlowTime / float64(analysis.TotalSlowQueries)
	}

	// Calculate slow query percentage
	if totalQueries > 0 {
		analysis.SlowQueryPercentage = float64(analysis.TotalSlowQueries) / float64(totalQueries) * 100
	}

	return analysis
}

// RecentSlowQueries returns slow queries detected in the last duration
func (sql *SlowQueryLogger) RecentSlowQueries(duration time.Duration) []*SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	threshold := time.Now().Add(-duration)
	var entries []*SlowQueryEntry

	for _, entry := range sql.slowQueries {
		if entry.LastOccurred.After(threshold) {
			entryCopy := *entry
			entries = append(entries, &entryCopy)
		}
	}

	return entries
}

// GetHistoryForQuery returns execution history for a specific query
func (sql *SlowQueryLogger) GetHistoryForQuery(queryHash string) []*SlowQueryEntry {
	sql.mu.RLock()
	defer sql.mu.RUnlock()

	var entries []*SlowQueryEntry
	for _, hist := range sql.history {
		if hist.QueryHash == queryHash {
			entryCopy := *hist
			entries = append(entries, &entryCopy)
		}
	}

	return entries
}

// Close closes the slow query log file
func (sql *SlowQueryLogger) Close() error {
	sql.mu.Lock()
	defer sql.mu.Unlock()

	if sql.logFile != nil {
		return sql.logFile.Close()
	}
	return nil
}

// GetLogPath returns the path to the slow query log file
func (sql *SlowQueryLogger) GetLogPath() string {
	return sql.logFilePath
}

// ClearOldEntries removes entries older than the specified duration
func (sql *SlowQueryLogger) ClearOldEntries(age time.Duration) {
	sql.mu.Lock()
	defer sql.mu.Unlock()

	threshold := time.Now().Add(-age)

	// Remove from history
	newHistory := make([]*SlowQueryEntry, 0)
	for _, entry := range sql.history {
		if entry.DetectedAt.After(threshold) {
			newHistory = append(newHistory, entry)
		}
	}
	sql.history = newHistory

	// Remove from slow queries map if all occurrences are old
	toDelete := make([]string, 0)
	for hash, entry := range sql.slowQueries {
		if entry.LastOccurred.Before(threshold) {
			toDelete = append(toDelete, hash)
		}
	}

	for _, hash := range toDelete {
		delete(sql.slowQueries, hash)
	}
}
