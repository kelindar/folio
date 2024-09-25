package render

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnum(t *testing.T) {
	test := &struct {
		Field string `validate:"oneof=one two three"`
	}{
		Field: "one",
	}

	rv := reflect.Indirect(reflect.ValueOf(test))
	field, _ := rv.Type().FieldByName("Field")
	fv := rv.FieldByName("Field")

	enum := newEnum(field, fv)
	assert.NotNil(t, enum)

	assert.Equal(t, "one", enum.Key())
	assert.Equal(t, "One", enum.Value())
	assert.Equal(t, []string{"one", "two", "three"}, enum.choices)

	var keys, values []string
	for k, v := range enum.Choices(nil, nil) {
		keys = append(keys, k)
		values = append(values, v)
	}

	assert.Equal(t, []string{"one", "two", "three"}, keys)
	assert.Equal(t, []string{"One", "Two", "Three"}, values)

}
