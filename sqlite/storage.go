package sqlite

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/kelindar/folio"
	_ "github.com/ncruces/go-sqlite3/embed" // cgo-free, uses wazero
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Record = folio.Object

// ---------------------------------- Open ----------------------------------

// rds represents a relational storage layer for resources.
type rds struct {
	db       *gorm.DB
	registry folio.Registry
}

// Open opens a storage database
func Open(dsn string, registry folio.Registry) (folio.Storage, error) {
	dblog := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             500 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open the database
	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{
		Logger: dblog,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: unable to open database: %w", err)
	}

	// Set the max open connections to 1 for sqlite so we can test concurrency
	if db.Name() == "sqlite" {
		db, _ := db.DB()
		db.SetMaxOpenConns(1)
	}

	// Auto-create the tables
	if err := registry.Range(func(k folio.Kind, _ reflect.Type) error {
		table := &row{}
		scope := db.Table(k.String()).Session(&gorm.Session{})

		// Fix the table name on the schema and auto-migrate
		scope.Statement.Parse(table)
		scope.Statement.Schema.Table = k.String()
		if err := scope.AutoMigrate(table); err != nil {
			return fmt.Errorf("storage: unable to auto-migrate table %s: %w", k, err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &rds{
		db:       db,
		registry: registry,
	}, nil
}

// OpenEphemeral opens a new emphemeral storage
func OpenEphemeral(registry folio.Registry) folio.Storage {
	s, err := Open(":memory:", registry)
	if err != nil {
		panic(err)
	}
	return s
}

// Close closes the storage gracefully.
func (s *rds) Close() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}

	return db.Close()
}

// tableOf returns the table for the specified resource kind
func (s *rds) tableOf(kind folio.Kind) *gorm.DB {
	return s.db.Table(kind.String())
}

// ---------------------------------- Query ----------------------------------

// query creates a query for the specified resource kind
func (s *rds) query(kind folio.Kind, q folio.Query) (*gorm.DB, error) {
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

	// Filter by organization & project
	query := s.tableOf(kind).Unscoped()
	if len(q.Namespaces) > 0 {
		query = query.Where("namespace IN (?)", q.Namespaces)
	}

	// Filter by state
	if len(q.States) > 0 {
		query = query.Where("state IN (?)", q.States)
	}

	// Filter by index
	if len(q.Indexes) > 0 {
		query = query.Where("indexed_by IN (?)", q.Indexes)
	}

	// Filter by filters (JSON Path)
	for path, filter := range q.Filters {
		switch path {
		case "":
		case "id",
			"createdAt", "updatedAt",
			"createdBy", "updatedBy":
			query = query.Where(snakeCase(path)+" in (?)", filter)
		default:
			query = query.Where(queryFilterByJSON(path, filter))
		}
	}

	// Order by
	for _, field := range q.SortBy {
		if len(field) > 0 {
			query = query.Order(queryOrder(field))
		}
	}

	// Skip
	if q.Offset > 0 {
		query = query.Offset(q.Offset)
	}

	// Take
	if q.Limit > 0 {
		query = query.Limit(q.Limit)
	}

	return query, nil
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
