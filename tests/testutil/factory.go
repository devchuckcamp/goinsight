package testutil

import (
	"context"
	"time"

	"github.com/chuckie/goinsight/internal/domain"
)

// Factory provides factory methods for creating test objects
type Factory struct {
	timestamp time.Time
}

// NewFactory creates a new test factory
func NewFactory() *Factory {
	return &Factory{
		timestamp: time.Date(2025, 12, 20, 10, 30, 0, 0, time.UTC),
	}
}

// MakeFeedback creates a test feedback item
func (f *Factory) MakeFeedback(id string, sentiment string) map[string]any {
	return map[string]any{
		"id":        id,
		"feedback":  "Test feedback for " + id,
		"sentiment": sentiment,
		"source":    "email",
		"created_at": f.timestamp,
	}
}

// MakeFeedbacks creates multiple test feedback items
func (f *Factory) MakeFeedbacks(count int, sentiment string) []map[string]any {
	results := make([]map[string]any, count)
	for i := 0; i < count; i++ {
		id := string(rune('0' + i))
		results[i] = f.MakeFeedback(id, sentiment)
	}
	return results
}

// MakeFeedbackEnriched creates a test enriched feedback item
func (f *Factory) MakeFeedbackEnriched(id string, sentiment string, productArea string) domain.FeedbackEnriched {
	return domain.FeedbackEnriched{
		ID:           id,
		CreatedAt:    f.timestamp,
		Source:       "email",
		ProductArea:  productArea,
		Sentiment:    sentiment,
		Priority:     1,
		Topic:        "general",
		Region:       "US",
		CustomerTier: "standard",
		Summary:      "Test summary for " + id,
	}
}

// MakeFeedbacksEnriched creates multiple test enriched feedback items
func (f *Factory) MakeFeedbacksEnriched(count int, sentiment string, productArea string) []domain.FeedbackEnriched {
	results := make([]domain.FeedbackEnriched, count)
	for i := 0; i < count; i++ {
		id := string(rune('0' + i))
		results[i] = f.MakeFeedbackEnriched(id, sentiment, productArea)
	}
	return results
}

// MakeAskRequest creates a test Ask request
func (f *Factory) MakeAskRequest(question string) domain.AskRequest {
	return domain.AskRequest{
		Question: question,
	}
}

// MakeAskResponse creates a test Ask response
func (f *Factory) MakeAskResponse(summary string) domain.AskResponse {
	return domain.AskResponse{
		Question:    "Test question",
		Summary:     summary,
		SQL:         "SELECT * FROM feedback",
		DataPreview: []map[string]any{},
		Recommendations: []string{
			"Improve service quality",
		},
		Actions: []domain.ActionItem{
			{
				Title:       "Follow up",
				Description: "Contact customers",
				Magnitude:   0.8,
			},
		},
	}
}

// MakeAccountRiskScore creates a test account risk score
func (f *Factory) MakeAccountRiskScore(accountID string, churnProbability float64) domain.AccountRiskScore {
	return domain.AccountRiskScore{
		AccountID:        accountID,
		ChurnProbability: churnProbability,
		HealthScore:      1.0 - churnProbability,
		RiskCategory:     getRiskCategory(churnProbability),
		PredictedAt:      f.timestamp,
		ModelVersion:     "v1.0",
	}
}

// MakeProductAreaImpact creates a test product area impact
func (f *Factory) MakeProductAreaImpact(productArea string, priorityScore float64) domain.ProductAreaImpact {
	return domain.ProductAreaImpact{
		ProductArea:       productArea,
		Segment:           "enterprise",
		PriorityScore:     priorityScore,
		FeedbackCount:     10,
		AvgSentimentScore: 0.6,
		NegativeCount:     4,
		CriticalCount:     1,
		PredictedAt:       f.timestamp,
		ModelVersion:      "v1.0",
	}
}

// getRiskCategory determines risk category from churn probability
func getRiskCategory(churnProbability float64) string {
	switch {
	case churnProbability >= 0.7:
		return "critical"
	case churnProbability >= 0.5:
		return "high"
	case churnProbability >= 0.3:
		return "medium"
	default:
		return "low"
	}
}

// MockDBConfig provides configuration for mock databases
type MockDBConfig struct {
	EnableTx     bool
	EnableEvents bool
	Timeout      time.Duration
}

// DefaultMockDBConfig returns default mock DB configuration
func DefaultMockDBConfig() MockDBConfig {
	return MockDBConfig{
		EnableTx:     true,
		EnableEvents: true,
		Timeout:      5 * time.Second,
	}
}

// ContextWithTimeout creates a context with timeout
func ContextWithTimeout(duration time.Duration) (context.Context, func()) {
	return context.WithTimeout(context.Background(), duration)
}

// ContextWithCancel creates a context with cancel function
func ContextWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// WaitForCondition waits for a condition to be true with timeout
func WaitForCondition(timeout time.Duration, check func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// AssertFieldEquals asserts that two values are equal
func AssertFieldEquals(t interface{ Errorf(string, ...interface{}) }, name string, got, want interface{}) {
	if got != want {
		t.Errorf("%s: expected %v, got %v", name, want, got)
	}
}

// AssertFieldNotEmpty asserts that a field is not empty
func AssertFieldNotEmpty(t interface{ Errorf(string, ...interface{}) }, name string, value interface{}) {
	if value == nil || value == "" || value == 0 {
		t.Errorf("%s: expected non-empty value", name)
	}
}
