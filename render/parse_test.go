package render

import (
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
