package render

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	Type        string   `json:"type"`
	Year        int      `json:"year"`
	Model       string   `json:"model"`
	Description string   `json:"description"`
	Company     string   `json:"company"`
	Engine      *Engine  `json:"engine,omitempty"`
	Engines     []Engine `json:"engines,omitempty"`
	CompanyInfo *Company `json:"companyInfo,omitempty"`
}

func TestUnmarshalForm(t *testing.T) {
	input := map[string]any{
		"kind":                           "car",
		"type":                           "sedan",
		"year":                           2000,
		"model":                          "Tesla",
		"description":                    "Electric car #1",
		"company":                        "urn:default:company:cs0m2m1hq4ujcu6kfr30",
		"engine.type":                    "petrol",
		"engine.power":                   0,
		"engines.0.type":                 "diesel",
		"engines.0.power":                150,
		"engines.1.type":                 "electric",
		"engines.1.power":                200,
		"companyInfo.name":               "Tech Corp",
		"companyInfo.address.street":     "123 Innovation Drive",
		"companyInfo.address.city":       "Techville",
		"companyInfo.departments.0.name": "Research and Development",
		"companyInfo.departments.0.employees.0.name": "Alice Smith",
		"companyInfo.departments.0.employees.0.role": "Engineer",
		"companyInfo.departments.1.name":             "Marketing",
		"companyInfo.departments.1.employees.0.name": "Bob Johnson",
		"companyInfo.departments.1.employees.0.role": "Manager",
		"optionalField": nil,
	}

	inputJSON, _ := json.Marshal(input)
	var car Car
	assert.NoError(t, decodeForm(bytes.NewBuffer(inputJSON), &car, ""))
	assert.Equal(t, "sedan", car.Type)
}

func TestJSONPath(t *testing.T) {
	rt := reflect.TypeOf(Car{})
	tests := []struct {
		path   string
		expect string
	}{
		{"type", "string"},
		{"year", "int"},
		{"model", "string"},
		{"engine", "struct"},
		{"engine.type", "string"},
		{"engine.power", "int"},
		{"engines", "slice"},
		{"engines.power", "int"},
		{"companyInfo.departments.employees.role", "string"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			typ, err := jsonPath(rt, test.path)
			assert.NoError(t, err)
			assert.Equal(t, test.expect, typ.Type.Kind().String())
		})
	}
}
