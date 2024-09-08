package object

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	ErrKindNotFound = errors.New("resource: kind not found")
)

// Registry represents a registry of various object kinds.
type Registry interface {
	Range(func(kind Kind, typ reflect.Type) error) error
	Register(Kind, reflect.Type)
	Resolve(Kind) (reflect.Type, error)
}

// entry represents a registry entry
type entry struct {
	Kind Kind         // Kind of the resource
	Type reflect.Type // Type of the resource
}

// registry represents a registry of various resource kinds.
type registry struct {
	mu   sync.RWMutex
	data map[Kind]entry
}

// NewRegistry creates a new registry.
func NewRegistry() Registry {
	return &registry{
		data: make(map[Kind]entry),
	}
}

// Register registers a resource kind into the specified registry.
func (c *registry) Register(kind Kind, typ reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[kind] = entry{
		Kind: kind,
		Type: typ,
	}
}

// Range iterates over all the registered resource kinds.
func (c *registry) Range(fn func(kind Kind, typ reflect.Type) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, e := range c.data {
		if err := fn(e.Kind, e.Type); err != nil {
			return err
		}
	}
	return nil
}

// Resolve returns the reflect.Type of the specified resource kind.
func (c *registry) Resolve(kind Kind) (reflect.Type, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if e, ok := c.data[kind]; ok {
		return e.Type, nil
	}
	return nil, fmt.Errorf("%w (%s)", ErrKindNotFound, kind)
}

// ---------------------------------- Generic ----------------------------------

// Register registers a resource kind into the specified registry.
func Register[T Object](c Registry) {
	typ := typeOfT[T]()
	kind, err := KindOf(typ)
	if err != nil {
		panic(err)
	}

	c.Register(kind, typ)
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
