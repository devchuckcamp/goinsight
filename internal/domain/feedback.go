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
