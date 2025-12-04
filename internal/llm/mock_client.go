package llm

import (
	"context"
	"fmt"
	"strings"
)

// MockClient is a mock implementation of the Client interface for testing
type MockClient struct{}

// NewMockClient creates a new mock LLM client
func NewMockClient() *MockClient {
	return &MockClient{}
}

// GenerateSQL returns a simple mock SQL query
func (m *MockClient) GenerateSQL(ctx context.Context, question string) (string, error) {
	// Return different queries based on the question to make it more realistic
	// All queries are safe SELECT statements
	questionLower := strings.ToLower(question)
	
	if strings.Contains(questionLower, "billing") || strings.Contains(questionLower, "payment") {
		return "SELECT product_area, topic, COUNT(*) as count FROM feedback_enriched WHERE product_area = 'billing' GROUP BY product_area, topic ORDER BY count DESC LIMIT 10;", nil
	}
	
	if strings.Contains(questionLower, "critical") || strings.Contains(questionLower, "priority") {
		return "SELECT * FROM feedback_enriched WHERE priority >= 4 ORDER BY priority DESC, created_at DESC LIMIT 20;", nil
	}
	
	if strings.Contains(questionLower, "enterprise") {
		return "SELECT * FROM feedback_enriched WHERE customer_tier = 'enterprise' ORDER BY created_at DESC LIMIT 15;", nil
	}
	
	if strings.Contains(questionLower, "performance") {
		return "SELECT * FROM feedback_enriched WHERE product_area = 'performance' ORDER BY priority DESC, created_at DESC LIMIT 15;", nil
	}
	
	if strings.Contains(questionLower, "sentiment") {
		return "SELECT sentiment, COUNT(*) as count FROM feedback_enriched GROUP BY sentiment ORDER BY count DESC;", nil
	}
	
	// Default safe query
	return "SELECT * FROM feedback_enriched ORDER BY created_at DESC LIMIT 10;", nil
}

// GenerateInsight returns a mock insight response
func (m *MockClient) GenerateInsight(ctx context.Context, question string, queryResults []map[string]any) (string, error) {
	mockResponse := fmt.Sprintf(`{
  "summary": "Mock analysis for: %s. Found %d records in the database.",
  "recommendations": [
    "This is a mock recommendation - configure OPENAI_API_KEY for real insights",
    "Review the data patterns shown in the preview",
    "Consider implementing the mock client interface with a real LLM provider"
  ],
  "actions": [
    {
      "title": "Configure OpenAI API Key",
      "description": "Set the OPENAI_API_KEY environment variable to enable real LLM-powered insights"
    }
  ]
}`, question, len(queryResults))
	return mockResponse, nil
}

// Generate returns a simple mock response for any prompt
func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	return `{"message": "Mock LLM client - configure a real LLM provider for actual responses"}`, nil
}
