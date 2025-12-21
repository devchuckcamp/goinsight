package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Client wraps a database connection
type Client struct {
	db *sql.DB
}

// NewClient creates a new database client with connection retry logic
func NewClient(databaseURL string) (*Client, error) {
	var db *sql.DB
	var err error

	// Retry logic for Docker environments where DB might not be ready immediately
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		// Test the connection
		err = db.Ping()
		if err == nil {
			break
		}

		if i < maxRetries-1 {
			fmt.Printf("Database not ready, retrying in 2 seconds... (attempt %d/%d)\n", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("Successfully connected to database")

	return &Client{db: db}, nil
}

// DB returns the underlying *sql.DB connection
// Useful for creating repositories or other low-level operations
func (c *Client) DB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// ExecuteQuery executes a SQL query and returns results as a slice of maps
func (c *Client) ExecuteQuery(query string) ([]map[string]any, error) {
	rows, err := c.db.Query(query)
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
				// Convert byte arrays to strings
				row[col] = string(v)
			case nil:
				// Keep null values as nil (will serialize to null in JSON)
				row[col] = nil
			default:
				// Keep all other types as-is
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

// HealthCheck verifies the database connection is healthy
func (c *Client) HealthCheck() error {
	return c.db.Ping()
}

// GetAccountRiskScore retrieves ML predictions for a specific account
func (c *Client) GetAccountRiskScore(accountID string) (*sql.Row, error) {
	query := `
		SELECT account_id, churn_probability, health_score, risk_category, predicted_at, model_version
		FROM account_risk_scores
		WHERE account_id = $1
	`
	return c.db.QueryRow(query, accountID), nil
}

// GetRecentNegativeFeedbackCount counts recent negative feedback for an account
// Note: This assumes feedback_enriched has an account_id column that needs to be added
func (c *Client) GetRecentNegativeFeedbackCount(accountID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM feedback_enriched
		WHERE customer_tier LIKE '%' || $1 || '%'
		AND sentiment = 'negative'
		AND created_at > NOW() - INTERVAL '30 days'
	`
	var count int
	err := c.db.QueryRow(query, accountID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count negative feedback: %w", err)
	}
	return count, nil
}

// GetProductAreaImpacts retrieves ML predictions for product area priorities
func (c *Client) GetProductAreaImpacts(segment string) ([]map[string]any, error) {
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

	return c.ExecuteQuery(query)
}
