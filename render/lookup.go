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
	Init(*Props) bool
	Choices() iter.Seq2[string, string]
	Current() (string, string)
	Len() int
}

// ---------------------------------- Enumerable ----------------------------------

// lookupEnum represents an lookupEnum lookup parsed from oneof tag.
type lookupEnum struct {
	current reflect.Value
	choices []string
}

// Parses oneof tag from validator e.g.: "required,oneof=male female prefer_not_to"
func decodeOneOf(tag string) []string {
	fields := strings.Split(tag, ",")
	for _, field := range fields {
		if strings.HasPrefix(field, "oneof=") {
			return strings.Split(field[6:], " ")
		}
	}
	return nil
}

// lookupForEnum creates a new lookup for an enum field.
func lookupForEnum(field reflect.StructField, rv reflect.Value) *lookupEnum {
	return &lookupEnum{
		current: rv,
		choices: decodeOneOf(field.Tag.Get("validate")),
	}
}

// Init initializes the lookup it returns false if the field is not valid.
func (o *lookupEnum) Init(props *Props) bool {
	return strings.Contains(props.Field.Tag.Get("validate"), "oneof=")
}

// Value returns the currently selected value.
func (o *lookupEnum) Current() (string, string) {
	return o.current.String(), convert.TitleCase(o.current.String())
}

// Choices returns the choices for the given state.
func (o *lookupEnum) Choices() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, choice := range o.choices {
			if !yield(choice, convert.TitleCase(choice)) {
				break
			}
		}
	}
}

// Len returns the number of choices.
func (o *lookupEnum) Len() int {
	return len(o.choices)
}

// ---------------------------------- Reference ----------------------------------

type lookupUrn struct {
	object  folio.Object
	storage folio.Storage
	kind    folio.Type
	query   folio.Query
}

func lookupForUrn(_ reflect.StructField, _ reflect.Value) *lookupUrn {
	return &lookupUrn{}
}

func (o *lookupUrn) Init(props *Props) bool {
	kind := props.Field.Tag.Get("kind")
	if props.Field.Type.Name() != "URN" || kind == "" {
		return false
	}

	// The kind must be resolved
	typ, err := props.Registry.Resolve(folio.Kind(kind))
	if err != nil {
		return false
	}

	// Fetch the object if it's available
	if urn, ok := props.Value.Interface().(folio.URN); ok {
		if value, err := props.Store.Fetch(urn); err == nil {
			o.object = value
		}
	}

	// Parse the query
	query, err := folio.ParseQuery(props.Field.Tag.Get("query"), props.Parent, folio.Query{
		Namespaces: []string{props.Parent.URN().Namespace},
	})
	if err != nil {
		panic(err) // Compiler error
	}

	//o.current = props.Value
	o.storage = props.Store
	o.kind = typ
	o.query = query
	return true
}

// Value returns the currently selected value.
func (o *lookupUrn) Current() (string, string) {
	if o.object == nil {
		return "", ""
	}
	return o.object.URN().String(), TitleOf(o.object)
}

// Choices returns the choices for the given state.
func (o *lookupUrn) Choices() iter.Seq2[string, string] {
	seq, err := o.storage.Search(o.kind.Kind, o.query)
	if err != nil {
		return func(yield func(string, string) bool) {}
	}

	return func(yield func(string, string) bool) {
		for obj := range seq {
			yield(obj.URN().String(), TitleOf(obj))
		}
	}
}

// Len returns the number of choices.
func (o *lookupUrn) Len() int {
	count, _ := o.storage.Count(o.kind.Kind, o.query)
	return count
}

// ---------------------------------- Lookup Functions ----------------------------------

// currentKey returns the current key from the lookup.
func currentKey(lookup Lookup) string {
	k, _ := lookup.Current()
	return k
}

// currentValue returns the current value from the lookup.
func currentValue(lookup Lookup) string {
	_, v := lookup.Current()
	return v
}
