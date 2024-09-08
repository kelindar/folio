package storage

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/creasty/defaults"
	"github.com/kelindar/ultima-web/internal/object"
	_ "github.com/ncruces/go-sqlite3/embed" // cgo-free, uses wazero
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	errNotFound = errors.New("storage: document was not found")
	errConflict = errors.New("storage: update of an outdated document")
)

// IsNotFound returns true if the specified error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// IsConflict returns true if the specified error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, errConflict)
}

// ---------------------------------- Contract ----------------------------------

type Record = object.Object

// Storage represents a storage layer for records.
type Storage interface {
	io.Closer
	Insert(v Record, createdBy string) (Record, error)
	Update(v Record, updatedBy string) (Record, error)
	Upsert(v Record, updatedBy string) (Record, error)
	Delete(urn object.URN, deletedBy string) (Record, error)
	Fetch(urn object.URN) (Record, error)
	Range(kind object.Kind, query Query) (iter.Seq[Record], error)
}

// ---------------------------------- Open ----------------------------------

// rds represents a relational storage layer for resources.
type rds struct {
	db       *gorm.DB
	registry object.Registry
}

// Open opens a storage database
func Open(dsn string, registry object.Registry) (Storage, error) {
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
	if err := registry.Range(func(k object.Kind, _ reflect.Type) error {
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
func OpenEphemeral(registry object.Registry) Storage {
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
func (s *rds) tableOf(kind object.Kind) *gorm.DB {
	return s.db.Table(kind.String())
}

// ---------------------------------- Query ----------------------------------

type Query struct {
	Namespaces []string            // Namespaces is a list of namespaces to filter by
	States     []string            // States is a list of states to filter by
	Indexes    []string            // Indexes is a list of indexes to filter by
	Filters    map[string][]string // Filters is a map of filters to apply
	SortBy     []string            `default:"[\"+id\"]"` // Sort is the set of fields to order by
	Offset     int                 `default:"0"`         // Offset is the number of records to skip
	Limit      int                 `default:"1000"`      // Limit is the maximum number of records to return
}

// query creates a query for the specified resource kind
func (s *rds) query(kind object.Kind, q Query) (*gorm.DB, error) {
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
