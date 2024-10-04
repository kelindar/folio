package folio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"reflect"

	"github.com/kelindar/folio/convert"
)

var (
	typeResource = reflect.TypeOf(Meta{})
	typeEmbedded = reflect.TypeOf(Embed{})
)

var (
	ErrNotFound = errors.New("storage: document was not found")
	ErrConflict = errors.New("storage: update of an outdated document")
)

// IsNotFound returns true if the specified error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflict returns true if the specified error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// ---------------------------------- Contract ----------------------------------

// Storage represents a storage layer for records.
type Storage interface {
	io.Closer
	Insert(v Object, createdBy string) (Object, error)
	Update(v Object, updatedBy string) (Object, error)
	Upsert(v Object, updatedBy string) (Object, error)
	Delete(urn URN, deletedBy string) (Object, error)
	Fetch(urn URN) (Object, error)
	Search(kind Kind, query Query) (iter.Seq[Object], error)
	Count(kind Kind, query Query) (int, error)
}

// Query represents a query to filter records.
type Query struct {
	Namespaces []string            // Namespaces is a list of namespaces to filter by
	States     []string            // States is a list of states to filter by
	Indexes    []string            // Indexes is a list of indexes to filter by
	Filters    map[string][]string // Filters is a map of filters to apply
	Match      string              // Match is the full-text search query
	SortBy     []string            `default:"[\"+id\"]"` // Sort is the set of fields to order by
	Offset     int                 `default:"0"`         // Offset is the number of records to skip
	Limit      int                 `default:"1000"`      // Limit is the maximum number of records to return
}

// ---------------------------------- Indexer ----------------------------------

// Indexer represents a resource that provides an index.
type Indexer interface {
	Index() string
}

// Embedded represents a generic embedded document for unmarshaling
type Embed struct {
	Value    Object `json:",inline"`
	Registry Registry
}

// UnmarshalJSON unmarshals the JSON into the embedded document
func (r *Embed) UnmarshalJSON(b []byte) error {
	if len(b) <= 4 { // "null" or empty
		return nil
	}

	o, err := FromJSON(r.Registry, b)
	if err != nil {
		return err
	}

	v, ok := o.(Object)
	if !ok {
		return fmt.Errorf("resource: unable to unmarshal, invalid type %T", o)
	}

	r.Value = v
	return nil
}

// MarshalJSON marshals the JSON from the embedded document
func (r Embed) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Value)
}

// walkEmbeds walks the struct and calls the specified function for each embedded resource
func walkEmbeds(r Registry, value any, fn func(reflect.Value)) error {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Invalid || v.IsZero() {
		return nil
	}

	switch v.Type().Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		v = reflect.Indirect(v)
		if v.Type().Kind() != reflect.Struct {
			return nil
		}
	default:
		return nil
	}

	rt := v.Type()
	for i := 0; i < v.NumField(); i++ {
		rv := v.Field(i)
		if rt.Field(i).Type == typeEmbedded {
			fn(rv)
			continue
		}

		// skip unexported fields
		if !rv.CanInterface() {
			continue
		}

		// Recursively walk the struct
		if err := walkEmbeds(r, rv.Interface(), fn); err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------------- Options & Metadata ----------------------------------

// Options represents the options for a document
type Options struct {
	Icon   string `json:"icon,omitempty"`   // Icon name from https://fonts.google.com/icons
	Title  string `json:"title,omitempty"`  // Title of the document (e.g. Person)
	Plural string `json:"plural,omitempty"` // Plural name of the document (e.g. People)
	Sort   string `json:"sort,omitempty"`   // Sort field
}

// defaultOptions returns the default options for the specified kind
func defaultOptions(kind Kind) Options {
	return Options{
		Icon:   "text_snippet",
		Title:  convert.TitleCase(string(kind)),
		Plural: convert.TitleCase(string(kind)),
		Sort:   string(kind),
	}
}
