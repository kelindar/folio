package docs

import (
	"fmt"
	"iter"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/render"
)

// ---------------------------------- Person ----------------------------------

type Person struct {
	folio.Meta `kind:"person" json:",inline"`

	Details render.Section `json:"-" name:"Personal Details" desc:"Personal details such as address and country"`
	Name    string         `json:"name" form:"rw" validate:"required"`
	Age     int            `json:"age" form:"rw" validate:"gte=0,lte=130"`
	Gender  string         `json:"gender" form:"rw" validate:"oneof=male female prefer_not_to"`
	Country string         `json:"country" form:"rw"`
	Address string         `json:"address" form:"rw"`
	Phone   string         `json:"phone" form:"rw"`

	Employment render.Section `json:"-" name:"Employment" desc:"Employment information for this person"`
	IsEmployed bool           `json:"isEmployed" form:"rw" desc:"Is the person employed?"`
	Company    string         `json:"company" form:"rw"`
	JobTitle   string         `json:"jobTitle" form:"rw"`
}

func NewPerson() *Person {
	p, err := folio.New("default", func(p *Person) error {
		p.Name = gofakeit.Name()
		p.Address = gofakeit.Address().Address
		p.Age = gofakeit.Number(18, 65)
		p.Phone = gofakeit.Phone()
		p.Company = gofakeit.Company()
		p.JobTitle = gofakeit.JobTitle()
		p.Country = gofakeit.Country()
		return nil
	})
	if err != nil {
		panic(err)
	}
	return p
}

func (p *Person) Title() string {
	return p.Name
}

func (p *Person) Subtitle() string {
	return fmt.Sprintf("%v years old, working as %s at %s", p.Age, p.JobTitle, p.Company)
}

func (p *Person) Badges() []string {
	return []string{p.Phone, p.Country}
}

func (p *Person) Icon() string {
	const icon = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyBAMAAADsEZWCAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAMUExURe/x82h3h7zDy83T2GnaqlsAAAAJcEhZcwAADsMAAA7DAcdvqGQAAABWSURBVDjL7dGxDcAgEANAmyxg2H/YACJK40ckFQWukE4vwI+TjZLfrMtzSJ9EkYwLrGgyk61cAuuYf1v7jBVWkZXegBN2KUbC3hhK2DV/7GdBhJNdAtxTeAm5sD159AAAAABJRU5ErkJggg=="
	return icon
}

func (p *Person) Status() string {
	return "Active"
}

// ---------------------------------- Company ----------------------------------

type Company struct {
	folio.Meta `kind:"company" json:",inline"`
	Name       string `json:"name" form:"rw" validate:"required"`
	Sector     string `json:"sector" form:"rw" validate:"required,oneof=tech finance health"`
	Year       int    `json:"year" form:"rw" validate:"gte=1800,lte=2021"`
}

func (c *Company) Title() string {
	return c.Name
}

func (c *Company) Subtitle() string {
	return fmt.Sprintf("Operating in the %s sector since %v", c.Sector, c.Year)
}

func NewCompany() *Company {
	c, err := folio.New("default", func(c *Company) error {
		c.Name = gofakeit.Company()
		c.Sector = gofakeit.RandomString([]string{"tech", "finance", "health"})
		c.Year = gofakeit.Number(1800, 2021)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return c
}

type CompanyLookup string

var _ render.Lookup = CompanyLookup("")

// Key returns the currently selected key.
func (c CompanyLookup) Key() string {
	return string(c)
}

// Value returns the currently selected value.
func (c CompanyLookup) Value() string {
	return string(c)
}

// Choices returns the choices for the given state.
func (c CompanyLookup) Choices(_ folio.Object, db folio.Storage) iter.Seq2[string, string] {
	return nil
}
