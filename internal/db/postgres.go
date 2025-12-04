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

			// Convert []byte to string for better JSON serialization
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
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
