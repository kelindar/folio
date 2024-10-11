package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	vd := NewValidator()
	errs, ok := vd.Validate(struct {
		Name   string `validate:"required" json:"name"`
		Age    int    `validate:"required" json:"age"`
		Height int    `validate:"required,gte=0" json:"height"`
	}{
		Name:   "John",
		Height: -1,
	})

	assert.False(t, ok)
	assert.Len(t, errs, 2)
	assert.Equal(t, "age is a required field", errs[0].String())
	assert.Equal(t, "height must be 0 or greater", errs[1].String())
	assert.Equal(t, "age", errs[0].Path)
	assert.Equal(t, "height", errs[1].Path)

}
