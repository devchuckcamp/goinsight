# Design Patterns - Usage Examples

This document provides practical examples for using the new design patterns implemented in GoInsight.

## Table of Contents
1. [Repository Pattern Examples](#repository-pattern-examples)
2. [Service Layer Examples](#service-layer-examples)
3. [Builder Pattern Examples](#builder-pattern-examples)
4. [Middleware Examples](#middleware-examples)
5. [Integration Examples](#integration-examples)

---

## Repository Pattern Examples

### Basic Query Execution

```go
package main

import (
    "context"
    "database/sql"
    "github.com/chuckie/goinsight/internal/repository"
)

func main() {
    db, _ := sql.Open("postgres", "postgresql://user:pass@localhost/goinsight")
    
    // Create repository
    repo := repository.NewPostgresFeedbackRepository(db)
    
    // Execute query
    ctx := context.Background()
    results, err := repo.QueryFeedback(ctx, "SELECT * FROM feedback_enriched LIMIT 10")
    if err != nil {
        panic(err)
    }
    
    // Use results
    for _, row := range results {
        fmt.Printf("%+v\n", row)
    }
}
```

### Account Risk Scoring

```go
func getAccountHealth(repo repository.FeedbackRepository, accountID string) {
    ctx := context.Background()
    
    // Get account risk score
    score, err := repo.GetAccountRiskScore(ctx, accountID)
    if err != nil {
        log.Fatal(err)
    }
    
    if score == nil {
        fmt.Println("Account not found")
        return
    }
    
    fmt.Printf("Account: %s\n", score.AccountID)
    fmt.Printf("Churn Probability: %.2f%%\n", score.ChurnProbability*100)
    fmt.Printf("Health Score: %d\n", score.HealthScore)
    fmt.Printf("Risk Category: %s\n", score.RiskCategory)
}
```

### Product Area Analysis

```go
func analyzeProductAreas(repo repository.FeedbackRepository) {
    ctx := context.Background()
    
    // Get impacts for enterprise segment
    impacts, err := repo.GetProductAreaImpacts(ctx, "enterprise")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Product Area Priorities (Enterprise):")
    for _, impact := range impacts {
        fmt.Printf("  %v\n", impact)
    }
    
    // Or get all segments
    allImpacts, _ := repo.GetProductAreaImpacts(ctx, "")
    fmt.Printf("Total areas across all segments: %d\n", len(allImpacts))
}
```

### Implementing a Custom Repository

```go
package repository

import (
    "context"
    "github.com/chuckie/goinsight/internal/domain"
)

// RedisRepository implements caching layer on top of database
type RedisRepository struct {
    db    FeedbackRepository
    cache *redis.Client
}

func NewRedisRepository(db FeedbackRepository, cache *redis.Client) *RedisRepository {
    return &RedisRepository{db: db, cache: cache}
}

func (r *RedisRepository) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
    // Try cache first
    cached, err := r.cache.Get(ctx, "account:"+accountID).Result()
    if err == nil {
        var score domain.AccountRiskScore
        json.Unmarshal([]byte(cached), &score)
        return &score, nil
    }
    
    // Fall back to database
    score, err := r.db.GetAccountRiskScore(ctx, accountID)
    if err == nil && score != nil {
        // Cache for 1 hour
        data, _ := json.Marshal(score)
        r.cache.Set(ctx, "account:"+accountID, string(data), time.Hour)
    }
    
    return score, err
}
```

---

## Service Layer Examples

### Complete Feedback Analysis Workflow

```go
package main

import (
    "context"
    "github.com/chuckie/goinsight/internal/service"
)

func analyzeFeedbackComplete(feedbackService *service.FeedbackService) {
    ctx := context.Background()
    
    // Ask a question about feedback
    question := "What are the most common issues reported by enterprise customers in the past 30 days?"
    
    response, err := feedbackService.AnalyzeFeedback(ctx, question)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Question:", response.Question)
    fmt.Println("Generated SQL:", response.SQL)
    fmt.Println("Data Preview:", response.DataPreview)
    fmt.Println("Summary:", response.Summary)
    fmt.Println("Recommendations:")
    for _, rec := range response.Recommendations {
        fmt.Printf("  - %s\n", rec)
    }
    fmt.Println("Proposed Actions:")
    for _, action := range response.Actions {
        fmt.Printf("  - %s: %s (Magnitude: %.2f)\n", action.Title, action.Description, action.Magnitude)
    }
}
```

### Creating Jira Tickets from Insights

```go
func createJiraTickets(feedbackService *service.FeedbackService) {
    ctx := context.Background()
    
    req := service.JiraTicketRequest{
        Summary:         "Customer billing issues causing churn",
        Recommendations: []string{"Fix payment processing", "Improve error messages"},
        Actions: []domain.ActionItem{
            {
                Title:       "Investigate payment failures",
                Description: "Debug why payments are failing for enterprise customers",
            },
            {
                Title:       "Improve error handling",
                Description: "Add better error messages for payment issues",
            },
        },
        Meta: service.JiraMetadata{
            ProjectKey:       "GOI",
            DefaultIssueType: "Task",
            DefaultLabels:    []string{"feedback", "urgent"},
        },
    }
    
    result, err := feedbackService.CreateJiraTickets(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created %d Jira tickets\n", len(result.Issues))
    for _, issue := range result.Issues {
        fmt.Printf("  - %s: %s\n", issue.Key, issue.Summary)
    }
}
```

### Service Composition

```go
func setupServices(db *sql.DB, llmClient llm.Client, jiraClient *jira.Client) *service.FeedbackService {
    // Create repository
    repo := repository.NewPostgresFeedbackRepository(db)
    
    // Optionally wrap with caching
    // cachedRepo := repository.NewRedisRepository(repo, redisClient)
    
    // Create service with repository
    feedbackService := service.NewFeedbackService(repo, llmClient, jiraClient)
    
    return feedbackService
}
```

---

## Builder Pattern Examples

### Simple Feedback Query

```go
package main

import "github.com/chuckie/goinsight/internal/builder"

func buildSimpleQuery() {
    query := builder.NewQueryBuilder().
        Select("id", "sentiment", "priority", "created_at").
        From("feedback_enriched").
        Where("sentiment = 'negative'").
        OrderBy("created_at", "DESC").
        Limit(20).
        Build()
    
    fmt.Println(query)
    // Output:
    // SELECT id, sentiment, priority, created_at
    // FROM feedback_enriched
    // WHERE sentiment = 'negative'
    // ORDER BY created_at DESC
    // LIMIT 20
}
```

### Specialized Feedback Query Builder

```go
func buildSpecializedQuery() {
    query := builder.NewFeedbackQueryBuilder().
        Select("id", "topic", "sentiment", "customer_tier").
        WithSentiment("negative").
        WithProductArea("billing").
        WithRegion("US").
        WithMinPriority(2).
        OrderBy("priority", "DESC").
        Limit(10).
        BuildFeedback()
    
    fmt.Println(query)
    // Output:
    // SELECT id, topic, sentiment, customer_tier
    // FROM feedback_enriched
    // WHERE sentiment = 'negative'
    // AND product_area = 'billing'
    // AND region = 'US'
    // AND priority >= 2
    // ORDER BY priority DESC
    // LIMIT 10
}
```

### Conditional Filters

```go
func buildConditionalQuery(filters map[string]interface{}) string {
    builder := builder.NewQueryBuilder().
        Select("*").
        From("feedback_enriched")
    
    // Conditionally add filters
    if sentiment, ok := filters["sentiment"].(string); ok && sentiment != "" {
        builder.Where("sentiment = '" + sentiment + "'")
    }
    
    if minPriority, ok := filters["minPriority"].(int); ok && minPriority > 0 {
        builder.WhereIf(minPriority > 0, "priority >= "+strconv.Itoa(minPriority))
    }
    
    if limit, ok := filters["limit"].(int); ok && limit > 0 {
        builder.Limit(limit)
    }
    
    return builder.Build()
}
```

### Pagination

```go
func paginateFeedback(pageNum, pageSize int) string {
    offset := (pageNum - 1) * pageSize
    
    return builder.NewQueryBuilder().
        Select("*").
        From("feedback_enriched").
        OrderBy("created_at", "DESC").
        Limit(pageSize).
        Offset(offset).
        Build()
    
    // Page 1: LIMIT 10 OFFSET 0
    // Page 2: LIMIT 10 OFFSET 10
    // Page 3: LIMIT 10 OFFSET 20
}
```

### Reusable Query Templates

```go
func getCommonQueries() {
    // Template 1: Recent negative feedback
    negativeQuery := builder.NewFeedbackQueryBuilder().
        WithSentiment("negative").
        OrderBy("created_at", "DESC").
        Limit(50).
        BuildFeedback()
    
    // Template 2: High priority issues
    highPriorityQuery := builder.NewFeedbackQueryBuilder().
        WithMinPriority(4).
        OrderBy("priority", "DESC").
        OrderBy("created_at", "DESC").
        BuildFeedback()
    
    // Template 3: Product area analysis
    productAreaQuery := builder.NewFeedbackQueryBuilder().
        Select("product_area", "COUNT(*)", "AVG(priority)").
        OrderBy("product_area", "ASC").
        BuildFeedback()
    
    return map[string]string{
        "recent_negative":   negativeQuery,
        "high_priority":     highPriorityQuery,
        "product_analysis":  productAreaQuery,
    }
}
```

---

## Middleware Examples

### Global Middleware Setup

```go
package http

import (
    "github.com/chuckie/goinsight/internal/http/middleware"
    "github.com/go-chi/chi/v5"
)

func setupMiddleware(router *chi.Mux) {
    // Add in order of execution
    router.Use(middleware.RecoveryMiddleware)           // Outermost - catch panics
    router.Use(middleware.LoggingMiddleware)            // Log all requests
    router.Use(middleware.TimingMiddleware)             // Measure duration
    router.Use(middleware.ValidateJSONMiddleware)       // Validate content type
}
```

### Custom Middleware

```go
// Add authentication middleware
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte(`{"error":"Missing authorization token"}`))
            return
        }
        
        // Validate token
        // ...
        
        next.ServeHTTP(w, r)
    })
}

// Add to router
router.Use(AuthMiddleware)
```

### Query Performance Monitoring

```go
func monitorQueryPerformance(repo repository.FeedbackRepository, query string) {
    ctx := context.Background()
    decorator := middleware.NewQueryExecutionDecorator()
    
    start := time.Now()
    results, err := repo.QueryFeedback(ctx, query)
    duration := time.Since(start)
    
    // Record metrics
    decorator.RecordExecution(duration, len(results), len(query))
    metrics := decorator.GetMetrics()
    
    fmt.Printf("Query Metrics: %+v\n", metrics)
    
    // Check for slow queries
    slowQueryChecker := middleware.SlowQueryThreshold(100 * time.Millisecond)
    slowQueryChecker(duration, query)
}
```

---

## Integration Examples

### Complete Application Setup

```go
package main

import (
    "database/sql"
    "github.com/chuckie/goinsight/internal/db"
    "github.com/chuckie/goinsight/internal/jira"
    "github.com/chuckie/goinsight/internal/llm"
    "github.com/chuckie/goinsight/internal/repository"
    "github.com/chuckie/goinsight/internal/service"
    "github.com/chuckie/goinsight/internal/http"
    "github.com/chuckie/goinsight/internal/http/middleware"
)

func setupApplication(databaseURL, llmProvider string) (*http.ServiceHandler, *chi.Mux, error) {
    // 1. Initialize database
    database, err := db.NewClient(databaseURL)
    if err != nil {
        return nil, nil, err
    }
    
    // 2. Initialize LLM client
    llmClient := llm.NewClient(llmProvider)
    
    // 3. Initialize optional Jira client
    jiraClient, _ := jira.NewClient()
    
    // 4. Create repository (data access layer)
    repo := repository.NewPostgresFeedbackRepository(database.DB)
    
    // 5. Create service (business logic layer)
    feedbackService := service.NewFeedbackService(repo, llmClient, jiraClient)
    
    // 6. Create HTTP handler (presentation layer)
    handler := http.NewServiceHandler(feedbackService, jiraClient)
    
    // 7. Setup router with middleware
    router := chi.NewRouter()
    setupMiddleware(router)
    
    // 8. Register routes
    router.Post("/api/ask", handler.Ask)
    router.Post("/api/jira-tickets", handler.CreateJiraTickets)
    router.Get("/api/health", handler.HealthCheck)
    
    return handler, router, nil
}

func setupMiddleware(router *chi.Mux) {
    router.Use(middleware.RecoveryMiddleware)
    router.Use(middleware.LoggingMiddleware)
    router.Use(middleware.TimingMiddleware)
    router.Use(middleware.ValidateJSONMiddleware)
}
```

### Testing with Mocks

```go
package service

import (
    "testing"
    "github.com/chuckie/goinsight/internal/repository"
    "github.com/chuckie/goinsight/internal/domain"
    "github.com/chuckie/goinsight/internal/llm"
)

type MockFeedbackRepository struct {
    QueryResults []map[string]any
    RiskScore    *domain.AccountRiskScore
}

func (m *MockFeedbackRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
    return m.QueryResults, nil
}

func (m *MockFeedbackRepository) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
    return m.RiskScore, nil
}

// ... implement other methods ...

type MockLLMClient struct {
    SQLResponse     string
    InsightResponse string
}

func (m *MockLLMClient) GenerateSQL(ctx context.Context, question string) (string, error) {
    return m.SQLResponse, nil
}

func (m *MockLLMClient) GenerateInsight(ctx context.Context, q string, data []map[string]any) (string, error) {
    return m.InsightResponse, nil
}

// Test function
func TestAnalyzeFeedback(t *testing.T) {
    mockRepo := &MockFeedbackRepository{
        QueryResults: []map[string]any{
            {"id": 1, "sentiment": "negative"},
        },
    }
    
    mockLLM := &MockLLMClient{
        SQLResponse: "SELECT * FROM feedback_enriched",
        InsightResponse: `{"summary": "Test", "recommendations": [], "actions": []}`,
    }
    
    service := NewFeedbackService(mockRepo, mockLLM, nil)
    
    response, err := service.AnalyzeFeedback(context.Background(), "test question")
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if response == nil {
        t.Fatal("Expected response, got nil")
    }
}
```

---

## Best Practices

1. **Always use Context** - Pass context through all layers for cancellation and timeouts
2. **Error Wrapping** - Use `fmt.Errorf` with `%w` for error context preservation
3. **Dependency Injection** - Pass dependencies through constructors, not globals
4. **Interface-based Design** - Code to interfaces, not concrete types
5. **Middleware Ordering** - Place cross-cutting concerns in logical order
6. **Query Builders** - Use builders for complex queries to improve readability

---

**Last Updated**: December 20, 2025
