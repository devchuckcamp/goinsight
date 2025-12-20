package builder

import (
	"fmt"
	"strings"
)

// QueryBuilder implements the Builder Pattern for constructing SQL queries incrementally
type QueryBuilder struct {
	selectClauses  []string
	fromClause     string
	whereClauses   []string
	orderByClauses []string
	limitValue     int
	offsetValue    int
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		selectClauses:  []string{},
		whereClauses:   []string{},
		orderByClauses: []string{},
		limitValue:     0,
		offsetValue:    0,
	}
}

// Select adds columns to the SELECT clause
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectClauses = append(qb.selectClauses, columns...)
	return qb
}

// From sets the main table for the query
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromClause = table
	return qb
}

// Where adds a WHERE condition (can be called multiple times for AND logic)
func (qb *QueryBuilder) Where(condition string) *QueryBuilder {
	if condition != "" {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	return qb
}

// WhereIf conditionally adds a WHERE clause only if the condition is true
func (qb *QueryBuilder) WhereIf(add bool, condition string) *QueryBuilder {
	if add && condition != "" {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	return qb
}

// OrderBy adds an ORDER BY clause
func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	if direction != "" && (direction == "ASC" || direction == "DESC") {
		qb.orderByClauses = append(qb.orderByClauses, fmt.Sprintf("%s %s", column, direction))
	} else {
		qb.orderByClauses = append(qb.orderByClauses, column)
	}
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitValue = limit
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetValue = offset
	return qb
}

// Build constructs and returns the final SQL query string
func (qb *QueryBuilder) Build() string {
	var query strings.Builder

	// SELECT clause
	if len(qb.selectClauses) == 0 {
		query.WriteString("SELECT *")
	} else {
		query.WriteString("SELECT ")
		query.WriteString(strings.Join(qb.selectClauses, ", "))
	}

	// FROM clause
	if qb.fromClause == "" {
		return "" // Invalid query without FROM
	}
	query.WriteString("\nFROM ")
	query.WriteString(qb.fromClause)

	// WHERE clause
	if len(qb.whereClauses) > 0 {
		query.WriteString("\nWHERE ")
		query.WriteString(strings.Join(qb.whereClauses, "\nAND "))
	}

	// ORDER BY clause
	if len(qb.orderByClauses) > 0 {
		query.WriteString("\nORDER BY ")
		query.WriteString(strings.Join(qb.orderByClauses, ", "))
	}

	// LIMIT clause
	if qb.limitValue > 0 {
		query.WriteString(fmt.Sprintf("\nLIMIT %d", qb.limitValue))
	}

	// OFFSET clause
	if qb.offsetValue > 0 {
		query.WriteString(fmt.Sprintf("\nOFFSET %d", qb.offsetValue))
	}

	return query.String()
}

// Reset clears all builder state for reuse
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.selectClauses = []string{}
	qb.fromClause = ""
	qb.whereClauses = []string{}
	qb.orderByClauses = []string{}
	qb.limitValue = 0
	qb.offsetValue = 0
	return qb
}

// FeedbackQueryBuilder is a specialized builder for feedback queries
type FeedbackQueryBuilder struct {
	*QueryBuilder
	sentiment   string
	productArea string
	region      string
	minPriority int
}

// NewFeedbackQueryBuilder creates a specialized query builder for feedback
func NewFeedbackQueryBuilder() *FeedbackQueryBuilder {
	return &FeedbackQueryBuilder{
		QueryBuilder: NewQueryBuilder(),
		minPriority:  0,
	}
}

// WithSentiment filters by sentiment (positive, negative, neutral)
func (fqb *FeedbackQueryBuilder) WithSentiment(sentiment string) *FeedbackQueryBuilder {
	fqb.sentiment = sentiment
	if sentiment != "" {
		fqb.Where(fmt.Sprintf("sentiment = '%s'", sentiment))
	}
	return fqb
}

// WithProductArea filters by product area
func (fqb *FeedbackQueryBuilder) WithProductArea(area string) *FeedbackQueryBuilder {
	fqb.productArea = area
	if area != "" {
		fqb.Where(fmt.Sprintf("product_area = '%s'", area))
	}
	return fqb
}

// WithRegion filters by region
func (fqb *FeedbackQueryBuilder) WithRegion(region string) *FeedbackQueryBuilder {
	fqb.region = region
	if region != "" {
		fqb.Where(fmt.Sprintf("region = '%s'", region))
	}
	return fqb
}

// WithMinPriority filters by minimum priority
func (fqb *FeedbackQueryBuilder) WithMinPriority(priority int) *FeedbackQueryBuilder {
	fqb.minPriority = priority
	if priority > 0 {
		fqb.Where(fmt.Sprintf("priority >= %d", priority))
	}
	return fqb
}

// BuildFeedback returns the SQL query string with default feedback table
func (fqb *FeedbackQueryBuilder) BuildFeedback() string {
	if fqb.fromClause == "" {
		fqb.From("feedback_enriched")
	}
	return fqb.Build()
}
