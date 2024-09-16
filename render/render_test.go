package render

import (
	"testing"

	"github.com/kelindar/folio"
	"github.com/stretchr/testify/assert"
)

type Person struct {
	folio.Meta `kind:"person" json:",inline"`
	Details    Section `json:"-" form:"ro" name:"Details" desc:"Details of the person."`
	Name       string  `json:"name" form:"rw" validate:"required"`
	Age        int     `json:"age" form:"rw" validate:"gte=0,lte=130"`
	Address    string  `json:"address" form:"rw"`
	Phone      string  `json:"phone" form:"rw"`
	Company    string  `json:"company" form:"rw"`
	JobTitle   string  `json:"jobTitle" form:"rw"`
	Country    string  `json:"country" form:"rw"`
	Gender     string  `json:"gender" form:"rw" validate:"oneof=male female prefer_not_to"`
	IsEmployed bool    `json:"isEmployed" form:"rw" desc:"Is the person employed?"`
}

func TestObject_View(t *testing.T) {
	components := Object(&Context{
		Mode: ModeView,
	}, newPerson())

	assert.Len(t, components, 10)
}

func TestObject_Edit(t *testing.T) {
	components := Object(&Context{
		Mode: ModeEdit,
	}, newPerson())

	assert.Len(t, components, 10)
}

func TestInspect_StringOf(t *testing.T) {
	assert.Equal(t, "John Doe", StringOf(newPerson(), "Name"))
	assert.Equal(t, "30", StringOf(newPerson(), "Age"))
}

func TestInspect_ListOf(t *testing.T) {
	l := []string{"one", "two", "three"}
	v := struct {
		Items []string
	}{Items: l}

	assert.Equal(t, l, ListOf(v, "Items"))
}

func newPerson() *Person {
	return &Person{
		Name:       "John Doe",
		Age:        30,
		Address:    "123 Main St",
		Phone:      "123-456-7890",
		Company:    "Acme Inc",
		JobTitle:   "Software Engineer",
		Country:    "USA",
		Gender:     "male",
		IsEmployed: true,
	}
}
