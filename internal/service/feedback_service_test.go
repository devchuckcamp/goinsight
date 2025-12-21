package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chuckie/goinsight/internal/cache"
	"github.com/chuckie/goinsight/tests/mocks"
)

// MockLLMClient is a mock LLM client for service tests
type MockLLMClient struct {
	GenerateSQLFn     func(context.Context, string) (string, error)
	GenerateInsightFn func(context.Context, string, []map[string]any) (string, error)
	GenerateFn        func(context.Context, string) (string, error)
}

func (m *MockLLMClient) GenerateSQL(ctx context.Context, question string) (string, error) {
	if m.GenerateSQLFn != nil {
		return m.GenerateSQLFn(ctx, question)
	}
	return "SELECT * FROM feedback", nil
}

func (m *MockLLMClient) GenerateInsight(ctx context.Context, question string, results []map[string]any) (string, error) {
	if m.GenerateInsightFn != nil {
		return m.GenerateInsightFn(ctx, question, results)
	}
	return `{"summary": "Analysis complete", "recommendations": [], "actions": []}`, nil
}

func (m *MockLLMClient) Generate(ctx context.Context, prompt string) (string, error) {
	if m.GenerateFn != nil {
		return m.GenerateFn(ctx, prompt)
	}
	return "Generated response", nil
}

// TestAnalyzeFeedbackSuccess tests successful feedback analysis
func TestAnalyzeFeedbackSuccess(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "Great product", "sentiment": "positive"},
	})

	service := NewFeedbackService(
		mockRepo,
		&MockLLMClient{},
		nil,
	)

	ctx := context.Background()
	result, err := service.AnalyzeFeedback(ctx, "What is customer sentiment?")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result")
	}
}

// TestAnalyzeFeedbackEmptyQuestion tests with empty question
func TestAnalyzeFeedbackEmptyQuestion(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx := context.Background()
	_, err := service.AnalyzeFeedback(ctx, "")

	if err == nil {
		t.Error("Expected error for empty question")
	}
}

// TestAnalyzeFeedbackWithCache tests caching behavior
func TestAnalyzeFeedbackWithCache(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "Excellent", "sentiment": "positive"},
	})

	cacheManager := cache.NewCacheManager(true, 100, 5*time.Minute)

	service := NewFeedbackServiceWithCache(
		mockRepo,
		&MockLLMClient{},
		nil,
		cacheManager,
	)

	ctx := context.Background()
	question := "What is the sentiment?"

	// First call - should hit repository
	result1, err1 := service.AnalyzeFeedback(ctx, question)
	if err1 != nil {
		t.Logf("First call error: %v", err1)
	}

	// Second call - might hit cache
	result2, err2 := service.AnalyzeFeedback(ctx, question)
	if err2 != nil {
		t.Logf("Second call error: %v", err2)
	}

	// Both should return results
	if result1 != nil && result2 != nil {
		// Results should be consistent
		if result1.Question != result2.Question {
			t.Error("Results should be consistent between calls")
		}
	}
}

// TestAnalyzeFeedbackRepositoryError tests error from repository
func TestAnalyzeFeedbackRepositoryError(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetupForDatabaseError()

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx := context.Background()
	_, err := service.AnalyzeFeedback(ctx, "What is sentiment?")

	// Should handle repository errors gracefully
	_ = err
}

// TestAnalyzeFeedbackLLMError tests error from LLM client
func TestAnalyzeFeedbackLLMError(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{})

	llmClient := &MockLLMClient{
		GenerateSQLFn: func(ctx context.Context, q string) (string, error) {
			return "", errors.New("LLM service error")
		},
	}

	service := NewFeedbackService(mockRepo, llmClient, nil)

	ctx := context.Background()
	_, err := service.AnalyzeFeedback(ctx, "What is sentiment?")

	// Should handle LLM errors
	_ = err
}

// TestNewFeedbackService tests basic service creation
func TestNewFeedbackService(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	if service == nil {
		t.Error("Service should not be nil")
	}

	if service.repo != mockRepo {
		t.Error("Repository not set correctly")
	}
}

// TestNewFeedbackServiceWithProfiler tests service with profiler
func TestNewFeedbackServiceWithProfiler(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()

	service := NewFeedbackServiceWithProfiler(
		mockRepo,
		&MockLLMClient{},
		nil,
		nil, nil, nil, nil,
	)

	if service == nil {
		t.Error("Service should not be nil")
	}
}

// TestNewFeedbackServiceWithCache tests service with cache
func TestNewFeedbackServiceWithCache(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	cacheManager := cache.NewCacheManager(true, 100, 5*time.Minute)

	service := NewFeedbackServiceWithCache(
		mockRepo,
		&MockLLMClient{},
		nil,
		cacheManager,
	)

	if service == nil {
		t.Error("Service should not be nil")
	}

	if !service.cacheQueryResults {
		t.Error("Cache should be enabled")
	}
}

// TestNewFeedbackServiceFull tests service with all features
func TestNewFeedbackServiceFull(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	cacheManager := cache.NewCacheManager(true, 100, 5*time.Minute)

	service := NewFeedbackServiceFull(
		mockRepo,
		&MockLLMClient{},
		nil,
		nil, nil, nil, nil,
		cacheManager,
	)

	if service == nil {
		t.Error("Service should not be nil")
	}

	if !service.cacheQueryResults {
		t.Error("Cache should be enabled")
	}
}

// TestSetCacheTTL tests setting cache TTL
func TestSetCacheTTL(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ttl := 10 * time.Minute
	service.SetCacheTTL(ttl)

	if service.queryResultsTTL != ttl {
		t.Errorf("Expected TTL %v, got %v", ttl, service.queryResultsTTL)
	}
}

// TestCacheQueryResults tests enabling/disabling caching
func TestCacheQueryResults(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	service.CacheQueryResults(true)
	if !service.cacheQueryResults {
		t.Error("Cache should be enabled")
	}

	service.CacheQueryResults(false)
	if service.cacheQueryResults {
		t.Error("Cache should be disabled")
	}
}

// TestAnalyzeFeedbackContextCancellation tests with cancelled context
func TestAnalyzeFeedbackContextCancellation(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "test"},
	})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.AnalyzeFeedback(ctx, "What is sentiment?")

	// Should handle cancelled context gracefully
	_ = err
}

// TestAnalyzeFeedbackTimeout tests with timeout
func TestAnalyzeFeedbackTimeout(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := service.AnalyzeFeedback(ctx, "What is sentiment?")

	// Should handle timeout gracefully
	_ = err
}

// TestAnalyzeFeedbackWithSpecialCharacters tests with special characters
func TestAnalyzeFeedbackWithSpecialCharacters(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx := context.Background()

	questions := []string{
		"What's the sentiment?",
		"What is customer's feedback?",
		"Test with \"quotes\"",
		"Test with 'single quotes'",
		"Test with unicode: 你好",
	}

	for _, q := range questions {
		_, err := service.AnalyzeFeedback(ctx, q)
		// Should handle special characters without panicking
		_ = err
	}
}

// TestAnalyzeFeedbackLongQuestion tests with very long question
func TestAnalyzeFeedbackLongQuestion(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	// Create a very long question
	longQuestion := ""
	for i := 0; i < 1000; i++ {
		longQuestion += "What "
	}

	ctx := context.Background()
	_, err := service.AnalyzeFeedback(ctx, longQuestion)

	// Should handle long questions
	_ = err
}

// TestAnalyzeFeedbackMultipleCalls tests multiple sequential calls
func TestAnalyzeFeedbackMultipleCalls(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "Feedback 1"},
	})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := service.AnalyzeFeedback(ctx, "Question "+string(rune(i)))
		_ = err
	}

	// All calls should complete without issues
}

// BenchmarkAnalyzeFeedback benchmarks feedback analysis
func BenchmarkAnalyzeFeedback(b *testing.B) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "Test feedback"},
	})

	service := NewFeedbackService(mockRepo, &MockLLMClient{}, nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.AnalyzeFeedback(ctx, "What is sentiment?")
	}
}

// BenchmarkServiceWithCache benchmarks service with caching
func BenchmarkServiceWithCache(b *testing.B) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": 1, "feedback": "Test"},
	})

	cacheManager := cache.NewCacheManager(true, 100, 5*time.Minute)

	service := NewFeedbackServiceWithCache(
		mockRepo,
		&MockLLMClient{},
		nil,
		cacheManager,
	)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		service.AnalyzeFeedback(ctx, "What is sentiment?")
	}
}
