package database

import (
	"fmt"
	"strings"
)

// JoinType represents the type of SQL join
type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	FullJoin  JoinType = "FULL OUTER JOIN"
)

// Join represents a single join operation
type Join struct {
	Type      JoinType
	Table     string
	Condition string
}

// QueryBuilder helps build complex queries with joins
type QueryBuilder struct {
	baseTable string
	columns   []string
	joins     []Join
	where     string
	whereArgs []interface{}
	orderBy   string
	groupBy   string
	having    string
	limit     string
	offset    string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(baseTable string) *QueryBuilder {
	return &QueryBuilder{
		baseTable: baseTable,
		columns:   []string{"*"},
		joins:     make([]Join, 0),
		whereArgs: make([]interface{}, 0),
	}
}

// Select sets the columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.columns = columns
	return qb
}

// Join adds a join to the query
func (qb *QueryBuilder) Join(joinType JoinType, table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, Join{
		Type:      joinType,
		Table:     table,
		Condition: condition,
	})
	return qb
}

// InnerJoin adds an INNER JOIN
func (qb *QueryBuilder) InnerJoin(table, condition string) *QueryBuilder {
	return qb.Join(InnerJoin, table, condition)
}

// LeftJoin adds a LEFT JOIN
func (qb *QueryBuilder) LeftJoin(table, condition string) *QueryBuilder {
	return qb.Join(LeftJoin, table, condition)
}

// RightJoin adds a RIGHT JOIN
func (qb *QueryBuilder) RightJoin(table, condition string) *QueryBuilder {
	return qb.Join(RightJoin, table, condition)
}

// FullJoin adds a FULL OUTER JOIN
func (qb *QueryBuilder) FullJoin(table, condition string) *QueryBuilder {
	return qb.Join(FullJoin, table, condition)
}

// Where sets the WHERE clause
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.where = condition
	qb.whereArgs = args
	return qb
}

// OrderBy sets the ORDER BY clause
func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// GroupBy sets the GROUP BY clause
func (qb *QueryBuilder) GroupBy(groupBy string) *QueryBuilder {
	qb.groupBy = groupBy
	return qb
}

// Having sets the HAVING clause
func (qb *QueryBuilder) Having(having string) *QueryBuilder {
	qb.having = having
	return qb
}

// Limit sets the LIMIT clause
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = fmt.Sprintf("%d", limit)
	return qb
}

// Offset sets the OFFSET clause
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = fmt.Sprintf("%d", offset)
	return qb
}

// Build constructs the final SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var query strings.Builder

	// SELECT clause
	query.WriteString("SELECT ")
	query.WriteString(strings.Join(qb.columns, ", "))

	// FROM clause
	query.WriteString(" FROM ")
	query.WriteString(qb.baseTable)

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(string(join.Type))
		query.WriteString(" ")
		query.WriteString(join.Table)
		query.WriteString(" ON ")
		query.WriteString(join.Condition)
	}

	// WHERE clause
	if qb.where != "" {
		query.WriteString(" WHERE ")
		query.WriteString(qb.where)
	}

	// GROUP BY clause
	if qb.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(qb.groupBy)
	}

	// HAVING clause
	if qb.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(qb.having)
	}

	// ORDER BY clause
	if qb.orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(qb.orderBy)
	}

	// LIMIT clause
	if qb.limit != "" {
		query.WriteString(" LIMIT ")
		query.WriteString(qb.limit)
	}

	// OFFSET clause
	if qb.offset != "" {
		query.WriteString(" OFFSET ")
		query.WriteString(qb.offset)
	}

	return query.String(), qb.whereArgs
}

// BuildCountQuery builds a COUNT query with the same joins and conditions
func (qb *QueryBuilder) BuildCountQuery() (string, []interface{}) {
	var query strings.Builder

	// SELECT COUNT(*)
	query.WriteString("SELECT COUNT(*)")

	// FROM clause
	query.WriteString(" FROM ")
	query.WriteString(qb.baseTable)

	// JOIN clauses
	for _, join := range qb.joins {
		query.WriteString(" ")
		query.WriteString(string(join.Type))
		query.WriteString(" ")
		query.WriteString(join.Table)
		query.WriteString(" ON ")
		query.WriteString(join.Condition)
	}

	// WHERE clause
	if qb.where != "" {
		query.WriteString(" WHERE ")
		query.WriteString(qb.where)
	}

	// GROUP BY clause (for count queries)
	if qb.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(qb.groupBy)
	}

	// HAVING clause
	if qb.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(qb.having)
	}

	return query.String(), qb.whereArgs
}
