package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	vd := NewValidator()
	errs, ok := vd.Validate(&struct {
		Name   string `is:"required" json:"name"`
		Age    int    `is:"required" json:"age"`
		Height int    `is:"required,min(0)" json:"height"`
	}{
		Name:   "John",
		Height: -1,
	})

	assert.False(t, ok)
	assert.Len(t, errs, 2)
	assert.Equal(t, "age is a required field", errs[0].String())
	assert.Equal(t, "height must be at least 0", errs[1].String())
	assert.Equal(t, "age", errs[0].Path.String())
	assert.Equal(t, "height", errs[1].Path.String())

}
