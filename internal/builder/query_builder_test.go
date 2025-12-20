package builder

import (
	"strings"
	"testing"
)

func TestQueryBuilderWithParams(t *testing.T) {
	qb := NewQueryBuilder()
	query, params := qb.
		Select("id", "name").
		From("users").
		WhereParam("email = %s", "test@example.com").
		WhereParam("status = %s", "active").
		BuildWithParams()

	expectedQuery := "SELECT id, name\nFROM users\nWHERE email = $1\nAND status = $2"
	if query != expectedQuery {
		t.Errorf("Expected query:\n%s\n\nGot:\n%s", expectedQuery, query)
	}

	if len(params) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(params))
	}

	if params[0] != "test@example.com" {
		t.Errorf("Expected first param to be 'test@example.com', got '%v'", params[0])
	}

	if params[1] != "active" {
		t.Errorf("Expected second param to be 'active', got '%v'", params[1])
	}
}

func TestFeedbackQueryBuilderWithSentiment(t *testing.T) {
	fqb := NewFeedbackQueryBuilder()
	query, params := fqb.
		WithSentiment("negative").
		BuildFeedbackWithParams()

	if !strings.Contains(query, "sentiment = $1") {
		t.Errorf("Expected query to contain 'sentiment = $1', got:\n%s", query)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}

	if params[0] != "negative" {
		t.Errorf("Expected param to be 'negative', got '%v'", params[0])
	}
}

func TestFeedbackQueryBuilderMultipleFilters(t *testing.T) {
	fqb := NewFeedbackQueryBuilder()
	query, params := fqb.
		WithSentiment("negative").
		WithProductArea("billing").
		WithRegion("US").
		BuildFeedbackWithParams()

	if !strings.Contains(query, "sentiment = $1") {
		t.Errorf("Expected query to contain 'sentiment = $1', got:\n%s", query)
	}

	if !strings.Contains(query, "product_area = $2") {
		t.Errorf("Expected query to contain 'product_area = $2', got:\n%s", query)
	}

	if !strings.Contains(query, "region = $3") {
		t.Errorf("Expected query to contain 'region = $3', got:\n%s", query)
	}

	if len(params) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(params))
	}

	if params[0] != "negative" {
		t.Errorf("Expected first param to be 'negative', got '%v'", params[0])
	}

	if params[1] != "billing" {
		t.Errorf("Expected second param to be 'billing', got '%v'", params[1])
	}

	if params[2] != "US" {
		t.Errorf("Expected third param to be 'US', got '%v'", params[2])
	}
}

func TestFeedbackQueryBuilderEmptyValues(t *testing.T) {
	fqb := NewFeedbackQueryBuilder()
	query, params := fqb.
		WithSentiment("").
		WithProductArea("").
		WithRegion("").
		BuildFeedbackWithParams()

	if strings.Contains(query, "sentiment") {
		t.Errorf("Expected query not to contain 'sentiment' filter, got:\n%s", query)
	}

	if len(params) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(params))
	}
}

func TestQueryBuilderBackwardCompatibility(t *testing.T) {
	// Test that Build() still works for simple queries without parameters
	qb := NewQueryBuilder()
	query := qb.
		Select("id", "name").
		From("users").
		Where("status = 'active'").
		Build()

	expectedQuery := "SELECT id, name\nFROM users\nWHERE status = 'active'"
	if query != expectedQuery {
		t.Errorf("Expected query:\n%s\n\nGot:\n%s", expectedQuery, query)
	}
}

func TestSQLInjectionPrevention(t *testing.T) {
	// Test that malicious input is properly parameterized
	maliciousInput := "'; DROP TABLE users; --"

	fqb := NewFeedbackQueryBuilder()
	query, params := fqb.
		WithSentiment(maliciousInput).
		BuildFeedbackWithParams()

	// The malicious input should be in params, not in the query string
	if strings.Contains(query, "DROP TABLE") {
		t.Errorf("Query contains unescaped malicious input:\n%s", query)
	}

	if len(params) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(params))
	}

	if params[0] != maliciousInput {
		t.Errorf("Expected param to be '%s', got '%v'", maliciousInput, params[0])
	}

	// Verify the query uses a placeholder
	if !strings.Contains(query, "sentiment = $1") {
		t.Errorf("Expected query to use placeholder $1, got:\n%s", query)
	}
}
