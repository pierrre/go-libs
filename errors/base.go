package errors

import (
	std_errors "errors"
	"fmt"
)

type baseError struct {
	s string
}

func newBase(s string) error {
	return &baseError{
		s: s,
	}
}

func (err *baseError) Error() string {
	return err.s
}

// New returns a new error with a message and a stack.
func New(msg string) error {
	return newError(msg)
}

// Newf returns a new error with a formatted message and a stack.
func Newf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return newError(msg)
}

func newError(msg string) error {
	err := newBase(msg)
	err = withStack(err, 3)
	return err
}

// As calls std_errors.As.
func As(err error, target interface{}) bool {
	return std_errors.As(err, target)
}

// Is calls std_errors.Is.
func Is(err, target error) bool {
	return std_errors.Is(err, target)
}

// Unwrap calls std_errors.Unwrap.
func Unwrap(err error) error {
	return std_errors.Unwrap(err)
}
