# Profiler Examples

Complete, runnable examples demonstrating the profiler components.

## Example 1: Basic Profiler Setup

```go
package main

import (
	"fmt"
	"time"

	"github.com/chuckie/goinsight/internal/profiler"
)

func main() {
	// Initialize with defaults
	config := profiler.DefaultConfig()
	config.SlowQueryThresholdMS = 100.0

	components, err := profiler.InitializeProfiler(config)
	if err != nil {
		panic(err)
	}
	defer components.Cleanup()

	// Simulate queries
	queries := []struct {
		sql  string
		time float64
	}{
		{"SELECT * FROM feedback WHERE id = 1", 45.5},
		{"SELECT * FROM feedback WHERE id = 1", 42.3},
		{"SELECT * FROM accounts WHERE status = 'active'", 250.5},
		{"SELECT * FROM products WHERE category = 'billing'", 980.2},
		{"SELECT * FROM feedback WHERE sentiment < 0", 1200.3},
	}

	for _, q := range queries {
		// Start profiling
		metrics := components.QueryProfiler.StartQueryExecution(q.sql)
		
		// Simulate execution
		time.Sleep(time.Duration(q.time) * time.Millisecond)
		
		// Record completion
		components.QueryProfiler.RecordQueryExecution(
			metrics,
			100,      // rows returned
			1,        // pool usage
			false,    // cache hit
			nil,      // no error
		)
	}

	// Get report
	report := components.QueryProfiler.GetProfileReport()
	fmt.Printf("Total Queries: %d\n", report.TotalQueries)
	fmt.Printf("Unique Queries: %d\n", report.UniqueQueries)
	fmt.Printf("Slow Queries: %d\n", report.SlowQueryCount)
	fmt.Printf("Average Time: %.2fms\n", report.AvgExecTimeMS)
	fmt.Printf("Cache Hit Rate: %.1f%%\n", report.CacheHitRate)
}
```

Output:
```
Total Queries: 5
Unique Queries: 3
Slow Queries: 2
Average Time: 493.76ms
Cache Hit Rate: 0.0%
```

## Example 2: Slow Query Analysis

```go
package main

import (
	"fmt"

	"github.com/chuckie/goinsight/internal/profiler"
)

func main() {
	components, _ := profiler.InitializeProfiler(profiler.DefaultConfig())
	defer components.Cleanup()

	// Record slow queries
	slowQueries := []struct {
		query string
		time  float64
	}{
		{"SELECT * FROM feedback WHERE sentiment < 0", 850.5},
		{"SELECT * FROM feedback WHERE sentiment < 0", 920.2},
		{"SELECT * FROM feedback WHERE sentiment < 0", 1100.3},
		{"SELECT * FROM accounts WHERE region = 'US'", 650.5},
		{"SELECT * FROM products WHERE category = 'billing'", 550.2},
	}

	for _, sq := range slowQueries {
		metrics := components.QueryProfiler.StartQueryExecution(sq.query)
		// Record with execution time that exceeds threshold
		components.QueryProfiler.RecordQueryExecution(
			metrics, 1000, 1, false, nil,
		)
		components.SlowQueryLog.RecordSlowQuery(
			metrics.QueryID,
			sq.query,
			metrics.QueryHash,
			sq.time,
			500.0,  // threshold
			1000,
		)
	}

	// Analyze patterns
	analysis := components.SlowQueryLog.GetAnalysis(5, 3921.7)
	fmt.Printf("Total Slow Queries: %d\n", analysis.TotalSlowQueries)
	fmt.Printf("Unique Slow Queries: %d\n", analysis.UniqueSlowQueries)
	fmt.Printf("Slow Query Percentage: %.1f%%\n", analysis.SlowQueryPercentage)
	fmt.Printf("Most Frequent Query:\n")
	fmt.Printf("  Query: %s\n", analysis.MostFrequentQuery.Query)
	fmt.Printf("  Occurrences: %d\n", analysis.MostFrequentQuery.Occurrences)

	// Get top 5 slowest
	slowest := components.SlowQueryLog.GetSlowestQueries(5)
	fmt.Printf("\nTop Slowest Queries:\n")
	for i, sq := range slowest {
		fmt.Printf("%d. Execution: %.1fms (exceeded by %.1fms)\n",
			i+1, sq.ExecutionMS, sq.ExceededByMS)
	}
}
```

Output:
```
Total Slow Queries: 5
Unique Slow Queries: 2
Slow Query Percentage: 100.0%
Most Frequent Query:
  Query: SELECT * FROM feedback WHERE sentiment < 0
  Occurrences: 3

Top Slowest Queries:
1. Execution: 1100.30ms (exceeded by 600.30ms)
2. Execution: 920.20ms (exceeded by 420.20ms)
3. Execution: 850.50ms (exceeded by 350.50ms)
4. Execution: 650.50ms (exceeded by 150.50ms)
5. Execution: 550.20ms (exceeded by 50.20ms)
```

## Example 3: Query Optimization Suggestions

```go
package main

import (
	"fmt"

	"github.com/chuckie/goinsight/internal/profiler"
)

func main() {
	components, _ := profiler.InitializeProfiler(profiler.DefaultConfig())
	defer components.Cleanup()

	// Problematic queries
	queries := []string{
		"SELECT * FROM feedback",  // SELECT *
		"SELECT * FROM feedback WHERE sentiment < 0 OR priority = 'high'",  // OR condition
		"SELECT f.id FROM feedback f FULL OUTER JOIN accounts a ON f.account_id = a.id",  // FULL JOIN
		"SELECT * FROM feedback WHERE id IN (SELECT feedback_id FROM issues)",  // Subquery
	}

	optimizer := components.QueryOptimizer

	for _, query := range queries {
		suggestions := optimizer.AnalyzeQuery(query, nil)
		
		if len(suggestions) > 0 {
			fmt.Printf("Query: %s\n", query[:50]+"...")
			fmt.Printf("Suggestions:\n")
			for _, s := range suggestions {
				fmt.Printf("  [%s] %s\n", s.Severity, s.Title)
				fmt.Printf("    Impact: %.0f%%\n", s.ImpactScore)
				fmt.Printf("    Fix: %s\n", s.Suggestion)
			}
			fmt.Println()
		}
	}
}
```

Output:
```
Query: SELECT * FROM feedback
Suggestions:
  [MEDIUM] SELECT * Usage
    Impact: 15%
    Fix: Specify only the columns you need: SELECT col1, col2, col3 FROM table

Query: SELECT * FROM feedback WHERE sentiment < 0 OR priority = 'high'
Suggestions:
  [MEDIUM] OR Conditions in WHERE Clause
    Impact: 25%
    Fix: Rewrite: col IN (val1, val2, val3) instead of col = val1 OR col = val2 OR col = val3

Query: SELECT f.id FROM feedback f FULL OUTER JOIN accounts a ON f.account_id = a.id
Suggestions:
  [MEDIUM] FULL JOIN Usage
    Impact: 30%
    Fix: Consider using UNION of LEFT and RIGHT JOINs, or restructure query logic

Query: SELECT * FROM feedback WHERE id IN (SELECT feedback_id FROM issues)
Suggestions:
  [MEDIUM] Subquery Usage Detected
    Impact: 35%
    Fix: Consider using JOINs instead of subqueries, or use CTEs (WITH clause) for better readability
```

## Example 4: FeedbackService Integration

```go
package main

import (
	"context"
	"fmt"

	"github.com/chuckie/goinsight/internal/profiler"
	"github.com/chuckie/goinsight/internal/service"
	// other imports...
)

func main() {
	// Initialize profiler
	components, _ := profiler.InitializeProfiler(profiler.DefaultConfig())
	defer components.Cleanup()

	// Create service with profiler
	feedbackService := service.NewFeedbackServiceWithProfiler(
		repo,
		llmClient,
		jiraClient,
		components.Logger,
		components.QueryProfiler,
		components.SlowQueryLog,
		components.QueryOptimizer,
	)

	// Use service normally
	ctx := context.Background()
	response, err := feedbackService.AnalyzeFeedback(ctx, "What are the top issues?")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Analysis Result:\n")
	fmt.Printf("SQL: %s\n", response.SQL)
	fmt.Printf("Summary: %s\n", response.Summary)
	fmt.Printf("Recommendations: %v\n", response.Recommendations)

	// Get metrics
	report := feedbackService.GetProfileReport()
	if report != nil {
		fmt.Printf("\nPerformance Metrics:\n")
		fmt.Printf("Total Queries: %d\n", report.TotalQueries)
		fmt.Printf("Avg Time: %.2fms\n", report.AvgExecTimeMS)
		fmt.Printf("Cache Hit Rate: %.1f%%\n", report.CacheHitRate)
	}

	// Get optimization suggestions
	suggestions := feedbackService.GetOptimizationSuggestions()
	for hash, suggs := range suggestions {
		fmt.Printf("\nQuery %s needs optimization:\n", hash[:8])
		for _, s := range suggs {
			fmt.Printf("  [%s] %s (%.0f%% improvement)\n",
				s.Severity, s.Title, s.ImpactScore)
		}
	}

	// Get slow query analysis
	analysis := feedbackService.GetSlowQueryAnalysis()
	if analysis != nil {
		fmt.Printf("\nSlow Query Analysis:\n")
		fmt.Printf("Total Slow: %d\n", analysis.TotalSlowQueries)
		fmt.Printf("Percentage: %.1f%%\n", analysis.SlowQueryPercentage)
	}
}
```

## Example 5: Continuous Monitoring

```go
package main

import (
	"fmt"
	"time"

	"github.com/chuckie/goinsight/internal/profiler"
	"github.com/chuckie/goinsight/internal/service"
)

func main() {
	components, _ := profiler.InitializeProfiler(profiler.DefaultConfig())
	defer components.Cleanup()

	feedbackService := service.NewFeedbackServiceWithProfiler(
		repo, llmClient, jiraClient,
		components.Logger,
		components.QueryProfiler,
		components.SlowQueryLog,
		components.QueryOptimizer,
	)

	// Monitor every 5 minutes
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Get current metrics
		report := feedbackService.GetProfileReport()
		if report == nil {
			continue
		}

		// Log metrics
		components.Logger.Info("Performance snapshot", map[string]interface{}{
			"total_queries":    report.TotalQueries,
			"unique_queries":   report.UniqueQueries,
			"slow_queries":     report.SlowQueryCount,
			"avg_time_ms":      report.AvgExecTimeMS,
			"cache_hit_rate":   report.CacheHitRate,
			"error_rate":       report.ErrorRate,
		})

		// Check for slow queries
		slowAnalysis := feedbackService.GetSlowQueryAnalysis()
		if slowAnalysis != nil && slowAnalysis.TotalSlowQueries > 0 {
			fmt.Printf("[ALERT] %d slow queries detected (%.1f%%)\n",
				slowAnalysis.TotalSlowQueries,
				slowAnalysis.SlowQueryPercentage,
			)

			// Show top offenders
			topSlow := feedbackService.GetSlowestQueries(3)
			for i, sq := range topSlow {
				fmt.Printf("  %d. %s (%.1fms)\n", i+1, sq.Query[:40], sq.ExecutionMS)
			}
		}

		// Get optimization suggestions
		suggestions := feedbackService.GetOptimizationSuggestions()
		if len(suggestions) > 0 {
			fmt.Printf("[OPTIMIZATION] %d queries have suggestions\n", len(suggestions))
		}
	}
}
```

## Example 6: Custom Configuration

```go
package main

import (
	"fmt"

	"github.com/chuckie/goinsight/internal/profiler"
)

func main() {
	// Custom config for strict SLA requirements
	config := profiler.ProfilerConfig{
		LogDirectory:                 "./metrics",
		EnableConsoleLogging:         true,
		MinLogLevel:                  profiler.DEBUG,
		SlowQueryThresholdMS:         200.0,  // Strict 200ms SLA
		MaxMetricsPerQuery:           200,    // Keep more history
		PerformanceDegradationFactor: 1.15,   // Alert at 15% degradation
		WarningThresholdMS:           300.0,
	}

	components, err := profiler.InitializeProfiler(config)
	if err != nil {
		panic(err)
	}
	defer components.Cleanup()

	fmt.Printf("Profiler initialized with strict SLA:\n")
	fmt.Printf("Slow query threshold: 200ms\n")
	fmt.Printf("Degradation factor: 15%%\n")
	fmt.Printf("Logs in: %s\n", config.LogDirectory)
}
```

## Example 7: Report Generation

```go
package main

import (
	"fmt"

	"github.com/chuckie/goinsight/internal/profiler"
)

func main() {
	components, _ := profiler.InitializeProfiler(profiler.DefaultConfig())
	defer components.Cleanup()

	// Execute some queries (simulated)
	// ... queries execution code ...

	// Generate comprehensive report
	report := components.QueryProfiler.GetProfileReport()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("QUERY PERFORMANCE REPORT")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Printf("Report Generated: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Println("EXECUTION STATISTICS")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Printf("Total Queries Executed:    %d\n", report.TotalQueries)
	fmt.Printf("Unique Query Patterns:     %d\n", report.UniqueQueries)
	fmt.Printf("Total Execution Time:      %.2f ms\n", report.TotalExecTimeMS)
	fmt.Printf("Average Query Time:        %.2f ms\n", report.AvgExecTimeMS)
	fmt.Println()

	fmt.Println("QUALITY METRICS")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Printf("Cache Hit Rate:            %.1f%%\n", report.CacheHitRate)
	fmt.Printf("Error Rate:                %.1f%%\n", report.ErrorRate)
	fmt.Printf("Slow Query Count:          %d\n", report.SlowQueryCount)
	fmt.Printf("Slow Query Percentage:     %.1f%%\n", float64(report.SlowQueryCount)/float64(report.TotalQueries)*100)
	fmt.Println()

	fmt.Println("LOGS LOCATION")
	fmt.Println("───────────────────────────────────────────────────────")
	fmt.Printf("Application Log:           %s\n", components.Logger.GetLogPath())
	fmt.Printf("Slow Query Log:            %s\n", components.SlowQueryLog.GetLogPath())
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
}
```

Output:
```
═══════════════════════════════════════════════════════
QUERY PERFORMANCE REPORT
═══════════════════════════════════════════════════════
Report Generated: 2025-12-20 10:30:45

EXECUTION STATISTICS
───────────────────────────────────────────────────────
Total Queries Executed:    150
Unique Query Patterns:     23
Total Execution Time:      12456.78 ms
Average Query Time:        83.05 ms

QUALITY METRICS
───────────────────────────────────────────────────────
Cache Hit Rate:            42.3%
Error Rate:                2.1%
Slow Query Count:          12
Slow Query Percentage:     8.0%

LOGS LOCATION
───────────────────────────────────────────────────────
Application Log:           ./logs/profiler.log
Slow Query Log:            ./logs/slow_queries.log

═══════════════════════════════════════════════════════
```

## Summary

These examples demonstrate:

✅ **Basic setup and configuration**  
✅ **Slow query tracking and analysis**  
✅ **Automatic optimization suggestions**  
✅ **FeedbackService integration**  
✅ **Continuous monitoring patterns**  
✅ **Custom configuration for different SLAs**  
✅ **Report generation**  

Use these as templates for integrating the profiler into your application.
