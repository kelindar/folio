package docs

import (
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
}

func NewPerson() *Person {
	p, err := object.NewWith("default", func(p *Person) error {
		p.Name = gofakeit.Name()
		p.Desc = gofakeit.Address().Address
		p.Age = gofakeit.Number(18, 65)
		p.Phone = gofakeit.Phone()
		p.Company = gofakeit.Company()
		p.JobTitle = gofakeit.JobTitle()
		return nil
	})
	if err != nil {
		panic(err)
	}
	return p
}
