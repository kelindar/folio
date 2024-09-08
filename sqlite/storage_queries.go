package sqlite

import (
	"fmt"
	"iter"

	"github.com/kelindar/folio"
)

// Insert inserts a new resource into the storage.
func (s *rds) Insert(v Record, createdBy string) (Record, error) {
	urn := v.URN()

	// Convert the resource to a database row
	row, err := rowOf(v)
	if err != nil {
		return nil, err
	}

	row.CreatedBy = createdBy
	row.UpdatedBy = createdBy

	// Insert the row
	table := s.tableOf(urn.Kind)
	if err := try(urn.String(), table.Create(row)); err != nil {
		return nil, err
	}

	// Unmarshal the row back to a resource
	return row.Unmarshal(s.registry)
}

// Update updates an existing resource in the storage.
func (s *rds) Update(v Record, updatedBy string) (Record, error) {
	_, version := v.Updated()
	urn := v.URN()
	table := s.tableOf(urn.Kind)

	// Retrieve the old object we are updating
	oldrec, err := s.Fetch(urn)
	if err != nil {
		return nil, err
	}

	// Convert the resource to a database row
	updated, err := rowOf(v)
	if err != nil {
		return nil, err
	}

	// Update the row using optimistic locking in case the object was updated
	// by another process/user between the time we retrieved it and now.
	result := table.
		Where("id = ?", urn.ID).
		Where("updated_at = ?", version).
		Updates(row{
			State:     updated.State,
			Data:      updated.Data,
			UpdatedBy: updatedBy,
		})

	switch {
	case result.Error != nil:
		return nil, err
	case result.RowsAffected == 0:
		return oldrec, fmt.Errorf("%w (%s)", folio.ErrConflict, oldrec.URN())
	}

	// Find the updated object
	return s.Fetch(urn)
}

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

// Delete deletes a resource from the storage and returns the deleted object.
func (s *rds) Delete(urn folio.URN, deletedBy string) (Record, error) {
	table := s.tableOf(urn.Kind).Unscoped() // permanently delete the object

	// Find the row, unscoped to include soft-deleted records
	var existing row
	if err := try(urn.String(), table.First(&existing, "id = ?", urn.ID)); err != nil {
		return nil, err
	}

	// Unmarshal the document we are about to delete
	deleted, err := existing.Unmarshal(s.registry)
	if err != nil {
		return nil, err
	}

	// Delete the row
	if err := try(urn.String(), table.Delete(&existing, "id = ?", urn.ID)); err != nil {
		return nil, err
	}

	// Return the deleted object
	return deleted, nil
}

// Fetch attempts to find a specific document in the storage layer.
func (s *rds) Fetch(urn folio.URN) (Record, error) {
	var found row
	if err := try(urn.String(), s.tableOf(urn.Kind).First(&found, "id = ?", urn.ID)); err != nil {
		return nil, err
	}

	// Unmarshal the row back to a object
	return found.Unmarshal(s.registry)
}

// Range performs a query against the storage layer and calls the specified
// function for each retrieved object.
func (s *rds) Range(kind folio.Kind, q folio.Query) (iter.Seq[Record], error) {
	query, err := s.query(kind, q)
	if err != nil {
		return nil, err
	}

	var result []row
	if err := try(kind.String(), query.Find(&result)); err != nil {
		return nil, err
	}

	var innerErr error
	return func(yield func(Record) bool) {
		for i := range result {
			record, err := result[i].Unmarshal(s.registry)
			if err != nil {
				innerErr = err
				return
			}

			if next := yield(record); !next {
				return
			}
		}
	}, innerErr
}

func errIter(err error) iter.Seq2[Record, error] {
	return func(yield func(Record, error) bool) {
		yield(nil, err)
	}
}
