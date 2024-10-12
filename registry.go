package folio

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
)

var (
	ErrKindNotFound = errors.New("resource: kind not found")
)

// Registry represents a registry of various object kinds.
type Registry interface {
	Types() iter.Seq[Type]
	Register(Type) error
	Resolve(Kind) (Type, error)
}

// Type represents a registration of a resource kind.
type Type struct {
	fields  map[Path]reflect.StructField
	Kind    Kind         // Kind of the resource
	Type    reflect.Type // Type of the resource
	Options              // Options of the resource
}

// Field retrieves a field information by the specified path.
func (t *Type) Field(path Path) (reflect.StructField, bool) {
	field, ok := t.fields[path]
	return field, ok
}

// registry represents a registry of various resource kinds.
type registry struct {
	mu   sync.RWMutex
	data map[Kind]Type
	sort []Type // sorted
}

// NewRegistry creates a new registry.
func NewRegistry() Registry {
	r := &registry{
		data: make(map[Kind]Type),
	}

	registerBuiltins(r)
	return r
}

// Register registers a resource kind into the specified registry.
func (c *registry) Register(typ Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Make sure the resource kind is an Object
	instance := reflect.New(typ.Type).Interface()
	if _, ok := instance.(Object); !ok {
		return fmt.Errorf("resource: unable to register '%s', not an Object", typ.Kind)
	}

	// Iterate over all the fields of the type and make sure any "query" tags can be parsed
	for i := 0; i < typ.Type.NumField(); i++ {
		field := typ.Type.Field(i)
		if field.Tag.Get("query") != "" {
			if _, err := ParseQuery(field.Tag.Get("query"), instance, Query{}); err != nil {
				return fmt.Errorf("resource: unable to register '%s', invalid query tag on '%s' field", typ.Kind, field.Name)
			}
		}
	}

	// Construct the fields map
	typ.fields = fieldsOf(typ.Type)

	//Register the resource kind and sort the data
	c.data[typ.Kind] = typ
	c.sort = slices.SortedFunc(maps.Values(c.data), func(a Type, b Type) int {
		return strings.Compare(
			fmt.Sprintf("%s/%s", a.Sort, a.Kind),
			fmt.Sprintf("%s/%s", a.Sort, a.Kind),
		)
	})
	return nil
}

// Types iterates over all the registered resource kinds.
func (c *registry) Types() iter.Seq[Type] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Values(c.sort)
}

// Resolve returns the reflect.Type of the specified resource kind.
func (c *registry) Resolve(kind Kind) (Type, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if t, ok := c.data[kind]; ok {
		return t, nil
	}
	return Type{}, fmt.Errorf("%w (%s)", ErrKindNotFound, kind)
}

// ---------------------------------- Generic ----------------------------------

// Register registers a resource kind into the specified registry.
func Register[T Object](c Registry, opts ...Options) {
	typ := typeOfT[T]()
	kind, err := KindOf(typ)
	if err != nil {
		panic(err)
	}

	options := defaultOptions(kind)
	if len(opts) > 0 {
		options = opts[0]
	}

	if err := c.Register(Type{
		Kind:    kind,
		Type:    typ,
		Options: options,
	}); err != nil {
		panic(err)
	}
}

// ---------------------------------- Reflection ----------------------------------

// cache to avoid unnecessary reflection every time
var cache sync.Map

// KindOfT returns the Kind of the object.
func KindOfT[T any]() (Kind, error) {
	return KindOf(typeOfT[T]())
}

// KindOf returns the Kind of the object.
func KindOf(typ reflect.Type) (Kind, error) {
	if kind, ok := cache.Load(typ); ok {
		return kind.(Kind), nil
	}

	res, ok := typ.FieldByName("Meta")
	switch {
	case !ok:
		return "", fmt.Errorf("resource: unable to register, not a valid resource (%s)", typ.Name())
	case res.Type.Kind() != reflect.Struct:
		return "", fmt.Errorf("resource: unable to register, not a valid resource (%s)", typ.Name())
	case res.Type != typeResource:
		return "", fmt.Errorf("resource: unable to register, not a valid resource (%s)", typ.Name())
	}

	// Parse the resource Kind from the json tag
	kind := Kind(strings.ToLower(res.Tag.Get("kind")))
	switch {
	case len(kind) == 0:
		return "", fmt.Errorf("resource: unable to register, no json tag found (%s)", typ.Name())
	case !regexName.MatchString(string(kind)):
		return "", fmt.Errorf("resource: unable to register, invalid resource kind (%s)", typ.Name())
	}

	// Cache the result
	cache.Store(typ, kind)
	return kind, nil
}

// typeOf returns the reflect.Type of the value provided
func typeOf(v any) reflect.Type {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		return rv.Type().Elem()
	default:
		return rv.Type()
	}
}

// typeOfT returns the reflect.Type of the type provided
func typeOfT[T any]() reflect.Type {
	var instance T
	return typeOf(instance)
}

// ---------------------------------- Path Scan ----------------------------------

// fieldsOf constructs a map from JSON paths to *reflect.StructField for the given type.
func fieldsOf(typ reflect.Type) map[Path]reflect.StructField {
	fields := make(map[Path]reflect.StructField, 16)
	visited := make(map[reflect.Type]bool, 16)
	walkType(typ, "", fields, visited)
	return fields
}

// walkType recursively traverses the type to build JSON paths.
func walkType(typ reflect.Type, current string, paths map[Path]reflect.StructField, visited map[reflect.Type]bool) {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// Avoid infinite recursion on recursive types
	if visited[typ] {
		return
	}

	visited[typ] = true
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name, inline := jsonName(field)
			switch {
			case inline:
				walkType(field.Type, current, paths, visited)
				continue
			case name == "":
				continue // Skip fields with JSON tag "-"
			case current != "":
				name = current + "." + name
			}

			// Store and recurse into the field
			paths[Path(name)] = field
			walkType(field.Type, name, paths, visited)
		}

	// Recurse into the element type without changing the path
	case reflect.Slice, reflect.Array, reflect.Map:
		walkType(typ.Elem(), current, paths, visited)
	default:
		// Ignore unsupported types
	}
}

// jsonName returns the JSON name of the field and whether it is inlined.
// A field is considered inlined if it's anonymous or has the `json:",inline"` tag.
func jsonName(field reflect.StructField) (name string, inline bool) {
	tag := strings.Split(field.Tag.Get("json"), ",")
	for _, part := range tag[1:] {
		if part == "inline" {
			inline = true
			break
		}
	}

	// If the field is anonymous, it's inlined
	if field.Anonymous {
		inline = true
	}

	switch {
	case len(tag) > 0 && tag[0] != "" && tag[0] != "-":
		return tag[0], inline
	case !inline && field.Anonymous:
		return field.Type.Name(), inline
	default:
		return name, inline
	}
}
