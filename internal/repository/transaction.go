package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/chuckie/goinsight/internal/domain"
)

// Transaction represents a database transaction
// Allows grouping multiple operations into a single atomic unit
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error

	// GetRepository returns a repository scoped to this transaction
	GetRepository() FeedbackRepository
}

// PostgresTransaction implements the Transaction interface
type PostgresTransaction struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback()
}

// GetRepository returns a repository scoped to this transaction
// Note: For now, returns a repository using the transaction connection
// In a production implementation, you'd create a special transactional repository
func (t *PostgresTransaction) GetRepository() FeedbackRepository {
	// Create a simple wrapper that executes queries on the transaction
	return &transactionalRepository{tx: t.tx}
}

// transactionalRepository is a helper for executing queries within a transaction
type transactionalRepository struct {
	tx *sql.Tx
}

// QueryFeedback executes a feedback query within the transaction
func (r *transactionalRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
	rows, err := r.tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]any)
		for i, col := range columns {
			val := values[i]
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

	return results, rows.Err()
}

// Stub implementations for other methods (not used in transactions typically)
func (r *transactionalRepository) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
	return nil, fmt.Errorf("not implemented in transaction")
}

func (r *transactionalRepository) GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error) {
	return 0, fmt.Errorf("not implemented in transaction")
}

func (r *transactionalRepository) GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error) {
	return nil, fmt.Errorf("not implemented in transaction")
}

func (r *transactionalRepository) GetFeedbackEnrichedCount(ctx context.Context) (int, error) {
	return 0, fmt.Errorf("not implemented in transaction")
}

// TransactionOptions holds configuration for transactions
type TransactionOptions struct {
	Isolation sql.IsolationLevel
	ReadOnly  bool
}

// DefaultTransactionOptions returns sensible defaults
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}
}

// BeginTransaction starts a new transaction
// Usage:
//   tx, err := repos.BeginTransaction(ctx, DefaultTransactionOptions())
//   if err != nil { return err }
//   defer tx.Rollback()
//   repo := tx.GetRepository()
//   // ... perform operations ...
//   return tx.Commit()
func (r *Repositories) BeginTransaction(ctx context.Context, opts TransactionOptions) (Transaction, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	txOpts := &sql.TxOptions{
		Isolation: opts.Isolation,
		ReadOnly:  opts.ReadOnly,
	}

	tx, err := r.db.BeginTx(ctx, txOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresTransaction{tx: tx}, nil
}
