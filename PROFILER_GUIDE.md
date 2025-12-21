# Query Performance Monitoring & Optimization

## Overview

Phase 2 introduces a comprehensive **Query Performance Monitoring & Optimization** system that provides real-time insights into query execution patterns, identifies bottlenecks, and suggests optimizations.

## Architecture

The profiler package (`internal/profiler/`) contains four core components working together:

```
┌─────────────────────────────────────────────────────────────┐
│          Query Performance Monitoring System                 │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐                       │
│  │   Logger     │    │QueryProfiler │                       │
│  │              │    │              │                       │
│  │  - JSON Logs │    │ - Execution  │                       │
│  │  - File Mgmt │    │   Timing     │                       │
│  │  - Rotation  │    │ - Hashing    │                       │
│  │  - Levels    │    │ - Stats      │                       │
│  └──────────────┘    └──────────────┘                       │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐                       │
│  │SlowQueryLog  │    │  Optimizer   │                       │
│  │              │    │              │                       │
│  │ - Detection  │    │ - Index recs │                       │
│  │ - Patterns   │    │ - Query recs │                       │
│  │ - Trends     │    │ - Impact     │                       │
│  │ - Analysis   │    │   scoring    │                       │
│  └──────────────┘    └──────────────┘                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
         ↑                                      ↑
         └──── Integrated via FeedbackService ─┘
```

## Components

### 1. Logger (`logger.go`)

**Purpose**: Structured logging with file persistence and extensibility

**Key Features**:
- **Structured JSON Output**: All logs are JSON-formatted for easy parsing
- **Log Levels**: DEBUG, INFO, WARN, ERROR with filtering
- **File Management**: 
  - Automatic log rotation (100MB default)
  - Keeps up to 5 rotated backups
  - Persistent storage in `./logs/profiler.log`
- **Dual Output**: Console + file logging (configurable)
- **Cloud-Ready**: Designed for extension to CloudWatch, DataDog, etc.

**Algorithms**:
- **Rotation**: When file size exceeds threshold, rotate to `.1`, `.2`, `.3`
- **Level Filtering**: Only logs meeting minimum level are written
- **Thread-Safe**: Mutex protection for concurrent logging

**Usage**:
```go
logger, _ := profiler.NewLogger("./logs", true) // Console + file logging
defer logger.Close()

logger.Info("Analysis started", map[string]interface{}{
    "question": "What's the revenue trend?",
})

logger.LogQueryExecution(
    queryID, query, queryHash, 
    executionTimeMS, rowsReturned, 
    err,
)
```

**Log Files**:
- `./logs/profiler.log` - Main application log
- `./logs/profiler.log.1` - Previous rotation
- `./logs/slow_queries.log` - Dedicated slow query log

### 2. QueryProfiler (`query_profiler.go`)

**Purpose**: Track and analyze query execution metrics

**Key Features**:
- **Execution Timing**: High-precision timing (start → end)
- **Query Hashing**: MD5 hash for pattern-based deduplication
- **Circular Buffers**: 
  - Store recent 100 executions per query
  - Bounded memory usage
  - O(1) append with memory cap
- **Aggregated Statistics**:
  - Min/Max/Avg execution times
  - Error rates
  - Cache hit rates
  - Execution counts
- **Performance Report**:
  - Total queries executed
  - Unique queries tracked
  - Slow query count
  - Cache hit percentage
  - Error rates

**Algorithms**:

1. **Query Hashing** (Deduplication):
   ```
   MD5(query_text) → query_hash
   Purpose: Group identical queries across executions
   ```

2. **Circular Buffer** (Memory Management):
   ```
   metrics[query_hash] = [QueryMetrics...]
   When len > maxSize: remove oldest, append newest
   Time: O(1), Space: O(maxSize)
   ```

3. **Statistics Aggregation**:
   ```
   For each execution:
     - Update min/max/avg times
     - Increment execution count
     - Track error rate
     - Calculate cache hit rate
   ```

**Usage**:
```go
profiler := profiler.NewQueryProfiler(logger, 500.0) // 500ms threshold

// Start profiling
metrics := profiler.StartQueryExecution(sqlQuery)

// Execute query (your code here)
rows, err := db.Query(sqlQuery)

// Record execution
profiler.RecordQueryExecution(
    metrics,           // metrics object
    rowsReturned,      // int64
    poolUsage,         // int (connection pool usage)
    cacheHit,          // bool
    err,               // error
)

// Get statistics
stats := profiler.GetStats(queryHash)
fmt.Printf("Avg: %.2fms, Min: %.2fms, Max: %.2fms\n",
    stats.AvgExecTimeMS, stats.MinExecTimeMS, stats.MaxExecTimeMS)

// Get comprehensive report
report := profiler.GetProfileReport()
fmt.Printf("Total queries: %d, Unique: %d, Slow: %d\n",
    report.TotalQueries, report.UniqueQueries, report.SlowQueryCount)
```

**Metrics Tracked**:
- Query ID (UUID)
- Query hash (MD5)
- Full query text
- Execution time (ms)
- Rows returned
- Connection pool usage
- Cache hit status
- Error details
- Timestamp

### 3. SlowQueryLogger (`slow_query_log.go`)

**Purpose**: Detect, track, and analyze slow query patterns

**Key Features**:
- **Slow Query Detection**: 
  - Configurable threshold (default: 500ms)
  - Real-time alerting capability
- **Pattern Analysis**:
  - Group identical slow queries
  - Track occurrence frequency
  - Identify trends
- **Performance Degradation Detection**:
  - Alert when query slows by 20% (configurable)
  - Historical comparison
  - Trend analysis
- **Persistent Logging**:
  - JSON-formatted slow query log
  - Historical data for trend analysis
  - Easy parsing for reporting
- **Circular History Buffer**:
  - Keep last 1000 slow query occurrences
  - Enables trend detection
  - Bounded memory usage

**Algorithms**:

1. **Slow Query Detection**:
   ```
   IF execution_time > threshold THEN
     record_slow_query()
     check_degradation()
     log_to_file()
   END
   ```

2. **Aggregation**:
   ```
   slow_queries[query_hash] = SlowQueryEntry {
     occurrences++,
     last_occurred = now,
     avg_time = recalculate(),
   }
   ```

3. **Degradation Detection**:
   ```
   IF current_exec_time > avg_exec_time * degradation_factor THEN
     alert_performance_degradation()
   END
   degradation_factor = 1.2 (20% threshold)
   ```

4. **Trend Analysis**:
   ```
   For each unique query:
     - Find all historical executions
     - Calculate average
     - Compare current to average
     - Detect patterns
   ```

**Usage**:
```go
slowLog, _ := profiler.NewSlowQueryLogger("./logs", 500.0)
defer slowLog.Close()

// Record a slow query (called by profiler)
slowLog.RecordSlowQuery(
    queryID,      // UUID
    query,        // SQL text
    queryHash,    // MD5 hash
    850.5,        // execution ms
    500.0,        // threshold ms
    1000,         // rows returned
)

// Get analysis
analysis := slowLog.GetAnalysis(totalQueries, totalExecTime)
fmt.Printf("Slow queries: %d (%.1f%%)\n",
    analysis.TotalSlowQueries,
    analysis.SlowQueryPercentage)

// Top 5 most frequent
frequent := slowLog.GetMostFrequentSlowQueries(5)

// Top 5 slowest
slowest := slowLog.GetSlowestQueries(5)

// Recent slow queries (last 1 hour)
recent := slowLog.RecentSlowQueries(time.Hour)
```

**Slow Query Entry Structure**:
```json
{
  "detected_at": "2025-12-20T10:30:45Z",
  "query_id": "550e8400-e29b-41d4-a716-446655440000",
  "query_hash": "a1b2c3d4e5f6",
  "query": "SELECT * FROM feedback WHERE sentiment < 0",
  "execution_ms": 850.5,
  "threshold_ms": 500.0,
  "exceeded_by_ms": 350.5,
  "rows_returned": 1000,
  "occurrences": 12,
  "last_occurred": "2025-12-20T10:30:45Z"
}
```

### 4. QueryOptimizer (`optimizer.go`)

**Purpose**: Analyze queries and suggest optimizations

**Key Features**:
- **Pattern Matching**:
  - SELECT * usage detection
  - Missing index identification
  - OR condition analysis
  - JOIN optimization opportunities
  - Subquery pattern detection
  - Full table scan indicators
- **Heuristic Analysis**:
  - Rule-based suggestion generation
  - Rule set optimizable at runtime
  - Extensible for custom rules
- **Statistical Analysis**:
  - Uses execution metrics from QueryProfiler
  - High error rate detection
  - Execution time analysis
- **Impact Scoring**:
  - 0-100 scale
  - Prioritizes recommendations
  - Guides optimization efforts
- **Severity Classification**:
  - CRITICAL: Immediate attention needed
  - HIGH: Significant impact (30-60% improvement)
  - MEDIUM: Moderate impact (15-35% improvement)
  - LOW: Minor improvements (<15%)

**Algorithms**:

1. **Pattern Matching** (Regex-based):
   ```
   SELECT * → UnusedColumn suggestion
   WHERE with OR → QueryRewrite suggestion
   JOIN ON columns → MissingIndex suggestion
   FULL JOIN → JoinOptimization suggestion
   Subqueries → NSubqueryUsage suggestion
   ```

2. **Column Extraction**:
   ```
   Input:  "table.column = value OR column = value"
   Regex:  (\w+)\.(\w+) | \b(\w+)\s*[=<>]
   Output: ["column"]
   ```

3. **Execution Analysis**:
   ```
   IF avg_exec_time > 1000ms THEN
     severity = CRITICAL
     impact_score = 60.0
     suggestion = "Review indexes and query plan"
   END
   ```

4. **Impact Scoring**:
   ```
   score = estimated_improvement_percentage
   Examples:
   - Missing index: 40%
   - Query rewrite: 25%
   - JOIN optimization: 50%
   - SELECT * elimination: 15%
   ```

**Usage**:
```go
optimizer := profiler.NewQueryOptimizer()
profiler := profiler.NewQueryProfiler(logger, 500.0)

// Analyze a slow query
stats := profiler.GetStats(queryHash)
suggestions := optimizer.AnalyzeQuery(sqlQuery, stats)

// Display by severity
critical := profiler.FilterSuggestionsBySeverity(suggestions, profiler.Critical)
for _, s := range critical {
    fmt.Printf("[%s] %s: %s\n", s.Severity, s.Title, s.Suggestion)
}

// Sort by impact
sorted := profiler.SortSuggestionsByImpact(suggestions)
fmt.Printf("Estimated improvement: %.0f%%\n", sorted[0].ImpactScore)

// Generate detailed report
report := optimizer.GenerateReport(sqlQuery, stats)
fmt.Printf("Critical: %d, High: %d, Medium: %d, Low: %d\n",
    report.CriticalCount, report.HighCount,
    report.MediumCount, report.LowCount)
```

**Suggestion Structure**:
```json
{
  "type": "missing_index",
  "severity": "high",
  "title": "Missing Index on WHERE Clause",
  "description": "Columns in WHERE clause should have indexes",
  "suggestion": "CREATE INDEX idx_sentiment ON feedback(sentiment)",
  "columns": ["sentiment"],
  "impact_score": 40.0
}
```

## Integration with FeedbackService

The profiler is seamlessly integrated into `FeedbackService`:

### Initialization
```go
// Initialize profiler components
profilerConfig := profiler.DefaultConfig()
profilerConfig.SlowQueryThresholdMS = 500.0
profilerConfig.EnableConsoleLogging = false

profilerComponents, _ := profiler.InitializeProfiler(profilerConfig)

// Create service with profiler
service := service.NewFeedbackServiceWithProfiler(
    repo,
    llmClient,
    jiraClient,
    profilerComponents.Logger,
    profilerComponents.QueryProfiler,
    profilerComponents.SlowQueryLog,
    profilerComponents.QueryOptimizer,
)
```

### Automatic Tracking
The `AnalyzeFeedback` method automatically:
1. **Starts profiling** when SQL is generated
2. **Records execution metrics** after query completes
3. **Logs slow queries** if threshold exceeded
4. **Generates optimization suggestions** for slow queries
5. **Logs all activities** with structured context

### Profiler Methods
```go
// Get current metrics
report := service.GetProfileReport()

// Get slow query analysis
analysis := service.GetSlowQueryAnalysis()

// Get optimization suggestions
suggestions := service.GetOptimizationSuggestions()

// Get specific slow query details
frequent := service.GetMostFrequentSlowQueries(5)
slowest := service.GetSlowestQueries(5)

// Reset metrics
service.ResetProfiler()
```

## Configuration

### Default Configuration
```go
config := profiler.DefaultConfig()
// LogDirectory: "./logs"
// EnableConsoleLogging: false
// MinLogLevel: INFO
// SlowQueryThresholdMS: 500.0
// MaxMetricsPerQuery: 100
// PerformanceDegradationFactor: 1.2
```

### Custom Configuration
```go
config := profiler.ProfilerConfig{
    LogDirectory:                 "./metrics",
    EnableConsoleLogging:         true,
    MinLogLevel:                  profiler.DEBUG,
    SlowQueryThresholdMS:         300.0,     // More aggressive
    MaxMetricsPerQuery:           200,
    PerformanceDegradationFactor: 1.15,      // 15% degradation alert
    WarningThresholdMS:           450.0,
}

components, _ := profiler.InitializeProfiler(config)
```

## Log Files

### Application Log (`./logs/profiler.log`)
```json
{"timestamp":"2025-12-20T10:30:45Z","level":"INFO","message":"Feedback analysis completed","question":"What are the top issues?","results":150,"actions":3,"exec_time_ms":245.5}
{"timestamp":"2025-12-20T10:31:02Z","level":"WARN","message":"Slow query detected","query_id":"550e8400-e29b-41d4-a716-446655440000","query_hash":"a1b2c3d4","execution_ms":850.5,"threshold_ms":500.0}
{"timestamp":"2025-12-20T10:31:15Z","level":"DEBUG","message":"Query optimization suggestions","query_id":"550e8400","suggestions":3}
```

### Slow Query Log (`./logs/slow_queries.log`)
```json
{"detected_at":"2025-12-20T10:31:02Z","query_id":"550e8400","execution_ms":850.5,"threshold_ms":500.0,"exceeded_by_ms":350.5,"occurrences":1}
{"detected_at":"2025-12-20T10:31:45Z","query_id":"550e8401","execution_ms":1200.3,"threshold_ms":500.0,"exceeded_by_ms":700.3,"occurrences":2}
```

## Performance Characteristics

### Memory Usage
- **QueryProfiler**: O(n × m) where n = unique queries, m = max metrics per query (default 100)
  - Example: 50 unique queries × 100 metrics ≈ 50KB
- **SlowQueryLogger**: O(s × 1000) where s = slow queries (history buffer = 1000)
  - Example: 10 slow queries × 1000 history ≈ 100KB
- **Logger**: O(1) - just keeps file handle open

### Processing Overhead
- **Query hashing**: O(n) where n = query length (MD5)
- **Metric recording**: O(1) - append to circular buffer
- **Statistics update**: O(1) - simple arithmetic
- **Slow query detection**: O(1) - threshold check

### Typical Metrics
- Per-query timing: < 1ms overhead
- Hash calculation: ~0.1ms for 500-char query
- Logging: ~0.2ms per write (buffered)

## Extension Points

The profiler is designed for extensibility:

1. **Cloud Logging**: Extend `Logger` to send to CloudWatch, DataDog, etc.
   ```go
   type CloudLogger struct {
       *Logger
       cloudClient CloudService
   }
   ```

2. **Custom Rules**: Add rules to `QueryOptimizer`
   ```go
   func (qo *QueryOptimizer) addCustomRule(pattern string, rule Rule)
   ```

3. **Alert Integrations**: Hook into SlowQueryLogger
   ```go
   func (sql *SlowQueryLogger) OnSlowQueryDetected(hook AlertFunc)
   ```

4. **Metrics Export**: Export profiler data to Prometheus/Grafana
   ```go
   func (p *ProfileReport) ToPrometheus() string
   ```

## Best Practices

1. **Initialize at Startup**:
   ```go
   components, _ := profiler.InitializeProfiler(config)
   defer components.Cleanup()
   ```

2. **Monitor Regularly**:
   ```go
   ticker := time.NewTicker(5 * time.Minute)
   defer ticker.Stop()
   for range ticker.C {
       report := service.GetProfileReport()
       fmt.Println(report)
   }
   ```

3. **Act on Suggestions**:
   ```go
   suggestions := service.GetOptimizationSuggestions()
   for hash, suggs := range suggestions {
       if len(suggs) > 0 {
           fmt.Printf("Query %s needs optimization:\n", hash)
           for _, s := range suggs {
               fmt.Printf("  - %s: %s\n", s.Title, s.Suggestion)
           }
       }
   }
   ```

4. **Reset Between Test Runs**:
   ```go
   service.ResetProfiler()
   ```

5. **Configure Thresholds Based on SLA**:
   ```go
   config.SlowQueryThresholdMS = 200.0 // 200ms SLA requirement
   ```

## Summary

The Query Performance Monitoring & Optimization system provides:

✅ **Real-time visibility** into query execution  
✅ **Automatic slow query detection** with pattern analysis  
✅ **Data-driven optimization suggestions** with impact scoring  
✅ **Persistent logging** in JSON format  
✅ **Cloud-ready architecture** for future extensions  
✅ **Zero-config initialization** with sensible defaults  
✅ **Integrated with FeedbackService** for automatic tracking  

This foundation enables continuous performance improvement and helps identify bottlenecks quickly.
