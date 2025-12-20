package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chuckie/goinsight/internal/domain"
)

// FeedbackRepository defines the interface for feedback data access operations
type FeedbackRepository interface {
	// QueryFeedback executes a feedback query and returns results
	QueryFeedback(ctx context.Context, query string, args ...any) ([]map[string]any, error)

	// GetAccountRiskScore retrieves ML predictions for a specific account
	GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error)

	// GetRecentNegativeFeedbackCount counts recent negative feedback for an account
	GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error)

	// GetProductAreaImpacts retrieves ML predictions for product area priorities
	GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error)

	// GetFeedbackEnrichedCount returns total count of enriched feedback records
	GetFeedbackEnrichedCount(ctx context.Context) (int, error)
}

// PostgresFeedbackRepository implements FeedbackRepository for PostgreSQL
type PostgresFeedbackRepository struct {
	db *sql.DB
}

// NewPostgresFeedbackRepository creates a new PostgreSQL feedback repository
func NewPostgresFeedbackRepository(db *sql.DB) *PostgresFeedbackRepository {
	return &PostgresFeedbackRepository{db: db}
}

// QueryFeedback executes a feedback query and returns results as maps
func (r *PostgresFeedbackRepository) QueryFeedback(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result slice
	var results []map[string]any

	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]

			// Handle different data types for better JSON serialization
			switch v := val.(type) {
			case []byte:
				row[col] = string(v)
			case nil:
				row[col] = nil
			default:
				row[col] = v
			}
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// GetAccountRiskScore retrieves ML predictions for a specific account
func (r *PostgresFeedbackRepository) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
	query := `
		SELECT account_id, churn_probability, health_score, risk_category, predicted_at, model_version
		FROM account_risk_scores
		WHERE account_id = $1
	`

	var score domain.AccountRiskScore
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&score.AccountID,
		&score.ChurnProbability,
		&score.HealthScore,
		&score.RiskCategory,
		&score.PredictedAt,
		&score.ModelVersion,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account risk score: %w", err)
	}

	return &score, nil
}

// GetRecentNegativeFeedbackCount counts recent negative feedback for an account
func (r *PostgresFeedbackRepository) GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM feedback_enriched
		WHERE customer_tier LIKE '%' || $1 || '%'
		AND sentiment = 'negative'
		AND created_at > NOW() - INTERVAL '30 days'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count negative feedback: %w", err)
	}
	return count, nil
}

// GetProductAreaImpacts retrieves ML predictions for product area priorities
func (r *PostgresFeedbackRepository) GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error) {
	var query string
	var args []any

	if segment != "" {
		query = `
			SELECT product_area, segment, priority_score, feedback_count, 
			       avg_sentiment_score, negative_count, critical_count, 
			       predicted_at, model_version
			FROM product_area_impact
			WHERE segment = $1
			ORDER BY priority_score DESC
		`
		args = append(args, segment)
	} else {
		query = `
			SELECT product_area, segment, priority_score, feedback_count,
			       avg_sentiment_score, negative_count, critical_count,
			       predicted_at, model_version
			FROM product_area_impact
			ORDER BY priority_score DESC
		`
	}

	return r.QueryFeedback(ctx, query)
}

// GetFeedbackEnrichedCount returns the total count of enriched feedback records
func (r *PostgresFeedbackRepository) GetFeedbackEnrichedCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM feedback_enriched`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count feedback: %w", err)
	}
	return count, nil
}
