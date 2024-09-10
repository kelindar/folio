package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/kelindar/folio"
)

// Upsert inserts or updates a resource in the storage.
func (s *rds) Upsert(v Record, updatedBy string) (Record, error) {
	_, err := s.Fetch(v.URN())
	switch {
	case folio.IsNotFound(err):
		return s.Insert(v, updatedBy)
	case err != nil:
		return nil, err
	default:
		return s.Update(v, updatedBy)
	}
}

// Insert inserts a new resource into the storage.
func (s *rds) Insert(v Record, createdBy string) (Record, error) {
	data, err := folio.ToJSON(v)
	if err != nil {
		return nil, err
	}

	// Prepare the statement
	urn := v.URN()
	now := time.Now()
	sql := `INSERT INTO ` + tableOf(urn.Kind) +
		` (id, namespace, state, indexed_by, data, created_by, updated_by, created_at, updated_at)` +
		` VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Insert the record
	if _, err := s.db.Exec(sql,
		urn.ID,
		urn.Namespace,
		v.Status(),
		indexOf(v),
		data,
		createdBy,
		createdBy, // same as created_by
		now.UnixNano(),
		now.UnixNano(), // same as created_at
	); err != nil {
		return nil, fmt.Errorf("storage: unable to insert, %w", err)
	}

	return withMeta(v, createdBy, createdBy, now, now), nil
}

// Update updates an existing resource in the storage.
func (s *rds) Update(v Record, updatedBy string) (Record, error) {
	data, err := folio.ToJSON(v)
	if err != nil {
		return nil, err
	}

	_, version := v.Updated()
	urn := v.URN()
	now := time.Now()
	sql := `UPDATE ` + tableOf(urn.Kind) +
		` SET state = ?, indexed_by = ?, data = ?, updated_by = ?, updated_at = ?` +
		` WHERE id = ? AND updated_at = ?`

	// Update the record
	r, err := s.db.Exec(sql, v.Status(), indexOf(v), data, updatedBy, now.UnixNano(), urn.ID, version.UnixNano())
	n, _ := r.RowsAffected()
	switch {
	case err != nil:
		return nil, fmt.Errorf("storage: unable to update, %w", err)
	case n == 0:
		return nil, fmt.Errorf("%w (%v)", folio.ErrConflict, urn.String())
	default:
		return s.Fetch(urn)
	}
}

// Fetch retrieves a resource by URN.
func (s *rds) Fetch(urn folio.URN) (Record, error) {
	selectSQL := `SELECT  data, created_by, updated_by, created_at, updated_at` +
		` FROM ` + tableOf(urn.Kind) + ` WHERE id = ?`

	row := s.db.QueryRow(selectSQL, urn.ID)
	obj, err := read(row.Scan, s.registry)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, fmt.Errorf("%w (%v)", folio.ErrNotFound, urn.String())
	case err != nil:
		return nil, fmt.Errorf("storage: unable to fetch, %w", err)
	default:
		return obj, nil
	}
}

// Delete deletes a resource from the storage.
func (s *rds) Delete(urn folio.URN, deletedBy string) (Record, error) {
	deleted, err := s.Fetch(urn)
	if err != nil {
		return nil, err
	}

	if _, err := s.db.Exec(`DELETE FROM `+tableOf(urn.Kind)+` WHERE id = ?`, urn.ID); err != nil {
		return nil, fmt.Errorf("failed to delete record: %w", err)
	}

	return deleted, nil
}

// Search performs a query against the storage layer and calls the specified
// function for each retrieved object.
func (s *rds) Search(kind folio.Kind, q folio.Query) (iter.Seq[Record], error) {
	rows, err := s.query(kind, q)
	if err != nil {
		return nil, err
	}

	var innerErr error
	return func(yield func(Record) bool) {
		defer rows.Close()
		for rows.Next() {
			obj, err := read(rows.Scan, s.registry)
			if err != nil {
				innerErr = fmt.Errorf("storage: unable to read, %w", err)
				return
			}

			if next := yield(obj); !next {
				return
			}
		}

	}, innerErr
}
