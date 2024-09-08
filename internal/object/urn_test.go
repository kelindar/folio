package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkURN/new-24         	 5232127	       227.1 ns/op	      24 B/op	       1 allocs/op
BenchmarkURN/parse-24       	 3904395	       310.8 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkURN(b *testing.B) {
	b.ReportAllocs()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewURN("my_project", "kind1")
		}
	})

	b.Run("parse", func(b *testing.B) {
		urn, _ := NewURN("my_project", "kind1")
		txt := urn.String()
		for i := 0; i < b.N; i++ {
			_, err := ParseURN(txt)
			assert.NoError(b, err)
		}
	})
}

func TestNewURN(t *testing.T) {
	urn, err := NewURN("my_project", "kind1")
	assert.NoError(t, err)
	assert.NotEmpty(t, urn)
	assert.Equal(t, "my_project", urn.Namespace)
	assert.Equal(t, "kind1", string(urn.Kind))
	assert.Len(t, urn.ID, 20)
	assert.NotEmpty(t, urn.String())
}

func TestMarshalURN(t *testing.T) {
	urn := "urn:my_project:deployment:clmai59hq4ui597oseng"
	parsed, err := ParseURN(urn)
	assert.NoError(t, err)

	data, err := parsed.MarshalJSON()
	assert.NoError(t, err)

	var decoded URN
	assert.NoError(t, decoded.UnmarshalJSON(data))
	assert.Equal(t, parsed, decoded)
}

func TestParseURN(t *testing.T) {
	urn := "urn:my_project:kind1:clmai59hq4ui597oseng"
	parsed, err := ParseURN(urn)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed)
	assert.Equal(t, "my_project", parsed.Namespace)
	assert.Equal(t, "kind1", string(parsed.Kind))
	assert.Equal(t, "clmai59hq4ui597oseng", parsed.ID)
	assert.Equal(t, urn, parsed.String())
	assert.True(t, parsed.IsValid())
}

func TestParseURN_WithoutID(t *testing.T) {
	urn := "urn:my_project:kind1:*"
	parsed, err := ParseURN(urn)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed)
	assert.Equal(t, "my_project", parsed.Namespace)
	assert.Equal(t, "kind1", string(parsed.Kind))
	assert.Len(t, parsed.ID, 20)
	assert.True(t, parsed.IsValid())
}

func TestParseURN_Shortest(t *testing.T) {
	urn := "urn:ab:cd:*"
	parsed, err := ParseURN(urn)
	assert.NoError(t, err)
	assert.NotEmpty(t, parsed)
	assert.Equal(t, "ab", parsed.Namespace)
	assert.Equal(t, "cd", string(parsed.Kind))
	assert.Len(t, parsed.ID, 20)
	assert.True(t, parsed.IsValid())
}

func TestParseURN_Errors(t *testing.T) {
	tests := []string{
		"",
		"urn",
		"urn:",
		"urn:project",
		"urn:project:kind",
		"urn:project:kind:",
		"urn:project:kind:clmai59hq4ui597oseng:",
		"urn:project:kind:clmai59hq4ui597oseng:extra",
		"urn:project:kind:clmai59hq4ui597oseng1",
		"urn:project:123:clmai59hq4ui597oseng",
		"urn:123:clmai59hq4ui597oseng",
		"urn:project:123:clmai59hq4ui597oseng",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			_, err := ParseURN(test)
			assert.Error(t, err)
		})
	}
}

func TestNewURN_Errors(t *testing.T) {
	tests := []struct {
		Project string
		Kind    string
	}{
		{"", ""},
		{"", ""},
		{"project", ""},
		{"project", "123"},
	}

	for _, test := range tests {
		t.Run(test.Project+":"+test.Kind, func(t *testing.T) {
			_, err := NewURN(test.Project, Kind(test.Kind))
			assert.Error(t, err)
		})
	}
}
