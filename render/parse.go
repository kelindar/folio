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

	obj := reflect.Indirect(reflect.ValueOf(dst))

	const appendop = 1e6 // starting from this index, it means we're appending to an array

	gjson.ParseBytes(input).ForEach(func(key, value gjson.Result) bool {
		/*key = strings.TrimPrefix(key.String(), prefix)
		if err = sjson.SetBytes(input, key, value.Value()); err != nil {
			err = fmt.Errorf("unable to decode %s: %w", key, err)
			return false
		}*/

		field, ok := typ.Field(Path(key.String()))
		if !ok {
			errDecode = fmt.Errorf("unable to find path %s", key.String())
			return false
		}

		fmt.Printf("[%s] value: %s, field: %s, type: %v\n", key.String(), value.String(), field.Name, field.Type.Kind())
		index := make([]int, 0, len(field.Index))

		// Ensure that the field is settable and the path can be reached. If not, allocate
		// everything along the way (analogous to MkDirAll)
		for subpath := range Path(key.String()).Walk() {
			fd, ok := typ.Field(subpath)
			if !ok {
				errDecode = fmt.Errorf("unable to find path %s", key.String())
				return false
			}

			index = append(index, fd.Index...)
			fmt.Printf(" - %s, field: %s, type: %v\n",
				subpath, fd.Name, fd.Type.Kind())

			fv := obj.FieldByIndex(index)

			// For slices and pointers, always create an empty value. The value then
			// will get populated by subsequent calls, in order of appearance.
			switch {
			case fv.Kind() == reflect.Ptr && fv.IsNil():
				fv.Set(reflect.New(fv.Type().Elem()))
			case fv.Kind() == reflect.Slice: //
				fv.Set(reflect.MakeSlice(fv.Type(), 0, 8))
				//case fv.Kind() == reflect.Slice && fd.Index[len(fd.Index)-1] == appendop:
				//	fv.Set(reflect.Append(fv, reflect.New(fv.Type().Elem())))

			}
		}

		rv := obj.FieldByIndex(index)

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
