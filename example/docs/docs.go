package docs

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/kelindar/folio"
)

// ---------------------------------- Person ----------------------------------

type Person struct {
	folio.Meta `kind:"person" json:",inline"`
	Name       string    `json:"name" form:"rw" validate:"required"`
	Age        int       `json:"age" form:"rw" validate:"gte=0,lte=130"`
	Gender     string    `json:"gender" form:"rw" validate:"oneof=male female prefer_not_to"`
	Country    string    `json:"country" form:"rw"`
	Address    string    `json:"address" form:"rw"`
	Phone      string    `json:"phone" form:"rw"`
	Boss       folio.URN `json:"boss" form:"rw" kind:"person"`
	IsEmployed bool      `json:"isEmployed" form:"rw" desc:"Is the person employed?"`
	JobTitle   string    `json:"jobTitle" form:"rw"`
	Workplace  folio.URN `json:"workplace" form:"rw" kind:"company" query:"namespace=*;match=Inc"`
}

func NewPerson() *Person {
	p, err := folio.New("default", func(p *Person) error {
		p.Name = gofakeit.Name()
		p.Address = gofakeit.Address().Address
		p.Age = gofakeit.Number(18, 65)
		p.Phone = gofakeit.Phone()
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
	return fmt.Sprintf("%v years old, working as %s", p.Age, p.JobTitle)
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

// ---------------------------------- Vehicle ----------------------------------

type Vehicle struct {
	folio.Meta  `kind:"vehicle" json:",inline"`
	Type        string    `json:"type" form:"rw" validate:"required,oneof=car bike truck"`
	Year        int       `json:"year" form:"rw" validate:"gte=1800,lte=2021"`
	Model       string    `json:"model" form:"rw" validate:"required"`
	Description string    `json:"description" form:"rw"`
	Company     folio.URN `json:"company" form:"rw" kind:"company"`
	Engine      struct {
		Type  string `json:"type" form:"rw" validate:"oneof=electric petrol diesel"`
		Power int    `json:"power" form:"rw" validate:"gte=0"`
	} `json:"engine" form:"rw"`
	Upholstery *struct {
		Type  string `json:"type" form:"rw" validate:"required,oneof=leather fabric"`
		Color string `json:"color" form:"rw" validate:"required"`
	} `json:"upholstery" form:"rw"`
	Owners []folio.URN `json:"owners" form:"rw" kind:"person"`
}

func (c *Vehicle) Title() string {
	return fmt.Sprintf("%s %s", c.Model, c.Type)
}

func (c *Vehicle) Subtitle() string {
	return fmt.Sprintf("Manufactured in %v", c.Year)
}
