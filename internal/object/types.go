package object

import (
	"encoding/json"
	"fmt"
	"reflect"
)

var (
	typeResource = reflect.TypeOf(Meta{})
	typeEmbedded = reflect.TypeOf(Embed{})
)

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
