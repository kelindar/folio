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
		Field string `is:"in(one|two|three)"`
	}{
		Field: "one",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := lookupForEnum(fv)
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

func TestLongEnum(t *testing.T) {
	test := &struct {
		Field string `is:"required,in(one_handed|two_handed|shoes|pants|shirt|helm|gloves|ring|talisman|neck|hair|waist|inner_torso|bracelet|face|facial_hair|middle_torso|earrings|arms|cloak|backpack|outer_torso|outer_legs|inner_legs|mount|shop_buy|shop_resale|shop_sell|bank|secure_trade)"`
	}{
		Field: "one_handed",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := lookupForEnum(fv)
	assert.True(t, enum.Init(&Props{
		Field: field,
	}))
	assert.NotNil(t, enum)

	var keys, values []string
	for k, v := range enum.Choices() {
		keys = append(keys, k)
		values = append(values, v)
	}

	assert.Equal(t, 30, len(keys))
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
			folio.Meta `kind:"mock" json:",inline"`
			Product    folio.URN `is:"required" kind:"product"`
		}{Product: phone.URN()}

		// Find the product field
		rv := reflect.Indirect(reflect.ValueOf(test))
		field, _ := rv.Type().FieldByName("Product")

		// Create a lookup for the urn & init
		enum := lookupForUrn()
		assert.True(t, enum.Init(&Props{
			Context: &Context{
				Registry: registry,
				Store:    db,
			},
			Value:  rv,
			Field:  field,
			Parent: test,
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
