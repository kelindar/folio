package walk

import (
	"reflect"
	"strings"
)

// IsEmpty checks whether a value is empty or not.
func IsEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !IsEmpty(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.String:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// NewIfNil creates a new instance of a value if it is nil.
func NewIfNil(v reflect.Value) reflect.Value {
	switch {
	case v.Kind() == reflect.Pointer && v.IsNil():
		v.Set(reflect.New(v.Type().Elem()))
	case v.Kind() == reflect.Slice && v.IsNil():
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	case v.Kind() == reflect.Map && v.IsNil():
		v.Set(reflect.MakeMap(v.Type()))
	}
	return v
}

// nameOf returns the JSON name of a field
func nameOf(field *reflect.StructField) string {
	name := field.Tag.Get("json")
	if name == "" || name == "-" {
		return field.Name
	}

	// Find the comma to separate name from options
	if commaIdx := strings.Index(name, ","); commaIdx != -1 {
		if commaIdx == 0 {
			return field.Name
		}

		name = name[:commaIdx]
	}

	if name == "-" {
		return field.Name
	}
	return name
}
