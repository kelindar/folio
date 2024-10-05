package render

import (
	"reflect"
	"strings"

	"github.com/kelindar/folio"
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
