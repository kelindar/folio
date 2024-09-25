package render

import (
	"iter"
	"reflect"
	"strings"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/convert"
)

// Lookup represents a lookup for a field.
type Lookup interface {
	Choices(folio.Object, folio.Storage) iter.Seq2[string, string]
	Key() string
	Value() string
}

type enum struct {
	current reflect.Value
	choices []string
}

func newEnum(field reflect.StructField, rv reflect.Value) *enum {
	return &enum{
		current: rv,
		choices: parseOneOf(field.Tag.Get("validate")),
	}
}

// Key returns the currently selected key.
func (o *enum) Key() string {
	return o.current.String()
}

// Value returns the currently selected value.
func (o *enum) Value() string {
	return convert.TitleCase(o.current.String())
}

// Choices returns the choices for the given state.
func (o *enum) Choices(_ folio.Object, _ folio.Storage) iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, choice := range o.choices {
			if !yield(choice, convert.TitleCase(choice)) {
				break
			}
		}
	}
}

// Parses oneof tag from validator e.g.: "required,oneof=male female prefer_not_to"
func parseOneOf(tag string) []string {
	fields := strings.Split(tag, ",")
	for _, field := range fields {
		if strings.HasPrefix(field, "oneof=") {
			return strings.Split(field[6:], " ")
		}
	}
	return nil
}
