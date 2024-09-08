package sqlite

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kelindar/folio"
)

// withMeta adds metadata to the record
func withMeta(v Record, createdBy, updatedBy string, createdAt, updatedAt time.Time) Record {
	rv := reflect.ValueOf(v).Elem()
	rv.FieldByName("CreatedBy").SetString(createdBy)
	rv.FieldByName("CreatedAt").SetInt(createdAt.UnixNano())
	rv.FieldByName("UpdatedBy").SetString(updatedBy)
	rv.FieldByName("UpdatedAt").SetInt(updatedAt.UnixNano())
	return v
}

func read(scan func(...any) error, r folio.Registry) (Record, error) {
	var record struct {
		ID        string
		Namespace string
		State     string
		Data      []byte
		CreatedBy string
		UpdatedBy string
		CreatedAt int64
		UpdatedAt int64
	}

	if err := scan(&record.Data,
		&record.CreatedBy, &record.UpdatedBy,
		&record.CreatedAt, &record.UpdatedAt,
	); err != nil {
		return nil, err
	}

	obj, err := folio.FromJSON(r, record.Data)
	if err != nil {
		return nil, err
	}

	return withMeta(obj,
		record.CreatedBy,
		record.UpdatedBy,
		time.Unix(0, record.CreatedAt),
		time.Unix(0, record.UpdatedAt),
	), nil
}

// ---------------------------------- Text ----------------------------------

var (
	matchFirst = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchEvery = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func snakeCase(str string) string {
	str = matchFirst.ReplaceAllString(str, "${1}_${2}")
	str = matchEvery.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(str)
}

// tableOf returns the table name for the specified resource kind
func tableOf(kind folio.Kind) string {
	return kind.String()
}

// defaultOf returns the default value for the specified type
func defaultOf[T any]() T {
	var v T
	return v
}
