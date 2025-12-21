package repository

import (
	"database/sql"
	"fmt"
)

// Repositories holds all repository instances for the application
// Implements dependency injection pattern for cleaner service initialization
type Repositories struct {
	Feedback FeedbackRepository
	db       *sql.DB
}

// NewRepositories creates and initializes all repository instances
// This is the single entry point for repository initialization
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Feedback: NewPostgresFeedbackRepository(db),
		db:       db,
	}
}

// RepositoryConfig holds configuration for repository initialization
type RepositoryConfig struct {
	// Connection pooling
	MaxOpenConnections int
	MaxIdleConnections int
	ConnMaxLifetime    int // seconds

	// Query execution
	QueryTimeout       int // seconds
	StatementCacheSize int
}

// DefaultRepositoryConfig returns sensible defaults for repository configuration
func DefaultRepositoryConfig() RepositoryConfig {
	return RepositoryConfig{
		MaxOpenConnections: 25,
		MaxIdleConnections: 5,
		ConnMaxLifetime:    3600, // 1 hour
		QueryTimeout:       30,   // 30 seconds
		StatementCacheSize: 100,
	}
}

// ApplyConfig applies configuration settings to the database connection
func (r *Repositories) ApplyConfig(config RepositoryConfig) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	r.db.SetMaxOpenConns(config.MaxOpenConnections)
	r.db.SetMaxIdleConns(config.MaxIdleConnections)
	// Note: SetConnMaxLifetime would be used here if needed
	// r.db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetime) * time.Second)

	return nil
}

// ValidateConnection checks if the database connection is healthy
func (r *Repositories) ValidateConnection() error {
	if r.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	if err := r.db.Ping(); err != nil {
		return fmt.Errorf("database connection check failed: %w", err)
	}

	return nil
}

// Close closes the database connection (call in defer)
func (r *Repositories) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
