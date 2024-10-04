package render

import (
	"reflect"
	"testing"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/sqlite"
	"github.com/stretchr/testify/assert"
)

type Product struct {
	folio.Meta `kind:"product" json:",inline"`
	Name       string
}

func TestEnum(t *testing.T) {
	test := &struct {
		Field string `validate:"oneof=one two three"`
	}{
		Field: "one",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := lookupForEnum(field, fv)
	assert.True(t, enum.Init(&Props{
		Field: field,
	}))
	assert.NotNil(t, enum)

	k, v := enum.Current()
	assert.Equal(t, "one", k)
	assert.Equal(t, "One", v)
	assert.Equal(t, []string{"one", "two", "three"}, enum.choices)

	var keys, values []string
	for k, v := range enum.Choices() {
		keys = append(keys, k)
		values = append(values, v)
	}

	assert.Equal(t, []string{"one", "two", "three"}, keys)
	assert.Equal(t, []string{"One", "Two", "Three"}, values)

}

func TestUrn(t *testing.T) {
	testStorage(func(db folio.Storage, registry folio.Registry) {
		phone, err := folio.New("default", func(obj *Product) error {
			obj.Name = "phone"
			return nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, phone)

		_, err = db.Insert(phone, "test")
		assert.NoError(t, err)
		assert.NotNil(t, phone)

		test := &struct {
			Product folio.URN `validate:"required" kind:"product,*"`
		}{Product: phone.URN()}

		// Find the product field
		rv := reflect.Indirect(reflect.ValueOf(test))
		field, _ := rv.Type().FieldByName("Product")
		fv := rv.FieldByName("Product")

		// Create a lookup for the urn & init
		enum := lookupForUrn(field, fv)
		assert.True(t, enum.Init(&Props{
			Value:    rv,
			Field:    field,
			Registry: registry,
			Store:    db,
		}))
		assert.NotNil(t, enum)
	})

}

func testStorage(fn func(db folio.Storage, registry folio.Registry)) {
	r := folio.NewRegistry()
	folio.Register[*Product](r)

	s := sqlite.OpenEphemeral(r)
	defer s.Close()
	fn(s, r)
}
