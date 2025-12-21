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

## Commands Reference

```bash
# All tests
go test ./...

# Verbose
go test -v ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Specific package
go test ./internal/cache

# Specific test
go test -run TestMemoryCache ./internal/cache

# Benchmarks
go test -bench=. -benchmem ./internal/cache

# Update golden files
go test -update ./...
```

## Performance Targets

| Operation | Target | Unit |
|-----------|--------|------|
| Cache Get | < 500 | ns/op |
| Cache Set | < 1,500 | ns/op |
| Cache Delete | < 2,000 | ns/op |
| Query Execution | < 50,000 | ns/op |
| Transaction Begin | < 5,000 | ns/op |

## Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Benchmark Best Practices](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Concurrent Code](https://go.dev/blog/pipelines)

---

**Last Updated**: December 20, 2025  
**Phase**: 5 - Enhanced Testing  
**Version**: 1.0
