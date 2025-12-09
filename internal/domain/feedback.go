package domain

import "time"

// FeedbackEnriched represents a row from the feedback_enriched table
type FeedbackEnriched struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	Source       string    `json:"source"`
	ProductArea  string    `json:"product_area"`
	Sentiment    string    `json:"sentiment"`
	Priority     int       `json:"priority"`
	Topic        string    `json:"topic"`
	Region       string    `json:"region"`
	CustomerTier string    `json:"customer_tier"`
	Summary      string    `json:"summary"`
}

// AskRequest represents the incoming request to /api/ask
type AskRequest struct {
	Question string `json:"question"`
}

// AskResponse represents the final response from /api/ask
type AskResponse struct {
	Question     string              `json:"question"`
	DataPreview  []map[string]any    `json:"data_preview"`
	Summary      string              `json:"summary"`
	Recommendations []string         `json:"recommendations"`
	Actions      []ActionItem        `json:"actions"`
}

// ActionItem represents a proposed action/ticket
type ActionItem struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Magnitude   float64 `json:"magnitude,omitempty"`
}

// InsightResult is the parsed result from the LLM's insight generation
type InsightResult struct {
	Summary         string       `json:"summary"`
	Recommendations []string     `json:"recommendations"`
	Actions         []ActionItem `json:"actions"`
}

// AccountRiskScore represents ML-based churn predictions for an account (from tens-insight)
type AccountRiskScore struct {
	AccountID        string    `json:"account_id"`
	ChurnProbability float64   `json:"churn_probability"`
	HealthScore      float64   `json:"health_score"`
	RiskCategory     string    `json:"risk_category"`
	PredictedAt      time.Time `json:"predicted_at"`
	ModelVersion     string    `json:"model_version"`
}

// ProductAreaImpact represents ML-based priority signals for product areas (from tens-insight)
type ProductAreaImpact struct {
	ProductArea       string    `json:"product_area"`
	Segment           string    `json:"segment"`
	PriorityScore     float64   `json:"priority_score"`
	FeedbackCount     int       `json:"feedback_count"`
	AvgSentimentScore float64   `json:"avg_sentiment_score"`
	NegativeCount     int       `json:"negative_count"`
	CriticalCount     int       `json:"critical_count"`
	PredictedAt       time.Time `json:"predicted_at"`
	ModelVersion      string    `json:"model_version"`
}

// AccountHealthResponse is the response for GET /api/accounts/{id}/health
type AccountHealthResponse struct {
	AccountID           string  `json:"account_id"`
	ChurnProbability    float64 `json:"churn_probability"`
	HealthScore         float64 `json:"health_score"`
	RiskCategory        string  `json:"risk_category"`
	RecentNegativeCount int     `json:"recent_negative_feedback_count"`
	PredictedAt         string  `json:"predicted_at"`
	ModelVersion        string  `json:"model_version"`
}

// ProductAreaPriorityResponse is the response for GET /api/priorities/product-areas
type ProductAreaPriorityResponse struct {
	ProductAreas []ProductAreaImpact `json:"product_areas"`
}
