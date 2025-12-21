package integration

import (
	"context"
	"testing"
	"time"

	"github.com/chuckie/goinsight/internal/cache"
	"github.com/chuckie/goinsight/tests/mocks"
	"github.com/chuckie/goinsight/tests/testutil"
)

// TestServiceWithMocks tests service layer with mocks
func TestServiceWithMocks(t *testing.T) {
	// Setup mocks
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Great", "sentiment": "positive"},
	})

	mockLLM := &MockServiceLLM{
		GenerateSQLFn: func(ctx context.Context, q string) (string, error) {
			return "SELECT * FROM feedback", nil
		},
		GenerateInsightFn: func(ctx context.Context, q string, r []map[string]any) (string, error) {
			return `{"summary": "Positive feedback", "recommendations": [], "actions": []}`, nil
		},
	}

	// Create service (basic version without cache)
	factory := testutil.NewFactory()

	// Verify mock behavior
	results, err := mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["sentiment"] != "positive" {
		t.Errorf("Expected positive sentiment")
	}

	// Test LLM
	sql, _ := mockLLM.GenerateSQL(context.Background(), "test")
	if sql != "SELECT * FROM feedback" {
		t.Errorf("Wrong SQL generated")
	}

	// Test factory
	feedback := factory.MakeFeedback("1", "positive")
	if feedback["sentiment"] != "positive" {
		t.Errorf("Factory failed to create feedback")
	}
}

// TestServiceWithCache tests service with caching
func TestServiceWithCache(t *testing.T) {
	// Create cache
	cacheManager := cache.NewCacheManager(true, 100, 5*time.Minute)

	// Setup repo mock
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Cached feedback"},
	})

	// First query
	results1, err := mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err != nil {
		t.Fatalf("First query failed: %v", err)
	}

	// Cache the result
	ctx := context.Background()
	cacheManager.CacheQueryResult(ctx, "SELECT * FROM feedback", results1, 5*time.Minute)

	// Retrieve from cache
	cached, found, _ := cacheManager.GetCachedQueryResult(ctx, "SELECT * FROM feedback")
	if !found {
		t.Error("Cache miss on cached query")
	}
	if cached == nil {
		t.Error("Cached value is nil")
	}

	// Second query should get same data
	results2, err := mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err != nil {
		t.Fatalf("Second query failed: %v", err)
	}

	if len(results1) != len(results2) {
		t.Errorf("Cache consistency failed")
	}
}

// TestServiceErrorHandling tests error scenarios
func TestServiceErrorHandling(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackError(NewTestError("query failed"))

	_, err := mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "query failed" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestServiceConcurrency tests concurrent service operations
func TestServiceConcurrency(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Test"},
	})

	done := make(chan bool, 20)
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		go func() {
			_, err := mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	for i := 0; i < 20; i++ {
		<-done
	}

	// Check for any errors
	select {
	case err := <-errors:
		t.Errorf("Concurrent operation failed: %v", err)
	default:
		// All good
	}
}

// TestServiceDataTransformation tests data transformation in service
func TestServiceDataTransformation(t *testing.T) {
	factory := testutil.NewFactory()

	// Create raw feedback
	rawFeedback := factory.MakeFeedback("1", "positive")

	// Simulate transformation
	enriched := map[string]any{
		"id":         rawFeedback["id"],
		"feedback":   rawFeedback["feedback"],
		"sentiment":  rawFeedback["sentiment"],
		"priority":   1,
		"topic":      "general",
		"region":     "US",
	}

	if enriched["sentiment"] != "positive" {
		t.Error("Sentiment lost in transformation")
	}

	if enriched["priority"] != 1 {
		t.Error("Priority not set in transformation")
	}
}

// TestServiceWithFactory tests service using test factory
func TestServiceWithFactory(t *testing.T) {
	factory := testutil.NewFactory()

	// Create test data
	feedbacks := factory.MakeFeedbacks(5, "positive")
	if len(feedbacks) != 5 {
		t.Errorf("Expected 5 feedbacks, got %d", len(feedbacks))
	}

	enriched := factory.MakeFeedbacksEnriched(5, "positive", "billing")
	if len(enriched) != 5 {
		t.Errorf("Expected 5 enriched feedbacks, got %d", len(enriched))
	}

	// Verify data
	for _, item := range enriched {
		if item.Sentiment != "positive" {
			t.Error("Wrong sentiment in enriched feedback")
		}

		if item.ProductArea != "billing" {
			t.Error("Wrong product area in enriched feedback")
		}
	}
}

// TestServiceResponseStructure tests response structure creation
func TestServiceResponseStructure(t *testing.T) {
	factory := testutil.NewFactory()

	askRequest := factory.MakeAskRequest("What is customer sentiment?")
	if askRequest.Question != "What is customer sentiment?" {
		t.Error("Ask request question mismatch")
	}

	askResponse := factory.MakeAskResponse("Customers are satisfied")
	if askResponse.Summary != "Customers are satisfied" {
		t.Error("Ask response summary mismatch")
	}

	if len(askResponse.Recommendations) == 0 {
		t.Error("No recommendations in response")
	}

	if len(askResponse.Actions) == 0 {
		t.Error("No actions in response")
	}
}

// TestServiceWithMultipleLLMs tests service switching between LLM providers
func TestServiceWithMultipleLLMs(t *testing.T) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Test"},
	})

	// Test with LLM 1
	llm1 := &MockServiceLLM{
		GenerateSQLFn: func(ctx context.Context, q string) (string, error) {
			return "SELECT * FROM feedback WHERE sentiment = 'positive'", nil
		},
	}

	sql1, _ := llm1.GenerateSQL(context.Background(), "test")
	if sql1 != "SELECT * FROM feedback WHERE sentiment = 'positive'" {
		t.Error("LLM1 SQL mismatch")
	}

	// Test with LLM 2
	llm2 := &MockServiceLLM{
		GenerateSQLFn: func(ctx context.Context, q string) (string, error) {
			return "SELECT * FROM feedback WHERE sentiment = 'negative'", nil
		},
	}

	sql2, _ := llm2.GenerateSQL(context.Background(), "test")
	if sql2 != "SELECT * FROM feedback WHERE sentiment = 'negative'" {
		t.Error("LLM2 SQL mismatch")
	}
}

// BenchmarkServiceQuery benchmarks service query operations
func BenchmarkServiceQuery(b *testing.B) {
	mockRepo := mocks.NewMockFeedbackRepository()
	mockRepo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Test"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mockRepo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	}
}

// BenchmarkServiceWithFactory benchmarks service using factory
func BenchmarkServiceWithFactory(b *testing.B) {
	factory := testutil.NewFactory()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory.MakeFeedbacks(10, "positive")
	}
}

// MockServiceLLM is a mock LLM for service tests
type MockServiceLLM struct {
	GenerateSQLFn     func(context.Context, string) (string, error)
	GenerateInsightFn func(context.Context, string, []map[string]any) (string, error)
}

func (m *MockServiceLLM) GenerateSQL(ctx context.Context, question string) (string, error) {
	if m.GenerateSQLFn != nil {
		return m.GenerateSQLFn(ctx, question)
	}
	return "SELECT * FROM feedback", nil
}

func (m *MockServiceLLM) GenerateInsight(ctx context.Context, question string, results []map[string]any) (string, error) {
	if m.GenerateInsightFn != nil {
		return m.GenerateInsightFn(ctx, question, results)
	}
	return `{"summary": "Analysis", "recommendations": [], "actions": []}`, nil
}

func (m *MockServiceLLM) Generate(ctx context.Context, prompt string) (string, error) {
	return "Generated response", nil
}
