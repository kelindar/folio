package render

import (
	"fmt"
	"io"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/errors"
	"github.com/kelindar/folio/internal/walk"
	"github.com/tidwall/gjson"
)

var (
	exIn = regexp.MustCompile(`^in\((.*)\)`)
)

// ---------------------------------- Struct Fields ----------------------------------

func isEnum(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("is"), "in(")
}

func isEmail(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("is"), "email")
}

func isRequired(field reflect.StructField) bool {
	return strings.Contains(field.Tag.Get("is"), "required")
}

// Parses oneof tag from validator e.g.: "required,oneof=male female prefer_not_to"
func decodeOneOf(field reflect.StructField) ([]string, bool) {
	if !isEnum(field) {
		return nil, false
	}

	fields := strings.Split(field.Tag.Get("is"), ",")
	for _, field := range fields {
		if out := exIn.FindAllStringSubmatch(field, -1); len(out) > 0 {
			return strings.Split(out[0][1], "|"), true
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
func hydrate(reader io.Reader, typ folio.Type, dst folio.Object, vd errors.Validator) (_ []errors.Validation, errDecode error) {
	input, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	lookup := make(map[Path]int, 16)
	reverse := make(slicePath, 0, 16)

	walk.Walk(dst, func(rv reflect.Value, field *reflect.StructField, path []string) error {
		//fmt.Printf(" - %s of %s (%s)\n", strings.Join(path, "."), v.Kind(), v.Type().Name())

		switch {

		// If we have a non-empty slice, reset it
		case rv.Kind() == reflect.Slice && rv.Len() > 0:
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 8))
		}

		return nil
	})

	// Reset slices, we need to do it only once to be able to overwrite
	gjson.ParseBytes(input).ForEach(func(key, value gjson.Result) bool {
		rv := reflect.Indirect(reflect.ValueOf(dst))
		for subpath := range Path(key.String()).Walk() {
			fd, ok := typ.Field(subpath)
			if !ok {
				errDecode = fmt.Errorf("unable to find path %s", key.String())
				return false
			}

			switch {
			case rv.Kind() == reflect.Struct:
				rv = newOrDefault(rv.FieldByIndex(fd.Index))

			// If we have a non-empty slice, reset it
			case rv.Kind() == reflect.Slice && rv.Len() > 0:
				rv.Set(reflect.MakeSlice(rv.Type(), 0, 8))
			}
		}
		return true
	})

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

			// fmt.Printf(" - %s of %s, field: %s, type: %v\n", subpath, rv.Kind(), fd.Name, fd.Type.Kind())

			switch rv.Kind() {
			case reflect.Struct:
				rv = newOrDefault(rv.FieldByIndex(fd.Index))
			case reflect.Slice:
				idx, ok := lookup[subpath]
				if !ok { // If not found, append a new element
					idx = rv.Len()
					lookup[subpath] = idx
					reverse = remap(reverse, subpath, idx)

					slice := reflect.Append(rv, reflect.New(rv.Type().Elem()).Elem())
					rv.Set(slice)
					rv = slice
				}

				// Point to the correct index within the slice
				rv = rv.Index(idx)
			}
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

	// Check if we have any errors during parsing
	if errDecode != nil && errDecode != skip {
		return nil, errDecode
	}

	// If we parsed successfully, validate the object
	if errs, ok := vd.Validate(dst); !ok {
		for i := range errs {
			errs[i].Path = errorPath(reverse, errs[i].Path)
		}
		return errs, nil
	}

	// Success
	return nil, nil
}

var skip = fmt.Errorf("skip")

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
	default:
		return rv
	}
}

type slicePath [][2]Path

func (a slicePath) Len() int           { return len(a) }
func (a slicePath) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a slicePath) Less(i, j int) bool { return len(a[i][1]) > len(a[j][1]) }

// swap iterates over the ordered mappings and replaces the first matching prefix
func swap(mappings slicePath, input Path, fromIdx int, toIdx int) Path {
	for _, m := range mappings {
		a := string(m[fromIdx])
		b := string(m[toIdx])

		if strings.HasPrefix(string(input), a) {
			return Path(strings.Replace(string(input), a, b, 1))
		}
	}
	return input
}

// remap updates the paths map by remapping the subpath with the given index
func remap(mappings slicePath, subpath Path, index int) slicePath {
	i := strings.LastIndex(string(subpath), ".")
	p := swap(mappings, Path(fmt.Sprintf("%s.%d", subpath[:i], index)), 1, 0)

	// Append the new path & keep sorted
	defer sort.Sort(mappings)
	return append(mappings, [2]Path{p, subpath})
}

// errorPath generates an error path by replacing the Key with the Prefix.
func errorPath(mappings slicePath, path Path) Path {
	return swap(mappings, path, 0, 1)
}
