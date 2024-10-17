// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Struct use tags for fields.
// result will be equal to `false` if there are any errors.
// todo currently there is no guarantee that errors will be returned in predictable order (tests may to fail)
func Struct(s any) (bool, error) {
	if s == nil {
		return true, nil
	}

	result := true
	var err error
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// we only accept structs
	if val.Kind() != reflect.Struct {
		return false, fmt.Errorf("function only accepts structs; got %s", val.Kind())
	}

	var errs Errors
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if typeField.PkgPath != "" {
			continue // Private field
		}
		structResult := true
		if valueField.Kind() == reflect.Interface {
			valueField = valueField.Elem()
		}
		if (valueField.Kind() == reflect.Struct ||
			(valueField.Kind() == reflect.Ptr && valueField.Elem().Kind() == reflect.Struct)) &&
			typeField.Tag.Get(rsTagName) != "-" {
			var err error
			structResult, err = Struct(valueField.Interface())
			if err != nil {
				err = withPath(err, typeField.Name)
				errs = append(errs, err)
			}
		}
		resultField, err2 := typeCheck(valueField, typeField, val, nil)
		if err2 != nil {

			// Replace structure name with JSON name if there is a tag on the variable
			jsonTag := nameOf(&typeField)
			if jsonTag != "" {
				switch jsonError := err2.(type) {
				case Error:
					jsonError.Name = jsonTag
					err2 = jsonError
				case Errors:
					for i2, err3 := range jsonError {
						switch customErr := err3.(type) {
						case Error:
							customErr.Name = jsonTag
							jsonError[i2] = customErr
						}
					}

					err2 = jsonError
				}
			}

			errs = append(errs, err2)
		}
		result = result && resultField && structResult
	}
	if len(errs) > 0 {
		err = errs
	}
	return result, err
}

func typeCheck(v reflect.Value, t reflect.StructField, o reflect.Value, options opts) (isValid bool, resultErr error) {
	if !v.IsValid() {
		return false, nil
	}

	tag := t.Tag.Get(rsTagName)

	// checks if the field should be ignored
	switch tag {
	case "":
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Map {
			if !fieldsRequiredByDefault {
				return true, nil
			}

			return false, errorf(t.Name, "required", "all fields must have at least one validation defined")
		}
	case "-":
		return true, nil
	}

	isRootType := false
	if options == nil {
		isRootType = true
		options = parseOpts(tag)
	}

	if isEmptyValue(v) {
		// an empty value is not validated, checks only required
		isValid, resultErr = checkRequired(t, options)
		for key := range options {
			delete(options, key)
		}
		return isValid, resultErr
	}

	var customTypeErrors Errors
	optionsOrder := options.orderedKeys()
	for _, validatorName := range optionsOrder {
		validatorStruct := options[validatorName]
		if validatefunc, ok := CustomTypeTagMap.Get(validatorName); ok {
			delete(options, validatorName)

			if result := validatefunc(v.Interface(), o.Interface()); !result {
				if len(validatorStruct.customErrorMessage) > 0 {
					customTypeErrors = append(customTypeErrors, errorf(t.Name, validatorName, validatorStruct.customErrorMessage, fmt.Sprint(v), validatorName))
					continue
				}

				customTypeErrors = append(customTypeErrors, errorf(t.Name, validatorName, "%s does not validate as %s", fmt.Sprint(v), validatorName))
			}
		}
	}

	if len(customTypeErrors.Errors()) > 0 {
		return false, customTypeErrors
	}

	if isRootType {
		// Ensure that we've checked the value by all specified validators before report that the value is valid
		defer func() {
			delete(options, "optional")
			delete(options, "required")

			if isValid && resultErr == nil && len(options) != 0 {
				optionsOrder := options.orderedKeys()
				for _, validator := range optionsOrder {
					isValid = false
					resultErr = errorf(t.Name, validator, "validator is invalid or can't be applied to the field: %q", validator)
					return
				}
			}
		}()
	}

	for _, validatorSpec := range optionsOrder {
		validatorStruct := options[validatorSpec]
		var negate bool
		validator := validatorSpec
		customMsgExists := len(validatorStruct.customErrorMessage) > 0

		// checks whether the tag looks like '!something' or 'something'
		if validator[0] == '!' {
			validator = validator[1:]
			negate = true
		}

		// checks for interface param validators
		for key, value := range InterfaceParamTagRegexMap {
			ps := value.FindStringSubmatch(validator)
			if len(ps) == 0 {
				continue
			}

			validatefunc, ok := InterfaceParamTagMap[key]
			if !ok {
				continue
			}

			delete(options, validatorSpec)

			field := fmt.Sprint(v)
			if result := validatefunc(v.Interface(), ps[1:]...); (!result && !negate) || (result && negate) {
				if customMsgExists {
					return false, errorf(t.Name, validator, validatorStruct.customErrorMessage, field, validator)
				}
				if negate {
					return false, errorf(t.Name, validatorSpec, "%s does validate as %s", field, validator)
				}

				return false, errorf(t.Name, validatorSpec, "%s does not validate as %s", field, validator)
			}
		}
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		// for each tag option checks the map of validator functions
		for _, validatorSpec := range optionsOrder {
			validatorStruct := options[validatorSpec]
			var negate bool
			validator := validatorSpec
			customMsgExists := len(validatorStruct.customErrorMessage) > 0

			// checks whether the tag looks like '!something' or 'something'
			if validator[0] == '!' {
				validator = validator[1:]
				negate = true
			}

			// checks for param validators
			for key, value := range ParamTagRegexMap {
				ps := value.FindStringSubmatch(validator)
				if len(ps) == 0 {
					continue
				}

				validatefunc, ok := ParamTagMap[key]
				if !ok {
					continue
				}

				delete(options, validatorSpec)

				switch v.Kind() {
				case reflect.String,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Float32, reflect.Float64:

					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field, ps[1:]...); (!result && !negate) || (result && negate) {
						if customMsgExists {
							return false, errorf(t.Name, validatorSpec, validatorStruct.customErrorMessage, field, validator)
						}
						if negate {
							return false, errorf(t.Name, validatorSpec, "%s does validate as %s", field, validator)
						}

						return false, errorf(t.Name, validatorSpec, "%s does not validate as %s", field, validator)
					}
				default:
					// type not yet supported, fail
					return false, errorf(t.Name, validatorSpec, "validator %s doesn't support kind %s for value %v", validator, v.Kind(), v)
				}
			}

			if validatefunc, ok := TagMap[validator]; ok {
				delete(options, validatorSpec)

				switch v.Kind() {
				case reflect.String,
					reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
					reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
					reflect.Float32, reflect.Float64:
					field := fmt.Sprint(v) // make value into string, then validate with regex
					if result := validatefunc(field); !result && !negate || result && negate {
						if customMsgExists {
							return false, errorf(t.Name, validatorSpec, validatorStruct.customErrorMessage, field, validator)
						}
						if negate {
							return false, errorf(t.Name, validatorSpec, "%s does validate as %s", field, validator)
						}

						return false, errorf(t.Name, validatorSpec, "%s does not validate as %s", field, validator)
					}
				default:
					//Not Yet Supported Types (Fail here!)
					return false, errorf(t.Name, validatorSpec, "validator %s doesn't support kind %s for value %v", validator, v.Kind(), v)
				}
			}
		}
		return true, nil
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return false, &UnsupportedTypeError{v.Type()}
		}
		var sv stringValues
		sv = v.MapKeys()
		sort.Sort(sv)
		result := true
		for i, k := range sv {
			var resultItem bool
			var err error
			if v.MapIndex(k).Kind() != reflect.Struct {
				resultItem, err = typeCheck(v.MapIndex(k), t, o, options)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = Struct(v.MapIndex(k).Interface())
				if err != nil {
					//err = withPath(err, t.Name+"."+sv[i].Interface().(string))
					err = withPath(err, nameOf(&t), sv[i].Interface().(string))
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil

	// If the value is a slice then check its elements
	case reflect.Slice, reflect.Array:
		result := true
		for i := 0; i < v.Len(); i++ {
			var resultItem bool
			var err error
			if v.Index(i).Kind() != reflect.Struct {
				resultItem, err = typeCheck(v.Index(i), t, o, options)
				if err != nil {
					return false, err
				}
			} else {
				resultItem, err = Struct(v.Index(i).Interface())
				if err != nil {
					//err = withPath(err, t.Name+"."+strconv.Itoa(i))
					err = withPath(err, nameOf(&t), strconv.Itoa(i))
					return false, err
				}
			}
			result = result && resultItem
		}
		return result, nil
	case reflect.Interface:
		// If the value is an interface then encode its element
		if v.IsNil() {
			return true, nil
		}
		return Struct(v.Interface())
	case reflect.Ptr:
		// If the value is a pointer then checks its element
		if v.IsNil() {
			return true, nil
		}
		return typeCheck(v.Elem(), t, o, options)
	case reflect.Struct:
		return true, nil
	default:
		return false, &UnsupportedTypeError{v.Type()}
	}
}

func stripParams(validatorString string) string {
	return rxParams.ReplaceAllString(validatorString, "")
}

// isEmptyValue checks whether value empty or not
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// nameOf returns the JSON name of a field
func nameOf(field *reflect.StructField) string {
	if field == nil {
		return ""
	}

	tag := field.Tag.Get("json")
	if tag == "" {
		return ""
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
func checkRequired(t reflect.StructField, options opts) (bool, error) {
	_, isRequired := options["required"]
	switch {
	case isRequired:
		return false, errorf(t.Name, "required", "non zero value required")
	default:
		return true, nil
	}
}
