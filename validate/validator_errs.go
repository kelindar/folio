// Fork of the validator package by https://github.com/asaskevich/govalidator
// Licensed under the MIT License.
// Copyright (c) 2014-2020 Alex Saskevich
// Copyright (c) 2024 Roman Atachiants

package validate

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ---------------------------------- Error ----------------------------------

func errorf(field *reflect.StructField, path []string, validator, message string, args ...any) Error {
	msg := fmt.Sprintf(message, args...)
	if index := strings.Index(msg, "%!(EXTRA"); index > 0 {
		msg = msg[:index]
	}

	return Error{
		error:     errors.New(msg),
		Name:      nameOf(field),
		Path:      path,
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
