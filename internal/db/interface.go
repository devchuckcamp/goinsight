package db

import (
	"database/sql"
)

// DatabaseClient is an interface for database operations
type DatabaseClient interface {
	ExecuteQuery(query string) ([]map[string]any, error)
	HealthCheck() error
	GetAccountRiskScore(accountID string) (*sql.Row, error)
	GetRecentNegativeFeedbackCount(accountID string) (int, error)
	GetProductAreaImpacts(segment string) ([]map[string]any, error)
	Close() error
}
