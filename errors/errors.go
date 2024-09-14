package errors

import (
	"fmt"
	"net/http"
)

// Error represents an error with an HTTP status code.
type Error struct {
	error
	Status int
}

// HTTP returns the HTTP status code for the error.
func (e *Error) HTTP() int {
	return e.Status
}

func New(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func Internal(format string, args ...any) error {
	return &Error{
		error:  fmt.Errorf(format, args...),
		Status: http.StatusInternalServerError,
	}
}

func BadRequest(format string, args ...any) error {
	return &Error{
		error:  fmt.Errorf(format, args...),
		Status: http.StatusBadRequest,
	}
}

func Unauthorized(format string, args ...any) error {
	return &Error{
		error:  fmt.Errorf(format, args...),
		Status: http.StatusUnauthorized,
	}
}

func Forbidden(format string, args ...any) error {
	return &Error{
		error:  fmt.Errorf(format, args...),
		Status: http.StatusForbidden,
	}
}

func NotFound(format string, args ...any) error {
	return &Error{
		error:  fmt.Errorf(format, args...),
		Status: http.StatusNotFound,
	}
}
