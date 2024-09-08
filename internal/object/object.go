package object

import (
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"
)

// Object represents an object in the system.
type Object interface {
	URN() URN                     // URN returns the uniform identifier of the object
	Status() string               // Status returns the current state
	Created() (string, time.Time) // Created returns createdBy and createdAt information
	Updated() (string, time.Time) // Updated returns updatedBy and updatedAt information
}

// Kind represents a resource Kind (e.g. "Document", "Sprite")
type Kind string

// String returns the string representation of the resource kind.
func (k Kind) String() string {
	return strings.ToLower(string(k))
}

// Meta represents a metadata of the object.
type Meta struct {
	ID        string   `json:"id"`                  // Globally unique identifier (e.g. "9m4e2mr0ui3e8a215n4g")
	Kind      Kind     `json:"kind"`                // Meta kind (e.g. "deployment")
	Namespace string   `json:"namespace"`           // Namespace of the object (e.g. "my_project")
	Desc      string   `json:"desc,omitempty"`      // Desc is a short description of the resource (e.g. "My deployment")
	Tags      []string `json:"tags,omitempty"`      // Tags are used to group resources
	State     string   `json:"state,omitempty"`     // State is the current state of the resource
	CreatedBy string   `json:"createdBy,omitempty"` // CreatedBy is the user who created the resource
	CreatedAt int64    `json:"createdAt,omitempty"` // CreatedAt is the time when the resource was created
	UpdatedBy string   `json:"updatedBy,omitempty"` // UpdatedBy is the user who last updated the resource
	UpdatedAt int64    `json:"updatedAt,omitempty"` // UpdatedAt is the time when the resource was last updated
}

// New creates a new instance of the specified resource kind.
func New[T Object](namespace string) (T, error) {
	typ := typeOfT[T]()
	kind, err := KindOf(typ)
	if err != nil {
		return *new(T), err
	}

	// Create a new URN (and validate it)
	urn, err := NewURN(namespace, kind)
	if err != nil {
		return *new(T), err
	}

	// Create a new instance
	instance := reflect.New(typ)
	resource := instance.Elem().FieldByName("Meta").Addr().Interface().(*Meta)
	resource.Namespace = urn.Namespace
	resource.Kind = urn.Kind
	resource.ID = urn.ID
	return instance.Interface().(T), nil
}

// NewWith creates a new instance of the specified resource kind and calls the
// specified function to initialize it.
func NewWith[T Object](namespace string, fn func(obj T) error) (T, error) {
	r, err := New[T](namespace)
	if err != nil {
		return *new(T), err
	}

	return r, fn(r)
}

// URN returns the URN of the object.
func (r *Meta) URN() URN {
	return URN{
		Namespace: r.Namespace,
		Kind:      r.Kind,
		ID:        r.ID,
	}
}

// Created returns who created the resource and when.
func (r *Meta) Created() (string, time.Time) {
	return r.CreatedBy, time.Unix(0, r.CreatedAt)
}

// Updated returns who updated the resource and when.
func (r *Meta) Updated() (string, time.Time) {
	return r.UpdatedBy, time.Unix(0, r.UpdatedAt)
}

// Status returns the current state of the resource.
func (r *Meta) Status() string {
	return r.State
}

// Metadata returns the metadata.
func (r *Meta) Metadata() *Meta {
	return r
}

// ---------------------------------- URN ----------------------------------

func FromKind(c Registry, kind Kind) (Object, error) {
	typ, err := c.Resolve(kind)
	if err != nil {
		return nil, err
	}

	// Create a new instance and prepare it for unmarshalling
	instance := reflect.New(typ).Interface().(Object)
	if err := walkEmbeds(c, instance, func(rv reflect.Value) {
		rv.Set(reflect.ValueOf(Embed{Registry: c}))
	}); err != nil {
		return nil, err
	}

	return instance, nil
}

// ---------------------------------- JSON ----------------------------------

// ToJSON encodes a resource to JSON
func ToJSON(v Object) ([]byte, error) {
	return json.Marshal(v)
}

// FromJSON parses a JSON file and returns a resource
func FromJSON(c Registry, data []byte) (Object, error) {
	var res Meta
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	// Create a new instance
	instance, err := FromKind(c, res.Kind)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into the instance
	if err := json.Unmarshal(data, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

// ReadJSON reads a JSON file and returns a resource
func ReadJSON(c Registry, reader io.Reader) (Object, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return FromJSON(c, data)
}
