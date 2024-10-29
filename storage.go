package folio

import "iter"

// ---------------------------------- Generic ----------------------------------

// Create creates a new resource and inserts it into the storage.
func Create[T Object](db Storage, constructor func(obj T) error, namespace, createdBy string) (T, error) {
	instance, err := New(namespace, constructor)
	if err != nil {
		return defaultOf[T](), err
	}

	return Insert(db, instance, createdBy)
}

// Insert inserts a new resource into the storage.
func Insert[T Object](db Storage, v T, createdBy string) (T, error) {
	out, err := db.Insert(v, createdBy)
	if err != nil {
		return defaultOf[T](), err
	}

	return out.(T), nil
}

// Update updates an existing resource in the storage.
func Update[T Object](db Storage, v T, updatedBy string) (T, error) {
	out, err := db.Update(v, updatedBy)
	if err != nil {
		return defaultOf[T](), err
	}

	return out.(T), nil
}

// Upsert inserts or updates a resource in the storage.
func Upsert[T Object](db Storage, v T, updatedBy string) (T, error) {
	out, err := db.Upsert(v, updatedBy)
	if err != nil {
		return defaultOf[T](), err
	}

	return out.(T), nil
}

// Delete deletes a resource from the storage and returns the deleted object.
func Delete[T Object](db Storage, urn URN, deletedBy string) (T, error) {
	out, err := db.Delete(urn, deletedBy)
	if err != nil {
		return defaultOf[T](), err
	}

	return out.(T), nil
}

// Fetch attempts to find a specific document in the storage layer.
func Fetch[T Object](db Storage, urn URN) (T, error) {
	v, err := db.Fetch(urn)
	if err != nil {
		return defaultOf[T](), err
	}

	return v.(T), nil
}

// Search performs a query against the storage layer.
func Search[T Object](db Storage, q Query) (iter.Seq[T], error) {
	kind, err := KindOfT[T]()
	if err != nil {
		return nil, err
	}

	cursor, err := db.Search(kind, q)
	if err != nil {
		return nil, err
	}

	return func(yield func(T) bool) {
		for v := range cursor {
			if next := yield(v.(T)); !next {
				return
			}
		}
	}, nil
}

// Count returns the number of records that match the specified query.
func Count[T Object](db Storage, q Query) (int, error) {
	kind, err := KindOfT[T]()
	if err != nil {
		return 0, err
	}

	return db.Count(kind, q)
}

// defaultOf returns the default value for the specified type
func defaultOf[T any]() T {
	var v T
	return v
}
