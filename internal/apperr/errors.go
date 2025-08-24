package apperr

import (
	"errors"
	"fmt"
)

// Code represents a stable, machine-readable error category.
type Code string

const (
	InvalidInput Code = "INVALID_INPUT"
	NotFound     Code = "NOT_FOUND"
	Conflict     Code = "CONFLICT"
	Unauthorized Code = "UNAUTHORIZED"
	Forbidden    Code = "FORBIDDEN"
	Internal     Code = "INTERNAL"
	Unavailable  Code = "UNAVAILABLE"
)

// Error is the application's rich error type.
type Error struct {
	Op      string      // operation name (optional)
	Code    Code        // error code
	Message string      // human-friendly message
	Err     error       // wrapped error (optional)
	Meta    interface{} // additional metadata (optional)
}

func (e *Error) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// E constructs an *Error. op, code and message are recommended; err and meta are optional.
func E(op string, code Code, err error, message string, meta ...interface{}) *Error {
	var m interface{}
	if len(meta) > 0 {
		m = meta[0]
	}
	return &Error{Op: op, Code: code, Err: err, Message: message, Meta: m}
}

// CodeOf returns the first Code found in the error chain, or Internal if none.
func CodeOf(err error) Code {
	var ae *Error
	if errors.As(err, &ae) {
		if ae.Code != "" {
			return ae.Code
		}
	}
	return Internal
}
