package render

import (
	"iter"
	"reflect"
	"strings"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/internal/convert"
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

// lookupForEnum creates a new lookup for an enum field.
func lookupForEnum(rv reflect.Value) *lookupEnum {
	return &lookupEnum{
		current: rv,
	}
}

// Init initializes the lookup it returns false if the field is not valid.
func (o *lookupEnum) Init(props *Props) bool {
	if choices, ok := decodeOneOf(props.Field); ok {
		o.choices = choices
		return true
	}
	return false
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

// ---------------------------------- Flags ----------------------------------

// lookupFlags represents a flags lookup parsed from flags tag for multiple selection.
type lookupFlags struct {
	current reflect.Value // slice of strings
	choices []string
}

// lookupForFlags creates a new lookup for a flags field.
func lookupForFlags(rv reflect.Value) *lookupFlags {
	return &lookupFlags{
		current: rv,
	}
}

// Init initializes the lookup it returns false if the field is not valid.
func (f *lookupFlags) Init(props *Props) bool {
	if choices, ok := decodeFlags(props.Field); ok {
		f.choices = choices
		return true
	}
	return false
}

// Current returns the currently selected values as a comma-separated string.
func (f *lookupFlags) Current() (string, string) {
	if f.current.Kind() != reflect.Slice {
		return "", ""
	}

	var selected []string
	for i := 0; i < f.current.Len(); i++ {
		selected = append(selected, f.current.Index(i).String())
	}

	if len(selected) == 0 {
		return "", "None selected"
	}

	// Convert to title case for display
	var displayValues []string
	for _, val := range selected {
		displayValues = append(displayValues, convert.TitleCase(val))
	}

	return strings.Join(selected, ","), strings.Join(displayValues, ", ")
}

// Choices returns the choices for the given state.
func (f *lookupFlags) Choices() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, choice := range f.choices {
			if !yield(choice, convert.TitleCase(choice)) {
				break
			}
		}
	}
}

// Len returns the number of choices.
func (f *lookupFlags) Len() int {
	return len(f.choices)
}

// Contains checks if a value is currently selected.
func (f *lookupFlags) Contains(value string) bool {
	if f.current.Kind() != reflect.Slice {
		return false
	}

	for i := 0; i < f.current.Len(); i++ {
		if f.current.Index(i).String() == value {
			return true
		}
	}
	return false
}

// ---------------------------------- Reference ----------------------------------

type lookupUrn struct {
	object  folio.Object
	storage folio.Storage
	kind    folio.Type
	query   folio.Query
}

func lookupForUrn() *lookupUrn {
	return &lookupUrn{}
}

func (o *lookupUrn) Init(props *Props) bool {
	kind := props.Field.Tag.Get("kind")
	if props.Value.Type().Name() != "URN" || kind == "" {
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
		Namespace: props.Context.Namespace,
	})
	if err != nil {
		return false // Invalid query
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

// ---------------------------------- URN Slice Reference ----------------------------------

// lookupUrnSlice represents a lookup for a slice of URNs
type lookupUrnSlice struct {
	objects []folio.Object
	storage folio.Storage
	kind    folio.Type
	query   folio.Query
	value   reflect.Value
}

// lookupForUrnSlice creates a new lookup for a slice of URNs
func lookupForUrnSlice() *lookupUrnSlice {
	return &lookupUrnSlice{}
}

// Init initializes the URN slice lookup
func (o *lookupUrnSlice) Init(props *Props) bool {
	kind := props.Field.Tag.Get("kind")

	// Ensure it's a slice of URNs
	if props.Value.Kind() != reflect.Slice || props.Value.Type().Elem().Name() != "URN" || kind == "" {
		return false
	}

	// The kind must be resolved
	typ, err := props.Registry.Resolve(folio.Kind(kind))
	if err != nil {
		return false
	}

	// Load up all referenced objects
	o.objects = make([]folio.Object, 0, props.Value.Len())
	for i := 0; i < props.Value.Len(); i++ {
		item := props.Value.Index(i)
		if urn, ok := item.Interface().(folio.URN); ok {
			if value, err := props.Store.Fetch(urn); err == nil {
				o.objects = append(o.objects, value)
			}
		}
	}

	// Parse the query
	query, err := folio.ParseQuery(props.Field.Tag.Get("query"), props.Parent, folio.Query{
		Namespace: props.Context.Namespace,
	})
	if err != nil {
		return false // Invalid query
	}

	o.value = props.Value
	o.storage = props.Store
	o.kind = typ
	o.query = query
	return true
}

// Current returns the currently selected value (empty for slices, as multiple values are selected)
func (o *lookupUrnSlice) Current() (string, string) {
	return "", "" // Multiple values selected
}

// Choices returns all available choices for the URN selector
func (o *lookupUrnSlice) Choices() iter.Seq2[string, string] {
	seq, err := o.storage.Search(o.kind.Kind, o.query)
	if err != nil {
		return func(yield func(string, string) bool) {}
	}

	// Create a map of already selected URNs for quick lookup
	selected := make(map[string]bool)
	for i := 0; i < o.value.Len(); i++ {
		item := o.value.Index(i)
		if urn, ok := item.Interface().(folio.URN); ok {
			selected[urn.String()] = true
		}
	}

	return func(yield func(string, string) bool) {
		for obj := range seq {
			urn := obj.URN().String()
			isSelected := selected[urn]
			if !yield(urn, TitleOf(obj)+" "+displaySelectedState(isSelected)) {
				break
			}
		}
	}
}

func (o *lookupUrnSlice) Contains(urn string) bool {
	for _, obj := range o.objects {
		if obj.URN().String() == urn {
			return true
		}
	}
	return false
}

// Helper to display if an item is already selected
func displaySelectedState(selected bool) string {
	if selected {
		return "(Selected)"
	}
	return ""
}

// Len returns the number of choices
func (o *lookupUrnSlice) Len() int {
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
