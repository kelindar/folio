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
	Register(Type)
	Resolve(Kind) (Type, error)
}

// Type represents a registration of a resource kind.
type Type struct {
	Kind    Kind         // Kind of the resource
	Type    reflect.Type // Type of the resource
	Options              // Options of the resource
}

// registry represents a registry of various resource kinds.
type registry struct {
	mu   sync.RWMutex
	data map[Kind]Type
	sort []Type // sorted
}

// NewRegistry creates a new registry.
func NewRegistry() Registry {
	return &registry{
		data: make(map[Kind]Type),
	}
}

// Register registers a resource kind into the specified registry.
func (c *registry) Register(typ Type) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[typ.Kind] = typ
	c.sort = slices.SortedFunc(maps.Values(c.data), func(a Type, b Type) int {
		return strings.Compare(
			fmt.Sprintf("%s/%s", a.Sort, a.Kind),
			fmt.Sprintf("%s/%s", a.Sort, a.Kind),
		)
	})
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

	c.Register(Type{
		Kind:    kind,
		Type:    typ,
		Options: options,
	})
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
