// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"fmt"
	"sort"
	"strings"
)

// Error returns string equivalent for reflect.Type
func (e *UnsupportedTypeError) Error() string {
	return "validator: unsupported type: " + e.Type.String()
}

func (sv stringValues) Len() int           { return len(sv) }
func (sv stringValues) Swap(i, j int)      { sv[i], sv[j] = sv[j], sv[i] }
func (sv stringValues) Less(i, j int) bool { return sv.get(i) < sv.get(j) }
func (sv stringValues) get(i int) string   { return sv[i].String() }

// Errors is an array of multiple errors and conforms to the error interface.
type Errors []error

// Errors returns itself.
func (es Errors) Errors() []error {
	return es
}

func (es Errors) Error() string {
	var errs []string
	for _, e := range es {
		errs = append(errs, e.Error())
	}
	sort.Strings(errs)
	return strings.Join(errs, ";")
}

// Error encapsulates a name, an error and whether there's a custom error message or not.
type Error struct {
	error
	Name      string
	Validator string
	Path      []string
}

func (e Error) Error() string {
	return strings.Join(e.Path, ".") + ": " + e.error.Error()
}

func errorf(name, validator, message string, args ...any) Error {
	return Error{
		error:     fmt.Errorf(message, args...),
		Name:      name,
		Validator: stripParams(validator),
	}
}

func stripParams(validatorString string) string {
	return rxParams.ReplaceAllString(validatorString, "")
}

func withPath(ex error, path ...string) error {
	switch err := ex.(type) {
	case Error:
		path = append(path, err.Path...)
		if err.Path == nil {
			path = append(path, err.Name)
		}
		err.Path = path
		return err
	case Errors:
		errors := err.Errors()
		for i, e := range errors {
			errors[i] = withPath(e, path...)
		}
		return err
	default:
		return ex
	}
}

// withName sets the name of the error (or sub-errors) to the given name.
func withName(err error, name string) error {
	switch ex := err.(type) {
	case Error:
		ex.Name = name
		return ex
	case Errors:
		for j, nested := range ex {
			switch v := nested.(type) {
			case Error:
				v.Name = name
				ex[j] = v
			}
		}

		return ex
	}
	return err
}
