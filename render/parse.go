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
