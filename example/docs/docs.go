package docs

import (
	"fmt"

	"github.com/brianvoe/gofakeit/v7"
	object "github.com/kelindar/folio"
)

type Person struct {
	object.Meta `kind:"person" json:",inline"`
	Name        string `json:"name"`
	Age         int    `json:"age"`
	Phone       string `json:"phone"`
	Company     string `json:"company"`
	JobTitle    string `json:"jobTitle"`
	Country     string `json:"country"`
}

func NewPerson() *Person {
	p, err := object.NewWith("default", func(p *Person) error {
		p.Name = gofakeit.Name()
		p.Desc = gofakeit.Address().Address
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
	return fmt.Sprintf("%d years old, working as %s at %s", p.Age, p.JobTitle, p.Company)
}

func (p *Person) Badges() []string {
	return []string{p.Phone, p.Country}
}

func (p *Person) Icon() string {
	const icon = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADIAAAAyBAMAAADsEZWCAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAMUExURe/x82h3h7zDy83T2GnaqlsAAAAJcEhZcwAADsMAAA7DAcdvqGQAAABWSURBVDjL7dGxDcAgEANAmyxg2H/YACJK40ckFQWukE4vwI+TjZLfrMtzSJ9EkYwLrGgyk61cAuuYf1v7jBVWkZXegBN2KUbC3hhK2DV/7GdBhJNdAtxTeAm5sD159AAAAABJRU5ErkJggg=="
	return icon
}
