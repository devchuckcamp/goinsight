package integration

import (
	"time"
)

// GetTestTime returns a consistent test time
func GetTestTime() time.Time {
	return time.Date(2025, 12, 20, 10, 30, 0, 0, time.UTC)
}

// TestError is a simple error implementation for testing
type TestError struct {
	message string
}

// NewTestError creates a new test error
func NewTestError(msg string) *TestError {
	return &TestError{message: msg}
}

// Error implements the error interface
func (e *TestError) Error() string {
	return e.message
}

// CreateTestFeedback creates a test feedback item with default values
func CreateTestFeedback(id string, sentiment string) map[string]any {
	return map[string]any{
		"id":       id,
		"feedback": "Test feedback",
		"sentiment": sentiment,
		"source":   "email",
		"created_at": GetTestTime(),
	}
}

// CreateTestFeedbacks creates multiple test feedback items
func CreateTestFeedbacks(count int, sentiment string) []map[string]any {
	results := make([]map[string]any, count)
	for i := 0; i < count; i++ {
		results[i] = CreateTestFeedback(string(rune(i)), sentiment)
	}
	return results
}

// CreateTestEnrichedFeedback creates test enriched feedback with defaults
func CreateTestEnrichedFeedback(id string, sentiment string, productArea string) map[string]any {
	return map[string]any{
		"id":            id,
		"sentiment":     sentiment,
		"product_area":  productArea,
		"priority":      1,
		"topic":         "general",
		"region":        "US",
		"customer_tier": "standard",
		"summary":       "Test summary",
		"created_at":    GetTestTime(),
	}
}
