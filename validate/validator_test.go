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
	ok, err := Validate(car)
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
	ok, err := Validate(car)
	assert.False(t, ok)
	assert.Nil(t, err)
}
