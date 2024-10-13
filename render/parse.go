package render

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/kelindar/folio"
	"github.com/tidwall/gjson"
)

// ---------------------------------- Struct Fields ----------------------------------

func isEnum(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("validate"), "oneof=")
}

func isEmail(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("validate"), "email")
}

func isRequired(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("validate"), "required")
}

// Parses oneof tag from validator e.g.: "required,oneof=male female prefer_not_to"
func decodeOneOf(field reflect.StructField) ([]string, bool) {
	if !isEnum(field) {
		return nil, false
	}

	fields := strings.Split(field.Tag.Get("validate"), ",")
	for _, field := range fields {
		if strings.HasPrefix(field, "oneof=") {
			return strings.Split(field[6:], " "), true
		}
	}
	return nil, false
}

func decodeKind(field reflect.StructField, registry folio.Registry) (folio.Type, bool) {
	text := field.Tag.Get("kind")
	typ, err := registry.Resolve(folio.Kind(text))
	if err != nil {
		return folio.Type{}, false
	}

	return typ, true
}

// ---------------------------------- Form ----------------------------------

// hydrate parses the input flat JSON form.
func hydrate(reader io.Reader, typ folio.Type, dst folio.Object) (errDecode error) {
	input, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	lookup := make(map[Path]int, 16)

	gjson.ParseBytes(input).ForEach(func(key, value gjson.Result) bool {
		rv := reflect.Indirect(reflect.ValueOf(dst))

		// Ensure that the field is settable and the path can be reached. If not, allocate
		// everything along the way (analogous to MkDirAll)
		for subpath := range Path(key.String()).Walk() {
			fd, ok := typ.Field(subpath)
			if !ok {
				errDecode = fmt.Errorf("unable to find path %s", key.String())
				return false
			}

			//fmt.Printf(" - %s of %s, field: %s, type: %v\n", subpath, rv.Kind(), fd.Name, fd.Type.Kind())

			switch rv.Kind() {
			case reflect.Struct:
				rv = rv.FieldByIndex(fd.Index)
				rv = newOrDefault(rv)
			case reflect.Ptr:
				rv = newOrDefault(rv)
			case reflect.Slice:
				idx, ok := lookup[subpath]
				if !ok { // If not found, append a new element
					idx = rv.Len()
					lookup[subpath] = idx
					rv.Set(reflect.Append(rv, reflect.New(rv.Type().Elem()).Elem()))
				}

				// Point to the correct index within the slice
				rv = rv.Index(idx)
			}

			//fmt.Print(" => ", rv.Type(), "\n")
		}

		// Try to decode value directly
		if errDecode = unmarshalJSON(rv, value.Raw); errDecode != skip {
			return errDecode == nil
		}

		// Try to decode value via a pointer receiver
		if errDecode = unmarshalJSON(rv.Addr(), value.Raw); errDecode != skip {
			return errDecode == nil
		}

		// Set the value based on the type
		switch rv.Kind() {
		case reflect.String:
			rv.SetString(value.String())
		case reflect.Int:
			rv.SetInt(value.Int())
		case reflect.Bool:
			rv.SetBool(value.Bool())
		case reflect.Float64:
			rv.SetFloat(value.Float())
		default:
			errDecode = fmt.Errorf("unable to decode %s: unsupported type %v", key.String(), rv.Kind())
		}

		return true
	})

	// Skip is safe to ignore
	if errDecode == skip {
		errDecode = nil
	}
	return
}

var skip = errors.New("skip")

func unmarshalJSON(dst reflect.Value, value string) error {
	out, ok := dst.Interface().(interface {
		UnmarshalJSON([]byte) error
	})
	if !ok {
		return skip
	}

	if err := out.UnmarshalJSON([]byte(value)); err != nil {
		return err
	}
	return nil

}

func newOrDefault(rv reflect.Value) reflect.Value {
	switch {
	case rv.Kind() == reflect.Pointer && rv.IsNil():
		instance := reflect.New(rv.Type().Elem())
		rv.Set(instance)
		return instance.Elem()
	case rv.Kind() == reflect.Pointer:
		return rv.Elem()
	//case rv.Kind() == reflect.Slice && rv.Len() == 0:
	//	return reflect.MakeSlice(rv.Type(), 0, 8)
	case rv.Kind() == reflect.Map && rv.IsNil():
		instance := reflect.MakeMap(rv.Type())
		rv.Set(instance)
		return instance
	default:
		return rv
	}
}

func instance(typ reflect.Type) reflect.Value {
	switch typ.Kind() {
	case reflect.Ptr:
		return reflect.New(typ.Elem())
	case reflect.Slice:
		return reflect.MakeSlice(typ, 0, 8)
	case reflect.Map:
		return reflect.MakeMap(typ)
	case reflect.Struct:
		return reflect.New(typ).Elem()
	default:
		return reflect.New(typ).Elem()
	}
}
