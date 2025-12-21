package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chuckie/goinsight/internal/cache"
	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/internal/jira"
	"github.com/chuckie/goinsight/internal/llm"
	"github.com/chuckie/goinsight/internal/profiler"
	"github.com/chuckie/goinsight/internal/repository"
)

// FeedbackService orchestrates business logic for feedback analysis
// with integrated performance monitoring, caching, and optimization
type FeedbackService struct {
	repo           repository.FeedbackRepository
	llmClient      llm.Client
	jiraClient     *jira.Client
	cacheManager   *cache.CacheManager
	logger         *profiler.Logger
	queryProfiler  *profiler.QueryProfiler
	slowQueryLog   *profiler.SlowQueryLogger
	queryOptimizer *profiler.QueryOptimizer

	// Cache configuration
	cacheQueryResults bool
	queryResultsTTL   time.Duration
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
) *FeedbackService {
	return &FeedbackService{
		repo:       repo,
		llmClient:  llmClient,
		jiraClient: jiraClient,
	}
}

// NewFeedbackServiceWithProfiler creates a new feedback service with profiling enabled
func NewFeedbackServiceWithProfiler(
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
	logger *profiler.Logger,
	queryProfiler *profiler.QueryProfiler,
	slowQueryLog *profiler.SlowQueryLogger,
	queryOptimizer *profiler.QueryOptimizer,
) *FeedbackService {
	return &FeedbackService{
		repo:           repo,
		llmClient:      llmClient,
		jiraClient:     jiraClient,
		logger:         logger,
		queryProfiler:  queryProfiler,
		slowQueryLog:   slowQueryLog,
		queryOptimizer: queryOptimizer,
	}
}

// NewFeedbackServiceWithCache creates a feedback service with caching enabled
func NewFeedbackServiceWithCache(
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
	cacheManager *cache.CacheManager,
) *FeedbackService {
	return &FeedbackService{
		repo:                 repo,
		llmClient:            llmClient,
		jiraClient:           jiraClient,
		cacheManager:         cacheManager,
		cacheQueryResults:    true,
		queryResultsTTL:      5 * time.Minute,
	}
}

// NewFeedbackServiceFull creates a feedback service with all features enabled
func NewFeedbackServiceFull(
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
	logger *profiler.Logger,
	queryProfiler *profiler.QueryProfiler,
	slowQueryLog *profiler.SlowQueryLogger,
	queryOptimizer *profiler.QueryOptimizer,
	cacheManager *cache.CacheManager,
) *FeedbackService {
	return &FeedbackService{
		repo:                 repo,
		llmClient:            llmClient,
		jiraClient:           jiraClient,
		logger:               logger,
		queryProfiler:        queryProfiler,
		slowQueryLog:         slowQueryLog,
		queryOptimizer:       queryOptimizer,
		cacheManager:         cacheManager,
		cacheQueryResults:    true,
		queryResultsTTL:      5 * time.Minute,
	}
}

// SetCacheTTL configures the cache TTL for query results
func (fs *FeedbackService) SetCacheTTL(ttl time.Duration) {
	fs.queryResultsTTL = ttl
}

// CacheQueryResults enables/disables query result caching
func (fs *FeedbackService) CacheQueryResults(enabled bool) {
	fs.cacheQueryResults = enabled
}

// QueryRequest represents a question to analyze
type QueryRequest struct {
	Question string
}

// AnalyzeFeedback orchestrates the full workflow: SQL generation, execution, and insight generation
// with integrated query profiling and performance monitoring
func (s *FeedbackService) AnalyzeFeedback(ctx context.Context, question string) (*domain.AskResponse, error) {
	// Validate input
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	// Step 0: Check cache for previously analyzed questions
	// (Cache is keyed by question text to allow caching of full insights)
	if s.cacheManager != nil && s.cacheQueryResults {
		cachedResponse, found, err := s.cacheManager.GetCachedQueryResult(ctx, question)
		if err == nil && found {
			// Cache hit - return cached response
			if response, ok := cachedResponse.(*domain.AskResponse); ok {
				return response, nil
			}
		}
	}

	// Step 1: Generate SQL from the question
	sqlQuery, err := s.llmClient.GenerateSQL(ctx, question)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to generate SQL", err, map[string]interface{}{
				"question": question,
			})
		}
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Step 2: Check cache for SQL query results (if different question generates same SQL)
	var queryResults []map[string]interface{}
	var metrics *profiler.QueryMetrics
	cachedResults := false

	if s.cacheManager != nil && s.cacheQueryResults {
		cachedData, found, err := s.cacheManager.GetCachedQueryResult(ctx, sqlQuery)
		if err == nil && found {
			if results, ok := cachedData.([]map[string]interface{}); ok {
				queryResults = results
				cachedResults = true
			}
		}
	}

	// Step 3: Validate SQL for safety
	if err := s.validateSQL(sqlQuery); err != nil {
		if s.logger != nil {
			s.logger.Warn("SQL validation failed", map[string]interface{}{
				"question": question,
				"sql":      sqlQuery,
				"error":    err.Error(),
			})
		}
		return nil, err
	}

	// Step 4: Execute the SQL query with profiling (if not cached)
	if !cachedResults {
		if s.queryProfiler != nil {
			metrics = s.queryProfiler.StartQueryExecution(sqlQuery)
		}

		queryResults, err = s.repo.QueryFeedback(ctx, sqlQuery)

		if s.queryProfiler != nil && metrics != nil {
			rowsReturned := int64(len(queryResults))
			poolUsage := 1 // Default value, can be enhanced with actual pool metrics
			s.queryProfiler.RecordQueryExecution(metrics, rowsReturned, poolUsage, false, err)

			// Check if query is slow and log accordingly
			execTimeMS := metrics.ExecutionTime.Seconds() * 1000
			if s.slowQueryLog != nil && execTimeMS > 500 {
				s.slowQueryLog.RecordSlowQuery(
					metrics.QueryID,
					sqlQuery,
					metrics.QueryHash,
					execTimeMS,
					500.0,
					rowsReturned,
				)
			}

			// Generate optimization suggestions if query is slow
			if s.queryOptimizer != nil && execTimeMS > 500 {
				stats := s.queryProfiler.GetStats(metrics.QueryHash)
				suggestions := s.queryOptimizer.AnalyzeQuery(sqlQuery, stats)

				if len(suggestions) > 0 && s.logger != nil {
					s.logger.Debug("Query optimization suggestions", map[string]interface{}{
						"query_id":     metrics.QueryID,
						"execution_ms": execTimeMS,
						"suggestions":  len(suggestions),
					})
				}
			}
		}

		if err != nil {
			if s.logger != nil {
				s.logger.Error("Query execution failed", err, map[string]interface{}{
					"sql": sqlQuery,
				})
			}
			return nil, fmt.Errorf("query execution failed: %w", err)
		}

		// Cache query results for future use
		if s.cacheManager != nil && s.cacheQueryResults {
			_ = s.cacheManager.CacheQueryResult(ctx, sqlQuery, queryResults, s.queryResultsTTL)
		}
	}

	// Step 5: Generate insights from the results
	insightJSON, err := s.llmClient.GenerateInsight(ctx, question, queryResults)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to generate insights", err, map[string]interface{}{
				"question": question,
				"results":  len(queryResults),
			})
		}
		return nil, fmt.Errorf("failed to generate insights: %w", err)
	}

	// Step 5: Parse the insight JSON
	var insightResult domain.InsightResult
	if err := json.Unmarshal([]byte(insightJSON), &insightResult); err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to parse insight response", err, map[string]interface{}{
				"question": question,
			})
		}
		return nil, fmt.Errorf("failed to parse insight response: %w", err)
	}

	// Step 6: Build the response
	dataPreview := queryResults
	if len(dataPreview) > 10 {
		dataPreview = dataPreview[:10]
	}

	response := &domain.AskResponse{
		Question:        question,
		SQL:             sqlQuery,
		DataPreview:     dataPreview,
		Summary:         insightResult.Summary,
		Recommendations: insightResult.Recommendations,
		Actions:         insightResult.Actions,
	}

	// Cache the complete response for future identical questions
	if s.cacheManager != nil && s.cacheQueryResults {
		_ = s.cacheManager.CacheQueryResult(ctx, question, response, s.queryResultsTTL)
	}

	if s.logger != nil {
		execTimeMs := 0.0
		if metrics != nil {
			execTimeMs = metrics.ExecutionTime.Seconds() * 1000
		}
		s.logger.Info("Feedback analysis completed", map[string]interface{}{
			"question":    question,
			"results":     len(queryResults),
			"actions":     len(insightResult.Actions),
			"exec_time_ms": execTimeMs,
		})
	}

	return response, nil
}

// validateSQL performs safety checks on the generated SQL query
func (s *FeedbackService) validateSQL(sqlQuery string) error {
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))

	// Ensure it's only a SELECT
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		return fmt.Errorf(
			"unable to generate a valid data query. "+
				"This API analyzes customer feedback data. "+
				"Please ask questions about feedback, such as: "+
				"'What are the most common billing issues?' or "+
				"'Show me negative feedback from enterprise customers.'",
		)
	}

	// Check for dangerous SQL statement keywords as standalone words
	dangerous := []string{
		"\\bDROP\\b", "\\bDELETE\\b", "\\bINSERT\\b", "\\bUPDATE\\b",
		"\\bALTER\\b", "\\bCREATE\\s+TABLE\\b", "\\bCREATE\\s+INDEX\\b",
		"\\bTRUNCATE\\b", "\\bEXEC\\b", "\\bEXECUTE\\b",
	}

	for _, pattern := range dangerous {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("internal error compiling SQL validation pattern: %w", err)
		}

		if re.MatchString(normalizedSQL) {
			return fmt.Errorf("generated query contains forbidden SQL statement")
		}
	}

	return nil
}

// GetAccountRiskScore retrieves ML predictions for a specific account
func (s *FeedbackService) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
	return s.repo.GetAccountRiskScore(ctx, accountID)
}

// GetRecentNegativeFeedbackCount counts recent negative feedback for an account
func (s *FeedbackService) GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error) {
	return s.repo.GetRecentNegativeFeedbackCount(ctx, accountID)
}

// GetProductAreaImpacts retrieves ML predictions for product area priorities
func (s *FeedbackService) GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error) {
	return s.repo.GetProductAreaImpacts(ctx, segment)
}

// JiraTicketRequest wraps the action items and metadata for Jira ticket creation
type JiraTicketRequest struct {
	Summary         string
	Recommendations []string
	Actions         []domain.ActionItem
	Meta            JiraMetadata
}

// JiraMetadata holds Jira-specific configuration
type JiraMetadata struct {
	ProjectKey       string
	DefaultIssueType string
	DefaultLabels    []string
}

// CreateJiraTickets converts insight actions into Jira tickets
func (s *FeedbackService) CreateJiraTickets(ctx context.Context, req JiraTicketRequest) (*domain.JiraCreationResult, error) {
	// Validate Jira is configured
	if s.jiraClient == nil {
		return nil, fmt.Errorf("jira integration is not configured")
	}

	// Validate request
	if len(req.Actions) == 0 {
		return nil, fmt.Errorf("no actions provided to convert into tickets")
	}

	// Validate required Jira meta and set defaults
	if strings.TrimSpace(req.Meta.ProjectKey) == "" {
		return nil, fmt.Errorf("jira project key is required")
	}
	if req.Meta.DefaultIssueType == "" {
		req.Meta.DefaultIssueType = "Story"
	}
	if len(req.Meta.DefaultLabels) == 0 {
		req.Meta.DefaultLabels = []string{"feedback", "ai-insight"}
	}

	// Calculate magnitude for each action item
	for i := range req.Actions {
		req.Actions[i].Magnitude = domain.CalculateMagnitude(
			req.Actions[i],
			req.Summary,
			req.Recommendations,
		)
	}

	// Convert request to JSON for LLM prompt
	domainReq := domain.JiraTicketRequest{
		Summary:         req.Summary,
		Recommendations: req.Recommendations,
		Actions:         req.Actions,
		Meta: domain.JiraTicketMeta{
			ProjectKey:       req.Meta.ProjectKey,
			DefaultIssueType: req.Meta.DefaultIssueType,
			DefaultLabels:    req.Meta.DefaultLabels,
		},
	}

	requestJSON, err := json.MarshalIndent(domainReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Use LLM to generate Jira ticket specifications
	prompt := llm.JiraTicketPrompt(string(requestJSON))
	ticketsJSON, err := s.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ticket specs: %w", err)
	}

	// Strip markdown code fences if present
	ticketsJSON = strings.TrimSpace(ticketsJSON)
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```json")
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSuffix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSpace(ticketsJSON)

	// Parse the ticket specifications
	var ticketsResp domain.JiraTicketsResponse
	if err := json.Unmarshal([]byte(ticketsJSON), &ticketsResp); err != nil {
		return nil, fmt.Errorf("failed to parse ticket specs: %w", err)
	}

	if len(ticketsResp.Tickets) == 0 {
		return nil, fmt.Errorf("LLM did not generate any ticket specifications")
	}

	// Create tickets in Jira
	result, err := s.jiraClient.CreateIssues(ticketsResp.Tickets)
	if err != nil {
		return nil, fmt.Errorf("failed to create jira tickets: %w", err)
	}

	return result, nil
}

// GetProfileReport returns the current profiling metrics and statistics
func (s *FeedbackService) GetProfileReport() *profiler.ProfileReport {
	if s.queryProfiler == nil {
		return nil
	}
	report := s.queryProfiler.GetProfileReport()
	return &report
}

// GetSlowQueryAnalysis returns analysis of detected slow queries
func (s *FeedbackService) GetSlowQueryAnalysis() *profiler.SlowQueryAnalysis {
	if s.slowQueryLog == nil {
		return nil
	}

	report := s.GetProfileReport()
	if report == nil {
		return nil
	}

	analysis := s.slowQueryLog.GetAnalysis(report.TotalQueries, report.TotalExecTimeMS)
	return &analysis
}

// GetOptimizationSuggestions returns optimization suggestions for slow queries
func (s *FeedbackService) GetOptimizationSuggestions() map[string][]profiler.OptimizationSuggestion {
	if s.queryOptimizer == nil || s.queryProfiler == nil {
		return nil
	}

	stats := s.queryProfiler.GetAggregateStats()
	suggestions := make(map[string][]profiler.OptimizationSuggestion)

	for hash, stat := range stats {
		suggestionList := s.queryOptimizer.AnalyzeQuery(stat.Query, stat)
		if len(suggestionList) > 0 {
			suggestions[hash] = suggestionList
		}
	}

	return suggestions
}

// GetMostFrequentSlowQueries returns the most frequently occurring slow queries
func (s *FeedbackService) GetMostFrequentSlowQueries(limit int) []*profiler.SlowQueryEntry {
	if s.slowQueryLog == nil {
		return nil
	}
	return s.slowQueryLog.GetMostFrequentSlowQueries(limit)
}

// GetSlowestQueries returns the queries with highest execution times
func (s *FeedbackService) GetSlowestQueries(limit int) []*profiler.SlowQueryEntry {
	if s.slowQueryLog == nil {
		return nil
	}
	return s.slowQueryLog.GetSlowestQueries(limit)
}

// ResetProfiler clears all profiling data
func (s *FeedbackService) ResetProfiler() {
	if s.queryProfiler != nil {
		s.queryProfiler.Reset()
	}
}
// === Cache Management Methods ===

// GetCacheStats returns current cache statistics
func (s *FeedbackService) GetCacheStats(ctx context.Context) cache.CacheStats {
	if s.cacheManager == nil {
		return cache.CacheStats{}
	}
	return s.cacheManager.GetCacheStats(ctx)
}

// ClearCache removes all cached entries
func (s *FeedbackService) ClearCache(ctx context.Context) error {
	if s.cacheManager == nil {
		return nil
	}
	return s.cacheManager.ClearCache(ctx)
}

// InvalidateQueryCache removes a cached query result
func (s *FeedbackService) InvalidateQueryCache(ctx context.Context, query string) error {
	if s.cacheManager == nil {
		return nil
	}
	return s.cacheManager.InvalidateQuery(ctx, query)
}

// InvalidateCachePattern removes cached entries matching a pattern
// Useful for: Invalidating all queries involving a table after updates
func (s *FeedbackService) InvalidateCachePattern(ctx context.Context, pattern string) error {
	if s.cacheManager == nil {
		return nil
	}
	return s.cacheManager.InvalidatePattern(ctx, pattern)
}

// IsCacheEnabled checks if caching is currently enabled
func (s *FeedbackService) IsCacheEnabled() bool {
	return s.cacheManager != nil && s.cacheQueryResults
}