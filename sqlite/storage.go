package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/creasty/defaults"
	"github.com/kelindar/folio"
	_ "github.com/ncruces/go-sqlite3/driver" // cgo-free, uses wazero
	_ "github.com/ncruces/go-sqlite3/embed"  // cgo-free, uses wazero
)

type Record = folio.Object

// rds represents a relational storage layer for resources.
type rds struct {
	db       *sql.DB
	registry folio.Registry
}

// Open opens a storage database
func Open(dsn string, registry folio.Registry) (folio.Storage, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: unable to open database: %w", err)
	}

	// Set max open connections for SQLite (as SQLite is not designed for many concurrent writes)
	db.SetMaxOpenConns(1)

	// Auto-create the tables
	if err := autoMigrate(db, registry); err != nil {
		return nil, err
	}

	return &rds{
		db:       db,
		registry: registry,
	}, nil
}

// OpenEphemeral opens an ephemeral storage
func OpenEphemeral(registry folio.Registry) folio.Storage {
	s, err := Open(":memory:", registry)
	if err != nil {
		panic(err)
	}
	return s
}

// Close closes the storage gracefully.
func (s *rds) Close() error {
	return s.db.Close()
}

// ---------------------------------- Query ----------------------------------

// query creates a query for the specified resource kind
func (s *rds) query(projection string, kind folio.Kind, q folio.Query) (*sql.Rows, error) {
	if err := defaults.Set(&q); err != nil {
		return nil, err
	}

	// Validate the query
	switch {
	case kind == "":
		return nil, fmt.Errorf("storage: kind is required")
	case q.Limit < 0:
		return nil, fmt.Errorf("storage: invalid limit %d", q.Limit)
	case q.Offset < 0:
		return nil, fmt.Errorf("storage: invalid offset %d", q.Offset)
	}

	// Build the query
	where := make([]string, 0, 4)
	args := make([]any, 0, 4)

	// Filter by namespaces
	if len(q.Namespace) > 0 {
		where = append(where, "namespace IN (?)")
		args = append(args, q.Namespace)
	}

	// Filter by states
	if len(q.States) > 0 {
		where = append(where, "state IN (?)")
		args = append(args, strings.Join(q.States, ","))
	}

	// Filter by filters (JSON Path)
	for path, filter := range q.Filters {
		if path != "" {
			where = append(where, queryFilterByJSON(path, filter))
		}
	}

	// If full-text search is requested, add it to the query using the corresponding _fts table
	if q.Match != "" {
		where = append(where, fmt.Sprintf(`id IN (SELECT id FROM %s_fts WHERE data match ?)`, tableOf(kind)))
		args = append(args, sanitizeTerm(q.Match))
	}

	// Build the SQL query string
	querySQL := fmt.Sprintf("%s FROM %s", projection, tableOf(kind))
	if len(where) > 0 {
		querySQL += ` WHERE ` + strings.Join(where, " AND ")
	}

	// Add sorting
	if len(q.SortBy) > 0 {
		sortFields := make([]string, 0, len(q.SortBy))
		for _, field := range q.SortBy {
			sortFields = append(sortFields, queryOrder(field))
		}
		querySQL += fmt.Sprintf(" ORDER BY %s", strings.Join(sortFields, ", "))
	}

	if q.Limit > 0 {
		querySQL += fmt.Sprintf(" LIMIT %d", q.Limit)
	}

	if q.Offset > 0 {
		querySQL += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	return s.db.Query(querySQL, args...)
}

// queryOrder converts the sort field to a query string
func queryOrder(field string) string {
	if len(field) == 0 {
		return ""
	}

	var suffix string
	switch field[0] {
	case '-':
		suffix = " DESC"
		field = field[1:]
	case '+':
		field = field[1:]
	}

	switch field {
	case "id",
		"createdBy", "createdAt",
		"updatedBy", "updatedAt",
		"updated_by", "updated_at",
		"created_by", "created_at":
		return snakeCase(field) + suffix
	default:
		return "json_extract(data, '$." + field + "')" + suffix
	}
}

// queryFilterByJSON returns a filter for the specified json path and values in SQLite
func queryFilterByJSON(path string, values []string) string {
	if len(path) == 0 || len(values) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("(json_extract(data, '$.")
	sb.WriteString(path)
	sb.WriteString("') IN (")

	// Write the values as a SQL 'IN' list
	for i, v := range values {
		sb.WriteString("'")
		sb.WriteString(v)
		sb.WriteString("'")
		if i < len(values)-1 {
			sb.WriteString(",")
		}
	}

	sb.WriteString("))")
	return sb.String()
}

// sanitizeTerm tokenizes and prepares the query for FTS5 with NEAR
func sanitizeTerm(query string) string {
	tokens := strings.Fields(query)
	for i, token := range tokens {
		tokens[i] = token + "*" // Add wildcard for partial matching
	}

	// Wrap with NEAR to match any of the tokens 30 words apart
	nearClause := "NEAR(" + strings.Join(tokens, " ") + ", 30)"
	return nearClause
}
