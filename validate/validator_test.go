package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

// Example structs to demonstrate the transformation
type Engine struct {
	Type  string `is:"in(V8|V6)" json:"type"`
	Power int    `json:"power"`
}

type Car struct {
	Type        string   `json:"type"`
	Year        int      `json:"year"`
	Model       string   `json:"model"`
	Description string   `json:"description"`
	Company     string   `json:"company"`
	Engine      *Engine  `json:"engine,omitempty"`
	Engines     []Engine `is:"required" json:"engines,omitempty"`
	CompanyInfo *Company `json:"companyInfo,omitempty"`
}

func TestValidate_Simple(t *testing.T) {
	car := &Car{
		Type:        "car",
		Year:        2018,
		Model:       "Mustang",
		Description: "A car",
		Company:     "Ford",
		Engines: []Engine{
			{Type: "V8", Power: 500},
		},
	}

	// Validate the struct
	ok, err := Struct(car)
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestValidate_Errors(t *testing.T) {
	car := &Car{
		Type:        "car",
		Year:        2018,
		Model:       "Mustang",
		Description: "A car",
		Company:     "Ford",
		Engines: []Engine{
			{Type: "V8", Power: 500},
			{Type: "V2", Power: 100},
		},
	}

	// Validate the struct
	ok, err := Struct(car)
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestNestedStruct(t *testing.T) {
	type EvenMoreNestedStruct struct {
		Bar string `is:"length(3|5)"`
	}
	type NestedStruct struct {
		Foo                 string `is:"length(3|5),required"`
		EvenMoreNested      EvenMoreNestedStruct
		SliceEvenMoreNested []EvenMoreNestedStruct
		MapEvenMoreNested   map[string]EvenMoreNestedStruct
	}
	type OuterStruct struct {
		Nested NestedStruct
	}

	var tests = []struct {
		param       interface{}
		expected    bool
		expectedErr string
	}{
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "",
			},
		}, false, "Nested.Foo: non zero value required"},
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "123",
			},
		}, true, ""},
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "123456",
			},
		}, false, "Nested.Foo: 123456 does not validate as length(3|5)"},
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "123",
				EvenMoreNested: EvenMoreNestedStruct{
					Bar: "123456",
				},
			},
		}, false, "Nested.EvenMoreNested.Bar: 123456 does not validate as length(3|5)"},
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "123",
				SliceEvenMoreNested: []EvenMoreNestedStruct{
					{
						Bar: "123456",
					},
				},
			},
		}, false, "Nested.SliceEvenMoreNested.0.Bar: 123456 does not validate as length(3|5)"},
		{OuterStruct{
			Nested: NestedStruct{
				Foo: "123",
				MapEvenMoreNested: map[string]EvenMoreNestedStruct{
					"Foo": {
						Bar: "123456",
					},
				},
			},
		}, false, "Nested.MapEvenMoreNested.Foo.Bar: 123456 does not validate as length(3|5)"},
	}

	for _, test := range tests {
		actual, err := Struct(test.param)
		assert.Equal(t, test.expected, actual)
		if test.expectedErr != "" {
			assert.Error(t, err)
			assert.EqualError(t, err, test.expectedErr)
		} else {
			assert.NoError(t, err)
		}
	}
}
