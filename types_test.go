package folio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbed(t *testing.T) {
	r := newRegistry()

	// Create a new app
	app, err := New[*Kind3]("my_project")
	assert.NoError(t, err)

	// Create a new deployment with the associated app
	dep, err := New("my_project", func(r *Kind2) error {
		r.App = app.URN()
		return nil
	})
	assert.NoError(t, err)

	req, err := New("my_project", func(r *Kind4) error {
		r.After = Embed{Value: dep}
		return nil
	})
	assert.NoError(t, err)

	// Marshal the request
	data, err := ToJSON(req)
	assert.NoError(t, err)

	// Unmarshal the request
	out, err := FromJSON(r, data)
	assert.NoError(t, err)
	decoded := out.(*Kind4).After.Value.(*Kind2)
	assert.NotNil(t, decoded)
	assert.Equal(t, dep.ID, decoded.ID)

}

// MockObject is a sample struct for testing purposes.
type MockObject struct {
	Meta   `kind:"mock" json:",inline"`
	Name   string
	Age    int
	Income float64
}

func TestParseQuery(t *testing.T) {
	tests := map[string]struct {
		query   string
		object  Object
		expect  Query
		invalid bool
	}{
		"invalid namespace format": {query: "namespace=;state=active", invalid: true},
		"invalid filter format":    {query: "namespace=company;filter=age;state=active", invalid: true},
		"invalid match format":     {query: "namespace=company;match=;", invalid: true},
		"empty query string": {
			query:  "",
			expect: Query{},
		},
		"valid query with all components": {
			query: "namespace=company;state=active;filter=age:30,income:1000;match={Name}",
			object: &MockObject{
				Name:   "Alice",
				Age:    30,
				Income: 50000.00,
			},
			expect: Query{
				Namespaces: []string{"company"},
				States:     []string{"active"},
				Filters: map[string][]string{
					"age":    {"30"},
					"income": {"1000"},
				},
				Match: "Alice",
			},
		},
		"single valid filter": {
			query: "namespace=company;state=active;filter=age:30;match={Name}",
			object: &MockObject{
				Name: "Alice",
			},
			expect: Query{
				Namespaces: []string{"company"},
				States:     []string{"active"},
				Filters: map[string][]string{
					"age": {"30"},
				},
				Match: "Alice",
			},
		},
		"multiple namespaces": {
			query: "namespace=company,person;state=active;filter=age:30;match={Name}",
			object: &MockObject{
				Name: "Alice",
			},
			expect: Query{
				Namespaces: []string{"company", "person"},
				States:     []string{"active"},
				Filters: map[string][]string{
					"age": {"30"},
				},
				Match: "Alice",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.object == nil {
				tt.object = &MockObject{}
			}

			result, err := ParseQuery(tt.query, tt.object, Query{})
			if tt.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestEncodeQuery(t *testing.T) {
	tests := map[string]Query{
		"": {},
		"namespace=company;state=active;filter=age:30;match=Alice;": {
			Namespaces: []string{"company"},
			States:     []string{"active"},
			Filters: map[string][]string{
				"age": {"30"},
			},
			Match: "Alice",
		},
		"namespace=company,person;state=active;filter=age:30;match=Alice;": {
			Namespaces: []string{"company", "person"},
			States:     []string{"active"},
			Filters: map[string][]string{
				"age": {"30"},
			},
			Match: "Alice",
		},
		"namespace=default;index=field1,field2;": {
			Namespaces: []string{"default"},
			Indexes:    []string{"field1", "field2"},
		},
	}

	for expect, query := range tests {
		t.Run(expect, func(t *testing.T) {
			assert.Equal(t, expect, query.String())
		})
	}
}
