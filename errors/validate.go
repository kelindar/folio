package errors

import (
	"strings"

	"github.com/kelindar/folio"
	"github.com/kelindar/folio/validate"
)

// Validator represents a validator
type Validator interface {
	Validate(value any) ([]Validation, bool)
}

// Validation represents a result of a validation.
type Validation struct {
	Path    folio.Path
	Message string
}

// String returns the string representation of the validation
func (v Validation) String() string {
	return v.Message
}

type validator struct{}

// NewValidator creates a new validator
func NewValidator() Validator {
	return &validator{}
}

// Validate validates the given value
func (v *validator) Validate(value any) ([]Validation, bool) {
	ok, err := validate.Struct(value)
	if ok {
		return nil, true
	}

	var out []Validation
	if errs, ok := err.(validate.Errors); ok {
		for _, err := range errs.Errors() {
			path := err.Path
			if len(path) == 0 {
				path = []string{err.Name}
			}

			out = append(out, Validation{
				Path:    folio.Path(strings.Join(path, ".")),
				Message: err.Message(),
			})
		}
	}
	return out, false
}
