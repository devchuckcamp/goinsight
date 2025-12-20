package profiler

import (
	"fmt"
	"regexp"
	"strings"
)

// OptimizationSuggestion represents a suggested optimization for a query
type OptimizationSuggestion struct {
	Type        SuggestionType `json:"type"`
	Severity    Severity       `json:"severity"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Suggestion  string         `json:"suggestion"`
	Columns     []string       `json:"columns,omitempty"`
	ImpactScore float64        `json:"impact_score"` // 0-100 estimated performance improvement
}

// SuggestionType categorizes the type of optimization
type SuggestionType string

const (
	MissingIndex      SuggestionType = "missing_index"
	IndexOptimization SuggestionType = "index_optimization"
	QueryRewrite      SuggestionType = "query_rewrite"
	JoinOptimization  SuggestionType = "join_optimization"
	WhereClauseIssue  SuggestionType = "where_clause_issue"
	FullTableScan     SuggestionType = "full_table_scan"
	NSubqueryUsage    SuggestionType = "n_subquery_usage"
	UnusedColumn      SuggestionType = "unused_column"
)

// Severity indicates the priority of the suggestion
type Severity string

const (
	Critical Severity = "critical"
	High     Severity = "high"
	Medium   Severity = "medium"
	Low      Severity = "low"
)

// QueryOptimizer analyzes queries and suggests optimizations
// Algorithms:
// 1. Pattern Matching: Uses regex to identify common inefficiencies
// 2. Heuristic Analysis: Applies rules based on query characteristics
// 3. Statistical Analysis: Considers execution metrics in recommendations
// 4. Index Recommendation: Suggests indexes based on WHERE/JOIN columns
type QueryOptimizer struct {
	// Compiled regex patterns for query analysis
	selectPattern      *regexp.Regexp
	fromPattern        *regexp.Regexp
	wherePattern       *regexp.Regexp
	joinPattern        *regexp.Regexp
	groupByPattern     *regexp.Regexp
	orderByPattern     *regexp.Regexp
	subqueryPattern    *regexp.Regexp
	wildCardPattern    *regexp.Regexp
	fullJoinPattern    *regexp.Regexp
	orConditionPattern *regexp.Regexp
}

// NewQueryOptimizer creates a new QueryOptimizer instance
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{
		selectPattern:      regexp.MustCompile(`(?i)SELECT\s+(.*?)\s+FROM`),
		fromPattern:        regexp.MustCompile(`(?i)FROM\s+(\w+)`),
		wherePattern:       regexp.MustCompile(`(?i)WHERE\s+(.+?)(?:GROUP BY|ORDER BY|LIMIT|$)`),
		joinPattern:        regexp.MustCompile(`(?i)(INNER\s+JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|CROSS\s+JOIN|FULL\s+JOIN)\s+(\w+)\s+ON\s+(.+?)(?:WHERE|GROUP BY|ORDER BY|LIMIT|$)`),
		groupByPattern:     regexp.MustCompile(`(?i)GROUP\s+BY\s+(.+?)(?:HAVING|ORDER BY|LIMIT|$)`),
		orderByPattern:     regexp.MustCompile(`(?i)ORDER\s+BY\s+(.+?)(?:LIMIT|$)`),
		subqueryPattern:    regexp.MustCompile(`(?i)\(\s*SELECT\s+.+?\s+\)`),
		wildCardPattern:    regexp.MustCompile(`(?i)SELECT\s+\*`),
		fullJoinPattern:    regexp.MustCompile(`(?i)FULL\s+(OUTER\s+)?JOIN`),
		orConditionPattern: regexp.MustCompile(`(?i)\s+OR\s+`),
	}
}

// AnalyzeQuery examines a query and returns optimization suggestions
// Algorithm:
// 1. Parse query components (SELECT, FROM, WHERE, etc.)
// 2. Apply pattern-matching rules to detect inefficiencies
// 3. Analyze execution metrics (execution time, rows)
// 4. Generate prioritized recommendations
func (qo *QueryOptimizer) AnalyzeQuery(query string, stats *QueryStats) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Normalize query for analysis
	normalizedQuery := strings.ToUpper(strings.TrimSpace(query))

	// Check for wildcard selections
	if qo.wildCardPattern.MatchString(normalizedQuery) {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        UnusedColumn,
			Severity:    Medium,
			Title:       "SELECT * Usage",
			Description: "Using SELECT * can include unnecessary columns, increasing data transfer and I/O",
			Suggestion:  "Specify only the columns you need: SELECT col1, col2, col3 FROM table",
			ImpactScore: 15.0,
		})
	}

	// Check for missing indexes on WHERE clause columns
	whereMatches := qo.wherePattern.FindStringSubmatch(normalizedQuery)
	if len(whereMatches) > 1 {
		whereClause := whereMatches[1]
		columns := qo.extractColumnsFromClause(whereClause)
		if len(columns) > 0 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        MissingIndex,
				Severity:    High,
				Title:       "Missing Index on WHERE Clause",
				Description: fmt.Sprintf("Columns in WHERE clause (%s) should have indexes for faster lookups", strings.Join(columns, ", ")),
				Suggestion:  fmt.Sprintf("Consider creating indexes: CREATE INDEX idx_name ON table(%s)", strings.Join(columns, ", ")),
				Columns:     columns,
				ImpactScore: 40.0,
			})
		}

		// Check for OR conditions that might prevent index usage
		if qo.orConditionPattern.MatchString(whereClause) {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        QueryRewrite,
				Severity:    Medium,
				Title:       "OR Conditions in WHERE Clause",
				Description: "OR conditions can prevent efficient index usage. Consider rewriting with IN clause or UNION",
				Suggestion:  "Rewrite: col IN (val1, val2, val3) instead of col = val1 OR col = val2 OR col = val3",
				ImpactScore: 25.0,
			})
		}
	}

	// Check for missing indexes on JOIN columns
	joinMatches := qo.joinPattern.FindAllStringSubmatch(normalizedQuery, -1)
	for _, match := range joinMatches {
		if len(match) > 3 {
			joinCondition := match[3]
			columns := qo.extractColumnsFromClause(joinCondition)
			if len(columns) > 0 {
				suggestions = append(suggestions, OptimizationSuggestion{
					Type:        MissingIndex,
					Severity:    High,
					Title:       "Missing Index on JOIN Column",
					Description: fmt.Sprintf("JOIN condition uses columns (%s) that should be indexed", strings.Join(columns, ", ")),
					Suggestion:  fmt.Sprintf("Create indexes on join columns: CREATE INDEX idx_name ON table(%s)", strings.Join(columns, ", ")),
					Columns:     columns,
					ImpactScore: 50.0,
				})
			}
		}
	}

	// Check for FULL JOIN (can be expensive)
	if qo.fullJoinPattern.MatchString(normalizedQuery) {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        JoinOptimization,
			Severity:    Medium,
			Title:       "FULL JOIN Usage",
			Description: "FULL OUTER JOIN can be expensive and may prevent index usage",
			Suggestion:  "Consider using UNION of LEFT and RIGHT JOINs, or restructure query logic",
			ImpactScore: 30.0,
		})
	}

	// Check for subqueries (especially in SELECT or WHERE)
	if qo.subqueryPattern.MatchString(normalizedQuery) {
		suggestions = append(suggestions, OptimizationSuggestion{
			Type:        NSubqueryUsage,
			Severity:    Medium,
			Title:       "Subquery Usage Detected",
			Description: "Subqueries, especially correlated subqueries, can be inefficient and executed multiple times",
			Suggestion:  "Consider using JOINs instead of subqueries, or use CTEs (WITH clause) for better readability",
			ImpactScore: 35.0,
		})
	}

	// Analyze execution metrics if available
	if stats != nil {
		// High execution time analysis
		if stats.AvgExecTimeMS > 1000 {
			suggestions = append(suggestions, OptimizationSuggestion{
				Type:        FullTableScan,
				Severity:    Critical,
				Title:       "Slow Query Execution",
				Description: fmt.Sprintf("Query averages %.2fms execution time. Likely performing full table scans", stats.AvgExecTimeMS),
				Suggestion:  "Review indexes on frequently queried columns and consider query rewriting",
				ImpactScore: 60.0,
			})
		}

		// Error rate analysis
		if stats.ErrorCount > 0 {
			errorRate := float64(stats.ErrorCount) / float64(stats.ExecutionCount) * 100
			if errorRate > 5 {
				suggestions = append(suggestions, OptimizationSuggestion{
					Type:        QueryRewrite,
					Severity:    High,
					Title:       fmt.Sprintf("High Error Rate (%.1f%%)", errorRate),
					Description: "Query has a high error rate which may indicate data quality or schema issues",
					Suggestion:  "Review query logic and data validation. Check for NULL handling and type mismatches",
					ImpactScore: 20.0,
				})
			}
		}
	}

	return suggestions
}

// extractColumnsFromClause extracts column names from a WHERE or JOIN condition
// Algorithm: Pattern matching to identify column references (table.column or column)
func (qo *QueryOptimizer) extractColumnsFromClause(clause string) []string {
	columnPattern := regexp.MustCompile(`(\w+)\.(\w+)|\b(\w+)\s*[=<>]`)
	matches := columnPattern.FindAllStringSubmatch(clause, -1)

	columnMap := make(map[string]bool)
	var columns []string

	for _, match := range matches {
		var column string
		if match[1] != "" && match[2] != "" {
			// table.column format
			column = match[2]
		} else if match[3] != "" {
			// column format
			column = match[3]
		}

		if column != "" && !columnMap[column] {
			columnMap[column] = true
			columns = append(columns, column)
		}
	}

	return columns
}

// GetSuggestionsByType filters suggestions by type
func FilterSuggestionsByType(suggestions []OptimizationSuggestion, suggestionType SuggestionType) []OptimizationSuggestion {
	var filtered []OptimizationSuggestion
	for _, s := range suggestions {
		if s.Type == suggestionType {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GetSuggestionsBySeverity filters suggestions by severity level
func FilterSuggestionsBySeverity(suggestions []OptimizationSuggestion, severity Severity) []OptimizationSuggestion {
	var filtered []OptimizationSuggestion
	for _, s := range suggestions {
		if s.Severity == severity {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// SortSuggestionsByImpact sorts suggestions by estimated impact score (descending)
func SortSuggestionsByImpact(suggestions []OptimizationSuggestion) []OptimizationSuggestion {
	// Create a copy to avoid modifying the original
	sorted := make([]OptimizationSuggestion, len(suggestions))
	copy(sorted, suggestions)

	// Simple bubble sort for clarity (in production, use sort.Slice)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].ImpactScore > sorted[i].ImpactScore {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// GenerateOptimizationReport creates a comprehensive optimization report
type OptimizationReport struct {
	Query        string
	Suggestions  []OptimizationSuggestion
	TotalImpact  float64 // Sum of impact scores
	CriticalCount int
	HighCount    int
	MediumCount  int
	LowCount     int
}

func (qo *QueryOptimizer) GenerateReport(query string, stats *QueryStats) OptimizationReport {
	suggestions := qo.AnalyzeQuery(query, stats)

	report := OptimizationReport{
		Query:       query,
		Suggestions: suggestions,
	}

	for _, s := range suggestions {
		report.TotalImpact += s.ImpactScore
		switch s.Severity {
		case Critical:
			report.CriticalCount++
		case High:
			report.HighCount++
		case Medium:
			report.MediumCount++
		case Low:
			report.LowCount++
		}
	}

	return report
}
