package folio

import (
	"reflect"
	"testing"

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
	assert.Equal(t, 1, count)

	// Get the reflect.Type of the specified resource kind
	typ, err := registry.Resolve("kind1")
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(Kind1{}), typ.Type)
}

// ---------------------------------- Test Types ----------------------------------

type Kind1 struct {
	Meta `kind:"kind1" json:",inline"`
	Name string
	Link URN `json:"link"`
}

type Kind2 struct {
	Meta `kind:"kind2" json:",inline"`
	Name string
	Env  string `json:"env"`
	App  URN    `json:"app"`
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

func newRegistry() Registry {
	registry := NewRegistry()
	Register[*Kind1](registry)
	Register[*Kind2](registry)
	Register[*Kind3](registry)
	Register[*Kind4](registry)
	return registry
}
