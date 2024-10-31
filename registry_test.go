package folio

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInvalidResource(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)

	assert.Panics(t, func() {
		Register[*struct {
			Meta
			Resource string `json:"xxx,inline"`
		}](r)
	})
}

func TestRegistryRange(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotPanics(t, func() {
		Register[*Kind1](registry)
	})

	// Create a new resource using generic function
	r, err := New[*Kind1]("my_project")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	// Range over the registered kinds
	count := 0
	for typ := range registry.Types() {
		assert.NotNil(t, typ)
		count++
	}
	assert.Equal(t, 2, count)

	// Get the reflect.Type of the specified resource kind
	typ, err := registry.Resolve("kind1")
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(Kind1{}), typ.Type)
}

func TestRegisterInvalid(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, r)
	assert.Error(t, r.Register(Type{
		Kind: "kind1",
		Type: reflect.TypeOf(time.Time{}),
	}))

	assert.Error(t, r.Register(Type{
		Kind: "kind1",
		Type: reflect.TypeOf(Kind5{}),
	}))
}

// ---------------------------------- Test Types ----------------------------------

type Kind1 struct {
	Meta `kind:"kind1" json:",inline"`
	Name string `json:"name"`
	Link URN    `json:"link"`
}

type Kind2 struct {
	Meta `kind:"kind2" json:",inline"`
	Name string
	Env  string `json:"env"`
	App  URN    `json:"app" kind:"kind1" query:"namespace=*;match=Inc"`
}

type Kind3 struct {
	Meta `kind:"kind3" json:",inline"`
}

type Kind4 struct {
	Meta   `kind:"kind4" json:",inline"`
	Name   string
	Before Embed `json:"before,omitempty"`
	After  Embed `json:"after,omitempty"`
}

type Kind5 struct {
	Meta `kind:"kind2" json:",inline"`
	Name string
	Env  string `json:"env"`
	App  URN    `json:"app" kind:"kind1" query:"xxx"`
}

func newRegistry() Registry {
	registry := NewRegistry()
	Register[*Kind1](registry)
	Register[*Kind2](registry)
	Register[*Kind3](registry)
	Register[*Kind4](registry)
	return registry
}

// ---------------------------------- Field Paths ----------------------------------

// Example structs to demonstrate the transformation
type Engine struct {
	Type  string `json:"type"`
	Power int    `json:"power"`
}

type Employee struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type Department struct {
	Name      string     `json:"name"`
	Employees []Employee `json:"employees,omitempty"`
}

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
}

type Company struct {
	Name        string       `json:"name"`
	Address     *Address     `json:"address,omitempty"`
	Departments []Department `json:"departments,omitempty"`
}

type Car struct {
	Meta        `kind:"car" json:",inline"`
	Type        string   `json:"type"`
	Year        int      `json:"year"`
	Model       string   `json:"model"`
	Description string   `json:"description"`
	Company     string   `json:"company"`
	Engine      *Engine  `json:"engine,omitempty"`
	Engines     []Engine `json:"engines,omitempty"`
	CompanyInfo *Company `json:"companyInfo,omitempty"`
}

func TestFieldsOf(t *testing.T) {
	registry := newRegistry()
	Register[*Kind1](registry)

	typ, err := registry.Resolve("kind1")
	assert.NoError(t, err)
	assert.NotNil(t, typ)

	f, ok := typ.Field("name")
	assert.True(t, ok)
	assert.Equal(t, "Name", f.Name)
	assert.Equal(t, 10, len(typ.fields))
}

func TestPath(t *testing.T) {
	p := Path("name")
	assert.Equal(t, "name", p.String())
	assert.Equal(t, "prefix-6e616d65", p.ID("prefix"))
	assert.Equal(t, "Name", p.Label())
}

func TestField_Slice(t *testing.T) {
	registry := newRegistry()
	Register[*Car](registry)

	typ, err := registry.Resolve("car")
	assert.NoError(t, err)
	assert.NotNil(t, typ)

	{ // Test the elem type
		f, ok := typ.Field("engines.41354")
		assert.True(t, ok)
		assert.Equal(t, reflect.Slice, f.Type.Kind())
	}

	{ // Test the slice itself
		f, ok := typ.Field("engines")
		assert.True(t, ok)
		assert.Equal(t, reflect.Slice, f.Type.Kind())
	}

	{ // Test the field of the elem
		f, ok := typ.Field("engines.41354.type")
		assert.True(t, ok)
		assert.Equal(t, reflect.String, f.Type.Kind())
	}

}
