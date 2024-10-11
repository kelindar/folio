package render

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/kelindar/folio"
	"github.com/tidwall/sjson"
)

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

// decodeForm parses the input flat JSON form.
func decodeForm(reader io.Reader, dst any, prefix string) (err error) {
	input, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	var data map[string]any
	if err := json.Unmarshal(input, &data); err != nil {
		return err
	}

	nested := make([]byte, 0, len(input)*2)
	for key, value := range data {
		key = strings.TrimPrefix(key, prefix)
		if nested, err = sjson.SetBytes(nested, key, value); err != nil {
			return fmt.Errorf("unable to decode %s: %w", key, err)
		}
	}

	return json.Unmarshal(nested, dst)
}

func validationPath(path string) string {
	// If the first letter is capitalized, it's a nested field
	if len(path) > 0 && path[0] >= 'A' && path[0] <= 'Z' {
		path = path[strings.Index(path, ".")+1:]
	}

	return strings.ReplaceAll(path, ".", "-")
}

// ---------------------------------- Path Scan ----------------------------------

// jsonPath finds a JSON path in the given type.
func jsonPath(parent reflect.Type, path string) (reflect.StructField, error) {
	parts := strings.Split(path, ".")
	rtype := parent

	var out reflect.StructField
	for _, component := range parts {
		switch rtype.Kind() {
		case reflect.Struct:
			// continue
		case reflect.Slice:
			rtype = rtype.Elem()
		case reflect.Map:
			rtype = rtype.Elem()
		case reflect.Ptr:
			rtype = rtype.Elem()
		default:
			return out, fmt.Errorf("type %s is not supported", rtype)
		}

		found := false
		count := rtype.NumField()

		for i := 0; i < count; i++ {
			if field := rtype.Field(i); jsonName(field) == component {
				found = true
				rtype = field.Type
				out = field
				if rtype.Kind() == reflect.Ptr {
					rtype = rtype.Elem()
				}
				break
			}
		}

		if !found {
			return out, fmt.Errorf("field %q not found in type %s", component, rtype)
		}
	}

	return out, nil
}

// jsonName returns the JSON name of the field.
func jsonName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}

	if v := strings.Split(tag, ","); len(v) > 0 && v[0] != "" {
		return v[0]
	}
	return field.Name
}
