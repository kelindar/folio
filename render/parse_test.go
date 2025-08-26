package render

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/errors"
	"github.com/stretchr/testify/assert"
)

// Example structs to demonstrate the transformation
type Engine struct {
	Type  string `json:"type"`
	Power int    `json:"power"`
}

type Employee struct {
	Name    string   `json:"name"`
	Role    string   `json:"role"`
	Aliases []string `json:"aliases"`
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
	folio.Meta  `kind:"car" json:",inline"`
	Type        string    `json:"type"`
	Year        int       `json:"year"`
	Model       string    `json:"model"`
	Description string    `json:"description"`
	Company     folio.URN `json:"company"`
	Engine      *Engine   `json:"engine,omitempty"`
	Engines     []Engine  `json:"engines,omitempty"`
	CompanyInfo *Company  `json:"companyInfo,omitempty"`
}

func TestUnmarshalForm(t *testing.T) {
	inputJSON := `{
		"kind": "car",
		"type": "sedan",
		"year": 2000,
		"model": "Tesla",
		"description": "Electric car #1",
		"company": "urn:default:company:cs0m2m1hq4ujcu6kfr30",
		"engine.type": "electric",
		"engine.power": 200,
		"engines.0.type": "diesel",
		"engines.0.power": 150,
		"engines.1.type": "electric",
		"engines.1.power": 200,
		"companyInfo.name": "Tech Corp",
		"companyInfo.address.street": "123 Innovation Drive",
		"companyInfo.address.city": "Techville",
		"companyInfo.departments.48972451.name": "Marketing",
		"companyInfo.departments.48972451.employees.523491677.name": "Bob Johnson",
		"companyInfo.departments.48972451.employees.523491677.role": "Supervisor",
		"companyInfo.departments.0.name": "Research and Development",
		"companyInfo.departments.0.employees.0.name": "Alice Smith",
		"companyInfo.departments.0.employees.0.role": "Engineer",
		"companyInfo.departments.0.employees.0.aliases[]": "Al",
		"companyInfo.departments.1.name": "Marketing",
		"companyInfo.departments.1.employees.0.name": "Sam Jackson",
		"companyInfo.departments.1.employees.0.role": "Manager",
		"companyInfo.departments.1.employees.0.aliases[]": ["Sammy", "Samuel"],
	}`

	registry := folio.NewRegistry()
	typ, _ := folio.Register[*Car](registry)

	var car Car
	_, err := hydrate(strings.NewReader(inputJSON), typ, &car, errors.NewValidator())
	assert.NoError(t, err)
	assert.Equal(t, "sedan", car.Type)
}

func TestUnmarshal_URN(t *testing.T) {
	inputJSON := `{
		"company": "urn:default:company:cs0m2m1hq4ujcu6kfr30",
	}`

	registry := folio.NewRegistry()
	typ, _ := folio.Register[*Car](registry)

	var car Car
	_, err := hydrate(strings.NewReader(inputJSON), typ, &car, errors.NewValidator())
	assert.NoError(t, err)
	assert.Equal(t, "urn:default:company:cs0m2m1hq4ujcu6kfr30", car.Company.String())
}

func TestUnmarshal_Struct(t *testing.T) {
	inputJSON := `{
		"engine.type": "electric",
		"engine.power": 200,
	}`

	registry := folio.NewRegistry()
	typ, _ := folio.Register[*Car](registry)

	var car Car
	_, err := hydrate(strings.NewReader(inputJSON), typ, &car, errors.NewValidator())
	assert.NoError(t, err)
	assert.Equal(t, "electric", car.Engine.Type)
	assert.Equal(t, 200, car.Engine.Power)
}

func TestUnmarshal_Slice(t *testing.T) {
	inputJSON := `{
		"engines.0.type": "diesel",
		"engines.0.power": 150,
		"engines.99999.type": "electric",
		"engines.99999.power": 200,
	}`

	registry := folio.NewRegistry()
	typ, _ := folio.Register[*Car](registry)

	var car Car
	_, err := hydrate(strings.NewReader(inputJSON), typ, &car, errors.NewValidator())
	assert.NoError(t, err)
	assert.Equal(t, "diesel", car.Engines[0].Type)
	assert.Equal(t, 150, car.Engines[0].Power)
	assert.Equal(t, "electric", car.Engines[1].Type)
	assert.Equal(t, 200, car.Engines[1].Power)
	assert.Equal(t, 2, len(car.Engines))
}

func TestUnmarshal_Slice_Overwrite(t *testing.T) {
	inputJSON := `{
		"engines.0.type": "diesel",
		"engines.0.power": 150,
		"engines.99999.type": "electric",
		"engines.99999.power": 200,
	}`

	registry := folio.NewRegistry()
	typ, _ := folio.Register[*Car](registry)

	car := Car{
		Engines: []Engine{
			{Type: "gasoline", Power: 100},
		},
	}
	_, err := hydrate(strings.NewReader(inputJSON), typ, &car, errors.NewValidator())
	assert.NoError(t, err)
	assert.Equal(t, "diesel", car.Engines[0].Type)
	assert.Equal(t, 150, car.Engines[0].Power)
	assert.Equal(t, "electric", car.Engines[1].Type)
	assert.Equal(t, 200, car.Engines[1].Power)
	assert.Equal(t, 2, len(car.Engines))
}

func TestDecodeRange(t *testing.T) {
	tests := map[string]struct {
		tag     string
		min     float64
		max     float64
		step    float64
		hasStep bool
	}{
		"no range tag":             {tag: `is:"required"`, min: 0, max: 0, step: 0, hasStep: false},
		"basic range":              {tag: `is:"range(0|100)"`, min: 0, max: 100, step: 1, hasStep: false},
		"range with step":          {tag: `is:"range(0|100|0.5)"`, min: 0, max: 100, step: 0.5, hasStep: true},
		"range with integer step":  {tag: `is:"range(1|10|1)"`, min: 1, max: 10, step: 1, hasStep: true},
		"complex tag with range":   {tag: `is:"required,range(0|100|0.1),min(0)"`, min: 0, max: 100, step: 0.1, hasStep: true},
		"negative range":           {tag: `is:"range(-10|10|0.5)"`, min: -10, max: 10, step: 0.5, hasStep: true},
		"decimal range":            {tag: `is:"range(0.5|10.5|0.25)"`, min: 0.5, max: 10.5, step: 0.25, hasStep: true},
		"invalid step zero":        {tag: `is:"range(0|100|0)"`, min: 0, max: 0, step: 0, hasStep: false},
		"invalid step negative":    {tag: `is:"range(0|100|-1)"`, min: 0, max: 0, step: 0, hasStep: false},
		"invalid step non-numeric": {tag: `is:"range(0|100|step)"`, min: 0, max: 0, step: 0, hasStep: false},
		"invalid min non-numeric":  {tag: `is:"range(min|100|1)"`, min: 0, max: 0, step: 0, hasStep: false},
		"invalid max non-numeric":  {tag: `is:"range(0|max|1)"`, min: 0, max: 0, step: 0, hasStep: false},
		"incomplete range":         {tag: `is:"range(0)"`, min: 0, max: 0, step: 0, hasStep: false},
		"empty range":              {tag: `is:"range()"`, min: 0, max: 0, step: 0, hasStep: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			field := reflect.StructField{
				Tag: reflect.StructTag(tc.tag),
			}
			min, max, step, hasStep := decodeRange(field)
			assert.Equal(t, tc.min, min, "min mismatch")
			assert.Equal(t, tc.max, max, "max mismatch")
			assert.Equal(t, tc.step, step, "step mismatch")
			assert.Equal(t, tc.hasStep, hasStep, "hasStep mismatch")
		})
	}
}

func TestRangeStepLimit(t *testing.T) {
	tests := map[string]struct {
		min          float64
		max          float64
		step         float64
		expectSlider bool
	}{
		"exactly 100 steps":        {min: 0, max: 100, step: 1, expectSlider: true},
		"101 steps":                {min: 0, max: 101, step: 1, expectSlider: false},
		"50 steps with decimals":   {min: 0, max: 10, step: 0.2, expectSlider: true},
		"200 steps":                {min: 0, max: 200, step: 1, expectSlider: false},
		"large range small step":   {min: 0, max: 1000000, step: 1, expectSlider: false},
		"reasonable decimal steps": {min: 0, max: 10, step: 0.1, expectSlider: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			steps := (tc.max - tc.min) / tc.step
			shouldUseSlider := steps <= 100
			assert.Equal(t, tc.expectSlider, shouldUseSlider, "step limit logic mismatch")
		})
	}
}
