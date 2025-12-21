# Repository Pattern Implementation Guide

## Overview

Phase 4 implements the Repository Pattern to provide a clean abstraction over data access logic. This decouples business logic from database implementation details, making the code more testable, maintainable, and flexible.

## Architecture

### Repository Components

```
internal/repository/
├── feedback_repository.go   # Interface and PostgreSQL implementation
├── factory.go               # Factory pattern for repository creation
├── query_builder.go         # Fluent query building interface
├── transaction.go           # Transaction support
└── mock_repository.go       # (Optional) Mock for testing
```

## Core Concepts

### 1. Repository Pattern

**Purpose**: Provide a collection-like interface to access data

**Key Benefits**:
- ✅ Abstraction over database implementation
- ✅ Easy to swap database backends
- ✅ Simplified unit testing with mocks
- ✅ Centralized data access logic
- ✅ Reduced code duplication

### 2. Repository Interface

```go
type FeedbackRepository interface {
    // QueryFeedback executes a feedback query
    QueryFeedback(ctx context.Context, query string) ([]map[string]any, error)

    // GetAccountRiskScore retrieves ML predictions
    GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error)

    // GetRecentNegativeFeedbackCount counts recent feedback
    GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error)

    // GetProductAreaImpacts retrieves ML predictions
    GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error)

    // GetFeedbackEnrichedCount returns total count
    GetFeedbackEnrichedCount(ctx context.Context) (int, error)
}
```

### 3. Repository Implementation

```go
type PostgresFeedbackRepository struct {
    db *sql.DB
}

// All interface methods are implemented here
// Encapsulates SQL query logic and database-specific details
```

## Components in Detail

### Factory Pattern

**Responsibility**: Create and configure repository instances

```go
type RepositoryFactory struct {
    db *sql.DB
}

factory := repository.NewRepositoryFactory(dbClient.DB())
repo := factory.CreateFeedbackRepository()
```

**Features**:
- Centralized repository creation
- Configuration management
- Connection validation
- Connection pooling setup

### Query Builder

**Responsibility**: Provide fluent interface for query construction

```go
// Fluent API for building queries
builder := repository.NewQueryBuilder()
query, params := builder.
    Select("id", "name", "email").
    From("feedback_enriched").
    Where("sentiment = $1", "negative").
    AndWhere("created_at > $2", timeThreshold).
    OrderBy("created_at DESC").
    Limit(10).
    Build()

results, err := repo.QueryFeedback(ctx, query)
```

**Benefits**:
- Type-safe query construction
- Prevents SQL injection (parameterized)
- Readable, chainable syntax
- Easier to maintain complex queries

### Transaction Support

**Responsibility**: Group operations into atomic units

```go
// Begin transaction
tx, err := factory.BeginTransaction(ctx, repository.DefaultTransactionOptions())
if err != nil {
    return err
}
defer tx.Rollback()

// Get repository within transaction
repo := tx.GetRepository()

// Perform operations...
// All changes are atomic - either all succeed or all fail

// Commit transaction
return tx.Commit()
```

**Configuration**:
```go
type TransactionOptions struct {
    Isolation sql.IsolationLevel
    ReadOnly  bool
}
```

### Repository Container

**Responsibility**: Hold all repository instances

```go
type RepositoryContainer struct {
    Feedback FeedbackRepository
}

container := repository.NewRepositoryContainer(dbClient.DB())
repo := container.Feedback  // Access any repository
```

## Integration

### Initialization in main.go

```go
// Create factory
factory := repository.NewRepositoryFactory(dbClient.DB())

// Configure
config := repository.DefaultRepositoryConfig()
factory.ApplyConfig(config)
factory.ValidateConnection()

// Create container
container := repository.NewRepositoryContainer(dbClient.DB())

// Use repositories
feedback := container.Feedback
```

### Service Layer Integration

Services delegate to repositories:

```go
type FeedbackService struct {
    repo repository.FeedbackRepository
    // ... other dependencies
}

func (s *FeedbackService) AnalyzeFeedback(ctx context.Context, question string) error {
    results, err := s.repo.QueryFeedback(ctx, sqlQuery)
    if err != nil {
        return fmt.Errorf("repository error: %w", err)
    }
    // ... process results
}
```

## Configuration

### DefaultRepositoryConfig

```go
type RepositoryConfig struct {
    MaxOpenConnections int  // 25 default
    MaxIdleConnections int  // 5 default
    ConnMaxLifetime    int  // 3600 seconds
    QueryTimeout       int  // 30 seconds
    StatementCacheSize int  // 100
}

config := repository.DefaultRepositoryConfig()
factory.ApplyConfig(config)
```

### Customization

```go
config := repository.RepositoryConfig{
    MaxOpenConnections: 50,
    MaxIdleConnections: 10,
    ConnMaxLifetime:    7200,
    QueryTimeout:       60,
    StatementCacheSize: 200,
}

factory.ApplyConfig(config)
```

## Usage Examples

### Basic Query Execution

```go
repo := factory.CreateFeedbackRepository()

// Direct SQL
results, err := repo.QueryFeedback(ctx, 
    "SELECT * FROM feedback_enriched WHERE sentiment = 'negative'")

// Using query builder
builder := repository.NewQueryBuilder()
query, _ := builder.
    SelectAll().
    From("feedback_enriched").
    Where("sentiment = $1", "negative").
    Build()

results, err := repo.QueryFeedback(ctx, query)
```

### Account Analysis

```go
// Get risk score for account
riskScore, err := repo.GetAccountRiskScore(ctx, "account-123")

// Count recent negative feedback
count, err := repo.GetRecentNegativeFeedbackCount(ctx, "customer-tier")

// Get product area impacts
impacts, err := repo.GetProductAreaImpacts(ctx, "enterprise")
```

### Transactional Operations

```go
tx, err := factory.BeginTransaction(ctx, repository.TransactionOptions{
    Isolation: sql.LevelReadCommitted,
    ReadOnly:  false,
})
if err != nil {
    return err
}
defer tx.Rollback()

repo := tx.GetRepository()

// Perform multiple operations within transaction
result1, err := repo.QueryFeedback(ctx, query1)
result2, err := repo.QueryFeedback(ctx, query2)

// All succeed or all fail
return tx.Commit()
```

## Testing

### Mocking Repositories

Create a mock for testing:

```go
type MockFeedbackRepository struct {
    QueryFeedbackFunc func(ctx context.Context, query string) ([]map[string]any, error)
    // ... other mock methods
}

func (m *MockFeedbackRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
    return m.QueryFeedbackFunc(ctx, query)
}

// Usage in tests
mockRepo := &MockFeedbackRepository{
    QueryFeedbackFunc: func(ctx context.Context, query string) ([]map[string]any, error) {
        return []map[string]any{
            {"id": "1", "sentiment": "negative"},
        }, nil
    },
}

service := service.NewFeedbackService(mockRepo, llmClient, jiraClient)
// ... test service behavior
```

### Integration Tests

```go
func TestFeedbackRepository(t *testing.T) {
    db, cleanup := setupTestDB()
    defer cleanup()

    repo := repository.NewPostgresFeedbackRepository(db)

    // Test QueryFeedback
    results, err := repo.QueryFeedback(ctx, "SELECT * FROM feedback_enriched LIMIT 1")
    assert.NoError(t, err)
    assert.NotEmpty(t, results)

    // Test GetAccountRiskScore
    score, err := repo.GetAccountRiskScore(ctx, "test-account")
    assert.NoError(t, err)
    // ... more assertions
}
```

## Performance Considerations

### Connection Pooling

Default configuration:
- Max Open Connections: 25
- Max Idle Connections: 5
- Connection Max Lifetime: 1 hour

Tune based on:
- Expected concurrent request load
- Database capacity
- Memory constraints

### Query Optimization

```go
// Use column selection instead of SELECT *
builder.Select("id", "sentiment", "created_at").From("feedback_enriched")

// Use WHERE clauses to filter early
builder.Where("created_at > $1", recent)

// Use LIMIT for large result sets
builder.Limit(100)

// Use appropriate indexes on frequently queried columns
```

### Transaction Isolation

Choose appropriate isolation level:

```go
// Default - allows dirty reads
sql.LevelDefault

// Prevents dirty reads
sql.LevelReadUncommitted

// Most common - prevents dirty & non-repeatable reads
sql.LevelReadCommitted

// Prevents most anomalies
sql.LevelRepeatableRead

// Highest isolation
sql.LevelSerializable
```

## Data Flow

```
HTTP Handler
    ↓
Service Layer
    ├── Cache Check
    ├── Repository Query
    │   ├── Connection Pool
    │   ├── Query Execution
    │   └── Result Mapping
    ├── Result Processing
    └── Cache Update
    ↓
HTTP Response
```

## Error Handling

Repository errors are wrapped with context:

```go
// From repository
return nil, fmt.Errorf("failed to get columns: %w", err)

// Propagated through service
if err != nil {
    if s.logger != nil {
        s.logger.Error("Query execution failed", err, ...)
    }
    return nil, fmt.Errorf("repository error: %w", err)
}

// Handled in handler
if err != nil {
    respondError(w, http.StatusInternalServerError, err.Error())
    return
}
```

## Extending the Repository

### Adding New Methods

```go
// 1. Add to interface
type FeedbackRepository interface {
    // ... existing methods
    
    // New method
    GetFeedbackByID(ctx context.Context, id string) (*domain.Feedback, error)
}

// 2. Implement in PostgresFeedbackRepository
func (r *PostgresFeedbackRepository) GetFeedbackByID(ctx context.Context, id string) (*domain.Feedback, error) {
    query := `SELECT id, summary, sentiment FROM feedback_enriched WHERE id = $1`
    row := r.db.QueryRowContext(ctx, query, id)
    // ... scan and return
}

// 3. Use in service
feedback, err := s.repo.GetFeedbackByID(ctx, feedbackID)
```

### Supporting Multiple Databases

```go
// Add new implementation for MongoDB
type MongoFeedbackRepository struct {
    client *mongo.Client
}

func (r *MongoFeedbackRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
    // MongoDB-specific implementation
}

// Factory handles creation
func (f *RepositoryFactory) CreateFeedbackRepository(backend string) FeedbackRepository {
    switch backend {
    case "postgres":
        return NewPostgresFeedbackRepository(f.db)
    case "mongo":
        return NewMongoFeedbackRepository(f.mongoClient)
    default:
        return NewPostgresFeedbackRepository(f.db)
    }
}
```

## Best Practices

✅ **Keep repositories simple** - Only data access logic
✅ **Use interfaces** - Swap implementations easily
✅ **Write to abstraction** - Services depend on interface, not implementation
✅ **Handle errors properly** - Wrap with context
✅ **Use connection pooling** - Configure appropriately
✅ **Test with mocks** - Avoid database dependencies in unit tests
✅ **Document queries** - Explain complex SQL
✅ **Validate input** - Prevent SQL injection
✅ **Use parameterized queries** - Always use placeholders ($1, $2, etc.)
✅ **Close transactions properly** - Use defer for cleanup

## Troubleshooting

### Connection Pool Exhaustion
- Increase `MaxOpenConnections`
- Check for connection leaks (connections not closed)
- Monitor active connections

### Slow Queries
- Add indexes on frequently queried columns
- Use EXPLAIN ANALYZE
- Consider query builder for optimization
- Cache frequently accessed data

### Transaction Deadlocks
- Maintain consistent transaction ordering
- Reduce transaction scope
- Use appropriate isolation level
- Monitor locks

## Migration from Direct DB Access

**Before (direct access)**:
```go
rows, _ := dbClient.Query("SELECT * FROM feedback")
```

**After (repository pattern)**:
```go
repo := factory.CreateFeedbackRepository()
results, _ := repo.QueryFeedback(ctx, "SELECT * FROM feedback")
```

**Benefits of migration**:
- Consistent interface
- Easier to test
- Better error handling
- Query optimization in one place
- Flexibility to change database

---

**Last Updated**: December 20, 2025
**Phase**: 4
**Status**: Complete
