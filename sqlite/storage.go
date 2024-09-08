package sqlite

import (
	"database/sql"
	"fmt"
	"reflect"
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
	if err := registry.Range(func(k folio.Kind, _ reflect.Type) error {
		_, err := db.Exec(createTableSQL(k.String()))
		return err
	}); err != nil {
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

// createTableSQL generates SQL to create a table for a specific resource kind.
func createTableSQL(tableName string) string {
	return fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		namespace TEXT,
		state TEXT,
		data JSON,
		created_by TEXT,
		updated_by TEXT,
		created_at INTEGER,
		updated_at INTEGER
	);`, tableName)
}

// ---------------------------------- Query ----------------------------------

// query creates a query for the specified resource kind
func (s *rds) query(kind folio.Kind, q folio.Query) (*sql.Rows, error) {
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
	if len(q.Namespaces) > 0 {
		where = append(where, "namespace IN (?)")
		args = append(args, strings.Join(q.Namespaces, ","))
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

	// Build the SQL query string
	querySQL := `SELECT data, created_by, updated_by, created_at, updated_at FROM ` + tableOf(kind)
	if len(where) > 0 {
		querySQL += ` WHERE ` + strings.Join(where, " AND ")
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
