package object

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkResource/new-24         	 3410836	       350.9 ns/op	     256 B/op	       3 allocs/op
*/
func BenchmarkResource(b *testing.B) {
	b.ReportAllocs()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			New[*Kind1]("my_project")
		}
	})
}

func TestNewWith(t *testing.T) {

	// Create a new resource using generic function
	r, err := NewWith("my_project", func(r *Kind1) error {
		r.Name = "my_name"
		r.Desc = "my_description"
		return nil
	})

	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.NotNil(t, r.URN())
	assert.Equal(t, Kind("kind1"), r.Kind)
	assert.Equal(t, "my_project", r.Namespace)
	assert.NotEmpty(t, r.ID)
	assert.Equal(t, "my_name", r.Name)
	assert.Equal(t, "my_description", r.Desc)
}

func TestMarshal(t *testing.T) {
	registry := newRegistry()
	o, err := New[*Kind3]("my_project")
	assert.NoError(t, err)
	assert.NotNil(t, o)
	o.ID = "xxx"

	// Marshal the resource to JSON
	encoded, err := ToJSON(o)
	assert.NoError(t, err)
	assert.JSONEq(t, `{
		"kind": "kind3",
		"namespace": "my_project",
		"id": "xxx"
	}`, string(encoded))

	// Unmarshal the JSON to a resource
	{
		decoded, err := FromJSON(registry, encoded)
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
		assert.Equal(t, o, decoded)
	}

	// Read the resource information
	{
		decoded, err := ReadJSON(registry, bytes.NewReader(encoded))
		assert.NoError(t, err)
		assert.NotNil(t, decoded)
		assert.Equal(t, o, decoded)
	}
}
