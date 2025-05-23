package validate

import (
	"fmt"
	"strconv"
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
	Type        string   `json:"type" is:"alphanum,!uuid"`
	Year        int      `json:"year" is:"min(2000)"`
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
		Type:        "car x",
		Year:        1999,
		Model:       "Mustang",
		Description: "A car",
		Company:     "Ford",
		Engines: []Engine{
			{Type: "V8", Power: 500},
			{Type: "V1", Power: 100},
			{Type: "V2", Power: 100},
		},
	}

	// Validate the struct
	ok, err := Struct(car)
	assert.False(t, ok)
	assert.NotNil(t, err)
	errs := err.(Errors).Errors()
	assert.Len(t, errs, 4)
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
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "",
				},
			},
			expected:    false,
			expectedErr: "Nested.Foo: Foo is a required field",
		},
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "123",
				},
			},
			expected:    true,
			expectedErr: "",
		},
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "123456",
				},
			},
			expected:    false,
			expectedErr: "Nested.Foo: Foo must be between 3 and 5",
		},
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "123",
					EvenMoreNested: EvenMoreNestedStruct{
						Bar: "123456",
					},
				},
			},
			expected:    false,
			expectedErr: "Nested.EvenMoreNested.Bar: Bar must be between 3 and 5",
		},
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "123",
					SliceEvenMoreNested: []EvenMoreNestedStruct{
						{
							Bar: "123456",
						},
					},
				},
			},
			expected:    false,
			expectedErr: "Nested.SliceEvenMoreNested.0.Bar: Bar must be between 3 and 5",
		},
		{
			param: &OuterStruct{
				Nested: NestedStruct{
					Foo: "123",
					MapEvenMoreNested: map[string]EvenMoreNestedStruct{
						"Foo": {
							Bar: "123456",
						},
					},
				},
			},
			expected:    false,
			expectedErr: "Nested.MapEvenMoreNested.Foo.Bar: Bar must be between 3 and 5",
		},
	}

	for _, test := range tests {
		actual, err := Struct(test.param)
		assert.Equal(t, test.expected, actual, "Validation result mismatch for input: %+v", test.param)
		if test.expectedErr != "" {
			assert.Error(t, err, "Expected error but got none for input: %+v", test.param)
			assert.EqualError(t, err, test.expectedErr, "Error message mismatch for input: %+v", test.param)
		} else {
			assert.NoError(t, err, "Unexpected error for input: %+v", test.param)
		}
	}
}

func TestRequired(t *testing.T) {
	type testByteArray [8]byte
	type testByteMap map[byte]byte
	type testByteSlice []byte
	type testStringStringMap map[string]string
	type testStringIntMap map[string]int

	testString := "foobar"
	testEmptyString := ""
	var tests = []struct {
		param    any
		expected bool
	}{

		{
			&struct {
				TestByteArray testByteArray `is:"required"`
			}{},
			false,
		},
		{
			&struct {
				TestByteArray testByteArray `is:"required"`
			}{
				testByteArray{},
			},
			false,
		},
		{
			&struct {
				TestByteMap testByteMap `is:"required"`
			}{},
			false,
		},
		{
			&struct {
				Pointer *string `is:"required"`
			}{
				Pointer: &testEmptyString,
			},
			false,
		},
		{
			&struct {
				Pointer *string `is:"required"`
			}{},
			false,
		},

		{
			&struct {
				Pointer *string `is:"required"`
			}{
				Pointer: &testString,
			},
			true,
		},
		{
			&struct {
				Addr Address `is:"required"`
			}{},
			false,
		},
		{
			&struct {
				Addr Address `is:"required"`
			}{
				Addr: Address{"", "123"},
			},
			true,
		},
		{
			&struct {
				Pointer *Address `is:"required"`
			}{},
			false,
		},
		{
			&struct {
				Pointer *Address `is:"required"`
			}{
				Pointer: &Address{"", "123"},
			},
			true,
		},
		{
			&struct {
				TestByteArray testByteArray `is:"required"`
			}{
				testByteArray{'1', '2', '3', '4', '5', '6', '7', 'A'},
			},
			true,
		},
		{
			&struct {
				TestByteSlice testByteSlice `is:"required"`
			}{},
			false,
		},
		{
			&struct {
				TestStringStringMap testStringStringMap `is:"required"`
			}{
				testStringStringMap{"test": "test"},
			},
			true,
		},
		{
			&struct {
				TestIntMap testStringIntMap `is:"required"`
			}{
				testStringIntMap{"test": 42},
			},
			true,
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, _ := Struct(test.param)
			assert.Equal(t, test.expected, actual, fmt.Sprintf("case: %+v", test.param))
		})
	}
}
