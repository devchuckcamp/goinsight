package repository

import (
	"fmt"
	"strings"
)

// QueryBuilder provides a fluent interface for building SQL queries
// Implements the Builder design pattern for query construction
type QueryBuilder struct {
	selectClause   string
	fromClause     string
	whereConditions []string
	orderByClause  string
	limitClause    string
	offsetClause   string
	params         []interface{}
	paramCount     int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		whereConditions: make([]string, 0),
		params:          make([]interface{}, 0),
		paramCount:      0,
	}
}

// Select sets the SELECT clause
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectClause = "SELECT " + strings.Join(columns, ", ")
	return qb
}

// SelectAll selects all columns
func (qb *QueryBuilder) SelectAll() *QueryBuilder {
	qb.selectClause = "SELECT *"
	return qb
}

// From sets the FROM clause
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromClause = "FROM " + table
	return qb
}

// Where adds a WHERE condition
// Use $1, $2, etc. as placeholders for parameters
func (qb *QueryBuilder) Where(condition string, params ...interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, condition)
	qb.params = append(qb.params, params...)
	qb.paramCount += len(params)
	return qb
}

// AndWhere adds an AND condition
func (qb *QueryBuilder) AndWhere(condition string, params ...interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, "AND "+condition)
	qb.params = append(qb.params, params...)
	qb.paramCount += len(params)
	return qb
}

// OrWhere adds an OR condition
func (qb *QueryBuilder) OrWhere(condition string, params ...interface{}) *QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, "OR "+condition)
	qb.params = append(qb.params, params...)
	qb.paramCount += len(params)
	return qb
}

// OrderBy sets the ORDER BY clause
func (qb *QueryBuilder) OrderBy(columns ...string) *QueryBuilder {
	qb.orderByClause = "ORDER BY " + strings.Join(columns, ", ")
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitClause = fmt.Sprintf("LIMIT %d", limit)
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetClause = fmt.Sprintf("OFFSET %d", offset)
	return qb
}

// Build constructs the final SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause (required)
	if qb.selectClause == "" {
		qb.SelectAll()
	}
	query.WriteString(qb.selectClause)

	// FROM clause (required)
	if qb.fromClause != "" {
		query.WriteString(" ")
		query.WriteString(qb.fromClause)
	}

	// WHERE clause
	if len(qb.whereConditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(qb.whereConditions, " "))
	}

	// ORDER BY clause
	if qb.orderByClause != "" {
		query.WriteString(" ")
		query.WriteString(qb.orderByClause)
	}

	// LIMIT clause
	if qb.limitClause != "" {
		query.WriteString(" ")
		query.WriteString(qb.limitClause)
	}

	// OFFSET clause
	if qb.offsetClause != "" {
		query.WriteString(" ")
		query.WriteString(qb.offsetClause)
	}

	return query.String(), qb.params
}

// String returns the SQL query as a string (without parameters)
func (qb *QueryBuilder) String() string {
	query, _ := qb.Build()
	return query
}

// Reset clears the builder for reuse
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.selectClause = ""
	qb.fromClause = ""
	qb.whereConditions = qb.whereConditions[:0]
	qb.orderByClause = ""
	qb.limitClause = ""
	qb.offsetClause = ""
	qb.params = qb.params[:0]
	qb.paramCount = 0
	return qb
}

// Example usage:
// builder := NewQueryBuilder()
// query, params := builder.
//     Select("id", "name", "email").
//     From("users").
//     Where("age > $1", 18).
//     AndWhere("status = $2", "active").
//     OrderBy("name ASC").
//     Limit(10).
//     Build()
// results := repo.QueryFeedback(ctx, query)
