# Design Patterns Implementation Guide

This document describes the design patterns implemented in the GoInsight application as outlined in FUTURE_FEATURES.md Phase 1.

## 📋 Overview

Four core design patterns have been implemented to improve code organization, maintainability, and scalability:

1. **Repository Pattern** - Data access abstraction
2. **Service Layer Pattern** - Business logic orchestration  
3. **Builder Pattern** - Query construction
4. **Decorator Pattern** - Cross-cutting concerns

---

## 🏛️ Repository Pattern

### Purpose
The Repository Pattern abstracts data access operations, separating data access logic from business logic. This makes code more testable and easier to swap database implementations.

### Location
`internal/repository/`

### Files
- **feedback_repository.go** - Interfaces and PostgreSQL implementation

### Key Interface
```go
type FeedbackRepository interface {
    QueryFeedback(ctx context.Context, query string) ([]map[string]any, error)
    GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error)
    GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error)
    GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error)
    GetFeedbackEnrichedCount(ctx context.Context) (int, error)
}
```

### Benefits
- ✅ Easy to swap implementations (e.g., PostgreSQL → MySQL, Redis cache layer)
- ✅ Simplified testing with mock repositories
- ✅ Centralized data access logic
- ✅ Better error handling and context propagation

### Usage Example
```go
// Inject into service
repo := repository.NewPostgresFeedbackRepository(db)
service := service.NewFeedbackService(repo, llmClient, jiraClient)

// Use in service
results, err := service.repo.QueryFeedback(ctx, sqlQuery)
```

---

## 🎯 Service Layer Pattern

### Purpose
The Service Layer contains all business logic, coordinating between repositories, LLM clients, and external services. Handlers delegate to services, keeping HTTP concerns separate from business logic.

### Location
`internal/service/`

### Files
- **feedback_service.go** - Feedback business logic and orchestration

### Key Service
```go
type FeedbackService struct {
    repo       repository.FeedbackRepository
    llmClient  llm.Client
    jiraClient *jira.Client
}
```

### Main Operations
- `AnalyzeFeedback()` - Complete workflow: SQL generation → execution → insight generation
- `CreateJiraTickets()` - Convert insights into actionable Jira tickets
- `GetAccountRiskScore()` - Retrieve account health metrics
- `GetProductAreaImpacts()` - Get product area prioritization

### Benefits
- ✅ Clean separation of concerns
- ✅ Testable business logic without HTTP dependencies
- ✅ Reusable across HTTP, gRPC, CLI interfaces
- ✅ Clear orchestration of complex workflows

### Usage Example
```go
// Service automatically orchestrates the workflow
response, err := feedbackService.AnalyzeFeedback(ctx, "Show me negative feedback")
if err != nil {
    // Handle error
    return err
}
// Response includes SQL, data preview, and AI-generated insights
```

---

## 🔨 Builder Pattern

### Purpose
The Builder Pattern enables constructing complex SQL queries incrementally with optional filters and clauses. This improves readability and maintainability of query construction.

### Location
`internal/builder/`

### Files
- **query_builder.go** - General and specialized query builders

### Key Classes
```go
type QueryBuilder struct { /* ... */ }
type FeedbackQueryBuilder struct { /* ... */ }
```

### Building Blocks
- `Select()` - Specify columns
- `From()` - Set table
- `Where()` - Add conditions (chainable)
- `WhereIf()` - Conditional filtering
- `OrderBy()` - Sort results
- `Limit()` - Restrict row count
- `Offset()` - Pagination

### Benefits
- ✅ Readable, chainable API
- ✅ Type-safe query construction
- ✅ Reusable query templates
- ✅ Easy to add optional filters

### Usage Example
```go
// General query builder
query := builder.NewQueryBuilder().
    Select("id", "sentiment", "priority").
    From("feedback_enriched").
    Where("sentiment = 'negative'").
    OrderBy("priority", "DESC").
    Limit(10).
    Build()

// Specialized feedback builder
fbQuery := builder.NewFeedbackQueryBuilder().
    WithSentiment("negative").
    WithProductArea("billing").
    WithMinPriority(3).
    OrderBy("created_at", "DESC").
    Limit(20).
    BuildFeedback()
```

---

## 🎭 Decorator Pattern (Middleware)

### Purpose
The Decorator Pattern adds cross-cutting concerns (logging, timing, validation, error handling) without modifying core handler logic. Middleware wraps handlers to compose behavior.

### Location
`internal/http/middleware/`

### Files
- **middleware.go** - Logging, timing, recovery, validation, query performance monitoring

### Key Middleware

#### LoggingMiddleware
Logs all incoming requests and responses with method, URI, status code, and duration.

#### TimingMiddleware
Measures request execution time and adds `X-Response-Time` header.

#### RecoveryMiddleware
Recovers from panics and logs them before returning 500 error.

#### ValidateJSONMiddleware
Validates incoming requests have `Content-Type: application/json`.

#### QueryExecutionDecorator
Measures query performance metrics:
- Execution time
- Rows returned
- Query size

### Benefits
- ✅ Cross-cutting concerns separated from handlers
- ✅ Composable middleware chain
- ✅ Easy to enable/disable features
- ✅ Reusable across multiple handlers

### Usage Example
```go
// In router setup
r.Use(middleware.LoggingMiddleware)
r.Use(middleware.TimingMiddleware)
r.Use(middleware.RecoveryMiddleware)
r.Use(middleware.ValidateJSONMiddleware)

// Query execution monitoring
decorator := middleware.NewQueryExecutionDecorator()
start := time.Now()
results, err := repo.QueryFeedback(ctx, query)
decorator.RecordExecution(time.Since(start), len(results), len(query))
metrics := decorator.GetMetrics()
```

---

## 🔄 Architecture Flow

```
HTTP Request
    ↓
Router → ServiceHandler
    ↓
Service Layer (FeedbackService)
    ├─→ Repository (FeedbackRepository)
    ├─→ LLM Client (GenerateSQL, GenerateInsight)
    ├─→ Jira Client (CreateIssues)
    └─→ Validation & Orchestration
    ↓
HTTP Response
```

---

## 📝 Migration Path from Legacy

### Old Architecture
```go
// Direct DB access in handlers
Handler → dbClient.ExecuteQuery()
```

### New Architecture
```go
// Service-based with repository
Handler → Service → Repository → DB
```

### Adapter for Gradual Migration
A `LegacyHandlerAdapter` is provided to support both old and new handlers during transition:

```go
adapter := NewLegacyHandlerAdapter(legacyHandler, repo, llmClient, jiraClient)
// Adapter wraps service layer inside legacy handler
```

---

## 🧪 Testing Benefits

Each layer can now be tested independently:

```go
// Unit test service without HTTP
func TestAnalyzeFeedback(t *testing.T) {
    mockRepo := &MockFeedbackRepository{}
    mockLLM := &MockLLMClient{}
    service := NewFeedbackService(mockRepo, mockLLM, nil)
    
    response, err := service.AnalyzeFeedback(ctx, "test question")
    assert.NoError(t, err)
    // Verify logic without database
}

// Integration test with real DB
func TestAnalyzeFeedbackIntegration(t *testing.T) {
    db := setupTestDB()
    repo := NewPostgresFeedbackRepository(db)
    service := NewFeedbackService(repo, realLLMClient, jiraClient)
    
    response, err := service.AnalyzeFeedback(ctx, "test question")
    assert.NoError(t, err)
    // Verify end-to-end workflow
}
```

---

## 📈 Future Enhancements

With this foundation, future improvements are easier:

1. **Caching Layer** - Add `CacheRepository` wrapper around `FeedbackRepository`
2. **Performance Monitoring** - Leverage `QueryExecutionDecorator` for metrics
3. **Additional Services** - Add `AnalyticsService`, `ReportingService`, etc.
4. **Query Optimization** - Use builders for automatic index recommendations
5. **Distributed Tracing** - Add middleware for request correlation IDs

---

## 🔗 Related Files

- Main handlers: [internal/http/handlers.go](../../internal/http/handlers.go)
- Service handlers: [internal/http/service_handler.go](../../internal/http/service_handler.go)
- Domain models: [internal/domain/](../../internal/domain/)
- Configuration: [internal/config/](../../internal/config/)
- LLM clients: [internal/llm/](../../internal/llm/)

---

**Last Updated**: December 20, 2025  
**Status**: ✅ Phase 1 Complete - Ready for Phase 2 (Cache Layer)
