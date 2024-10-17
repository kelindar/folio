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
	Name      string
	Err       error
	Validator string
	Path      []string
}

func (e Error) Error() string {
	errName := e.Name
	if len(e.Path) > 0 {
		errName = strings.Join(append(e.Path, e.Name), ".")
	}

	return errName + ": " + e.Err.Error()
}

func errorf(name, validator, message string, args ...any) Error {
	return Error{
		Name:      name,
		Err:       fmt.Errorf(message, args...),
		Validator: stripParams(validator),
	}
}

// TruncatingErrorf removes extra args from fmt.Errorf if not formatted in the str object
func TruncatingErrorf(str string, args ...interface{}) error {
	n := strings.Count(str, "%s")
	return fmt.Errorf(str, args[:n]...)
}

func withPath(err error, path ...string) error {
	switch err2 := err.(type) {
	case Error:
		err2.Path = append(path, err2.Path...)
		err2.Path = append(path, err2.Name)
		return err2
	case Errors:
		errors := err2.Errors()
		for i, err3 := range errors {
			errors[i] = withPath(err3, path...)
		}
		return err2
	}
	return err
}
