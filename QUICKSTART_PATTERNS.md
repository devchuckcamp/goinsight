# Design Patterns - Developer Quick Start

A quick reference guide for developers using the new design patterns in GoInsight.

## TL;DR

- **Repository** = Data access layer
- **Service** = Business logic layer
- **Builder** = Query construction
- **Middleware** = Cross-cutting concerns

---

## 5-Minute Setup

### 1. Initialize Application
```go
// main.go
import (
    "database/sql"
    "github.com/chuckie/goinsight/internal/db"
    "github.com/chuckie/goinsight/internal/repository"
    "github.com/chuckie/goinsight/internal/service"
    "github.com/chuckie/goinsight/internal/http"
)

func main() {
    // 1. Database
    dbClient, _ := db.NewClient(os.Getenv("DATABASE_URL"))
    
    // 2. Repository
    repo := repository.NewPostgresFeedbackRepository(dbClient.DB)
    
    // 3. Service
    svc := service.NewFeedbackService(repo, llmClient, jiraClient)
    
    // 4. Handler
    handler := http.NewServiceHandler(svc, jiraClient)
    
    // 5. Start server
    http.ListenAndServe(":8080", setupRouter(handler))
}
```

### 2. Use Service Layer
```go
// Analyze feedback
response, err := feedbackService.AnalyzeFeedback(
    context.Background(),
    "Show me negative billing feedback",
)

// Create Jira tickets
tickets, err := feedbackService.CreateJiraTickets(
    context.Background(),
    jiraTicketRequest,
)
```

### 3. Build Queries
```go
// Simple builder
query := builder.NewQueryBuilder().
    Select("id", "sentiment").
    From("feedback_enriched").
    Where("sentiment = 'negative'").
    Limit(10).
    Build()

// Specialized feedback builder with parameterized queries
fbQuery, params := builder.NewFeedbackQueryBuilder().
    WithSentiment("negative").
    WithProductArea("billing").
    Limit(20).
    BuildFeedbackWithParams()
```

---

## Common Tasks

### Query Feedback Data
```go
// Direct repository access
results, _ := repo.QueryFeedback(ctx, "SELECT * FROM feedback_enriched LIMIT 10")

// Or through service
response, _ := feedbackService.AnalyzeFeedback(ctx, "Show me feedback")
```

### Get Account Health Score
```go
score, _ := feedbackService.GetAccountRiskScore(ctx, "account-123")
fmt.Printf("Churn Risk: %.2f%%\n", score.ChurnProbability*100)
```

### Analyze Product Areas
```go
impacts, _ := feedbackService.GetProductAreaImpacts(ctx, "enterprise")
for _, impact := range impacts {
    fmt.Println(impact)
}
```

### Create Jira Tickets
```go
req := service.JiraTicketRequest{
    Summary: "Billing issues",
    Actions: []domain.ActionItem{
        {Title: "Fix payments", Description: "..."},
    },
}
result, _ := feedbackService.CreateJiraTickets(ctx, req)
```

---

## Adding New Repository Methods

```go
// 1. Add to interface in repository/feedback_repository.go
type FeedbackRepository interface {
    // ... existing methods ...
    GetTopicCounts(ctx context.Context) (map[string]int, error)
}

// 2. Implement for PostgreSQL
func (r *PostgresFeedbackRepository) GetTopicCounts(ctx context.Context) (map[string]int, error) {
    query := `SELECT topic, COUNT(*) FROM feedback_enriched GROUP BY topic`
    results, err := r.QueryFeedback(ctx, query)
    // ... process and return ...
}

// 3. Use in service
func (s *FeedbackService) GetTopicAnalysis(ctx context.Context) (map[string]int, error) {
    return s.repo.GetTopicCounts(ctx)
}

// 4. Add handler
func (h *ServiceHandler) GetTopicAnalysis(w http.ResponseWriter, r *http.Request) {
    topics, _ := h.feedbackService.GetTopicAnalysis(r.Context())
    respondJSON(w, http.StatusOK, topics)
}
```

---

## Adding New Services

```go
// Create internal/service/analytics_service.go
type AnalyticsService struct {
    repo repository.FeedbackRepository
}

func NewAnalyticsService(repo repository.FeedbackRepository) *AnalyticsService {
    return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) GetMetrics(ctx context.Context) (*Metrics, error) {
    count, _ := s.repo.GetFeedbackEnrichedCount(ctx)
    // ... more metrics ...
    return &Metrics{TotalFeedback: count}, nil
}
```

---

## Adding Custom Middleware

```go
func RateLimitMiddleware(maxRequests int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(maxRequests), maxRequests)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                respondError(w, http.StatusTooManyRequests, "Rate limit exceeded")
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Add to router
router.Use(RateLimitMiddleware(100))
```

---

## Testing

### Mock Repository
```go
type MockRepository struct {
    QueryResults []map[string]any
    RiskScore    *domain.AccountRiskScore
}

func (m *MockRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
    return m.QueryResults, nil
}

// Use in test
mockRepo := &MockRepository{
    QueryResults: []map[string]any{{"id": 1}},
}
svc := service.NewFeedbackService(mockRepo, mockLLM, nil)
```

### Test Service
```go
func TestAnalyzeFeedback(t *testing.T) {
    mockRepo := &MockRepository{...}
    mockLLM := &MockLLMClient{...}
    
    svc := service.NewFeedbackService(mockRepo, mockLLM, nil)
    
    response, err := svc.AnalyzeFeedback(context.Background(), "test")
    assert.NoError(t, err)
    assert.NotNil(t, response)
}
```

---

## Advanced Builder Usage

### Conditional Filters
```go
builder := builder.NewQueryBuilder().
    Select("*").
    From("feedback_enriched")

// Only add filter if present
if sentiment != "" {
    builder.Where("sentiment = '" + sentiment + "'")
}

query := builder.Build()
```

### Dynamic Pagination
```go
func paginate(pageNum, pageSize int) string {
    offset := (pageNum - 1) * pageSize
    return builder.NewQueryBuilder().
        Select("*").
        From("feedback_enriched").
        OrderBy("created_at", "DESC").
        Limit(pageSize).
        Offset(offset).
        Build()
}
```

### Reusable Templates
```go
var queryTemplates = map[string]*builder.QueryBuilder{
    "recent_negative": builder.NewFeedbackQueryBuilder().
        WithSentiment("negative").
        OrderBy("created_at", "DESC"),
    
    "high_priority": builder.NewFeedbackQueryBuilder().
        WithMinPriority(4).
        OrderBy("priority", "DESC"),
}
```

---

## Error Handling

### Service Layer
```go
response, err := feedbackService.AnalyzeFeedback(ctx, question)
if err != nil {
    // Service already wraps errors with context
    // No need for additional wrapping
    log.Printf("Analysis failed: %v", err)
    return fmt.Errorf("feedback analysis failed: %w", err)
}
```

### HTTP Handler
```go
func (h *ServiceHandler) Ask(w http.ResponseWriter, r *http.Request) {
    response, err := h.feedbackService.AnalyzeFeedback(r.Context(), question)
    if err != nil {
        // Error message is user-friendly
        respondError(w, http.StatusInternalServerError, err.Error())
        return
    }
    respondJSON(w, http.StatusOK, response)
}
```

---

## Performance Monitoring

### Query Timing
```go
decorator := middleware.NewQueryExecutionDecorator()
start := time.Now()

results, _ := repo.QueryFeedback(ctx, query)

decorator.RecordExecution(time.Since(start), len(results), len(query))
metrics := decorator.GetMetrics()
log.Printf("Query metrics: %+v", metrics)
```

### Slow Query Detection
```go
checkSlowQuery := middleware.SlowQueryThreshold(100 * time.Millisecond)
checkSlowQuery(duration, sqlQuery)
// Logs if duration > threshold
```

---

## Debugging

### Enable Logging
```go
// In handlers
log.Printf("SQL Generated: %s", sqlQuery)
log.Printf("Query Results: %d rows", len(results))
log.Printf("Insight Generated: %s", insightResult.Summary)
```

### Query Builder Debug
```go
builder := builder.NewFeedbackQueryBuilder().
    WithSentiment("negative").
    WithProductArea("billing")

query, params := builder.BuildFeedbackWithParams()
fmt.Println(query, params) // Print generated SQL and parameters
```

### Service Call Tracing
```go
start := time.Now()
fmt.Printf("[%s] Calling AnalyzeFeedback\n", start.Format(time.RFC3339))

response, err := feedbackService.AnalyzeFeedback(ctx, question)

fmt.Printf("[%v] AnalyzeFeedback completed in %v\n", 
    time.Now().Format(time.RFC3339), 
    time.Since(start),
)
```

---

## Migration from Legacy Handler

### Old Way (Still Works)
```go
// internal/http/handlers.go
func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
    // Direct database access
    results, _ := h.dbClient.ExecuteQuery(sqlQuery)
    // ... process directly ...
}
```

### New Way (Recommended)
```go
// internal/http/service_handler.go
func (h *ServiceHandler) Ask(w http.ResponseWriter, r *http.Request) {
    // Use service layer
    response, _ := h.feedbackService.AnalyzeFeedback(r.Context(), question)
    // ... handler only returns response ...
}
```

### Gradual Migration
```go
// Use LegacyHandlerAdapter during transition
adapter := NewLegacyHandlerAdapter(legacyHandler, repo, llmClient, jiraClient)
// Adapter wraps new service in legacy handler interface
```

---

## Checklist for New Features

- [ ] Create repository method for data access
- [ ] Add service method for business logic
- [ ] Create handler method for HTTP endpoint
- [ ] Add middleware if cross-cutting concern
- [ ] Write unit test with mock repository
- [ ] Write integration test with real DB
- [ ] Document in DESIGN_PATTERNS_EXAMPLES.md
- [ ] Update DESIGN_PATTERNS.md if introducing new patterns

---

## Quick Reference

| Layer | Responsibility | Location |
|-------|-----------------|----------|
| **Repository** | Data access | `internal/repository/` |
| **Service** | Business logic | `internal/service/` |
| **Handler** | HTTP endpoints | `internal/http/service_handler.go` |
| **Builder** | Query construction | `internal/builder/` |
| **Middleware** | Cross-cutting concerns | `internal/http/middleware/` |

---

## Resources

- [Full Design Patterns Guide](./DESIGN_PATTERNS.md)
- [Detailed Examples](./DESIGN_PATTERNS_EXAMPLES.md)
- [Roadmap](./FUTURE_FEATURES.md)
- [Architecture](./ARCHITECTURE.md)

---

**Last Updated**: December 20, 2025
