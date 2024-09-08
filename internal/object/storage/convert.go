package storage

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/kelindar/ultima-web/internal/object"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// row represents a database row
type row struct {
	ID        string         `gorm:"primaryKey"`
	Namespace string         `gorm:"index"`
	State     string         `gorm:"index"`
	Data      datatypes.JSON `gorm:"size:1048576"`
	IndexedBy string         `gorm:"index"`
	CreatedBy string
	UpdatedBy string
	CreatedAt time.Time `gorm:"autoUpdateTime:nano"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:nano"`
}

// rowOf converts a resource to a database row
func rowOf(v Record) (*row, error) {
	data, err := object.ToJSON(v)
	if err != nil {
		return nil, err
	}

	// If the resource can be indexed, retrieve the indexedBy value
	indexedBy := ""
	if idx, ok := v.(object.Indexer); ok {
		indexedBy = idx.Index()
	}

	// Retrieve created and updated dates
	urn := v.URN()
	createdBy, _ := v.Created()
	updatedBy, _ := v.Updated()
	return &row{
		ID:        urn.ID,
		Namespace: urn.Namespace,
		State:     v.Status(),
		Data:      data,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
		IndexedBy: indexedBy,
	}, nil
}

// Unmarshal converts the database row to a object. It uses the specified registry
// to find the resource type.
func (r *row) Unmarshal(c object.Registry) (Record, error) {
	v, err := object.FromJSON(c, r.Data)
	if err != nil {
		return nil, err
	}

	// Merge the row into the resource
	rv := reflect.ValueOf(v).Elem()
	rv.FieldByName("UpdatedBy").SetString(r.UpdatedBy)
	rv.FieldByName("UpdatedAt").SetInt(r.UpdatedAt.UnixNano())
	rv.FieldByName("CreatedBy").SetString(r.CreatedBy)
	rv.FieldByName("CreatedAt").SetInt(r.CreatedAt.UnixNano())
	return v, nil
}

// ---------------------------------- Rowset ----------------------------------

// unmarshalRecords converts an array of rows to an array of records
func unmarshalRecords[T Record](r []row, c object.Registry) ([]T, error) {
	out := make([]T, 0, len(r))
	for i := range r {
		v, err := r[i].Unmarshal(c)
		if err != nil {
			return nil, err
		}

		x, ok := v.(T)
		if !ok {
			return nil, fmt.Errorf("storage: unable to unmarshal %v to %T", r[i].ID, v)
		}

		out = append(out, x)
	}
	return out, nil
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

// Converts the result error to an error
func try(name string, result *gorm.DB) error {
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return fmt.Errorf("%w (%s)", errNotFound, name)
	case result.Error != nil:
		return result.Error
	default:
		return nil
	}
}

// defaultOf returns the default value for the specified type
func defaultOf[T any]() T {
	var v T
	return v
}
