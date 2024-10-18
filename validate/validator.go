// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/kelindar/folio/internal/walk"
)

// Validation function that replaces the original Struct function
func Struct(s interface{}) (bool, error) {
	if s == nil {
		return true, nil
	}

	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return false, fmt.Errorf("function only accepts non-nil pointer to struct; got %s", val.Kind())
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return false, fmt.Errorf("function only accepts pointer to struct; got %s", val.Kind())
	}

	var errs Errors

	err := walk.Walk(s, func(v reflect.Value, field *reflect.StructField, path []string) error {
		// Skip if field is nil (e.g., map or slice elements)
		if field == nil {
			return nil
		}

		// Get the validation tag from the field
		tag := field.Tag.Get(rsTagName)
		if tag == "-" {
			return nil
		}

		// Parse options from the tag
		options := parseOpts(tag)

		// Check for empty values and required fields
		if walk.IsEmpty(v) {
			isValid, err := checkRequired(field, options, path)
			if !isValid {
				if name := nameOf(field); name != "" {
					err = withName(err, name)
				}
				errs = append(errs, err)
			}
			return nil
		}

		delete(options, "required")

		// Perform validation based on options
		isValid, err := typeCheck(v, field, options, path)
		if !isValid && err != nil {
			if name := nameOf(field); name != "" {
				err = withName(err, name)
			}
			errs = append(errs, err)
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	if len(errs) > 0 {
		return false, errs
	}

	return true, nil
}

// typeCheck performs validation on a single field value
func typeCheck(v reflect.Value, field *reflect.StructField, options opts, path []string) (bool, error) {
	optionsOrder := options.orderedKeys()
	for _, validator := range optionsOrder {

		// Parse the validator and parameters
		matches := rxValidator.FindStringSubmatch(validator)
		if len(matches) == 0 {
			err := errorf(field, path, validator, "validator is invalid or can't be applied to the field: %q", validator)
			return false, err
		}

		// Find an appropriate validator by its name
		if vd, ok := lookup(matches[1]); ok {
			delete(options, validator)

			// Perform the validation
			params := strings.Split(matches[2], "|")
			if v := fmt.Sprint(v.Interface()); !vd.Validate(v, params...) {
				args := make([]any, 0, len(params)+1)
				args = append(args, nameOf(field))
				for _, param := range params {
					args = append(args, param)
				}

				return false, errorf(field, path, validator, vd.format, args...)
			}
		} else {
			err := errorf(field, path, validator, "unknown validator %q for field %s", validator, field.Name)
			return false, err
		}
	}

	// Check for any unprocessed options
	delete(options, "required")
	if len(options) != 0 {
		optionsOrder := options.orderedKeys()
		for _, validator := range optionsOrder {
			err := errorf(field, path, validator, "validator is invalid or can't be applied to the field: %q", validator)
			return false, err
		}
	}

	return true, nil
}

// parseOpts parses a struct tag `valid:required~Some error message,length(2|3)` into map[string]string{"required": "Some error message", "length(2|3)": ""}
func parseOpts(tag string) opts {
	opts := make(opts)
	vals := strings.Split(tag, ",")

	for i, option := range vals {
		option = strings.TrimSpace(option)
		validations := strings.Split(option, "~")
		if !isValidTag(validations[0]) {
			continue
		}

		switch {
		case len(validations) == 2:
			opts[validations[0]] = tagOption{validations[0], validations[1], i}
		default:
			opts[validations[0]] = tagOption{validations[0], "", i}
		}
	}
	return opts
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// checkRequired checks whether a field is required or not
func checkRequired(t *reflect.StructField, options opts, path []string) (bool, error) {
	_, isRequired := options["required"]
	switch {
	case isRequired:
		return false, errorf(t, path, "required", "%s is a required field", nameOf(t))
	default:
		return true, nil
	}
}

// nameOf returns the JSON name of a field
func nameOf(field *reflect.StructField) string {
	if field == nil {
		return ""
	}

	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	// JSON name always comes first. If there's no options then split[0] is
	// JSON name, if JSON name is not set, then split[0] is an empty string.
	split := strings.SplitN(tag, ",", 2)
	name := split[0]

	// However it is possible that the field is skipped when
	// (de-)serializing from/to JSON, in which case assume that there is no
	// tag name to use
	if name == "-" {
		return ""
	}
	return name
}

// ---------------------------------- Validators API ----------------------------------

var validators sync.Map

// Register registers a new validator function
func Register(name, format string, fn Func) {
	RegisterN(name, format, func(str string, _ ...string) bool {
		return fn(str)
	})
}

// RegisterN registers a new validator function with additional parameters
func RegisterN(name, format string, fn FuncN) {
	validators.Store(name, &validator{
		name:   name,
		format: format,
		fn:     fn,
	})

	// Register the negated version of the validator
	negated := fmt.Sprintf("!%s", name)
	validators.Store(negated, &validator{
		name:   negated,
		format: strings.ReplaceAll(format, "must ", "must not "),
		fn:     func(str string, params ...string) bool { return !fn(str, params...) },
	})
}

type validator struct {
	name   string // Name of the validator
	format string // Format of the error message
	fn     FuncN  // The function to call
}

func (v *validator) Validate(str string, params ...string) bool {
	return v.fn(str, params...)
}

func lookup(name string) (*validator, bool) {
	if v, ok := validators.Load(name); ok {
		return v.(*validator), true
	}
	return nil, false
}
