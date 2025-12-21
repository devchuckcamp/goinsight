# Testing Guide

**Version**: 1.0
**Phase**: 5
**Last Updated**: December 20, 2025

## Overview

This guide covers the comprehensive testing framework for goinsight, including unit tests, integration tests, benchmarks, and best practices.

## Table of Contents

1. [Test Structure](#test-structure)
2. [Running Tests](#running-tests)
3. [Unit Tests](#unit-tests)
4. [Integration Tests](#integration-tests)
5. [Benchmarks](#benchmarks)
6. [Mocks and Fixtures](#mocks-and-fixtures)
7. [Test Coverage](#test-coverage)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## Test Structure

### File Organization

```
.
├── internal/
│   ├── repository/
│   │   ├── feedback_repository.go
│   │   ├── feedback_repository_test.go      ← Unit tests
│   │   ├── transaction.go
│   │   └── transaction_test.go              ← Unit tests
│   ├── cache/
│   │   ├── memory_cache.go
│   │   ├── memory_cache_test.go             ← Unit tests
│   │   └── cache_test.go                    ← Interface tests
│   └── service/
│       ├── feedback_service.go
│       └── feedback_service_test.go         ← Unit tests
├── tests/
│   ├── integration/                         ← Integration tests
│   │   ├── repository_test.go
│   │   └── service_test.go
│   ├── mocks/                               ← Mock implementations
│   │   └── feedback_repository.go
│   └── testutil/                            ← Test utilities
│       ├── db.go
│       └── factory.go
└── Makefile                                 ← Test commands
```

### Test Naming Convention

- **Unit Tests**: `TestFeatureName` (e.g., `TestQueryFeedback`)
- **Table-Driven Tests**: Test with sub-tests using `t.Run()`
- **Benchmark Tests**: `BenchmarkFeatureName` (e.g., `BenchmarkQueryFeedback`)
- **Integration Tests**: `TestIntegration*` (e.g., `TestIntegrationRepository`)

## Running Tests

### Run All Tests

```bash
# Run all tests in the repository
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Run Specific Tests

```bash
# Run tests in a package
go test ./internal/repository

# Run a specific test
go test ./internal/repository -run TestQueryFeedback

# Run tests matching a pattern
go test ./... -run "^TestMemoryCache"
```

### Run Benchmarks

```bash
# Run benchmarks only
go test -bench=. -benchmem ./internal/cache

# Run specific benchmark
go test -bench=BenchmarkMemoryCacheSet -benchmem ./internal/cache

# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/cache

# View profile
go tool pprof cpu.prof
```

### Using Makefile

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
make bench

# Run integration tests
make test-integration
```

## Unit Tests

### Repository Tests

**Location**: `internal/repository/feedback_repository_test.go`

Tests the PostgresFeedbackRepository implementation:

```go
func TestQueryFeedback(t *testing.T) {
    // Test various SQL queries
    // Verify error handling
    // Check result formatting
}
```

**Coverage Areas**:
- ✅ Basic query execution
- ✅ Error handling (invalid SQL, connection errors)
- ✅ Result mapping and type conversion
- ✅ Context handling and timeouts
- ✅ Performance benchmarks

### Cache Tests

**Location**: `internal/cache/memory_cache_test.go`

Tests the MemoryCache implementation:

```go
func TestMemoryCacheSet(t *testing.T) {
    // Test setting various data types
    // Verify TTL handling
    // Check eviction policies
}
```

**Coverage Areas**:
- ✅ Set/Get/Delete operations
- ✅ TTL expiration
- ✅ LRU eviction
- ✅ Concurrent access (thread safety)
- ✅ Cache statistics
- ✅ Performance benchmarks

### Transaction Tests

**Location**: `internal/repository/transaction_test.go`

Tests transaction support:

```go
func TestBeginTransaction(t *testing.T) {
    // Test transaction creation
    // Verify commit/rollback
    // Check ACID compliance
}
```

**Coverage Areas**:
- ✅ Transaction creation
- ✅ Commit operations
- ✅ Rollback operations
- ✅ ACID properties (Atomicity, Consistency, Isolation, Durability)
- ✅ Error handling during transactions

## Integration Tests

### Setup

**Location**: `tests/integration/`

Integration tests require a test database:

```go
func TestIntegrationRepository(t *testing.T) {
    // Create test database
    tdb, err := testutil.NewTestDBWithConnection(t, testDatabaseURL)
    if err != nil {
        t.Fatalf("Failed to create test DB: %v", err)
    }
    defer tdb.Close()

    // Run tests
}
```

### Running Integration Tests

```bash
# Start test database
docker-compose -f tests/docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./tests/integration/

# Stop test database
docker-compose -f tests/docker-compose.test.yml down
```

### Test Database

Configure via environment variables:

```bash
export TEST_DATABASE_URL="postgres://user:password@localhost:5432/goinsight_test?sslmode=disable"
go test -v ./tests/integration/
```

## Benchmarks

### Running Benchmarks

```bash
# Run all benchmarks with memory stats
go test -bench=. -benchmem ./internal/cache

# Run benchmarks with specific time
go test -bench=. -benchtime=10s ./internal/cache

# Run benchmarks and save results
go test -bench=. -benchmem -benchstat ./internal/cache > bench.txt
```

### Interpreting Results

```
BenchmarkMemoryCacheSet-8       1000000      1234 ns/op      256 B/op       5 allocs/op
                    ↑           ↑            ↑               ↑              ↑
                 Name        Iterations   Time/Op        Bytes/Op      Allocations
```

**Key Metrics**:
- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Number of allocations per operation (lower is better)

### Benchmark Examples

#### Cache Performance

```bash
$ go test -bench=BenchmarkMemoryCache -benchmem ./internal/cache

BenchmarkMemoryCacheSet-8             1000000      1234 ns/op      256 B/op       5 allocs/op
BenchmarkMemoryCacheGet-8             5000000       234 ns/op       48 B/op       1 allocs/op
BenchmarkMemoryCacheDelete-8          1000000      1567 ns/op      256 B/op       6 allocs/op
```

#### Repository Performance

```bash
$ go test -bench=BenchmarkQueryFeedback -benchmem ./internal/repository

BenchmarkQueryFeedback-8               100000    12340 ns/op    2048 B/op      23 allocs/op
```

## Mocks and Fixtures

### Mock Repository

**Location**: `tests/mocks/feedback_repository.go`

Use mocks for service testing without database:

```go
func TestFeedbackService(t *testing.T) {
    // Create mock
    mockRepo := mocks.NewMockFeedbackRepository()
    mockRepo.SetQueryFeedbackResult([]map[string]any{
        {"id": 1, "feedback": "great product"},
    })

    // Use in service
    service := service.NewFeedbackService(mockRepo, ...)
    result, err := service.AnalyzeFeedback(ctx, "test question")

    // Assert
    if !mockRepo.QueryFeedbackCalled {
        t.Error("QueryFeedback should have been called")
    }
}
```

### Configuring Mocks

```go
// Success scenario
mockRepo.SetupForSuccess()

// Failure scenario
mockRepo.SetupForFailure(errors.New("database error"))

// Timeout scenario
mockRepo.SetupForTimeout()

// Custom results
mockRepo.SetQueryFeedbackResult(results)
mockRepo.SetAccountRiskScoreResult(riskScore)

// Track calls
if mockRepo.QueryFeedbackCalled {
    t.Logf("Query called %d times", mockRepo.QueryFeedbackCallCount)
}
```

### Test Fixtures

**Location**: `tests/fixtures/seed.sql`

Pre-populated test data:

```sql
-- seed.sql
INSERT INTO feedback (id, account_id, sentiment, content) VALUES
    (1, 'acc-123', 'negative', 'Product is slow'),
    (2, 'acc-123', 'positive', 'Good support');
```

Loading fixtures:

```go
func TestWithFixtures(t *testing.T) {
    tdb, _ := testutil.NewTestDBWithConnection(t, dbURL)
    
    // Load seed data
    tdb.ExecSQL("CREATE TABLE feedback (...)")
    tdb.ExecSQL(readFixtures("seed.sql"))
    
    // Run tests
}
```

## Test Coverage

### Measuring Coverage

```bash
# Generate coverage report
go test -cover ./...

# Detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Package-level coverage
go test -coverprofile=coverage.out ./internal/cache
go tool cover -html=coverage.out
```

### Coverage Goals

| Component | Target |
|-----------|--------|
| Repository | 85%+ |
| Cache | 90%+ |
| Service | 80%+ |
| Overall | 80%+ |

### Coverage Report

```bash
# View coverage in HTML
go test -coverprofile=coverage.out -html ./internal/cache

# View coverage for specific functions
go test -v -cover ./internal/cache | grep -E "coverage|ok"
```

## Best Practices

### 1. Table-Driven Tests

```go
// Good: Clear, maintainable, extensible
func TestQueryFeedback(t *testing.T) {
    tests := []struct {
        name        string
        query       string
        shouldError bool
    }{
        {"empty query", "", true},
        {"valid query", "SELECT * FROM feedback", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2. Cleanup and Defer

```go
// Good: Proper cleanup
func TestWithDatabase(t *testing.T) {
    tdb, _ := testutil.NewTestDBWithConnection(t, dbURL)
    defer tdb.Close()
    
    // Database guaranteed to close
}
```

### 3. Sub-tests for Organization

```go
// Good: Organized testing
t.Run("SetUp", func(t *testing.T) {
    // Setup tests
})

t.Run("Operations", func(t *testing.T) {
    // Operation tests
})

t.Run("Cleanup", func(t *testing.T) {
    // Cleanup tests
})
```

### 4. Error Messages

```go
// Good: Clear error messages
t.Errorf("Query() returned %v, want %v", got, want)
t.Fatalf("Failed to connect to database: %v", err)
```

### 5. Test Isolation

```go
// Good: Each test is independent
func TestCacheIsolation(t *testing.T) {
    // Create fresh cache for each test
    cache := NewMemoryCache(10, 5*time.Minute)
    
    // Test in isolation
    cache.Set(ctx, "key", "value", 0)
}
```

## Troubleshooting

### Tests Failing Inconsistently

**Problem**: Tests pass sometimes, fail other times

**Solutions**:
- Check for race conditions: `go test -race ./...`
- Verify timing assumptions (TTL, delays)
- Ensure test isolation (shared state)
- Check concurrent access patterns

### Database Connection Failures

**Problem**: Integration tests can't connect to database

**Solutions**:
```bash
# Verify database is running
docker ps

# Check connection string
echo $TEST_DATABASE_URL

# Test connection manually
psql $TEST_DATABASE_URL -c "SELECT 1"

# Restart database
docker-compose down && docker-compose up
```

### Memory Leaks in Tests

**Problem**: Tests use excessive memory

**Solutions**:
```bash
# Run with memory profiling
go test -memprofile=mem.prof ./internal/cache
go tool pprof mem.prof

# Check for goroutine leaks
go test -v ./internal/cache | grep "goroutines"
```

### Slow Tests

**Problem**: Tests are taking too long

**Solutions**:
```bash
# Find slow tests
go test -v -timeout 30s ./... -count 3

# Profile slow tests
go test -cpuprofile=cpu.prof -run SlowTest ./...
go tool pprof cpu.prof
```

## CI/CD Integration

### GitHub Actions

**.github/workflows/test.yml**:
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go test -v ./...
      - run: go test -race ./...
      - run: go test -cover ./...
```

### Local Pre-commit Hook

**.git/hooks/pre-commit**:
```bash
#!/bin/sh
go test ./... || exit 1
go test -race ./... || exit 1
```

## Test Utilities and Factories

### Using the Test Factory

The `tests/testutil/factory.go` provides convenient factory methods for creating test objects:

```go
import "github.com/chuckie/goinsight/tests/testutil"

// Create test data
factory := testutil.NewFactory()

// Create single feedback item
feedback := factory.MakeFeedback("1", "positive")

// Create multiple feedback items
feedbacks := factory.MakeFeedbacks(10, "positive")

// Create enriched feedback
enriched := factory.MakeFeedbackEnriched("1", "positive", "billing")

// Create API request/response
request := factory.MakeAskRequest("What is sentiment?")
response := factory.MakeAskResponse("Customers are satisfied")
```

### Using Mock Database

```go
import "github.com/chuckie/goinsight/tests/testutil"

// Create mock database
db := testutil.NewMockDatabase()

// Set up query results
db.SetQueryResult("SELECT * FROM feedback", []map[string]any{
    {"id": "1", "feedback": "Great", "sentiment": "positive"},
})

// Execute query
results, err := db.ExecuteQuery("SELECT * FROM feedback")
```

### Test Data Fixtures

Test fixtures are available in `tests/fixtures/seed.sql`:

```bash
# Load fixtures into test database
psql -f tests/fixtures/seed.sql
```

Fixtures include:
- Sample feedback items (5 entries)
- Enriched feedback data
- Account risk scores
- Product area impacts

## Integration Testing

### Running Integration Tests

```bash
# Run all integration tests
go test ./tests/integration -v

# Run specific integration test
go test ./tests/integration -run TestRepositoryIntegration -v

# Run with timeout
go test ./tests/integration -timeout 120s -v
```

### Writing Integration Tests

Integration tests combine multiple components:

```go
func TestIntegrationFlow(t *testing.T) {
    // Setup
    mockRepo := mocks.NewMockFeedbackRepository()
    factory := testutil.NewFactory()
    
    // Create test data
    testData := factory.MakeFeedbacks(5, "positive")
    mockRepo.SetQueryFeedbackResult(testData)
    
    // Execute
    results, err := mockRepo.QueryFeedback(context.Background(), "SELECT *")
    
    // Assert
    if err != nil || len(results) != 5 {
        t.Fail()
    }
}
```

## Using Make Targets

The `Makefile.test` provides convenient test targets:

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Run only unit tests
make test-unit

# Run only integration tests  
make test-integration

# Generate coverage report
make test-coverage

# Generate HTML coverage report
make test-coverage-html

# Run benchmarks
make benchmark

# Run specific package tests
make test-http
make test-service
make test-cache
make test-repository

# Run tests with race detector
make test-race

# Clean test artifacts
make clean-test

# View all available targets
make help-test
```

## CI/CD Integration

### GitHub Actions

Tests run automatically on:
- Push to main, develop, or feature branches
- Pull requests

The workflow file `.github/workflows/test.yml` runs:
1. **Unit Tests** - Fast tests for pure functions
2. **Integration Tests** - Tests with dependencies
3. **Coverage Analysis** - Code coverage reporting
4. **Race Detector** - Concurrency issues detection
5. **Benchmarks** - Performance tracking (main branch only)
6. **Linting** - Code quality checks

### View Results

Coverage reports are sent to Codecov:
- https://codecov.io/gh/chuckie/goinsight

## Common Testing Patterns

### Table-Driven Tests

```go
func TestRepository(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    int
        wantErr bool
    }{
        {"valid query", "SELECT *", 1, false},
        {"invalid query", "INVALID", 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := repo.QueryFeedback(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if len(got) != tt.want {
                t.Errorf("got %d, want %d", len(got), tt.want)
            }
        })
    }
}
```

### Testing with Context

```go
func TestWithTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    results, err := repo.QueryFeedback(ctx, query)
    if err != nil {
        t.Fatalf("query failed: %v", err)
    }
}
```

### Testing Concurrency

```go
func TestConcurrent(t *testing.T) {
    done := make(chan bool, 10)
    
    for i := 0; i < 10; i++ {
        go func() {
            _, err := repo.QueryFeedback(context.Background(), query)
            if err != nil {
                t.Error(err)
            }
            done <- true
        }()
    }
    
    for i := 0; i < 10; i++ {
        <-done
    }
}
```

## Coverage Requirements

- **Overall Coverage Target**: 70%+
- **Package Coverage Targets**:
  - `internal/http`: 80%+
  - `internal/service`: 75%+
  - `internal/cache`: 85%+
  - `internal/repository`: 70%+ (database-dependent)

### Generate Coverage Report

```bash
# Generate coverage data
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check specific package coverage
go tool cover -func=coverage.out | grep internal/http
```

## Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Benchmark Best Practices](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Concurrent Code](https://go.dev/blog/pipelines)

---

**Last Updated**: December 20, 2025  
**Phase**: 5 - Enhanced Testing  
**Version**: 1.0
