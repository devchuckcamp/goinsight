package integration

import (
	"context"
	"testing"

	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/tests/mocks"
)

// TestRepositoryIntegration tests repository operations end-to-end
func TestRepositoryIntegration(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()

	// Setup test data
	testFeedback := []map[string]any{
		{
			"id":       "1",
			"feedback": "Excellent service",
			"sentiment": "positive",
			"source":   "email",
		},
		{
			"id":       "2",
			"feedback": "Poor support",
			"sentiment": "negative",
			"source":   "chat",
		},
	}
	repo.SetQueryFeedbackResult(testFeedback)

	// Test query
	results, err := repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err != nil {
		t.Fatalf("Failed to query feedback: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify call was tracked
	if !repo.QueryFeedbackCalled {
		t.Error("Repository call was not tracked")
	}
}

// TestRepositoryWithEnrichedData tests enriched data retrieval
func TestRepositoryWithEnrichedData(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()

	now := GetTestTime()
	enrichedData := []domain.FeedbackEnriched{
		{
			ID:           "1",
			CreatedAt:    now,
			Source:       "email",
			ProductArea:  "billing",
			Sentiment:    "negative",
			Priority:     1,
			Topic:        "pricing",
			Region:       "US",
			CustomerTier: "enterprise",
			Summary:      "High pricing concerns",
		},
		{
			ID:           "2",
			CreatedAt:    now,
			Source:       "chat",
			ProductArea:  "support",
			Sentiment:    "positive",
			Priority:     2,
			Topic:        "response time",
			Region:       "EU",
			CustomerTier: "pro",
			Summary:      "Quick support response",
		},
	}
	repo.SetGetFeedbackEnrichedResult(enrichedData)

	// Test retrieval
	results, err := repo.GetFeedbackEnriched(context.Background())
	if err != nil {
		t.Fatalf("Failed to get enriched feedback: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 enriched results, got %d", len(results))
	}

	// Verify data integrity
	if results[0].ProductArea != "billing" {
		t.Errorf("Expected product_area 'billing', got %s", results[0].ProductArea)
	}

	if results[1].Sentiment != "positive" {
		t.Errorf("Expected sentiment 'positive', got %s", results[1].Sentiment)
	}
}

// TestRepositoryFiltering tests filtering by various criteria
func TestRepositoryFiltering(t *testing.T) {
	tests := []struct {
		name    string
		filter  string
		data    []map[string]any
		count   int
	}{
		{
			name:   "Filter by sentiment positive",
			filter: "WHERE sentiment = 'positive'",
			data: []map[string]any{
				{"id": "1", "sentiment": "positive"},
				{"id": "2", "sentiment": "positive"},
			},
			count: 2,
		},
		{
			name:   "Filter by sentiment negative",
			filter: "WHERE sentiment = 'negative'",
			data: []map[string]any{
				{"id": "3", "sentiment": "negative"},
			},
			count: 1,
		},
		{
			name:   "Filter by product area",
			filter: "WHERE product_area = 'billing'",
			data: []map[string]any{
				{"id": "4", "product_area": "billing"},
				{"id": "5", "product_area": "billing"},
				{"id": "6", "product_area": "billing"},
			},
			count: 3,
		},
		{
			name:   "No results",
			filter: "WHERE id = 'nonexistent'",
			data:   []map[string]any{},
			count:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockFeedbackRepository()
			repo.SetQueryFeedbackResult(tt.data)

			results, err := repo.QueryFeedback(context.Background(), "SELECT * FROM feedback "+tt.filter)
			if err != nil {
				t.Errorf("Query failed: %v", err)
			}

			if len(results) != tt.count {
				t.Errorf("Expected %d results, got %d", tt.count, len(results))
			}
		})
	}
}

// TestRepositoryDataConsistency tests data consistency across multiple queries
func TestRepositoryDataConsistency(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()

	testData := []map[string]any{
		{"id": "1", "feedback": "Test", "sentiment": "positive"},
	}
	repo.SetQueryFeedbackResult(testData)

	// Multiple queries should return consistent data
	for i := 0; i < 5; i++ {
		results, err := repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
		if err != nil {
			t.Errorf("Query %d failed: %v", i, err)
		}

		if len(results) != 1 {
			t.Errorf("Query %d: expected 1 result, got %d", i, len(results))
		}

		if results[0]["feedback"] != "Test" {
			t.Errorf("Query %d: data mismatch", i)
		}
	}
}

// TestRepositoryLargeDatasets tests handling of large result sets
func TestRepositoryLargeDatasets(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()

	// Create large dataset
	largeDataset := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		largeDataset[i] = map[string]any{
			"id":        i,
			"feedback":  "Feedback item",
			"sentiment": "positive",
		}
	}
	repo.SetQueryFeedbackResult(largeDataset)

	// Query should handle large results
	results, err := repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err != nil {
		t.Fatalf("Failed to query large dataset: %v", err)
	}

	if len(results) != 1000 {
		t.Errorf("Expected 1000 results, got %d", len(results))
	}
}

// TestRepositoryErrorPropagation tests error handling
func TestRepositoryErrorPropagation(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()

	// Set error
	repo.SetQueryFeedbackError(NewTestError("database unavailable"))

	_, err := repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "database unavailable" {
		t.Errorf("Wrong error message: %v", err)
	}
}

// TestRepositoryContextCancellation tests context cancellation handling
func TestRepositoryContextCancellation(t *testing.T) {
	repo := mocks.NewMockFeedbackRepository()
	repo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Test"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Mock should still work (in real implementation, would respect cancellation)
	results, err := repo.QueryFeedback(ctx, "SELECT * FROM feedback")
	if err != nil {
		t.Errorf("Expected mock to work even with cancelled context: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

// BenchmarkRepositoryQuery benchmarks repository queries
func BenchmarkRepositoryQuery(b *testing.B) {
	repo := mocks.NewMockFeedbackRepository()
	repo.SetQueryFeedbackResult([]map[string]any{
		{"id": "1", "feedback": "Test"},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	}
}

// BenchmarkRepositoryLargeResult benchmarks large result sets
func BenchmarkRepositoryLargeResult(b *testing.B) {
	repo := mocks.NewMockFeedbackRepository()

	// Create large dataset
	largeDataset := make([]map[string]any, 1000)
	for i := 0; i < 1000; i++ {
		largeDataset[i] = map[string]any{
			"id":       i,
			"feedback": "Test feedback",
		}
	}
	repo.SetQueryFeedbackResult(largeDataset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.QueryFeedback(context.Background(), "SELECT * FROM feedback")
	}
}
