package testutil

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// TestDB represents a test database connection
type TestDB struct {
	DB       *sql.DB
	Cleanup  func() error
	URL      string
	testName string
	t        *testing.T
}

// NewTestDB creates a new test database connection
// For unit tests, use NewTestDBMock() instead
func NewTestDB(t *testing.T) *TestDB {
	// This is a placeholder that will use mock database
	// For actual integration tests, override this with real connection
	return &TestDB{
		DB:       nil,
		testName: t.Name(),
		t:        t,
	}
}

// NewTestDBWithConnection creates a test DB with actual connection
func NewTestDBWithConnection(t *testing.T, databaseURL string) (*TestDB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Test the connection with retries
	var lastErr error
	for i := 0; i < 5; i++ {
		err := db.Ping()
		if err == nil {
			break
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}

	if lastErr != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping test database: %w", lastErr)
	}

	// Configure connection pool for tests
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &TestDB{
		DB:       db,
		URL:      databaseURL,
		testName: t.Name(),
		t:        t,
		Cleanup: func() error {
			return db.Close()
		},
	}, nil
}

// ExecSQL executes a SQL statement for testing
func (tdb *TestDB) ExecSQL(query string) error {
	if tdb.DB == nil {
		return fmt.Errorf("test database not initialized")
	}

	_, err := tdb.DB.Exec(query)
	return err
}

// ClearTable removes all rows from a table
func (tdb *TestDB) ClearTable(tableName string) error {
	return tdb.ExecSQL(fmt.Sprintf("DELETE FROM %s", tableName))
}

// ClearAllTables removes all rows from all tables (careful with this)
func (tdb *TestDB) ClearAllTables(tableNames ...string) error {
	for _, table := range tableNames {
		if err := tdb.ClearTable(table); err != nil {
			return err
		}
	}
	return nil
}

// BeginTx starts a test transaction
func (tdb *TestDB) BeginTx() (*sql.Tx, error) {
	if tdb.DB == nil {
		return nil, fmt.Errorf("test database not initialized")
	}
	return tdb.DB.Begin()
}

// Close closes the test database connection
func (tdb *TestDB) Close() error {
	if tdb.Cleanup != nil {
		return tdb.Cleanup()
	}
	if tdb.DB != nil {
		return tdb.DB.Close()
	}
	return nil
}

// Logf logs a message for debugging tests
func (tdb *TestDB) Logf(format string, args ...interface{}) {
	tdb.t.Logf("[%s] %s", tdb.testName, fmt.Sprintf(format, args...))
}

// TestHelper provides common test utilities
type TestHelper struct {
	t *testing.T
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{t: t}
}

// AssertNoError fails the test if err is not nil
func (h *TestHelper) AssertNoError(err error, msg string) {
	if err != nil {
		h.t.Errorf("%s: %v", msg, err)
	}
}

// AssertError fails the test if err is nil
func (h *TestHelper) AssertError(err error, msg string) {
	if err == nil {
		h.t.Errorf("%s: expected error but got nil", msg)
	}
}

// AssertEqual fails the test if expected != actual
func (h *TestHelper) AssertEqual(expected, actual interface{}, msg string) {
	if expected != actual {
		h.t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertNotEqual fails the test if expected == actual
func (h *TestHelper) AssertNotEqual(expected, actual interface{}, msg string) {
	if expected == actual {
		h.t.Errorf("%s: expected not equal to %v, but got %v", msg, expected, actual)
	}
}

// AssertTrue fails the test if condition is false
func (h *TestHelper) AssertTrue(condition bool, msg string) {
	if !condition {
		h.t.Errorf("%s: expected true, got false", msg)
	}
}

// AssertFalse fails the test if condition is true
func (h *TestHelper) AssertFalse(condition bool, msg string) {
	if condition {
		h.t.Errorf("%s: expected false, got true", msg)
	}
}
