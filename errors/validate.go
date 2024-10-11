package errors

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Validator represents a validator
type Validator interface {
	Validate(value any) ([]Validation, bool)
}

// Validation represents a result of a validation.
type Validation struct {
	Path    string
	Message string
}

// String returns the string representation of the validation
func (v Validation) String() string {
	return v.Message
}

type enValidator struct {
	vd *validator.Validate
	ut ut.Translator
}

// NewValidator creates a new validator
func NewValidator() Validator {
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		if name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]; name != "-" {
			return name
		}
		return ""
	})
	en_translations.RegisterDefaultTranslations(validate, trans)
	return &enValidator{
		vd: validate,
		ut: trans,
	}
}

// Validate validates the given value
func (v *enValidator) Validate(value any) ([]Validation, bool) {
	err := v.vd.Struct(value)
	if err == nil {
		return nil, true
	}

	var result []Validation
	for _, err := range err.(validator.ValidationErrors) {
		result = append(result, Validation{
			Path:    validationPath(err.Namespace()),
			Message: err.Translate(v.ut),
		})
	}
	return result, false
}

func validationPath(path string) string {
	// If the first letter is capitalized, it's a nested field
	if len(path) > 0 && path[0] >= 'A' && path[0] <= 'Z' {
		path = path[strings.Index(path, ".")+1:]
	}

	return strings.ReplaceAll(path, ".", "-")
}
