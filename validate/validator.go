// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"fmt"
	"reflect"
	"strings"
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

		/*

			if isEmptyValue(v) {
				// an empty value is not validated, checks only required
				isValid, resultErr = checkRequired(t, options)
				for key := range options {
					delete(options, key)
				}
				return isValid, resultErr
			}

		*/

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
	for _, validatorSpec := range optionsOrder {
		validatorStruct := options[validatorSpec]
		var negate bool
		validator := validatorSpec
		withCustomMessage := len(validatorStruct.message) > 0

		// Check for negation
		if validator[0] == '!' {
			validator = validator[1:]
			negate = true
		}

		// Check for param validators
		matched := false
		for key, value := range ParamTagRegexMap {
			ps := value.FindStringSubmatch(validator)
			if len(ps) == 0 {
				continue
			}
			matched = true
			validatefunc, ok := ParamTagMap[key]
			if !ok {
				continue
			}

			delete(options, validatorSpec)

			fieldStr := fmt.Sprint(v.Interface())
			if result := validatefunc(fieldStr, ps[1:]...); (!result && !negate) || (result && negate) {
				var err error
				if withCustomMessage {
					err = errorf(field, path, validatorSpec, validatorStruct.message, fieldStr, validator)
				} else {
					if negate {
						err = errorf(field, path, validatorSpec, "%s does validate as %s", fieldStr, validator)
					} else {
						err = errorf(field, path, validatorSpec, "%s does not validate as %s", fieldStr, validator)
					}
				}
				return false, err
			}
		}

		if matched {
			continue
		}

		// Check for standard validators
		if validatefunc, ok := TagMap[validator]; ok {
			delete(options, validatorSpec)

			fieldStr := fmt.Sprint(v.Interface())
			if result := validatefunc(fieldStr); !result && !negate || result && negate {
				var err error
				if withCustomMessage {
					err = errorf(field, path, validatorSpec, validatorStruct.message, fieldStr, validator)
				} else {
					if negate {
						err = errorf(field, path, validatorSpec, "%s does validate as %s", fieldStr, validator)
					} else {
						err = errorf(field, path, validatorSpec, "%s does not validate as %s", fieldStr, validator)
					}
				}
				return false, err
			}
		} else {
			// Unknown validator
			err := errorf(field, path, validatorSpec, "unknown validator %q for field %s", validator, field.Name)
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
