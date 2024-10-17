// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ErrUnsupported is a wrapper for reflect.Type
type ErrUnsupported struct {
	Type reflect.Type
}

// Error returns string equivalent for reflect.Type
func (e *ErrUnsupported) Error() string {
	return "validator: unsupported type: " + e.Type.String()
}

// ---------------------------------- Error ----------------------------------

func errorf(field *reflect.StructField, validator, message string, args ...any) Error {
	return Error{
		error:     fmt.Errorf(message, args...),
		Name:      nameOf(field),
		Validator: stripParams(validator),
	}
}

// Error encapsulates a name, an error and whether there's a custom error message or not.
type Error struct {
	error
	Name      string
	Validator string
	Path      []string
}

// Error returns the string representation of the error.
func (e Error) Error() string {
	return strings.Join(e.Path, ".") + ": " + e.Message()
}

// Message returns the error message.
func (e Error) Message() string {
	return e.error.Error()
}

func stripParams(validatorString string) string {
	return rxParams.ReplaceAllString(validatorString, "")
}

// ---------------------------------- Errors ----------------------------------

// Errors is an array of multiple errors and conforms to the error interface.
type Errors []error

// Errors returns all contained errors, recursively.
func (e Errors) Errors() []Error {
	var errs []Error
	for _, e := range e {
		switch ex := e.(type) {
		case Error:
			errs = append(errs, ex)
		case Errors:
			errs = append(errs, ex.Errors()...)
		}
	}
	return errs
}

// Error returns a string representation of the errors.
func (es Errors) Error() string {
	var errs []string
	for _, e := range es {
		errs = append(errs, e.Error())
	}
	sort.Strings(errs)
	return strings.Join(errs, ";")
}

// ---------------------------------- Path & Name ----------------------------------

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
		errors := err
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
